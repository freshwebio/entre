package entre

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_BasicAuth(t *testing.T) {
	recorder := httptest.NewRecorder()
	ba := NewBasicAuth("user", "password")
	e := New()
	e.Push(ba)
	e.PushHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	req, err := http.NewRequest("GET", "http://localhost:8384/test", nil)
	if err != nil {
		t.Error(err)
	}
	e.ServeHTTP(recorder, req)
	// In the first case we expect a 401 unauthorised as basic auth credentials
	// are not provided.
	expect(t, recorder.Code, http.StatusUnauthorized)
	// Re-use the same request but give it our basic auth details.
	req.SetBasicAuth("user", "password")
	recorder = httptest.NewRecorder()
	e.ServeHTTP(recorder, req)
	expect(t, recorder.Code, http.StatusNotFound)
}
