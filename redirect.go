package fweight

import (
	"net/http"
)

type redirinterface interface {
	Header() http.Header
	WriteHeader(int)
}

//implements http.ResponseWriter but really only pretends
type redirWrapper struct {
	redirinterface
}

const apichange = "The 'net/http'.Redirect functionality appears to have changed, please contact the package owner"

func (redirWrapper) Write([]byte) (int, error) { panic(apichange) }

//Redirect replies to the request with a redirect to url, which may be a path relative
//to the request path. This function is more generic than the net/http implementation, but
//is functionally identical.
func Redirect(w interface {
	Header() http.Header
	WriteHeader(int)
}, r *http.Request, urlStr string, code int) {
	http.Redirect(redirWrapper{w}, r, urlStr, code)
}
