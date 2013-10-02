package fweight

import (
	"net/http"
)
type Extension interface {
	TransformRequest(rw http.ResponseWriter,
		rq *http.Request)(http.ResponseWriter, *http.Request)
}


