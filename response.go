package entre

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// Response provides a wrapper around http.ResponseWriter
// to make it easier for middleware to evaluate and retrieve
// information from responses.
type Response interface {
	http.ResponseWriter
	http.Flusher
	// Status returns the code of the response or 200 if the status code
	// has not yet been written to the response.
	Status() int
	// Written determines whether or not response has been written to.
	Written() bool
	// BodyLength retrieves the size of the response body.
	BodyLength() int
	// Before provides a way for functions to be called before a response is written.
	// This comes in to play for tasks like setting headers.
	Before(func(Response))
}

// NewResponse provides a wrapper response instance for the provided resposne writer.
func NewResponse(w http.ResponseWriter) Response {
	resp := &response{
		ResponseWriter: w,
	}
	if _, ok := w.(http.CloseNotifier); ok {
		return &responseCloseNotifier{resp}
	}
	return resp
}

type response struct {
	http.ResponseWriter
	status int
	length int
	before []func(Response)
}

func (r *response) WriteHeader(s int) {
	r.status = s
	r.callBefore()
	r.ResponseWriter.WriteHeader(s)
}

func (r *response) Write(b []byte) (int, error) {
	// When the response hasn't been written to write a 200 status code.
	if !r.Written() {
		r.WriteHeader(http.StatusOK)
	}
	size, err := r.ResponseWriter.Write(b)
	r.length += size
	return size, err
}

func (r *response) Status() int {
	return r.status
}

func (r *response) BodyLength() int {
	return r.length
}

func (r *response) Written() bool {
	return r.status != 0
}

func (r *response) Before(before func(Response)) {
	r.before = append(r.before, before)
}

func (r *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support hijacking")
	}
	return hijacker.Hijack()
}

func (r *response) callBefore() {
	for i := len(r.before) - 1; i >= 0; i-- {
		r.before[i](r)
	}
}

func (r *response) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		if !r.Written() {
			r.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

type responseCloseNotifier struct {
	*response
}

func (w *responseCloseNotifier) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
