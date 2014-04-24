//Package resource provides some boilerplate code
//for serving static resources.
package resource

import (
	"github.com/TShadwell/fweight/route"
	"net/http"
	"strings"
	"time"
)

var initTime = time.Now()

type Resource struct {
	Name     string //name or extension, used for sniffing
	Value    string
	Modified time.Time //defaults to the time program was started
}

func (r Resource) RouteHTTP(_ *http.Request) route.Router { return route.Handle(r) }
func (r Resource) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	if r.Modified.Equal(time.Time{}) {
		r.Modified = initTime
	}
	http.ServeContent(w, rq, r.Name, r.Modified, strings.NewReader(r.Value))
}
