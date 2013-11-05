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

type RequestCompleteHooker interface {
	RequestCompleted(rw http.ResponseWriter, rq *http.Request)
}

type IgnorePort struct {
}

func (i IgnorePort) TransformRequest(rw http.ResponseWriter, rq *http.Request) (http.ResponseWriter, *http.Request) {
	if spl := strings.SplitN(rq.Host, ":", 2); len(spl) > 1 {
		rq.Host = spl[0]
	}
	return rw, rq
}

type Compression struct {
	compressor
}

func (c Compression) RequestCompleted(rw http.ResponseWriter, rq *http.Request) {
	rw.(compressionWrap).Close()
}

type compressor interface {
	io.Writer
	Close() error
	Flush() error
}

type compressionWrap struct {
	rw         http.ResponseWriter
	compressor compressor
}

func (c compressionWrap) Close() {
	c.compressor.Flush()
	c.compressor.Close()
}

func (c compressionWrap) Write(b []byte) (n int, err error) {

	var bt = b
	var cnt int = 0
	for ; cnt < len(b); bt = bt[cnt:] {
		var thisW int
		thisW, err = c.compressor.Write(bt)
		cnt += thisW
		if err != nil {
			return cnt, err
		}
	}

	return cnt, err
}

func (c compressionWrap) Header() http.Header {
	return c.rw.Header()
}

func (c compressionWrap) WriteHeader(an int) {
	c.rw.WriteHeader(an)
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
				c.compressor = compressor
				return compressionWrap{rw, compressor}, rq
			}
		}
	}

	return rw, rq
}
