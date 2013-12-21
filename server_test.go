package fweight

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type plainHandler string

func (p plainHandler) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte(p))
}

var NotFoundHandler = plainHandler("Not Found")
var InternalServerErrorHandler = plainHandler("Internal Server Error")

func TestWorking(t *testing.T) {
	if err := test(); err != nil {
		t.Fatal(err)
	}
}

func test() (err error) {
	const (
		aresp plainHandler = "hi this is a"
		bresp plainHandler = "hi this is b"
	)
	p := Pipeline{
		Base: RouteHandler{
			Router: PathRouter{
				"a": Handle(aresp),
				"jo": PathRouter{
					"b": Handle(bresp),
				},
			},
			NotFound: NotFound,
			Recover:  HandleRecovery,
		},
		Middleware: []Middleware{
			// this is intelligent so the tester won't get compressed output
			Compression,
		},
	}
	if err = Expected(&p, "http://anything/a", []byte(aresp)); err != nil {
		return
	}

	if err = Expected(&p, "http://anything/jo/b", []byte(bresp)); err != nil {
		return
	}

	return
}

func technicalInformation(rq *http.Request) string {
	return fmt.Sprintf(
		`Technical Information:
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
}

func Expected(h http.Handler,
	url string, expected []byte) (err error) {
	r := httptest.NewRecorder()
	rq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	h.ServeHTTP(r, rq)
	fmt.Printf("%+v\n", r.HeaderMap)
	if r.Code != int(StatusOK) || !bytes.Equal(r.Body.Bytes(), expected) {
		return errors.New(fmt.Sprintf("Failed with code %d - expected %+q recieved %+q\n"+technicalInformation(rq), r.Code, expected, r.Body.Bytes()))
	}
	return
}
