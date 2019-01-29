// Package errors represents API error interface accessors for GOGO.
package errors

// An Error wraps lower level errors with code, message and an original error.
// The underlying concrete error type may also satisfy other interfaces which
// can be to used to obtain more specific information about the error.
//
// Calling Error() or String() will always include the full information about
// an error based on its underlying type.
//
// Example:
//
//     output, err := svc.Api(ctx, input)
//     if err != nil {
//         if gerr, ok := err.(errors.Error); ok {
//             // Get error details
//             log.Println("Error:", gerr.Code(), gerr.Message())
//
//             // Prints out full error message, including original error if there was one.
//             log.Println("Error:", gerr.Error())
//
//             // Get original error
//             if origErr := gerr.OrigErr(); origErr != nil {
//                 // operate on original error.
//             }
//         } else {
//             fmt.Println(err.Error())
//         }
//     }
//
type Error interface {
	// Satisfy the generic error interface.
	error

	// Returns the short phrase depicting the classification of the error.
	Code() string

	// Returns the error details message.
	Message() string

	// Returns the original error if one was set.  Nil is returned if not set.
	OrigErr() error
}

// New returns an Error object described by the code, message, and origErr.
//
// If origErr satisfies the Error interface it will not be wrapped within a new
// Error object and will instead be returned.
func New(code, message string, origErr error) Error {
	var errs []error
	if origErr != nil {
		errs = append(errs, origErr)
	}

	return newBaseError(code, message, errs)
}

// BatchedErrors is a batch of errors which also wraps lower level errors with
// code, message, and original errors. Calling Error() will include all errors
// that occurred in the batch.
type BatchedErrors interface {
	// Satisfy the base Error interface.
	Error

	// Returns the original error if one was set.  Nil is returned if not set.
	OrigErrs() []error
}

// NewBatchedErrors returns an BatchedErrors with a collection of errors as an
// array of errors.
func NewBatchedErrors(code, message string, errs []error) BatchedErrors {
	return newBaseError(code, message, errs)
}

// A RequestFailure is an interface to extract request failure information from
// an Error such as the request ID of the failed request returned by a service.
// RequestFailures may not always have a requestID value if the request failed
// prior to reaching the service such as a connection error.
//
// A RequestFailure implements StatusCoder of gogo which allows custom response
// code in the flying.
//
// Example:
//
//     output, err := svc.Api(ctx, input)
//     if err != nil {
//         if httpErr, ok := err.(RequestFailure); ok {
//             log.Println("Request failed", httpErr.Code(), httpErr.Message(), httpErr.RequestID())
//         } else {
//             log.Println("Error:", err.Error())
//         }
//     }
//
// Combined with errors.Error:
//
//    output, err := svc.Api(ctx, input)
//    if err != nil {
//        if gerr, ok := err.(errors.Error); ok {
//            // Generic Error with Code, Message, and original error (if any)
//            log.Println(gerr.Code(), gerr.Message(), gerr.OrigErr())
//
//            if httpErr, ok := err.(errors.RequestFailure); ok {
//                // A service error occurred
//                log.Println(httpErr.StatusCode(), httpErr.RequestID())
//            }
//        } else {
//            log.Println(err.Error())
//        }
//    }
//
type RequestFailure interface {
	Error

	// The status code of the HTTP response.
	StatusCode() int

	// The request ID returned by the service for a request failure. This will
	// be empty if no request ID is available such as the request failed due
	// to a connection error.
	RequestID() string
}

// NewRequestFailure returns a new request error wrapper for the given Error
// provided.
func NewRequestFailure(err Error, statusCode int, requestID string) RequestFailure {
	return newRequestError(err, statusCode, requestID)
}

// A WrappedRequestFailure is an interface to extract request failure infomation customable
// at runtime.
type WrappedRequestFailure interface {
	RequestFailure

	// Allows injecting http response status code at runtime
	WithStatusCode(statusCode int) WrappedRequestFailure

	// Allows injecting request id at runtime
	WithRequestID(requestID string) WrappedRequestFailure

	// Allows injecting error at runtime
	WithError(err error) WrappedRequestFailure
}

// NewWrappedRequestFailure returns a new wrapped request failure error.
func NewWrappedRequestFailure(statusCode int, code, message string) WrappedRequestFailure {
	return newWrappedRequestError(statusCode, code, message)
}
