package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseWriter(t *testing.T) {
	w := NewResponseWriter(httptest.NewRecorder())

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("hello"))
	w.Header().Set("Foo", "bar")

	if w.Header().Get("Foo") != "bar" {
		t.Error("w.Header().Set not correct")
	}
	if w.Status() != http.StatusBadRequest {
		t.Errorf("w.status not correct: %d, w.status is %d", w.Status(), w.status)
	}
}
