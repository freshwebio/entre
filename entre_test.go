package entre

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/julienschmidt/httprouter"
)

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

// Ensure our entre middleware chain correctly returns all of its handlers.
func Test_Handlers(t *testing.T) {
	resp := httptest.NewRecorder()
	e := New()
	handlers := e.handlers
	expect(t, 0, len(handlers))
	e.Push(HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		w.WriteHeader(http.StatusOK)
	}))
	// Expect the new length of handlers to be 1 with the new handler we
	// have just added.
	handlers = e.handlers
	expect(t, 1, len(handlers))
	// Make sure our first handler is working as expected.
	handlers[0].ServeHTTP(resp, (*http.Request)(nil), httprouter.Params{}, nil)
	expect(t, resp.Code, http.StatusOK)
}

func Test_EntreServeHTTP(t *testing.T) {
	res := ""
	resp := httptest.NewRecorder()
	e := New()
	e.Push(HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		res += "my "
		next(w, r)
		res += "result"
	}))
	e.Push(HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		res += "awesome and "
		next(w, r)
		res += "new "
	}))
	e.Push(HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		res += "epic "
		w.WriteHeader(http.StatusBadRequest)
	}))
	e.ServeHTTP(resp, (*http.Request)(nil))
	expect(t, res, "my awesome and epic new result")
	expect(t, resp.Code, http.StatusBadRequest)
}

func Test_EntrePushNil(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("Expected entre.Push(nil) to panic, but for some reason it kept it's cool")
		}
	}()

	e := New()
	e.Push(nil)
}

func Test_EntreServe(t *testing.T) {
	// Ensure we can simply serve an entre stack.
	go New().Serve(":8483")
}
