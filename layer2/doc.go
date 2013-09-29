/*
	Package layer2 provides higher level web primitives,
	providing functions to convert information into specific MIME types,
	as well as an error handling system based on this ability.

	Succinctly, layer2 provides a data transformation centric system, including:
		- A system that extracts mediatype information from Accept: headers
			  and extensions, which is used to transform data stored in interface{} types.
		- An ErrorHandler that:
			- Transforms the error via a func(interface{}) interface{}
			- Marshals the response into the expected media type.
		- A Handler that:
			- Expands on the typical http.Handler to


*/
package layer2

import (
	"net/http"
)

/*
	A Handler uses interfaces to transform data.

	If the response implements `error`, HandleHTTPError is called
	with the error.

	If the error impliments HTTPStatusCode, the
	status code of the response is set to it.

	If the response impliments Transformer,
	the data that would be marshaled is the result
	of calling LayerDataTransform.

	If the response impliments HeaderSetter,
	the headers of the response are (over)written
	with that value.
*/
type Handler func(rq *http.Request) interface{}

func (h Handler) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	v := h(rq)
	switch n := v.(type) {
	case error:
	}
}

type Transformer interface {
	LayerDataTransform() interface{}
}

type HeaderSetter interface {
	HTTPHeader() http.Header
}


func HandlerOf(h Handler) http.Handler {
	return nil
}
