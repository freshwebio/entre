package entre

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_PanicRecovery(t *testing.T) {
	buf := bytes.NewBufferString("")
	recorder := httptest.NewRecorder()
	calledHandler := false
	r := NewPanicRecovery(true)
	r.Logger = log.New(buf, "|-entre-|", 0)
	r.ErrorHandlerFunc = func(i interface{}) {
		calledHandler = true
	}
	e := New()
	e.Push(r)
	e.PushHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		panic("You have caused a panic")
	}))
	e.ServeHTTP(recorder, (*http.Request)(nil))
	expect(t, recorder.Header().Get("Content-Type"), "text/plain; charset=utf-8")
	expect(t, recorder.Code, http.StatusInternalServerError)
	expect(t, calledHandler, true)
	refute(t, recorder.Body.Len(), 0)
	refute(t, len(buf.String()), 0)
}

func Test_PanicRecoveryWithContentType(t *testing.T) {
	recorder := httptest.NewRecorder()
	r := NewPanicRecovery(true)
	r.Logger = log.New(bytes.NewBuffer([]byte{}), "|-entre-|", 0)
	e := New()
	e.Push(r)
	e.PushHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		panic("You have caused a panic")
	}))
	e.ServeHTTP(recorder, (*http.Request)(nil))
	expect(t, recorder.Header().Get("Content-Type"), "application/json; charset=utf-8")
}

func Test_PanicRecoveryErrorCallback(t *testing.T) {
	buf := bytes.NewBufferString("")
	recorder := httptest.NewRecorder()
	r := NewPanicRecovery(true)
	r.Logger = log.New(buf, "|-entre-|", 0)
	r.ErrorHandlerFunc = func(i interface{}) {
		panic("Your callback has caused a bit of a panic")
	}
	e := New()
	e.Push(r)
	e.PushHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		panic("You have caused a panic")
	}))
	e.ServeHTTP(recorder, (*http.Request)(nil))
	expect(t, strings.Contains(buf.String(), "Your callback has caused a bit of a panic"), true)
}
