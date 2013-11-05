package fweight

import (
	"log"
	"net/http"
)

var contentType = http.CanonicalHeaderKey("Content-Type")

type ErrorHandler interface {
	HandleHTTPError(e error, rw http.ResponseWriter, req *http.Request)
}

//ExtendedErr is the suggested format for errors in the server.
//Err represents the HTTP error code, Cause is a publicly available informational
//string, and AdditionalInformation represents debug information that should be logged.
type ExtendedErr struct {
	Err
	Cause                 string
	AdditionalInformation interface{}
}

//Function Error formats the ExtendedError in the format
//Err.String() + (e.Cause != ""?" :" + e.Cause:"").
func (e ExtendedErr) Error() (o string) {
	o = e.Err.Error()
	if e.Cause != "" {
		o += ": " + e.Cause
	}
	return
}

//The Server type provides the ServeHTTP function for fweight.
//The http.ResponseWriters served by Server are extended fweight.ResponseWriters,
//which can be used to access the Server's additional error handling functionality
//whilst being backward compatible with typical http.Handlers and HandlerFuncs.
//
//If you wish to add your own extra functionality, a non-nil Extension field is
//used on all requests/responseWriters before they are passed to handlers.
//The Compression extension should be used as an example of an Extension.
type Server struct {
	Router
	//The ErrorHandler is sent an error with an ExtendedErr type.
	//For StatusInternalServerError, the AdditionalInformation field contains
	//whatever information the subrouter paniced with.
	ErrorHandler
	Extensions []Extension
}

func (s *Server) Wrap(rw http.ResponseWriter, rq *http.Request) (o ResponseWriter) {
	o, ok := o.(ResponseWriter)
	if ok {
		return
	}
	return responseWrapper{
		s,
		rw,
	}
}

type responseWrapper struct {
	*Server
	http.ResponseWriter
}

/*
	Write writes the data to the connection as part of an HTTP reply.
	If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
	before writing the data. If the header does not contain a Content-Type line, Write adds
	a Content-Type set to the result of passing the initial 512 bytes to DetectContentType.
*/
func (r responseWrapper) Write(b []byte) (int, error) {
	return 0, nil
}

//Set the Router for this server. The Router handles all
//routing
func (s *Server) Route(sdr Router) *Server {
	s.Router = sdr
	return s
}

func isHandler(r Router) (b bool) {
	_, b = r.(Handler)
	return
}

/*
	Serves HTTP.

	http.ListenAndServe(":8080", new(Server))
*/
func (s *Server) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	if s != nil && s.Extensions != nil {
		for _, e := range s.Extensions {
			if k, ok := e.(RequestCompleteHooker); ok {
				defer func() {
					k.RequestCompleted(rw, rq)
				}()
			}
			rw, rq = e.TransformRequest(rw, rq)

		}

	}
	//A nil server will route all requests to NotFound.
	if s == nil || s.Router == nil {
		s.HandleHTTPError(Err(StatusNotFound), rw, rq)
		return
	}

	//Traverse the router tree
	for router := s.Router.RouteHTTP(rq); ; router = router.RouteHTTP(rq) {

		//If we have a nil router, serve a 404.
		if router == nil {
			s.HandleHTTPError(
				Err(StatusNotFound),
				rw,
				rq,
			)
			return

		}

		//if the type of the router is a Handler, we can terminate
		if hl, ok := router.(Handler); ok {
			//defer a function to recover panics within child functions.
			defer func() {
				if e := recover(); e != nil {
					s.HandleHTTPError(
						ExtendedErr{
							Err: Err(StatusInternalServerError),
							AdditionalInformation: e,
						},
						rw,
						rq,
					)
				}
				return
			}()

			//serve the response.
			hl.ServeHTTP(rw, rq)
			return
		}

	}

}

func (s *Server) ServeError(e Err, rw http.ResponseWriter, rq *http.Request) {
	var eh ErrorHandler
	if s.ErrorHandler != nil {
		eh = s.ErrorHandler
	} else {
		eh = DefaultErrorHandler
	}

	eh.HandleHTTPError(e, rw, rq)
}

func (s *Server) HandleHTTPError(e error, rw http.ResponseWriter, rq *http.Request) {
	if s != nil && s.ErrorHandler != nil {
		s.ErrorHandler.HandleHTTPError(e, rw, rq)
		return
	}
	DefaultErrorHandler.HandleHTTPError(e, rw, rq)
}

var DefaultErrorHandler ErrorHandler = defaultErrorHandler{}

type defaultErrorHandler struct{}

func (d defaultErrorHandler) HandleHTTPError(e error, rw http.ResponseWriter, rq *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("A panic was encountered when a response was attempted %+q\n", e)
		}
	}()
	rw.Header().Add(contentType, "text/plain")
	if v, ok := e.(ExtendedErr); ok {
		log.Printf("ERROR ("+rq.URL.String()+"): %+q %+q\n", v.Error(), v.AdditionalInformation)
	} else {
		log.Printf("ERROR ("+rq.URL.String()+"): %+q\n", e.Error())
	}
	rw.Write(errorBytes(e))
	return
}

func errorBytes(e error) []byte {
	return []byte("Error:" + e.Error())
}
