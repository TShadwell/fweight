package fweight

import (
	"net/http"
)

type (
	//A Router value is used to find the correct http.Handler
	//for this http.Request.
	//
	//The routing terminates when Router is Handler, or if
	//Router is nil; a nil value causes NotFound to be served
	//by the server.
	Router interface {
		RouteHTTP(*http.Request) Router
	}

	PathingRouter interface {
		Child(subpath string) (n Router, remainingSubpath string)
	}

	ExtraFunctions interface {
		HandleHTTPError(e error, rw http.ResponseWriter, rq *http.Request)
	}

	ResponseWriter interface {
		http.ResponseWriter
		//The Server's HandleHTTPError function.
		ExtraFunctions
	}
)

type (
	MarshalFunc func(i interface{}) ([]byte, error)
)

func (m MarshalFunc) Marshal(i interface{}) ([]byte, error) {
	return m(i)
}
