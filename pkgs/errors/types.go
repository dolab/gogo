package errors

import (
	"encoding/json"
	"fmt"
)

// SprintError returns a json string of the formatted error code.
//
// Both extra and origErr are optional. If they are included their lines
// will be added, but if they are not included their lines will be ignored.
func SprintError(code, message, extra string, origErr error) string {
	data := map[string]interface{}{
		"Code":    code,
		"Message": message,
	}
	if len(extra) > 0 {
		data["Extra"] = extra
	}
	if origErr != nil {
		if err, ok := origErr.(Error); ok {
			data["OrigError"] = err
		} else {
			data["OrigError"] = fmt.Sprintf("%T(%s)", origErr, origErr)
		}
	}

	b, _ := json.Marshal(data)
	return string(b)
}

// A baseError wraps the code and message which defines an error. It also
// can be used to wrap an original error object.
//
// Should be used as the root for errors satisfying the errors.Error. Also
// for any error which does not fit into a specific error wrapper type.
type baseError struct {
	// Classification of error
	code string

	// Detailed information about error
	message string

	// Optional original error this error is based off of. Allows building
	// chained errors.
	errs []error
}

// newBaseError returns an error object for the code, message, and errors.
//
// code is a short no whitespace phrase depicting the classification of
// the error that is being created.
//
// message is the free flow string containing detailed information about the
// error.
//
// origErrs is the error objects which will be nested under the new errors to
// be returned.
func newBaseError(code, message string, origErrs []error) *baseError {
	b := &baseError{
		code:    code,
		message: message,
		errs:    origErrs,
	}

	return b
}

// Error returns the string representation of the error.
//
// See ErrorWithExtra for formatting.
//
// Satisfies the error interface.
func (e *baseError) Error() string {
	size := len(e.errs)
	if size > 0 {
		return SprintError(e.code, e.message, "", errorList(e.errs))
	}

	return SprintError(e.code, e.message, "", nil)
}

// String returns the string representation of the error.
// Alias for Error to satisfy the stringer interface.
func (e *baseError) String() string {
	return e.Error()
}

// Code returns the short phrase depicting the classification of the error.
func (e *baseError) Code() string {
	return e.code
}

// Message returns the error details message.
func (e *baseError) Message() string {
	return e.message
}

// OrigErr returns the original error if one was set. Nil is returned if no
// error was set. This only returns the first element in the list. If the full
// list is needed, use BatchedErrors.
func (e *baseError) OrigErr() error {
	switch len(e.errs) {
	case 0:
		return nil
	case 1:
		return e.errs[0]
	default:
		if err, ok := e.errs[0].(Error); ok {
			return NewBatchedErrors(err.Code(), err.Message(), e.errs[1:])
		}

		return NewBatchedErrors("BatchedErrors", "multiple errors occurred", e.errs)
	}
}

// OrigErrs returns the original errors if one was set. An empty slice is
// returned if no error was set.
func (e *baseError) OrigErrs() []error {
	return e.errs
}

func (e *baseError) MarshalJSON() ([]byte, error) {
	return []byte(e.Error()), nil
}

// So that the Error interface type can be included as an anonymous field
// in the requestError struct and not conflict with the error.Error() method.
type innerError Error

// A requestError wraps a request or service error.
//
// Composed of baseError for code, message, and original error.
type requestError struct {
	innerError

	statusCode int
	requestID  string
}

// newRequestError returns a wrapped error with additional information for
// request status code, and service requestID.
//
// Should be used to wrap all request which involve service requests. Even if
// the request failed without a service response, but had an HTTP status code
// that may be meaningful.
//
// Also wraps original errors via the baseError.
func newRequestError(err Error, statusCode int, requestID string) *requestError {
	return &requestError{
		innerError: err,
		statusCode: statusCode,
		requestID:  requestID,
	}
}

