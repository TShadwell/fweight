package object

import (
	"net/http"
)

type ResponseWriter interface {
	Header() http.Header
	WriteHeader(int)
}

type Getter interface {
	Get(ResponseWriter, *http.Request) interface{}
}

type GetterFunc func(ResponseWriter, *http.Request) interface{}

func (g GetterFunc) Get(rs ResponseWriter, r *http.Request) interface{} {
	return g(rs, r)
}

func NewGetter(g interface {
	Get() interface{}
}) Getter {
	return GetterFunc(func(_ ResponseWriter, _ *http.Request) interface{} {
		return g.Get()
	})
}
