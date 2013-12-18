package fweight

import (
	"net/http"
)

/*
A RouteHandler implements an http.Handler for a tree of Routers.
When the RouteHandler is unable to find an http.Handler, or
encounters an error while handling a response, it uses the NotFound
and InternalServerError handlers.

These functions run at the position of this RouteHandler in the pipeline,
meaning compression and security middleware that wraps this http.Handler
will still be executed.
*/
type RouteHandler struct {
	Router
	NotFound            func(rq *http.Request) http.Handler
	InternalServerError func(i interface{}) http.Handler
}

func (r RouteHandler) HandleNotFound(rq *http.Request) http.Handler {
	if r.NotFound == nil {
		panic(Err(StatusNotFound))
	}
	return r.NotFound(rq)
}

func (r RouteHandler) HandleInternalServerError(i interface{}) http.Handler {
	if r.HandleInternalServerError == nil {
		panic(Err(StatusInternalServerError))
	}
	return r.HandleInternalServerError(i)
}

/*
	ServeHTTP traverses this RouteHandler's Router tree.
	If a failure is encountered, if RouteHandler.Failure is non-nil, it
	is called for an http.Handler which it uses to handle the response,
	Handling at this RouteHandler's position in the middleware stack.
	If it is nil, ServeHTTP instead panics with an ExtendedErr.
*/
func (s RouteHandler) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	for router := s.Router.RouteHTTP(rq); ; router = router.RouteHTTP(rq) {

		//If we have a nil router, serve a 404.
		if router == nil {
			s.HandleNotFound(rq).ServeHTTP(rw, rq)
			return

		}

		//if the type of the router is a Handler, we can terminate
		if hl, ok := router.(Handler); ok {
			//defer a function to recover panics within child functions.
			defer func() {
				if e := recover(); e != nil {
					s.HandleInternalServerError(e).ServeHTTP(rw, rq)
				}
				return
			}()

			//serve the response.
			hl.ServeHTTP(rw, rq)
			return
		}

	}

}
