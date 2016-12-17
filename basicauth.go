package entre

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// BasicAuth provides the basic authentication middleware.
type BasicAuth struct {
	user     string
	password string
}

// NewBasicAuth creates a new basic auth instance with the provided
// username and password.
func NewBasicAuth(user string, password string) *BasicAuth {
	return &BasicAuth{user, password}
}

func (b *BasicAuth) ServeHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next http.HandlerFunc) {
	usr, pass, hasAuth := r.BasicAuth()
	if hasAuth && usr == b.user && pass == b.password {
		next(w, r)
	} else {
		w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}
