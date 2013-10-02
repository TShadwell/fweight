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
		rq *http.Request)(http.ResponseWriter, *http.Request)
}

type Compression struct {
	Extension
}


type compressionWrap struct {
	http.ResponseWriter
	compressor io.Writer
}

func (c compressionWrap) Write(b []byte) (n int, err error) {
	return c.compressor.Write(b)
}

func(c Compression) TransformRequest(rw http.ResponseWriter, rq *http.Request) (http.ResponseWriter, *http.Request){
	if ae := rq.Header.Get("Accept-Encoding");ae== ""{
		return rw, rq
	} else {
		var compressor io.Writer
		for _, encoding :=  range strings.Split(strings.ToLower(ae), ","){
			switch encoding {
			case "gzip":
				compressor=gzip.NewWriter(rw)
				rw.Header().Set("Content-Encoding", "gzip")
			case "deflate":
				var err error
				compressor, err = flate.NewWriter(rw, -1)
				if err != nil{
					compressor = nil
				} else {
					rw.Header().Set("Content-Encoding", "deflate")
				}
			}
			if compressor != nil{
				return compressionWrap{rw, compressor}, rq
			}
		}
	}


	return rw, rq
}
