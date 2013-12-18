package fweight

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

/*
	Compressor implements a http.Handler that provides an intelligent
	compression middleware, detecting and leveraging acceptable
	Content-Encodings.
*/
type Compressor struct {
	Handler http.Handler
}

func NewCompressor(h http.Handler) Compressor {
	return Compressor{h}
}

var Compression Middleware = MiddlewareFunc(func(h http.Handler) http.Handler {
	return NewCompressor(h)
})

type (
	compressor interface {
		io.Writer
		Close() error
		Flush() error
	}
	httpCompressor struct {
		compressor
		http.ResponseWriter
	}
)

func (h httpCompressor) Write(b []byte) (int, error) {
	return h.compressor.Write(b)
}

func (c Compressor) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	var closure = func() {
		c.Handler.ServeHTTP(rw, rq)
	}
	defer closure()
	var ae string
	if ae = rq.Header.Get("Accept-Encoding"); ae == "" {
		return
	}
	var compressor interface {
		io.Writer
		Close() error
		Flush() error
	}
	var header string
	for _, encoding := range strings.Split(strings.ToLower(ae), ",") {
		switch encoding {
		case "gzip":
			compressor = gzip.NewWriter(rw)
			header = "gzip"
		case "deflate":
			var err error
			if compressor, err = flate.NewWriter(rw, -1); err != nil {
				compressor = nil
			} else {
				header = "deflate"
			}
		}
	}

	if compressor != nil {
		rw.Header().Set("Content-Encoding", header)
		closure = func() {
			c.Handler.ServeHTTP(httpCompressor{
				compressor:     compressor,
				ResponseWriter: rw,
			}, rq)
			defer compressor.Flush()
			defer compressor.Close()
		}
	}
}
