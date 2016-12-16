package entre

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Entre provides the functionality
// for the entre middleware handler.
type Entre struct {
	handlers []Handler
	mw       middleware
}

// Use takes a handler and adds it to the handler list.
func (e *Entre) Use(h Handler) {
	if h == nil {
		panic("A valid handler must be provided, not nil")
	}
	e.handlers = append(e.handlers, handler)
	e.mw = build(e.handlers)
}

func New(mw ...Handler) *Entre {
	return &Entre{}
}

func (e *Entre) Run() {
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

// HandlerFunc provides the definition
type HandlerFunc func(http.ResponseWriter, *http.Request, httprouter.Params, http.HandlerFunc)

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
	h(w, r, ps, next)
}

// UseHandler provides a way to wrap a http.Handler in an entre.Handler to be used
// as middleware.
func UseHandler(h http.Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		h.ServeHTTP(w, r)
		next(w, r)
	})
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
	return middleware{handlers[0], &next}
}

func terminalMiddleware() middleware {
	return middleware{
		HandlerFunc(func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {}),
		&middleware{},
	}
}
