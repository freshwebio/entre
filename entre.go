package entre

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Entre struct {
	handlers []http.Handler
	mw       []middleware
}

func (e *Entre) Use() {
}

func New(mw ...Handler) Entre {

}

func (e *Entre) Run() {
}

func (e *Entre) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (e *Entre) AsHttpRouterRoute() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		e.Serve()
	}
}

type Handler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, next Handler)
}

type HandlerFunc func(http.ResponseWriter, *http.Request, httprouter.Params, http.HandlerFunc)

func (h HandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, next Handler) {
	h(rw, r, ps, next)
}

func UseHandler(h http.Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next Handler) {
		h.ServeHttp(w, r)
		next(w, r, ps, next)
	})
}

func UseHttpRouterHandler(h httprouter.Handle) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
		h(w, r, ps)
		next(w, r, ps, next)
	})
}

type middleware struct {
	handler Handler
	next    *middleware
}

func build(handlers []Handler) middleware {
}

func voidMiddleware() middleware {
	return middleware{
		HandlerFunc(func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, next Handler) {}),
		&middleware{},
	}
}
