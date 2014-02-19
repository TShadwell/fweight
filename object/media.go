package object

import (
	"mime"
	"path"
	"strings"
)

type MediaType string

//returns the restricted pattern for path.Match
func restricted(mediaType string) (patt string) {
	for _, chr := range mediaType {
		switch chr {
		case '?', '\\', '[':
			patt += "\\"
		}
		patt += string(chr)
	}
	return
}

//Returns the MIME type string associated with this path
func pathMime(p string) (typ string) {
	ext := path.Ext(p)
	if ext == "" {
		return
	}

	typ = mime.TypeByExtension(ext)
	return
}

type ContentType struct {
	MediaType
	Params map[string]string
}

//Function ParseContentType returns an array of ContentTypes corresponding to the
//Content-Type header string cts.
func ParseContentType(cts string) (c []ContentType, err error) {
	if cts == "" {
		return
	}

	ctts := strings.Split(cts, ",")
	c = make([]ContentType, len(cts))
	for i, v := range ctts {
		c[i].MediaType, c[i].Params = pmt(v)
	}
	return
}

func pmt(v string) (mt MediaType, p map[string]string) {
	var s string
	s, p, _ = mime.ParseMediaType(v)
	mt = MediaType(s)
	return
}
