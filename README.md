# entre
[![GoDoc](https://godoc.org/github.com/freshwebio/entre?status.svg)](http://godoc.org/github.com/freshwebio/entre)
Another one of those Go Middleware libraries

The purpose of entre is to provide a lightweight middleware solution which works nicely both ways with
httprouter as well as a range of other routers that simply use the core go http.Handler pattern.

There are other middleware libraries which are pretty great in terms of what you can throw their way with adapters
for handlers implementing generic http.Handler interface.

For instance:
``` go
func myHandler(w http.ResponseWriter, r *http.Request) {
  // Do handling here
}

middleman.New(Middleware1, Middleware2, middleman.UseHandler(http.HandlerFunc(myHandler)))
```

Though none that I've come across work well with providing a the middleware object as a handler for a router which expects handlers of a different shape from the http.Handler interface.

For instance:
``` go
func myHandler(w http.ResponseWriter, r *http.Request, ctx mypkg.Context, params mypkg.Params) {
  // Do the handling here
}

mwHandler := middleman.New(Middleware1, Middleware2, middleman.UseMyRouterHandler(myHandler))
myRouter.Get("/entity/", mwHandler.AsMyRouterHandler())
```

That's where entre comes in to play.

## Installing entre
```
go get github.com/freshwebio/entre
```
To use this package use:
```
import "github.com/freshwio/entre"
```

## Handlers

You can use entre to provide middleware stacks for specific routes, route groups (where router provides route grouping) or to be used
as the top level middleware for an application's core router.

### Top level middleware stack

### Specific routes or route groups

### Built-in support for httprouter

You can use most handlers accepted by most go routers as a handler within an entre stack instance
as with other middleware libraries.

Entre supports and provides an adapter for 2 main different types of handlers.

The first being the go core library http.Handler
``` go
router := httprouter.New()

func myHttpHandler(w http.ResponseWriter, r *http.Request) {
  // Do handling here
}

e := entre.New(Middleware1, Middleware2, entre.UseHandler(HandlerFunc(myHttpHandler)))
router.Handler("POST", "/entity/", e)
```

The second being the httprouter.Handle type which is simply a function which takes an extra parameter
than that of the core http.Handler ServeHTTP function.

Entre provides built-in support to provide a middleware stack as a httprouter.Handle handler in order to
retain and pass the httprouter.Params object through to the final handler in the chain.

``` go
router := httprouter.New()

func myHttpRouterHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
  // Do handling here
}

e := entre.New(Middleware1, Middleware2, entre.UseHTTPRouterHandler(myHttpRouterHandler))
router.POST("/entity/", e.ForHTTPRouter())
```

In the above providing the final handler using the UseHTTPRouterHandler wrapper method
and the ForHTTPRouter method in providing the entre stack as a httprouter handler are both needed
in order to pass the correct httprouter.Params object through the chain to the final handler.

## Serving your app with entre

You can run your core web server from entre like the following:
``` go
router := httprouter.New()
e := entre.Bundled()
e.PushHandler(router)
e.Serve(":8383")
```

Or you could simply use entre as the main handler like the following:
``` go
router := httprouter.New()
e := entre.Bundled()
e.PushHandler(router)
http.ListenAndServe(":3000", e)
```
## Bundled middleware
Entre comes with three built-in middleware items, you can set up an entre stack
with the default middleware like so:
``` go
e := entre.Bundled()
```
### Logging
### Basic Authentication
### Panic recovery
## Further support
So far the implementation of entre will support most routers.
Special adaptation was needed to integrate with the httprouter package both ways.
If there is a router that takes a handler with a different shape from that of the standard http.Handler
and you think entre should support it, create an issue on the repository.
