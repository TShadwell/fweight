package fweight

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestContinuity(t *testing.T) {
	const lent = 100000
	bt := make([]byte, lent)
	//	bt:=[]byte(`hello world`)
	for i := range bt {
		bt[i] = byte(rand.Int())
	}
	sv := new(Server).Route(
		PathRouter{
			"file": PathRouter{
				"cake": HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
					rw.Write(bt)
				}),
			},
		},
	)
	req, err := http.NewRequest("GET", "http://anything/file/cake", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	sv.ServeHTTP(w, req)

	verifyEquality(t, w.Body.Bytes(), bt)

}

func verifyEquality(t *testing.T, a, b []byte) {
	if !bytes.Equal(a, b) {
		var count uint
		for i, v := range a {
			if v == b[i] {
				count++
			} else {
				t.Logf("Byte incorrect at [%v] - %+q is not %+q", i, v, b[i])
			}
		}
		//t.Logf("Not passing! A was:\n%+q\nB was:\n%+q", b, a)
		t.Fatal("Failed continuity test,", count, "bytes correct.")
	}
}

//TODO: originalRequest ?

func TestFile(t *testing.T) {
	const path http.Dir = "./_testfiles/"
	const filename = "testfile.txt"
	const host = "foo.com"
	sv := (&Server{
		Extensions: []Extension{
			Compression{},
		},
	}).Route(
		SubdomainRouter{
			host: HandlerOf(http.FileServer(path)),
		},
	)
	req, err := http.NewRequest("GET", "http://"+host+"/"+filename, nil)
	w := httptest.NewRecorder()
	sv.ServeHTTP(w, req)
	file, err := os.Open(string(path) + filename)
	if err != nil {
		t.Logf("Error attempting to open test file %+q.\n", path+filename)
		t.Fatal(err)
	}
	fc, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatalf("Error attempting to read test file %+q.\n", err)
	}

	verifyEquality(t, w.Body.Bytes(), fc)

	q := httptest.NewServer(sv)
	defer q.Close()
	//Enter the WHY IS THIS NOT WORKING section.
	if req, err = http.NewRequest("GET", q.URL+"/"+filename, nil); err != nil {
		t.Fatal(err)
	}
	req.Host = host
	rs, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	rp, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	verifyEquality(t, rp, fc)
	t.Logf("Got:%+q", rp)
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
