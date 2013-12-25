package fweight

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type (
	compressor interface {
		io.Writer
		Close() error
		Flush() error
	}
	compression    func(io.Writer) compressor
	httpCompressor struct {
		header      int
		compression compression
		rw          http.ResponseWriter
		c           compressor
		buf         bytes.Buffer
	}
)

func (h *httpCompressor) Header() http.Header {
	return h.rw.Header()
}

func (h *httpCompressor) WriteHeader(i int) {
	h.header = i
}

func (h *httpCompressor) Write(b []byte) (int, error) {
	if h.c == nil {
		h.c = h.compression(&h.buf)
	}
	fmt.Printf("[!!] Bytes being written: %+q\n", b)
	return h.c.Write(b)
}

func (h *httpCompressor) Close() {
	if h.c != nil {
		h.rw.WriteHeader(h.header)
		if err := h.c.Flush(); err != nil {
			panic(err)
		}
		if _, err := io.Copy(h.rw, &h.buf); err != nil {
			panic(err)
		}

		if err := h.c.Close(); err != nil {
			panic(err)
		}
	}
}

func newGzip(r io.Writer) (c compressor) {
	return gzip.NewWriter(r)
}

func newFlate(r io.Writer) (c compressor) {
	var err error
	if c, err = flate.NewWriter(r, -1); err != nil {
		panic(fmt.Sprintf("[!] Flate error encountered: %+q", err.Error()))
	}
	return
}

/*
	Compressor implements a http.Handler that provides an intelligent
	compression middleware, detecting and leveraging acceptable
	Content-Encodings.
*/
var Compression Middleware = MiddlewareFunc(compressionMiddleware)

func compressionMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		var dontRecover bool
		defer func() {
			if debug {
				log.Println("[?] Compressor recovery closure running.")
			}
			if !dontRecover {
				if debug {
					log.Println("[?] Compressor recovery closure running- recovering.")
				}
				h.ServeHTTP(rw, rq)
			}
		}()
		var ae string
		if ae = rq.Header.Get("Accept-Encoding"); ae == "" {
			if debug {
				log.Printf("[?] Not compressing - no Accept-Encoding specified. Headers were: %+v", rq.Header)
			}
			return
		}
		var compression compression
		var header string
	LOOP:
		for _, encoding := range strings.Split(strings.ToLower(ae), ",") {
			switch encoding {
			case "gzip":
				compression = newGzip
				header = "gzip"
				break LOOP
			case "deflate":
				compression = newFlate
				header = "deflate"
				break LOOP
			}
		}
		if compression != nil {
			fW := &httpCompressor{
				compression: compression,
				rw:          rw,
			}
			defer fW.Close()
			dontRecover = true
			h.ServeHTTP(fW, rq)
			if fW.compression != nil {
				rw.Header().Set("Content-Encoding", header)
				if debug {
					log.Printf("[?] Responding to %s with Content-Encoding %s.", rq.URL.Host+rq.URL.Path, header)
				}
			} else if debug {
				log.Println("[~] Not written to, not encoding body.")
			}
		} else if debug {
			log.Printf("[?] No acceptable `Accept-Encoding`s - choices were: %+q", ae)
		}
	})
}
