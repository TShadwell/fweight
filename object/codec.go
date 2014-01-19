/*
	Package object provides a mime-independant
	abstraction for writing handlers.
		var Archetype = object.Archetype{
			"application/json": Json,
			"application/xml":  Xml,
		}

		type someInformation struct {
			X string
		}

		hn := Archetype.Router(GetterFunc(func(rq *http.Request) interface{}{
			return someInformation{
				X: "words",
			}
		}))

		//[...]

		hn.ServeHTTP(rw, rq)

		//or use framework eight

		fw.DomainRouter{
			"localhost": fw.PathRouter{
				"user": hn,
			}
		}
*/
package object

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/TShadwell/fweight"
	htmltemplate "html/template"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"
	texttemplate "text/template"
)

var once sync.Once

var defaultArchetype *Archetype

var DefaultArchetype = func() *Archetype {
	once.Do(func() {
		defaultArchetype = &Archetype{
			ContentMarshaler: ContentMarshaler{
				"":                 Json,
				"application/json": Json,
				"application/xml":  Xml,
			},
		}
	})
	return defaultArchetype
}

type MarshalFunc func(v interface{}, m MediaType,
	params map[string]string) (data []byte, contentType string, err error)

type (
	MediaType string
)

type Getter interface {
	Get(*http.Request) interface{}
}

type GetterFunc func(*http.Request) interface{}

func (g GetterFunc) Get(r *http.Request) interface{} {
	return g(r)
}

type ErrorGetter interface {
	GetError() (interface{}, error)
}

//The ContentMarshaler binds MarshalFuncs to Content-Types.
//The empty string ("") represents the default Content-Type.
type ContentMarshaler map[MediaType]MarshalFunc

//the Archetype shadows MarshalRouters.
type Archetype struct {
	ContentMarshaler
}

func (a Archetype) Router(g Getter) MarshalRouter {
	return MarshalRouter{
		Archetype:        &a,
		ContentMarshaler: nil,
		Getter:           g,
	}
}

//MarshalRouter implements a http.Handler and a fweight.Router.
type MarshalRouter struct {
	Archetype *Archetype
	ContentMarshaler
	Getter
}

func (m MarshalRouter) marshalFunc(mt MediaType) (mf MarshalFunc) {
	var ok bool

	if strings.Contains(string(mt), "*") {
		//perform additional matching
		//build pattern string
		var pattern string
		for _, chr := range mt {
			if chr != '*' {
				pattern += "\\"
			}
			pattern += string(chr)
		}

		for k, v := range m.ContentMarshaler {
			//impossible to produce malformed pattern
			if ok, _ := path.Match(pattern, string(k)); ok {
				return v
			}
		}

		for k, v := range m.Archetype.ContentMarshaler {
			if ok, _ := path.Match(pattern, string(k)); ok {
				return v
			}
		}
	}

	if mf, ok = m.ContentMarshaler[mt]; ok {
		return
	}

	if mf, ok = m.Archetype.ContentMarshaler[mt]; mt == "" && (!ok) {
		panic(fmt.Sprintf("No appropriate handler and no fallback -- %+v", m.Archetype))
	}

	return m.Archetype.ContentMarshaler[mt]
}

func (m MarshalRouter) mime(r *http.Request) (mf MarshalFunc, mt MediaType, params map[string]string, err error) {

	const ctt = "Accept"

	switch {
	case true:
		ext := path.Ext(r.URL.Path)
		if ext == "" {
			break
		}

		typ := mime.TypeByExtension(ext)
		if typ == "" {
			break
		}

		if v, ok := r.Header[ctt]; ok && (len(v) > 0) {
			r.Header[ctt][0] = strings.TrimRight(typ+","+r.Header[ctt][0], ",")
			break
		}

		r.Header[ctt] = []string{
			typ,
		}

	}
	if ct, ok := r.Header["Accept"]; ok && (len(ct) > 0) {

		var cts []contentType

		cts, err = parseContentType(ct[0])
		if err != nil {
			if debug {
				log.Printf("Failed processing MIME type %+q:", ct[0])
			}
			return
		}

		if debug {
			log.Printf("Recieved Accept types: %+v\n", cts)
		}

		for _, v := range cts {
			if mf = m.marshalFunc(v.mediaType); mf != nil {
				mt = v.mediaType
				params = v.params
				break
			}
		}
		return
	}
	mf = m.marshalFunc("")
	return
}

func (m MarshalRouter) RouteHTTP(rq *http.Request) fweight.Router {
	mf, mt, params, err := m.mime(rq)
	if err != nil {
		panic(err)
	}
	return fweight.HandleFunc(func(rw http.ResponseWriter, rq *http.Request) {
		data, contentType, err := mf(m.Get(rq), mt, params)
		if err != nil {
			panic(err)
		}
		rw.Header().Add("Content-Type", contentType)
		_, err = rw.Write(data)
		if err != nil {
			panic(err)
		}
	})
}

func (m MarshalRouter) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	m.RouteHTTP(rq).(fweight.Handler).ServeHTTP(rw, rq)
}

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

	contentType = "application/xml"
	data, err = xml.Marshal(v)
	return
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
