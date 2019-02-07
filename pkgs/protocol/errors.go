package protocol

import (
	"net/http"

	"github.com/dolab/gogo/pkgs/errors"
)

// Protocol errors
var (
	ErrCloseResponse           = errors.NewWrappedRequestFailure(http.StatusGone, "CloseResponse", "Failed to close response body.")
	ErrContextTimeout          = errors.NewWrappedRequestFailure(http.StatusBadRequest, "ContextTimeout", "Request has aborted because context was done.")
	ErrDoPOSTRequest           = errors.NewWrappedRequestFailure(http.StatusGone, "DoPOSTRequest", "Failed to do POST request.")
	ErrIncompleteRequest       = errors.NewWrappedRequestFailure(http.StatusBadRequest, "IncompleteRequest", "The request did not provide the number of bytes specified by the Content-Length HTTP header.")
	ErrIncompleteResponse      = errors.NewWrappedRequestFailure(http.StatusBadRequest, "IncompleteResponse", "The response did not provide the number of bytes specified by the Content-Length HTTP header.")
	ErrInternalError           = errors.NewWrappedRequestFailure(http.StatusInternalServerError, "InternalError", "We encountered an internal error. Please try again.")
	ErrInvalidPOSTRequest      = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InvalidPOSTRequest", "Could not build http POST request.")
	ErrInvalidProtocol         = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InvalidProtocol", "The protocol resolved from request Content-Type header is not available.")
	ErrInvalidResponseMessage  = errors.NewWrappedRequestFailure(http.StatusBadGateway, "InvalidMessage", "The message returned from service dit not validate against protocol marshaler.")
	ErrInvalidMarshaler        = errors.NewWrappedRequestFailure(http.StatusInternalServerError, "InvalidMarshaler", "The marshaler of protocol returned error with message.")
	ErrMalformedRequestMessage = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MalformedRequestMessage", "The message you provided was not well-formed or did not validate against our published schema.")
	ErrMalformedTimeLayout     = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MalformedTimeLayout", "The date time you provided was not well-formed or did not validate against our published schema.")
	ErrNotImplemented          = errors.NewWrappedRequestFailure(http.StatusNotImplemented, "NotImplemented", "The service invoked is not implemented.")
	ErrRequestTimeout          = errors.NewWrappedRequestFailure(http.StatusBadRequest, "RequestTimeout", "Your socket connection to the server was not read from or written to within the timeout period.")
	ErrRequestTimeTooSkewed    = errors.NewWrappedRequestFailure(http.StatusForbidden, "RequestTimeTooSkewed", "The difference between the request time and the server's time is too large.")
	ErrWritePOSTResponse       = errors.NewWrappedRequestFailure(http.StatusGone, "WritePOSTResponse", "Failed to write response body.")
)

// A RequestFailure is a shadow of error.RequestFailure for protocol
type RequestFailure interface {
	errors.RequestFailure
}
