package entre

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

// Entre provides the functionality
// for the entre middleware handler.
type Entre struct {
	handlers []Handler
	mw       middleware
}

// New deals with creating a new Entre with the provided middleware.
func New(mw ...Handler) *Entre {
	e := &Entre{}
	e.handlers = mw
	e.mw = build(e.handlers)
	return e
}

// Bundled creates a new entre middleware stack from the bundled middleware.
func Bundled(printStack bool, user string, pass string) *Entre {
	e := &Entre{}
	e.handlers = append(e.handlers, NewLogger())
	e.handlers = append(e.handlers, NewBasicAuth(user, pass))
	e.handlers = append(e.handlers, NewPanicRecovery(printStack))
	e.mw = build(e.handlers)
	return e
}

// Basic create a new entre middleware stack from the bundled middleware
// taking no parameters. This produces a stack with a logging middleware
// and a panic recovery middleware which prints the panic stack trace to the response.
func Basic() *Entre {
	e := &Entre{}
	e.handlers = append(e.handlers, NewLogger())
	e.handlers = append(e.handlers, NewPanicRecovery(true))
	e.mw = build(e.handlers)
	return e
}

// Push takes a handler and adds it to the handler list.
func (e *Entre) Push(h Handler) {
	if h == nil {
		panic("A valid handler must be provided, not nil")
	}
	e.handlers = append(e.handlers, h)
	e.mw = build(e.handlers)
}

// PushFunc adds a handler function of the entre handler type
// to the stack of middleware.
func (e *Entre) PushFunc(hf func(http.ResponseWriter, *http.Request, httprouter.Params, http.HandlerFunc)) {
	e.Push(HandlerFunc(hf))
}

// PushHandler adds a http.Handler to the stack of middleware.
func (e *Entre) PushHandler(h http.Handler) {
	e.Push(UseHandler(h))
}

// PushHandlerFunc adds a http.HandlerFunc based handler on to our stack of middleware.
func (e *Entre) PushHandlerFunc(hf func(http.ResponseWriter, *http.Request)) {
	e.Push(UseHandler(http.HandlerFunc(hf)))
}

// Serve deals with setting up with the web server
// to listen to the provided port.
func (e *Entre) Serve(addr string) {
	l := log.New(os.Stdout, "|-entre-| ", 0)
	l.Printf("listening on %s", addr)
	l.Fatal(http.ListenAndServe(addr, e))
}

// ServeHTTP deals with invoking the entre middleware chain
// for the standard http.Handler integration.
func (e *Entre) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We don't have to clone the middleware in this instance as we aren't
	// integrating directly with httprouter so don't need to add any state
	// to the middleware.
	e.mw.ServeHTTP(w, r)
}

// ServeHTTPForHTTPRouter is the endpoint handler for integration with the httprouter
// router.
func (e *Entre) ServeHTTPForHTTPRouter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Ensure the route parameters for httprouter usage are set to those of the current request
	// in a cloned middleware instance.
	mw := e.mw.Clone()
	mw.ps = ps
	mw.ServeHTTP(w, r)
}

// ForHTTPRouter provides an Entre object as a httprouter handler
// for application's using the httprouter for routing.
func (e *Entre) ForHTTPRouter() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		e.ServeHTTPForHTTPRouter(w, r, ps)
	}
}

// Handler provides the base definition for an entre handler that
// provides the core middleware functionality.
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, params httprouter.Params, next http.HandlerFunc)
}

// HandlerFunc provides the definition for a handler function.
type HandlerFunc func(http.ResponseWriter, *http.Request, httprouter.Params, http.HandlerFunc)

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
	h(w, r, ps, next)
}

// NextHandlerFunc provides the definition for a handler function which does not make use of httprouter.Params,
// this is for middleware components built for other routers that have include a reference to the next handler in the chain.
type NextHandlerFunc func(http.ResponseWriter, *http.Request, http.HandlerFunc)

func (h NextHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	h(w, r, next)
}
// NextHandler provides the definition for a handler that is not coupled with the httprouter package
// but allows us to make use of middleware primarily written for libraries like negroni.
type NextHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

// UseHandler provides a way to wrap a http.Handler in an entre.Handler to be used
// as middleware.
func UseHandler(h http.Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		h.ServeHTTP(w, r)
		next(w, r)
	})
}

// UseNextHandlerFunc allows us to use a handler which allows calling of the next handler in the chain without
// the expectation of router params.
func UseNextHandlerFunc(h NextHandlerFunc) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		h.ServeHTTP(w, r, next)
		next(w, r)
	})
}

// UseNextHandler allows entre to take handlers with a next argument.
func UseNextHandler(h NextHandler) Handler {
	return UseNextHandlerFunc(h.ServeHTTP)
}

// UseHTTPRouterHandler wraps a httprouter handler so it can be used as the part
// of then entre middleware chain.
func UseHTTPRouterHandler(h httprouter.Handle) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		h(w, r, ps)
		next(w, r)
	})
}

type middleware struct {
	handler Handler
	next    *middleware
	// On every new request this needs to be updated for each middleware
	// to work correctly with the httprouter router.
	ps httprouter.Params
}

// ServeHTTP begins the execution of the chain of middleware.
func (m middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clone the next middleware in the chain to pass on the route parameters.
	nextmw := m.next.Clone()
	nextmw.ps = m.ps
	m.handler.ServeHTTP(w, r, m.ps, nextmw.ServeHTTP)
}

// Clone deals with making a clone of the current middleware.
// This is needed as with every request disposable middleware objects are needed
// in order to keep consistent while dynamically populating route parameters
// for applications utilising the httprouter package.
// To handle simultaneous requests it's simply easier to spawn temporary middleware
// in ensuring the wrong route parameters don't get passed through the chain.
func (m middleware) Clone() *middleware {
	return &middleware{
		handler: m.handler,
		next:    m.next,
	}
}

func build(handlers []Handler) middleware {
	var next middleware
	var hlen = len(handlers)
	if hlen > 1 {
		next = build(handlers[1:])
	} else if hlen == 0 {
		return terminalMiddleware()
	} else {
		next = terminalMiddleware()
	}
	return middleware{handlers[0], &next, httprouter.Params{}}
}

func terminalMiddleware() middleware {
	return middleware{
		handler: HandlerFunc(func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {}),
		next:    &middleware{nil, nil, httprouter.Params{}},
	}
}
