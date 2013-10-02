package fweight

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmptyServer(t *testing.T) {
	var s *Server
	for _, v := range []http.Handler{new(Server), s} {
		//An empty server and a nil server should route all
		//to NotFound.
		sv := httptest.NewServer(
			v,
		)
		defer sv.Close()

		res, err := http.Get(sv.URL)
		defer res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		bt, err := ioutil.ReadAll(res.Body)
		if string(bt) != string(errorBytes(Err(StatusNotFound))) {
			t.FailNow()
		}
	}
}

const testString = "hello"

var writeTestString http.Handler = http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
	rw.Write([]byte(testString))
})

type Integ struct {
	*testing.T
}

func (t Integ) integTestString(res *http.Response, err error) {
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	cts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(cts) != testString {
		t.Fatal("The string returned (\"" + string(cts) + "\") was not the expected \"" + testString + "\" when accessing " +
			res.Request.URL.String() + ".")
	}

}

func TestVarious(t *testing.T) {
	const a, b, c, d = "alpha", "beta",
		"gamma", "delta"
	sv := httptest.NewServer(
		new(Server).Route(
			SubdomainRouter{
				/*
					beta.alpha -> test
				*/
				a: SubdomainRouter{
					b: HandlerOf(writeTestString),
				},
			},
		),
	)
	defer sv.Close()
}

func TestSimplePath(t *testing.T) {
	sv := httptest.NewServer(
		new(Server).
			//test.* GET should result in testStrng
			Route(AnyDomainThen(new(SubdomainRouter).Domain(
			"test",
			make(VerbRouter).Get(HandlerOf(writeTestString)),
		),
		)))
	defer sv.Close()
	req, err := http.NewRequest("GET", sv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "test.anything"
	res, err := http.DefaultClient.Do(req)
	defer res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	cts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(cts) != testString {
		t.Fatal("The string returned (\"" + string(cts) + "\") was not the expected \"" + testString + "\" when accessing " +
			res.Request.URL.String() + ".")
	}
}

func TestVerbs(t *testing.T) {
	for _, v := range []string{"GET", "PUT", "POST", "PATCH", "EGG", "BREW"} {
		srv := httptest.NewServer(
			new(Server).
				Route(AnyDomainThen(make(SubdomainRouter).
				Domain("test", make(VerbRouter).Verb(v, HandlerOf(writeTestString))))))
		defer srv.Close()
		req, err := http.NewRequest(v, srv.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = "test.localhost"
		res, err := http.DefaultClient.Do(req)
		defer res.Body.Close()

		if err != nil {
			t.Fatal(err)
		}

		cts, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(cts) != testString {
			t.Fatal("The string returned (\"" + string(cts) + "\") was not the expected " + testString)
		}
	}
}
