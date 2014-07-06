package compression

import (
	"compress/flate"
	"compress/gzip"
	"github.com/TShadwell/fweight"
	"io"
	"net/http"
	"strings"
	"sync"
)

type (
	Compressor interface {
		io.Writer
		Close() error
	}

	Compression func(io.Writer) Compressor
)

var compressions = map[string]Compression{
	"gzip":  Gzip,
	"flate": Flate,
}

var rwm sync.RWMutex

/*
	Registers a new compression with this package. Gzip and Flate are already registered.
	Name should be canonicalised to lower case.
*/
func Register(name string, c Compression) {
	rwm.Lock()
	compressions[name] = c
	rwm.Unlock()
}

func Gzip(r io.Writer) Compressor {
	return gzip.NewWriter(r)
}

func Flate(r io.Writer) (c Compressor) {
	var err error
	if c, err = flate.NewWriter(r, -1); err != nil {
		panic(err)
	}
	return
}

//writer is a replacement http.ResponseWriter
//used to intercept and compress written bytes.
type writer struct {
	rw http.ResponseWriter
	Compression
	c Compressor
}

func (w *writer) Header() http.Header {
	return w.rw.Header()
}

func (w *writer) Write(b []byte) (int, error) {
	//only load compression on first write.
	//this should prevent errors when status disallows body.
	if w.c == nil {
		w.c = w.Compression(w.rw)
	}
	return w.c.Write(b)
}

func (w *writer) WriteHeader(i int) {
	w.rw.WriteHeader(i)
}

func (w *writer) flush() {
	if w.c != nil {
		if err := w.c.Close(); err != nil {
			panic(err)
		}
	}
}

var Middleware = fweight.MiddlewareFunc(func(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var flush func()

		defer func() {
			h.ServeHTTP(w, r)
			if flush != nil {
				flush()
			}
		}()

		encs := r.Header.Get("Accept-Encoding")
		if encs == "" {
			return
		}

		var compression Compression
		var encoding string

		rwm.RLock()
		for _, encoding = range strings.Split(strings.ToLower(encs), ",") {
			if compression = compressions[encoding]; compression != nil {
				break
			}
		}
		rwm.RUnlock()

		if compression == nil {
			return
		}

		//copy the old writer
		ow := w

		//create our writer, which compresses.
		uw := writer{
			rw:          ow,
			Compression: compression,
		}

		//replace the writer (for the deferred Handler)
		w = &uw

		w.Header().Set("Content-Encoding", encoding)
		w.Header().Set("Vary", "Accept-Encoding")

		//once it's written to our writer, flush it
		flush = uw.flush
	})
})

func init() {

}
