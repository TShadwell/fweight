package route

import (
	"net/http"
)

const debug = false
const failOnPanic = false

type (
	//A Router value is used to find the correct http.Handler
	//for this http.Request.
	//
	//The routing terminates when Router is Handler, or if
	//Router is nil.
	Router interface {
		RouteHTTP(*http.Request) Router
	}
	RouterFunc func(*http.Request) Router
)

func (r RouterFunc) RouteHTTP(rq *http.Request) Router {
	return r(rq)
}
