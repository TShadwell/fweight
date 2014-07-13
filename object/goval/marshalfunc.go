package goval

import (
	"github.com/TShadwell/fweight/object"
	"io"
	"strings"
)

//A text/plain MarshalFunc that encodes the response as a Go type declaration.
var MarshalFunc object.MarshalFunc = func(r object.Responder, rq object.Request) (err error) {
	r.ContentType("text/plain;charset=utf8")
	_, err = io.Copy(r, strings.NewReader(Val(r.I)))
	return
}
