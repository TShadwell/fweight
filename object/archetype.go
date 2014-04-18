package object

import (
	"github.com/TShadwell/fweight/route"
	"log"
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

func RouterFunc(g GetterFunc) *HTTPHandler {
	return DefaultArchetype.Router(GetterFunc(g))
}

func Router(g Getter) *HTTPHandler {
	return DefaultArchetype.Router(g)
}

func (a *Archetype) Handler() Handler {
	return Handler{
		Archetype: a,
	}
}

func (a *Archetype) Router(g Getter) *HTTPHandler {
	return &HTTPHandler{
		Getter: g,
		Handler: Handler{
			Archetype: a,
		},
	}
}

func (a *Archetype) RouterFunc(g GetterFunc) *HTTPHandler {
	return a.Router(GetterFunc(g))
}

type HTTPHandler struct {
	Getter
	Handler
}

func (h *HTTPHandler) Bind(m MediaType, mf MarshalFunc) {
	if h.Handler.ContentMarshaler == nil {
		h.Handler.ContentMarshaler = make(ContentMarshaler)
	}
	h.Handler.ContentMarshaler[m] = mf
}

func (h HTTPHandler) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	h.Handler.ServeObject(h.Getter.Get(rw, rq), rw, rq)
}

func (h HTTPHandler) RouteHTTP(_ *http.Request) route.Router {
	return route.Handle(h)
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

	if debug {
		log.Printf("supported: %+v %+v\n", h.ContentMarshaler, h.Archetype.ContentMarshaler)
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

/*
Function BindFweight binds this Archetype to Fweight's OptionsHandler
and MethodNotAllowed handlers.
*/
func (a *Archetype) Bind() {
	hnd := a.Handler()
	route.OptionsHandler = hnd.ServeObject
	route.MethodNotAllowed = hnd.ServeObject
}
