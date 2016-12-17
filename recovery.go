package entre

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/julienschmidt/httprouter"
)

// PanicRecovery is the middleware that handles recovery from panics.
type PanicRecovery struct {
	Logger           LoggerIface
	PrintStack       bool
	ErrorHandlerFunc func(interface{})
	StackAll         bool
	StackSize        int
}

// NewPanicRecovery deals with create a new instance to be used in a middleware stack.
func NewPanicRecovery(printStack bool) *PanicRecovery {
	return &PanicRecovery{
		Logger:     log.New(os.Stdout, "|-entre-|", 0),
		PrintStack: printStack,
		StackAll:   false,
		StackSize:  1024 * 12,
	}
}

func (pr *PanicRecovery) ServeHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}
			w.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, pr.StackSize)
			stack = stack[:runtime.Stack(stack, pr.StackAll)]
			f := "PANIC: %s\n%s"
			pr.Logger.Printf(f, err, stack)
			if pr.PrintStack {
				fmt.Fprintf(w, f, err, stack)
			}
			if pr.ErrorHandlerFunc != nil {
				func() {
					defer func() {
						if err := recover(); err != nil {
							pr.Logger.Printf("provided ErrorHandlerFunc %s had a panic and the stack trace is:\n%s", err, debug.Stack())
							pr.Logger.Printf("%s\n", debug.Stack())
						}
					}()
					pr.ErrorHandlerFunc(err)
				}()
			}
		}
	}()
	next(w, r)
}
