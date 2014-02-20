package route

import (
	"fmt"
	"github.com/TShadwell/fweight"
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
	fweight.Router
	//NotFound is called when a route for the request could not be found.
	//Recover is called when there is a panic in a Router.
	NotFound http.Handler
	Recover  RecoverHandler
}

/*
	NotFound is a convenience http.Handler for RouteHandler's NotFound.
*/
var NotFound http.Handler = http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
	rw.Header().Add("Content-Type", "text/plain")
	rw.WriteHeader(int(fweight.StatusNotFound))
	fmt.Fprintf(
		rw,
		`A resource could not be found to match your request.
	Technical Information:
	Request URI: %+q
	Method: %+q
	Protocol: %+q
	Headers: %+v
	ContentLength: %v
	Remote Address: %+q`,
		rq.Host+rq.URL.Path,
		rq.Method,
		rq.Proto,
		rq.Header,
		rq.ContentLength,
		rq.RemoteAddr,
	)
})

/*
	HandleRecovery is a convenience RecoverHandler for development builds.
*/
var HandleRecovery RecoverHandler = RecoverHandlerFunc(func(i interface{}) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		rw.Header().Add("Content-Type", "text/plain")
		rw.WriteHeader(int(fweight.StatusInternalServerError))
		fmt.Fprintf(
			rw,
			`An Internal Server Error was encountered while handling your request:
			%+q`,
			fmt.Sprint(i),
		)
	})
})

type (
	RecoverHandler interface {
		//'i' is the data the server paniced with.
		ServeRecover(i interface{}) http.Handler
	}

	RecoverHandlerFunc func(interface{}) http.Handler
)

func (r RecoverHandlerFunc) ServeRecover(i interface{}) http.Handler {
	return r(i)
}

func (r RouteHandler) HandleNotFound(rq *http.Request) http.Handler {
	if r.NotFound == nil {
		panic(fweight.Err(fweight.StatusNotFound))
	}
	return r.NotFound
}

func (r RouteHandler) HandleInternalServerError(i interface{}) http.Handler {
	if r.HandleInternalServerError == nil {
		panic(fweight.Err(fweight.StatusInternalServerError))
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
			if debug {
				fmt.Println("[!] Routing failed. Serving 404.")
			}
			s.HandleNotFound(rq).ServeHTTP(rw, rq)
			return

		}

		//if the type of the router is a Handler, we can terminate
		if hl, ok := router.(fweight.Handler); ok {
			//defer a function to recover panics within child functions.
			if !failOnPanic {
				defer func() {
					if e := recover(); e != nil {
						if debug {
							fmt.Printf("[!] Recovered from panic %v\n", e)
						}
						s.HandleInternalServerError(e).ServeHTTP(rw, rq)
					}
					return
				}()
			}

			//serve the response.
			hl.ServeHTTP(rw, rq)
			return
		}

	}

}
