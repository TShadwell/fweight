package fweight

import (
	"net/http"
)

type Status uint16

const (
	StatusContinue           Status = 100
	StatusSwitchingProtocols Status = 101

	StatusOK                   Status = 200
	StatusCreated              Status = 201
	StatusAccepted             Status = 202
	StatusNonAuthoritativeInfo Status = 203
	StatusNoContent            Status = 204
	StatusResetContent         Status = 205
	StatusPartialContent       Status = 206

	StatusMultipleChoices   Status = 300
	StatusMovedPermanently  Status = 301
	StatusFound             Status = 302
	StatusSeeOther          Status = 303
	StatusNotModified       Status = 304
	StatusUseProxy          Status = 305
	StatusTemporaryRedirect Status = 307

	StatusBadRequest                   Status = 400
	StatusUnauthorized                 Status = 401
	StatusPaymentRequired              Status = 402
	StatusForbidden                    Status = 403
	StatusNotFound                     Status = 404
	StatusMethodNotAllowed             Status = 405
	StatusNotAcceptable                Status = 406
	StatusProxyAuthRequired            Status = 407
	StatusRequestTimeout               Status = 408
	StatusConflict                     Status = 409
	StatusGone                         Status = 410
	StatusLengthRequired               Status = 411
	StatusPreconditionFailed           Status = 412
	StatusRequestEntityTooLarge        Status = 413
	StatusRequestURITooLong            Status = 414
	StatusUnsupportedMediaType         Status = 415
	StatusRequestedRangeNotSatisfiable Status = 416
	StatusExpectationFailed            Status = 417
	StatusTeapot                       Status = 418

	StatusInternalServerError     Status = 500
	StatusNotImplemented          Status = 501
	StatusBadGateway              Status = 502
	StatusServiceUnavailable      Status = 503
	StatusGatewayTimeout          Status = 504
	StatusHTTPVersionNotSupported Status = 505
)

//Function String returns the status text associated with this Status
//that would be returned by http.StatusText.
func (s Status) String() string {
	return http.StatusText(int(s))
}

func (s Status) HTTPStatusCode() Status {
	return s
}

type HTTPStatus interface {
	HTTPStatusCode() Status
}

/*
	Returns an Err as the concrete value
	of an error interface, or nil.
*/
func (s Status) Err() error {
	if s > 400 {
		return Err(s)
	}
	return nil
}

//Type Err represents an HTTP error code (Status > 400).
type Err Status

func (e Err) HTTPStatusCode() Status {
	return Status(e)
}

func (e Err) Error() string {
	return Status(e).String()
}
