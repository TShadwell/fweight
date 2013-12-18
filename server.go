package fweight

import (
	"net/http"
)


//A Pipeline represents an http.Handler wrapped by a series of
//Middlewares.
//
//The resultant http.Handler is lazily computed, but can be 
//(re)computed with the ReloadMiddleware function.
type Pipeline struct {
	Base http.Handler
	pipeline http.Handler
	//Middleware is excecuted in slice order.
	Middleware []Middleware
}

func (p *Pipeline) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	if p.pipeline == nil {
		p.ReloadMiddleware()
	}
	p.pipeline.ServeHTTP(rw, rq)
}

func (p *Pipeline) ReloadMiddleware() {
	var newPipeline http.Handler = p.Base
	for _, v := range p.Middleware {
		newPipeline = v.Middleware(newPipeline)
	}
	p.pipeline = newPipeline
}

type Middleware interface {
	Middleware(http.Handler) http.Handler
}

type MiddlewareFunc func(http.Handler) http.Handler

func(m MiddlewareFunc) Middleware(h http.Handler) http.Handler {
	return m(h)
}

func isHandler(r Router) (b bool) {
	_, b = r.(Handler)
	return
}
