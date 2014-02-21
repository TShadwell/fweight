package object

import (
	"net/http"
)

type Getter interface {
	Get(*http.Request) interface{}
}

type GetterFunc func(*http.Request) interface{}

func (g GetterFunc) Get(r *http.Request) interface{} {
	return g(r)
}

func NewGetter(g interface{
	Get() interface{}
}) Getter {
	return GetterFunc(func(_ *http.Request) interface{}{
		return g.Get()
	})
}
