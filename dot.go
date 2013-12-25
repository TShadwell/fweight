package fweight

import (
	"mime"
	"net/http"
	"path"
	"strings"
)

//DotContent implements a http.Handler that provides
//a middleware that adds the extension's media type
//to the ContentType of the request, and removes the extension.
var DotContent Middleware = MiddlewareFunc(func(h http.Handler) http.Handler {
	const ctt = "Accept"
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		defer h.ServeHTTP(rw, rq)
		//attempt to get extension
		ext := path.Ext(rq.URL.Path)
		if ext == "" {
			return
		}

		//attempt to get media type
		typ := mime.TypeByExtension(ext)
		if typ == "" {
			return
		}

		rq.URL.Path = strings.TrimRight(rq.URL.Path, ext)

		if v, ok := rq.Header[ctt]; ok && (len(v) > 0) {
			rq.Header[ctt][0] = strings.TrimRight(typ+rq.Header[ctt][0], ",")
			return
		}
		rq.Header[ctt] = []string{
			typ,
		}
	})
})
