package fweight

import (
	"net/http"
)

type (
	//A Router value is used to find the correct http.Handler
	//for this http.Request.
	//
	//The routing terminates when Router is Handler, or if
	//Router is nil.
	Router interface {
		RouteHTTP(*http.Request) Router
	}
)
