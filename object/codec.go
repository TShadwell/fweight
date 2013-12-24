package object

/*
	a := Archetype {
		"text/plain":plain,
	}
*/

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/TShadwell/fweight"
	htmltemplate "html/template"
	"net/http"
	"path"
	"strings"
	texttemplate "text/template"
)

type MarshalFunc func(v interface{}, m MediaType,
	params map[string]string) (data []byte, contentType string, err error)

type (
	MediaType string
)

type Getter interface {
	Get() interface{}
}

type GetterFunc func() interface{}

func (g GetterFunc) Get() interface{} {
	return g()
}

type ErrorGetter interface {
	GetError() (interface{}, error)
}

//The ContentMarshaler binds MarshalFuncs to Content-Types.
//The empty string ("") represents the default Content-Type.
type ContentMarshaler map[MediaType]MarshalFunc

//the Archetype shadows MarshalRouters.
type Archetype ContentMarshaler

func (a Archetype) Router(g Getter) MarshalRouter {
	return MarshalRouter{
		Archetype:        a,
		ContentMarshaler: nil,
		Getter:           g,
	}
}

//MarshalRouter implements a http.Handler and a fweight.Router.
type MarshalRouter struct {
	Archetype Archetype
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
			//impossible to produce malformatted pattern
			if ok, _ := path.Match(pattern, string(k)); ok {
				return v
			}
		}

		for k, v := range m.Archetype {
			if ok, _ := path.Match(pattern, string(k)); ok {
				return v
			}
		}
	}

	if mf, ok = m.ContentMarshaler[mt]; ok {
		return
	}

	return m.Archetype[mt]
}

func (m MarshalRouter) mime(r *http.Request) (mf MarshalFunc, mt MediaType, params map[string]string, err error) {
	if ct, ok := r.Header["Content-Type"]; ok && (len(ct) > 0) {
		var cts []contentType
		cts, err = parseContentType(ct[0])
		if err != nil {
			return
		}

		for _, v := range cts {
			if m.marshalFunc(v.mediaType) != nil {
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
		data, contentType, err := mf(m.Get(), mt, params)
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

//HTMLTemplate returns a MarshalFunc that executes the data on template `t` using the html/template
//package
func HTMLTemplate(t *htmltemplate.Template) MarshalFunc {
	return func(v interface{}, _ MediaType,
		_ map[string]string) (data []byte, contentType string, err error) {
		contentType = "text/html;charset=utf8"
		var bf bytes.Buffer
		err = t.Execute(&bf, v)
		if err != nil {
			panic(err)
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
		contentType = "text/html;charset=utf8"
		var bf bytes.Buffer
		err = t.Execute(&bf, v)
		if err != nil {
			panic(err)
		}
		data = bf.Bytes()
		return
	}
}

var Json MarshalFunc = func(v interface{}, _ MediaType,
	_ map[string]string) (data []byte, contentType string, err error) {
	contentType = "application/json;charset=utf8"
	data, err = json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return
}

var Xml MarshalFunc = func(v interface{}, _ MediaType,
	_ map[string]string) (data []byte, contentType string, err error) {

	contentType = "application/xml"
	data, err = xml.Marshal(v)
	if err != nil {
		panic(err)
	}
	return
}

var DefaultArchetype = Archetype{
	"application/json": Json,
	"application/xml":  Xml,
}
