package fweight

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
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

func TestSubdomains(t *testing.T) {
	if err := test(); err != nil {
		t.Fatal(err)
	}
}

var random = rand.New(rand.NewSource(3478001))

func wantThis() (hnd Handler, pass func([]byte) bool) {
	buf := make([]byte, 100, 100)
	for i := range buf {
		buf[i] = uint8(random.Intn(8))
	}
	hnd = HandleFunc(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, err := rw.Write(buf)
		if err != nil {
			panic(err)
		}
	}))

	pass = func(b []byte) bool {
		return bytes.Equal(b, buf)
	}

	return
}

func testSubdomains() (err error) {

	hnd1, tf1 := wantThis()
	hnd2, tf2 := wantThis()

	rt := SubdomainRouter{
		"any": hnd1,
		"many": SubdomainRouter{
			"all": hnd2,
		},
	}

	p := RouteHandler{
		Router: SubdomainRouter{
			"cool.com": rt,
		},
		NotFound: NotFound,
		Recover:  HandleRecovery,
	}

	if err = Expected(&p, "http://any.cool.com", tf1); err != nil {
		return
	}

	if err = Expected(&p, "http://all.many.cool.com", tf2); err != nil {
		return
	}

	return

}

func test() (err error) {

	hnd1, tf1 := wantThis()
	hnd2, tf2 := wantThis()

	p := Pipeline{
		Base: RouteHandler{
			Router: PathRouter{
				"a": hnd1,
				"jo": PathRouter{
					"b": hnd2,
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
	if err = Expected(&p, "http://anything/a", tf1); err != nil {
		return
	}

	if err = Expected(&p, "http://anything/jo/b", tf2); err != nil {
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
	url string, expected func([]byte) bool) (err error) {
	r := httptest.NewRecorder()
	rq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	h.ServeHTTP(r, rq)
	if r.Code != int(StatusOK) || !expected(r.Body.Bytes()) {
		return errors.New(fmt.Sprintf("Failed with code %d - expected %+q recieved %+q\n"+technicalInformation(rq), r.Code, expected, r.Body.Bytes()))
	}
	return
}
