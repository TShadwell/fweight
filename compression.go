package fweight

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"log"
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
	var dontRecover bool
	defer func() {
		log.Println("[?] Compressor recovery closure running.")
		if !dontRecover {
			log.Println("[?] Compressor recovery closure running- recovering.")
			c.Handler.ServeHTTP(rw, rq)
		}
	}()
	var ae string
	if ae = rq.Header.Get("Accept-Encoding"); ae == "" {
		if debug {
			log.Printf("[?] Not compressing - no Accept-Encoding specified. Headers were: %+v", rq.Header)
		}
		return
	}
	var compressor interface {
		io.Writer
		Close() error
		Flush() error
	}
	var header string
LOOP:
	for _, encoding := range strings.Split(strings.ToLower(ae), ",") {
		switch encoding {
		case "gzip":
			compressor = gzip.NewWriter(rw)
			header = "gzip"
			break LOOP
		case "deflate":
			var err error
			if compressor, err = flate.NewWriter(rw, -1); err != nil {
				compressor = nil
			} else {
				header = "deflate"
			}
			break LOOP
		}
	}

	if compressor != nil {
		rw.Header().Set("Content-Encoding", header)
		if debug {
			log.Println("[?] Responding with Content-Encoding " + header + ".")
		}
		dontRecover = true
		c.Handler.ServeHTTP(httpCompressor{
			compressor:     compressor,
			ResponseWriter: rw,
		}, rq)
		if err := compressor.Flush(); err != nil {
			panic(err)
		}
		if err := compressor.Close(); err != nil {
			panic(err)
		}
	} else if debug {
		log.Printf("[?] No acceptable `Accept-Encoding`s - choices were: %+q", ae)
	}
}
