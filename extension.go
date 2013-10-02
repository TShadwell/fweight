package fweight

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type Extension interface {
	TransformRequest(rw http.ResponseWriter,
		rq *http.Request) (http.ResponseWriter, *http.Request)
}

type Compression struct {
	Extension
	ErrorHandler func(error)
}

type compressor interface {
	io.Writer
	Close() error
	Flush() error
}

func (c compressionWrap) handleError(e error) {
	if c.errorHandler != nil {
		c.errorHandler(e)
	}
}

type compressionWrap struct {
	http.ResponseWriter
	compressor   compressor
	errorHandler func(error)
}

func (c compressionWrap) Write(b []byte) (i int, e error) {
	i, e = c.compressor.Write(b)
	if e != nil {
		c.handleError(e)
		return
	}
	if e = c.compressor.Flush(); e != nil {
		c.handleError(e)
		return
	}
	if e = c.compressor.Close(); e != nil {
		c.handleError(e)
		return
	}
	return
}

func (c Compression) TransformRequest(rw http.ResponseWriter, rq *http.Request) (http.ResponseWriter, *http.Request) {
	if ae := rq.Header.Get("Accept-Encoding"); ae == "" {
		return rw, rq
	} else {
		var compressor compressor
		for _, encoding := range strings.Split(strings.ToLower(ae), ",") {
			switch encoding {
			case "gzip":
				compressor = gzip.NewWriter(rw)
				rw.Header().Set("Content-Encoding", "gzip")
			case "deflate":
				var err error
				compressor, err = flate.NewWriter(rw, -1)
				if err != nil {
					compressor = nil
				} else {
					rw.Header().Set("Content-Encoding", "deflate")
				}
			}
			if compressor != nil {
				return compressionWrap{rw, compressor, c.ErrorHandler}, rq
			}
		}
	}

	return rw, rq
}
