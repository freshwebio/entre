package entre

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Logger(t *testing.T) {
	buf := bytes.NewBufferString("")
	recorder := httptest.NewRecorder()
	l := NewLogger()
	l.LoggerIface = log.New(buf, "|-entre-|", 0)
	e := New()
	e.Push(l)
	e.PushHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	req, err := http.NewRequest("GET", "http://localhost:8384/test", nil)
	if err != nil {
		t.Error(err)
	}
	e.ServeHTTP(recorder, req)
	expect(t, recorder.Code, http.StatusNotFound)
	refute(t, len(buf.String()), 0)
}
