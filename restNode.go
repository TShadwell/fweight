
// +build ignore

package fweight

import (
	"mime"
	"net/http"
	"strings"
)

var contentTypeHeader = http.CanonicalHeaderKey("Content-Type")
var acceptHeader = http.CanonicalHeaderKey("Accept")

type Mime struct {
	Mediatype string
	Params    map[string]string
	Logger    ErrorHandler
}

type convertorParams struct {
	Convertor
	Params map[string]string
}

func (m Mime) Format() string {
	return mime.FormatMediaType(m.Mediatype, m.Params)
}

/*
	The MimeManager binds Content-Types to Convertor functions
	which take structs and write them to ResponseWriters.

	MimeNodes can be created from a MimeManager that are used
	to route requests to functions by their Content-Type and extensions.

	When a http error occurs, the Err will be passed to the convertor.
*/
type MimeManager struct {
	Map map[string]convertorParams
}

func (c convertorParams) Mime(mediatype string) Mime {
	return Mime{
		Mediatype: mediatype,
		Params:    c.Params,
	}
}

const defaultMime = "text/html"

func (m MimeManager) Any() (Mime, Convertor) {
	if cP, ok := m.Map[defaultMime]; ok {
		return cP.Mime(defaultMime), cP.Convertor
	}
	for mm, v := range m.Map {
		return v.Mime(mm), v.Convertor
	}
	return Mime{}, nil
}

func (m MimeManager) ConvertorMime(h http.Header) (mi Mime, c Convertor) {
	if acceptRaw := h.Get(acceptHeader); acceptRaw != "" {
		//Go through each MIME
		for _, v := range strings.Split(acceptRaw, ",") {
			//Split off the params, it's too complicated to work on at the moment.
			v = strings.SplitN(v, ";", 2)[0]

			if v == "*/*" {
				return m.Any()
			}
			//If we do not have a wildcard mime, it is fast and streightforward to look
			//for a convertor with the correct MIME.
			if strings.Contains(v, "*") {
				//Strip * on either side to get
				//the search string.
				v = strings.Replace(v, "*", "", -1)

				for k, cP := range m.Map {
					if strings.Contains(k, v) {
						return cP.Mime(k), cP.Convertor
					}
				}
			}
		}
		return
	}
	return m.Any()
}

func (m *MimeManager) AddConvertor(mediatype Mime, c Convertor) *MimeManager {
	if m.Map == nil {
		m.Map = make(map[string]convertorParams)
	}
	m.Map[mediatype.Mediatype] = convertorParams{
		Convertor: c,
		Params:    mediatype.Params,
	}
	return m
}

type (
	Convertor interface {
		Convert(interface{}) []byte
	}
	ResponseWrapper interface {
		Header() http.Header
		WriteHeader(int)
	}
	Generator interface {
		Generate(ResponseWrapper, *http.Request) interface{}
	}
	ConvertorFunc func(interface{}) []byte
	GeneratorFunc func(ResponseWrapper, *http.Request) interface{}
)

func (g GeneratorFunc) Generate(r ResponseWrapper, rq *http.Request) interface{} {
	return g(r, rq)
}

func (c ConvertorFunc) Convert(i interface{}) []byte {
	return c(i)
}

func (m *MimeManager) BindConvertor(c Convertor, Mime Mime) {
	m.Map[Mime.Mediatype] = convertorParams{
		Convertor: c,
		Params:    Mime.Params,
	}
}

func (m *MimeManager) handleErrMime(e Err, mime Mime, c Convertor, rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set(contentTypeHeader, mime.Format())
	rw.Write(c.Convert(e))
}

func (m *MimeManager) HandleErr(e Err, rw http.ResponseWriter, req *http.Request) {
	mm, cP := m.ConvertorMime(req.Header)
	if cP == nil {
		mm, cP = m.Any()
	}
	m.handleErrMime(e, mm, cP, rw, req)
}

type HTTPError struct {
	Status string
	Code   int
}

type Error struct {
	Error string
}

func (m *MimeManager) HandleHTTPError(e error, rw http.ResponseWriter, req *http.Request) {
	mm, cP := m.ConvertorMime(req.Header)
	if cP == nil {
		mm, cP = m.Any()
	}
	rw.Header().Set(contentTypeHeader, mm.Format())
	if v, ok := e.(Err); ok {
		rw.Write(
			cP.Convert(HTTPError{
				Status: e.Error(),
				Code:   int(v),
			}),
		)
		return
	}
	rw.Write(cP.Convert(Error{e.Error()}))
}

type MimeHandler struct {
	Manager   *MimeManager
	Overrides map[string]convertorParams
	Generator
}

func (m *MimeHandler) AddOverride(mimetype Mime, c Convertor) *MimeHandler {
	if m.Overrides == nil {
		m.Overrides = make(map[string]convertorParams)
	}
	m.Overrides[mimetype.Mediatype] = convertorParams{
		Convertor: c,
		Params:    mimetype.Params,
	}
	return m
}

func (m MimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mm, cP := m.Manager.ConvertorMime(r.Header)
	//Handle NotAcceptable
	if cP == nil {
		mm, cP = m.Manager.Any()
		m.Manager.handleErrMime(StatusNotAcceptable.Err().(Err), mm, cP, w, r)
		return
	}
	w.Header().Set(contentTypeHeader, mm.Format())
	w.Write(cP.Convert(m.Generate(w, r)))
}

func (m *MimeManager) Handle(g Generator) MimeHandler {
	return MimeHandler{
		Manager:   m,
		Generator: g,
	}
}
