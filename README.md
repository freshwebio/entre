# entre
Another one of those Go Middleware libraries

The purpose of entre is to provide a lightweight middleware solution which works nicely both ways with
httprouter as well as a range of other routers that simply use the core go http.Handler pattern.

There are other middleware libraries which are pretty great in terms of what you can throw their way with adapters
for handlers implementing generic http.Handler interface.

For instance:
``` go
middleman.New(Middleware1, Middleware2, middleman.UseHandler(myHandler))
```

Though none that I've come across work well with providing a the middleware object as a handler for a router which expects handlers of a different shape from the http.Handler interface.

For instance:
``` go
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

## Running entre

## Bundled middleware

### Logging
### Basic Authentication
### Panic recovery
