package fweight

import (
	"net/http"
	"time"
)

//Functin Modified is a more general implement of net/http's ServeContent.
//Modifed handles If-Modified-Since requests. ETag, Content-Range and the rest can be handled by a Middleware.  A zero lastModified
//is assumed to mean not modified.
func Modified(w interface {
	Header() http.Header
	WriteHeader(int)
}, r *http.Request, lastModified time.Time) (modified bool) {
	if lastModified.IsZero() {
		return false
	}


	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && lastModified.Before(t.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
	return false
}
