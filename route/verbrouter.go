package route

import (
	"github.com/TShadwell/fweight"
	"net/http"
)

func GetOnly(r fweight.Router) Verb {
	return Verb{
		"GET": r,
	}
}

var _ fweight.Router = make(Verb)

var OptionsHandler func(i interface{}, rw http.ResponseWriter, rq *http.Request)
var MethodNotAllowed func(i interface{}, rw http.ResponseWriter, rq *http.Request)

func (v Verb) Verbs() (ops []string) {
	ops = make([]string, len(v))
	if OptionsHandler != nil {
		ops = append(ops, "OPTIONS")
	}
	var i uint
	for k := range v {
		ops[i] = k
		i++
	}
	return
}

func (v Verb) RouteHTTP(rq *http.Request) fweight.Router {
	if OptionsHandler != nil && rq.Method == "OPTIONS" {
		return fweight.Handle(http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
			OptionsHandler(v.Verbs(), rw, rq)
		}))
	}
	if r := v.self()[rq.Method]; r != nil {
		return r
	} else {
		return fweight.Handle(http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
			rw.WriteHeader(405)

			MethodNotAllowed(v.Verbs(), rw, rq)
		}))
	}
}

//A Router that routes based on verbs and provides
type Verb map[string]fweight.Router

func (p Verb) self() map[string]fweight.Router {
	if p == nil {
		p = make(Verb)
	}
	return p
}

//func Verb adds an HTTP verb to this VerbRouter.
func (p Verb) Verb(verb string, hf fweight.Router) Verb {
	p = p.self()
	p[verb] = hf
	return p
}

func (p Verb) Get(hf fweight.Router) Verb {
	p.Verb("GET", hf)
	return p
}

func (p Verb) Post(hf fweight.Router) Verb {
	p.Verb("POST", hf)
	return p
}

func (p Verb) Head(hf fweight.Router) Verb {
	p.Verb("HEAD", hf)
	return p
}

func (p Verb) Put(hf fweight.Router) Verb {
	p.Verb("PUT", hf)
	return p
}

func (p Verb) Delete(hf fweight.Router) Verb {
	p.Verb("DELETE", hf)
	return p
}

func (p Verb) Options(hf fweight.Router) Verb {
	p.Verb("OPTIONS", hf)
	return p
}

func (p Verb) Patch(hf fweight.Router) Verb {
	p.Verb("PATCH", hf)
	return p
}
