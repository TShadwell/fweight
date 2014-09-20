package object

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"github.com/TShadwell/jsarray"
	htmltemplate "html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"path"
	"strings"
	texttemplate "text/template"
	"time"
)

type ContentMarshaler map[MediaType]MarshalFunc

type Request struct {
	*http.Request
	Params    map[string]string
	MediaType MediaType
}

type Responder struct {
	I interface{}
	http.ResponseWriter
	cttset bool
}

//The first call to this function sets the content type of the response.
func (r Responder) ContentType(ctt string) {
	if r.ResponseWriter.Header().Get("Content-Type") == "" {
		r.ResponseWriter.Header().Set("Content-Type", ctt)
	}
}

type MarshalFunc func(r Responder, rq Request) (err error)

//HTMLTemplate returns a MarshalFunc that executes the data on template `t` using the html/template
//package
func HTMLTemplate(t *htmltemplate.Template) MarshalFunc {
	return func(r Responder, rq Request) (err error) {
		r.ContentType("text/html;charset=utf8")

		err = t.Execute(r, r.I)
		return
	}
}

//TextTemplate returns a MarshalFunc that executes the data on template `t` using the text/template
//package.
func TextTemplate(t *texttemplate.Template) MarshalFunc {
	return func(r Responder, rq Request) (err error) {
		r.ContentType("text/plain;charset=utf8")
		err = t.Execute(r, r.I)

		return
	}
}

var Json MarshalFunc = func(r Responder, rq Request) error {
	r.ContentType("application/json;charset=utf8")

	return json.NewEncoder(r).Encode(r.I)
}

var Xml MarshalFunc = func(r Responder, rq Request) error {
	r.ContentType("application/xml;charset=utf8")

	return xml.NewEncoder(r).Encode(r.I)
}

var nullbytes = []byte("null")

//See github.com/TShadwell/jsarray for details.
var JsonArray MarshalFunc = func(r Responder, rq Request) (err error) {
	r.ContentType("application/json;charset=utf8")

	data, err := jsarray.Marshal(r.I)
	if err != nil {
		return
	}

	_, err = io.Copy(r, bytes.NewReader(data))
	return
}

//Plaintext. The underlying value of v must be string.
var Plain MarshalFunc = func(r Responder, rq Request) (err error) {
	r.ContentType("text/plain;charset=utf8")

	_, err = io.Copy(r, strings.NewReader(r.I.(string)))
	return
}

//Overwrites the ContentType of a MarshalFunc.
func (m MarshalFunc) ContentType(ctt string) MarshalFunc {
	return func(r Responder, rq Request) (err error) {
		r.ContentType(ctt)
		err = m(r, rq)
		return
	}
}

const planesStr string = "🛧🛪🛨🛦🛫🛩"

var planes = []rune(planesStr)
var nPlanes int

//Planetext.
//
//🛧🛪🛨🛦🛫🛦 🛫🛨 🛨✈🛦🛧🛦 ✈🛬🛫🛫🛬🛫 🛩🛦✈🛩🛬🛬🛨 🛫 🛬🛨✈ 🛩 🛨
var Plane MarshalFunc = func(r Responder, rq Request) (err error) {

	rand.Seed(time.Now().Unix())

	r.ContentType("text/plane")

	rn := make([]rune, 100)
	for i, ed := 0, cap(rn); i < ed; i++ {
		rn[i] = planes[rand.Intn(nPlanes)]
	}

	_, err = io.Copy(r, strings.NewReader(string(rn)))

	return
}

var Gob MarshalFunc = func(r Responder, rq Request) (err error) {
	r.ContentType("application/gob")
	err = gob.NewEncoder(r).Encode(r.I)

	return
}

//Marshaler returns the MarshalFunc present in c that matches one of the
//ContentTypes with pattern matching (text/*) in order of preference,
//as well as the chosen ContentType.
//If mf is nil ct may be some random data.
func (c ContentMarshaler) Marshaler(cts ...ContentType) (mf MarshalFunc, ct ContentType) {
	var k MediaType
	for _, ct = range cts {
		patt := restricted(string(ct.MediaType))
		for k, mf = range c {
			if ok, _ := path.Match(patt, string(k)); ok {
				return
			}
		}
	}
	mf = nil
	return
}

func Marshaler(cs []ContentMarshaler, cts ...ContentType) (mf MarshalFunc, ct ContentType) {
	var k MediaType
	for _, ct = range cts {
		patt := restricted(string(ct.MediaType))
		for _, c := range cs {
			for k, mf = range c {
				if k != "" {
					if ok, _ := path.Match(patt, string(k)); ok {
						return
					}
				}
			}
		}
	}
	mf = nil
	return
}

//RequestMarshalFunc returns the MarshalFunc mf that matches the *http.Request rq,
//as well as the selected ContentType.
//The extension of the request path is assumed to be the most preferable MIME type.
func (c ContentMarshaler) RequestMarshaler(r *http.Request) (mf MarshalFunc, ct ContentType) {
	types, _ := ParseContentType(r.Header.Get("Accept"))
	if m := pathMime(r.URL.Path); m != "" {
		types = []ContentType{{MediaType: MediaType(m)}}
	}

	return c.Marshaler(types...)
}

//RequestMarshaler returns the first MarshalFunc mf that matches the MIME types specified in
//the *http.Request from the []ContentMarshaler cs in order of preference, as well as the selected ContentType.
//The extension of the request path is assumed to be the most preferable MIME type.
func RequestMarshaler(r *http.Request, cs ...ContentMarshaler) (mf MarshalFunc, c ContentType) {
	types, _ := ParseContentType(r.Header.Get("Accept"))
	if debug {
		log.Printf("<- Accept: %+v", r.Header.Get("Accept"))
	}
	if m := pathMime(r.URL.Path); m != "" {
		a, b := pmt(m)
		types = append([]ContentType{{a, b}}, types...)
		if debug {
			log.Printf("%+v", types)
		}
	}

	mf, c = Marshaler(cs, types...)

	if debug {
		log.Printf("Picked content type: %s\nMarshaler: %+v", c, mf)
	}
	return

}

func init() {
	nPlanes = len(planes)
}
