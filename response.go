package main

import (
	"net/http"
)

// ResponseWriter is a wrapper to net/http.ResponseWriter, which records the status code
type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func NewResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{rw, 200}
}

func (w ResponseWriter) Status() int {
	return w.status
}

func (w ResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w ResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

func (w ResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
