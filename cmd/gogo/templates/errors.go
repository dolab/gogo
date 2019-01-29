package templates

var (
	errorsTemplate = `package defaults

import (
	"net/http"

	"github.com/dolab/gogo/pkgs/errors"
)

var (
	ErrAccessDenied               = errors.NewWrappedRequestFailure(http.StatusForbidden, "AccessDenied", "Access Denied")
	ErrAccountProblem             = errors.NewWrappedRequestFailure(http.StatusForbidden, "AccountProblem", "There is a problem with your account that prevents the operation from completing successfully. Please use Contact Us.")
	ErrBadDigest                  = errors.NewWrappedRequestFailure(http.StatusBadRequest, "BadDigest", "The Content-MD5 you specified did not match what we received.")
	ErrCredentialsNotSupported    = errors.NewWrappedRequestFailure(http.StatusBadRequest, "CredentialsNotSupported", "This request does not support credentials.")
	ErrCrossLocationProhibited    = errors.NewWrappedRequestFailure(http.StatusForbidden, "CrossLocationProhibited", "Cross-location is not allowed.")
	ErrEntityTooSmall             = errors.NewWrappedRequestFailure(http.StatusBadRequest, "EntityTooSmall", "The proposed request body is smaller than the minimum allowed object size.")
	ErrEntityTooLarge             = errors.NewWrappedRequestFailure(http.StatusBadRequest, "EntityTooLarge", "The proposed request body exceeds the maximum allowed object size.")
	ErrExpiredToken               = errors.NewWrappedRequestFailure(http.StatusBadRequest, "ExpiredToken", "The provided token has expired.")
	ErrIllegalVersioningException = errors.NewWrappedRequestFailure(http.StatusBadRequest, "IllegalVersioningException", "The versioning specified in the request is invalid.")
	ErrIncompleteBody             = errors.NewWrappedRequestFailure(http.StatusBadRequest, "IncompleteBody", "The request did not provide the number of bytes specified by the Content-Length HTTP header.")
	ErrInlineDataTooLarge         = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InlineDataTooLarge", "Inline data exceeds the maximum allowed size.")
	ErrInternalError              = errors.NewWrappedRequestFailure(http.StatusInternalServerError, "InternalError", "We encountered an internal error. Please try again.")
	ErrInvalidArgument            = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InvalidArgument", "Invalid Argument")
	ErrInvalidDigest              = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InvalidDigest", "The Content-MD5 you specified is not valid.")
	ErrInvalidEncryptionAlgorithm = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InvalidEncryptionAlgorithm", "The encryption request you specified is not valid. The valid value is AES256.")
	ErrInvalidLocationConstraint  = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InvalidLocationConstraint", "The specified location constraint is not valid.")
	ErrInvalidRange               = errors.NewWrappedRequestFailure(http.StatusRequestedRangeNotSatisfiable, "InvalidRange", "The requested range cannot be satisfied.")
	ErrInvalidRequest             = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InvalidRequest", "All requests must be made over an HTTPS connection.")
	ErrInvalidSecurity            = errors.NewWrappedRequestFailure(http.StatusForbidden, "InvalidSecurity", "The provided security credentials are not valid.")
	ErrInvalidToken               = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InvalidToken", "The provided token is malformed or otherwise invalid.")
	ErrInvalidURI                 = errors.NewWrappedRequestFailure(http.StatusBadRequest, "InvalidURI", "Couldn't parse the specified URI.")
	ErrMalformedJSON              = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MalformedJSON", "The JSON you provided was not well-formed or did not validate against our published schema.")
	ErrMalformedPOSTRequest       = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MalformedPOSTRequest", "The body of your POST request is not well-formed multipart/form-data.")
	ErrMalformedXML               = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MalformedXML", "The XML you provided was not well-formed or did not validate against our published schema.")
	ErrMalformedTimeLayout        = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MalformedTimeLayout", "The date time you provided was not well-formed or did not validate against our published schema.")
	ErrMaxMessageLengthExceeded   = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MaxMessageLengthExceeded", "Your request was too big.")
	ErrMetadataTooLarge           = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MetadataTooLarge", "Your metadata headers exceed the maximum allowed metadata size.")
	ErrMethodNotAllowed           = errors.NewWrappedRequestFailure(http.StatusMethodNotAllowed, "MethodNotAllowed", "The specified method is not allowed against this resource.")
	ErrMissingAttachment          = errors.NewWrappedRequestFailure(http.StatusNotAcceptable, "MissingAttachment", "A request attachment was expected, but none were found.")
	ErrMissingContentLength       = errors.NewWrappedRequestFailure(http.StatusLengthRequired, "MissingContentLength", "You must provide the Content-Length HTTP header.")
	ErrMissingRequestBody         = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MissingRequestBody", "Request body is empty.")
	ErrMissingSecurityElement     = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MissingSecurityElement", "The request is missing a security element.")
	ErrMissingSecurityHeader      = errors.NewWrappedRequestFailure(http.StatusBadRequest, "MissingSecurityHeader", "Your request is missing a required header.")
	ErrNoSuchVersion              = errors.NewWrappedRequestFailure(http.StatusNotFound, "NoSuchVersion", "The version ID specified in the request does not match an existing version.")
	ErrNotImplemented             = errors.NewWrappedRequestFailure(http.StatusNotImplemented, "NotImplemented", "A header you provided implies functionality that is not implemented.")
	ErrNotModified                = errors.NewWrappedRequestFailure(http.StatusNotModified, "NotModified", "Not Modified")
	ErrPermanentRedirect          = errors.NewWrappedRequestFailure(http.StatusMovedPermanently, "PermanentRedirect", "The resource you are attempting to access must be addressed using the specified endpoint.")
	ErrPreconditionFailed         = errors.NewWrappedRequestFailure(http.StatusPreconditionFailed, "PreconditionFailed", "At least one of the preconditions you specified did not hold.")
	ErrRedirect                   = errors.NewWrappedRequestFailure(http.StatusTemporaryRedirect, "Redirect", "Temporary redirect.")
	ErrRequestTimeout             = errors.NewWrappedRequestFailure(http.StatusBadRequest, "RequestTimeout", "Your socket connection to the server was not read from or written to within the timeout period.")
	ErrRequestTimeTooSkewed       = errors.NewWrappedRequestFailure(http.StatusForbidden, "RequestTimeTooSkewed", "The difference between the request time and the server's time is too large.")
	ErrSignatureDoesNotMatch      = errors.NewWrappedRequestFailure(http.StatusForbidden, "SignatureDoesNotMatch", "The request signature we calculated does not match the signature you provided.")
	ErrServiceUnavailable         = errors.NewWrappedRequestFailure(http.StatusServiceUnavailable, "ServiceUnavailable", "Reduce your request rate.")
	ErrSlowDown                   = errors.NewWrappedRequestFailure(http.StatusServiceUnavailable, "SlowDown", "Reduce your request rate.")
	ErrTemporaryRedirect          = errors.NewWrappedRequestFailure(http.StatusTemporaryRedirect, "TemporaryRedirect", "You are being redirected to the bucket while DNS updates.")
	ErrUnexpectedContent          = errors.NewWrappedRequestFailure(http.StatusBadRequest, "UnexpectedContent", "This request does not support content.")
)
`
)
