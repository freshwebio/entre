package entre

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/julienschmidt/httprouter"
)

/* Test Helpers */
func expect(t *testing.T, result interface{}, expected interface{}) {
	if expected != result {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", expected, reflect.TypeOf(expected), result, reflect.TypeOf(result))
	}
}

func refute(t *testing.T, result interface{}, notExpected interface{}) {
	if notExpected == result {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", notExpected, reflect.TypeOf(notExpected), result, reflect.TypeOf(result))
	}
}

type testNextHandler struct{}

func (h *testNextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("X-Test-Handler-Key", "testhandlervalue")
	next(w, r)
}

// Ensure our entre middleware chain correctly returns all of its handlers.
func Test_Handlers(t *testing.T) {
	resp := httptest.NewRecorder()
	e := New()
	handlers := e.handlers
	expect(t, 0, len(handlers))
	e.Push(HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	nextHandlerFunc := NextHandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		w.Write([]byte("This is content from the NextHandlerFunc"))
	})
	e.Push(UseNextHandlerFunc(nextHandlerFunc))

	nextHandler := &testNextHandler{}
	e.Push(UseNextHandler(nextHandler))

	// Expect the new length of handlers to be 2 with the new handler we
	// have just added.
	handlers = e.handlers
	expect(t, len(handlers), 3)
	// Make sure our handlers are working as expected.
	handlers[0].ServeHTTP(resp, (*http.Request)(nil), httprouter.Params{}, nil)
	// NextHandlerFunc handlers require a next handler in testing as will be called due
	// to it not being a direct entre handler.
	handlers[1].ServeHTTP(resp, (*http.Request)(nil), httprouter.Params{}, func(w http.ResponseWriter, r *http.Request) {})
	expect(t, resp.Code, http.StatusBadRequest)
	expect(t, string(resp.Body.Bytes()), "This is content from the NextHandlerFunc")

	// Now for the NextHandler middleware.
	handlers[2].ServeHTTP(resp, (*http.Request)(nil), httprouter.Params{}, func(w http.ResponseWriter, r *http.Request) {})
	expect(t, resp.Header().Get("X-Test-Handler-Key"), "testhandlervalue")
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
