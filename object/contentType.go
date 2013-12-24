package object

import (
	"mime"
	"strings"
)

type contentType struct {
	mediaType MediaType
	params    map[string]string
}

func parseContentType(ct string) (c []contentType, err error) {
	if ct == "" {
		return
	}

	cts := strings.Split(ct, ",")
	c = make([]contentType, len(cts))
	for i, v := range cts {
		var (
			mt string
			pr map[string]string
		)
		mt, pr, err = mime.ParseMediaType(v)
		if err != nil {
			return
		}
		c[i] = contentType{
			mediaType: MediaType(mt),
			params:    pr,
		}
	}
	return
}
