package object

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"github.com/TShadwell/jsarray"
	htmltemplate "html/template"
	"log"
	"net/http"
	"path"
	texttemplate "text/template"
)

type ContentMarshaler map[MediaType]MarshalFunc

type MarshalFunc func(v interface{}, m MediaType,
	params map[string]string) (data []byte, contentType string, err error)

//HTMLTemplate returns a MarshalFunc that executes the data on template `t` using the html/template
//package
func HTMLTemplate(t *htmltemplate.Template) MarshalFunc {
	return func(v interface{}, _ MediaType,
		_ map[string]string) (data []byte, contentType string, err error) {
		contentType = "text/html;charset=utf8"
		var bf bytes.Buffer
		err = t.Execute(&bf, v)
		if err != nil {
			return
		}
		data = bf.Bytes()
		return
	}
}

//TextTemplate returns a MarshalFunc that executes the data on template `t` using the text/template
//package.
func TextTemplate(t *texttemplate.Template) MarshalFunc {
	return func(v interface{}, _ MediaType,
		_ map[string]string) (data []byte, contentType string, err error) {
		contentType = "text/plain;charset=utf8"
		var bf bytes.Buffer
		err = t.Execute(&bf, v)
		if err != nil {
			return
		}
		data = bf.Bytes()
		return
	}
}

var Json MarshalFunc = func(v interface{}, _ MediaType,
	_ map[string]string) (data []byte, contentType string, err error) {
	contentType = "application/json;charset=utf8"
	data, err = json.Marshal(v)

	return
}

var Xml MarshalFunc = func(v interface{}, _ MediaType,
	_ map[string]string) (data []byte, contentType string, err error) {

	contentType = "application/xml;charset=utf8"
	data, err = xml.Marshal(v)
	return
}

var nullbytes = []byte("null")

//See github.com/TShadwell/jsarray for details.
var JsonArray MarshalFunc = func(v interface{}, _ MediaType,
	_ map[string]string) (data []byte, contentType string, err error) {

	contentType = "application/json;charset=utf8"
	data, err = jsarray.Marshal(v)
	return
}

var plain MarshalFunc = func(v interface{}, _ MediaType,
	_ map[string]string) ([]byte, string, error) {
	return []byte(v.(string)), "text/plain", nil
}

var Gob MarshalFunc = func(v interface{}, _ MediaType,
	_ map[string]string) (data []byte, contentType string, err error) {

	contentType = "application/gob"
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(v)
	if err != nil {
		return
	}

	data = buf.Bytes()
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
		log.Printf("%+v", r.Header.Get("Accept"))
	}
	if m := pathMime(r.URL.Path); m != "" {
		a, b := pmt(m)
		types = append([]ContentType{{a, b}}, types...)
		if debug {
			log.Printf("%+v", types)
		}
	}

	mf, c = Marshaler(cs, types...)
	return

}
