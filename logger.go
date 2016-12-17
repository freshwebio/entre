package entre

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

// LoggerIface provides a minimal interface with our middleware logger
// with the funcionality of a simple middleware logger.
type LoggerIface interface {
	Println(...interface{})
	Printf(string, ...interface{})
}

// Logger is the type which provides our core logging middleware.
type Logger struct {
	LoggerIface
}

// NewLogger creates a new logger middleware instance.
func NewLogger() *Logger {
	return &Logger{log.New(os.Stdout, "|-entre-|", 0)}
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
	startTime := time.Now()
	l.Printf("Began %s %s", r.Method, r.URL.Path)
	next(w, r)
	resp := w.(Response)
	l.Printf("Completed with %v %s response in %v", resp.Status(), http.StatusText(resp.Status()), time.Since(startTime))
}
