package main

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

// see https://github.com/valyala/fasthttp/issues/66

type GracefulListener struct {
	// inner listener
	ln net.Listener

	// maximum wait time for graceful shutdown
	maxWaitTime time.Duration

	// this channel is closed during graceful shutdown on zero open connections.
	done chan struct{}

	// the number of open connections
	connsCount uint64

	// becomes non-zero when graceful shutdown starts
	shutdown uint64
}

// NewGracefulListener wraps the given listener into 'graceful shutdown' listener.
func newGracefulListener(ln net.Listener, maxWaitTime time.Duration) net.Listener {
	return &GracefulListener{
		ln:          ln,
		maxWaitTime: maxWaitTime,
		done:        make(chan struct{}),
	}
}

func (ln *GracefulListener) Accept() (net.Conn, error) {
	c, err := ln.ln.Accept()

	if err != nil {
		return nil, err
	}

	atomic.AddUint64(&ln.connsCount, 1)

	return &gracefulConn{
		Conn: c,
		ln:   ln,
	}, nil
}

func (ln *GracefulListener) Addr() net.Addr {
	return ln.ln.Addr()
}

// Close closes the inner listener and waits until all the pending open connections
// are closed before returning.
func (ln *GracefulListener) Close() error {
	err := ln.ln.Close()

	if err != nil {
		return nil
	}

	return ln.waitForZeroConns()
}

func (ln *GracefulListener) waitForZeroConns() error {
	atomic.AddUint64(&ln.shutdown, 1)

	if atomic.LoadUint64(&ln.connsCount) == 0 {
		close(ln.done)
		return nil
	}

	select {
	case <-ln.done:
		return nil
	case <-time.After(ln.maxWaitTime):
		return fmt.Errorf("cannot complete graceful shutdown in %s", ln.maxWaitTime)
	}
}

func (ln *GracefulListener) closeConn() {
	connsCount := atomic.AddUint64(&ln.connsCount, ^uint64(0))

	if atomic.LoadUint64(&ln.shutdown) != 0 && connsCount == 0 {
		close(ln.done)
	}
}

type gracefulConn struct {
	net.Conn
	ln *GracefulListener
}

func (c *gracefulConn) Close() error {
	err := c.Conn.Close()

	if err != nil {
		return err
	}

	c.ln.closeConn()

	return nil
}
