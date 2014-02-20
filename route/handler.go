package route

import (
	"net/http"
)

//Function HandlerOf returns a Handler of a http.Handler, which
//can be used to terminatr and handle routes.
func Handle(h http.Handler) Handler {
	return Handler{
		h,
	}
}

func HandleFunc(hf http.HandlerFunc) Handler {
	return Handle(hf)
}

//A handler represents a termination of the Router B+ tree.
type Handler struct {
	http.Handler
}

func (h Handler) RouteHTTP(rq *http.Request) Router {
	return h
}