// Error returns the string representation of the error.
// Satisfies the error interface.
func (e *requestError) Error() string {
	var extra string
	if len(e.requestID) > 0 {
		extra = fmt.Sprintf("status code: %d, request id: %s", e.statusCode, e.requestID)
	} else {
		extra = fmt.Sprintf("status code: %d", e.statusCode)
	}

	return SprintError(e.Code(), e.Message(), extra, e.OrigErr())
}

// String returns the string representation of the error.
// Alias for Error to satisfy the stringer interface.
func (e *requestError) String() string {
	return e.Error()
}

// StatusCode returns the wrapped status code for the error
func (e *requestError) StatusCode() int {
	return e.statusCode
}

// RequestID returns the wrapped requestID
func (e *requestError) RequestID() string {
	return e.requestID
}

// OrigErrs returns the original errors if one was set. An empty slice is
// returned if no error was set.
func (e *requestError) OrigErrs() []error {
	if b, ok := e.innerError.(BatchedErrors); ok {
		return b.OrigErrs()
	}

	return []error{e.OrigErr()}
}

func (e *requestError) MarshalJSON() ([]byte, error) {
	return []byte(e.Error()), nil
}

// A wrappedRequestError wraps a request or service error for default, and
// exposes methods for custom at runtime.
//
// Composed of baseError for code, message, and original error.
type wrappedRequestError struct {
	*requestError
	err error
}

// newWrappedRequestError returns a wrapped error with requestError.
//
// Should be used to wrap all request which involve service requests. Even if
// the request failed without a service response, but had an HTTP status code
// that may be meaningful.
//
// Also wraps original errors via the baseError.
func newWrappedRequestError(statusCode int, code, message string) *wrappedRequestError {
	return &wrappedRequestError{
		requestError: newRequestError(newBaseError(code, message, nil), statusCode, ""),
	}
}

// Error returns the string representation of the error.
// Satisfies the error interface.
func (e *wrappedRequestError) Error() string {
	var extra string
	if len(e.requestID) > 0 {
		extra = fmt.Sprintf("status code: %d, request id: %s", e.statusCode, e.requestID)
	} else {
		extra = fmt.Sprintf("status code: %d", e.statusCode)
	}

	return SprintError(e.Code(), e.Message(), extra, e.OrigErr())
}

// String returns the string representation of the error.
// Alias for Error to satisfy the stringer interface.
func (e *wrappedRequestError) String() string {
	return e.Error()
}

// OrigErr returns the original error if one was set. Nil is returned if no
// error was set. This only returns the first element in the list. If the full
// list is needed, use BatchedErrors.
func (e *wrappedRequestError) OrigErr() error {
	if e.err != nil {
		return e.err
	}

	return e.requestError.OrigErr()
}

// OrigErrs returns the original errors if one was set. An empty slice is
// returned if no error was set.
func (e *wrappedRequestError) OrigErrs() []error {
	errs := e.requestError.OrigErrs()
	if e.err != nil {
		errs = append(errs, e.err)
	}

	return errs
}

func (e *wrappedRequestError) MarshalJSON() ([]byte, error) {
	return []byte(e.Error()), nil
}

// WithStatusCode allows injecting http status code at runtime
func (e *wrappedRequestError) WithStatusCode(statusCode int) WrappedRequestFailure {
	e.statusCode = statusCode

	return e
}

// WithRequestID allows injecting request id at runtime
func (e *wrappedRequestError) WithRequestID(requestID string) WrappedRequestFailure {
	e.requestID = requestID

	return e
}

// WithError allows injecting error at runtime
func (e *wrappedRequestError) WithError(err error) WrappedRequestFailure {
	e.err = err

	return e
}

// An error list that satisfies the golang interface
type errorList []error

// Error returns the string representation of the error.
//
// Satisfies the error interface.
func (e errorList) Error() string {
	msg := ""
	// How do we want to handle the array size being zero
	if size := len(e); size > 0 {
		for i := 0; i < size; i++ {
			msg += fmt.Sprintf("%s", e[i].Error())

			// We check the next index to see if it is within the slice.
			// If it is, then we append a semicolon.
			if i+1 < size {
				msg += "; "
			}
		}
	}

	return msg
}
