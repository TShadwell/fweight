package object

import (
	"github.com/TShadwell/fweight"
	"net/http"
	"sync"
)

var once sync.Once

var DefaultArchetype = Archetype{
	ContentMarshaler: ContentMarshaler{
		"":                 Json,
		"application/json": Json,
		"application/xml":  Xml,
	},
}

type Archetype struct {
	ContentMarshaler
}

func (a *Archetype) Handler() Handler {
	return Handler{
		Archetype: a,
	}
}

func (a *Archetype) Router(g Getter) HTTPHandler {
	return HTTPHandler{
		Getter: g,
		Handler: Handler{
			Archetype: a,
		},
	}
}

type HTTPHandler struct {
	Getter
	Handler
}

func (h HTTPHandler) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	h.Handler.ServeObject(h.Getter.Get(rq), rw, rq)
}

func (h HTTPHandler) RouteHTTP(_ *http.Request) fweight.Router {
	return fweight.Handle(h)
}

type Handler struct {
	*Archetype
	ContentMarshaler
}

func (h Handler) ServeObject(o interface{}, rw http.ResponseWriter, rq *http.Request) {
	var ms []ContentMarshaler
	if h.Archetype != nil {
		ms = []ContentMarshaler{
			h.ContentMarshaler,
			h.Archetype.ContentMarshaler,
		}
	} else {
		ms = []ContentMarshaler{h.ContentMarshaler}
	}

	mf, ct := RequestMarshaler(rq, ms...)
	switch {
	case mf == nil:
		rw.WriteHeader(406)
		if h.ContentMarshaler != nil {
			if mf = h.ContentMarshaler[""]; mf != nil {
				break
			}
		}
		if h.Archetype != nil {
			if mf = h.Archetype.ContentMarshaler[""]; mf != nil {
				break
			}
		}
		//wtf man
		mf = plain
		o = "None of the specified Content-Types supported."
	}

	dt, ctt, err := mf(o, ct.MediaType, ct.Params)
	if err != nil {
		panic(err)
	}
	rw.Header().Add("Content-Type", ctt)
	rw.Write(dt)
}
