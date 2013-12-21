package fweight

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

type Handler struct {
	http.Handler
}

func (h Handler) RouteHTTP(rq *http.Request) Router {
	return h
}
