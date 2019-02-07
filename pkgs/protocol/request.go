package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/dolab/gogo/pkgs/errors"
)

// A Request is http.Client of protocol
type Request struct {
	Protocoler
	HTTPClient

	mux sync.RWMutex
}

func (r *Request) WithHTTPClient(client HTTPClient) *Request {
	r.mux.Lock()
	r.HTTPClient = client
	r.mux.Unlock()

	return r
}

// Do is common code to make a request to the remote gogo rpc service.
func (r *Request) Do(ctx context.Context, absUrl string, in, out interface{}) (err error) {
	// is context done?
	if err := ctx.Err(); err != nil {
		return ErrContextTimeout.WithError(err)
	}

	// build request
	payload, err := r.Marshal(in)
	if err != nil {
		return ErrMalformedRequestMessage.WithError(err)
	}

	req, err := newRequest(ctx, absUrl, r.ContentType(), bytes.NewBuffer(payload))
	if err != nil {
		return ErrInvalidPOSTRequest.WithError(err)
	}

	// is context done?
	if err := ctx.Err(); err != nil {
		return ErrContextTimeout.WithError(err)
	}

	// send request
	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return ErrDoPOSTRequest.WithError(err)
	}
	defer func() {
		tmperr := resp.Body.Close()
		if err == nil && tmperr != nil {
			err = ErrCloseResponse.WithError(tmperr)
		}
	}()

	// is context done?
	if err = ctx.Err(); err != nil {
		return ErrContextTimeout.WithError(err)
	}

	// is response 2xx?
	if resp.StatusCode/100 != 2 {
		return newErrorFromResponse(resp)
	}

	// build response
	err = r.Unmarshal(resp.Body, out)
	if err != nil {
		return ErrInvalidResponseMessage.WithError(err)
	}

	// is context done?
	if err = ctx.Err(); err != nil {
		return ErrContextTimeout.WithError(err)
	}
	return nil
}

// The standard library will, by default, redirect requests (including POSTs) if it gets a 302 or
// 303 response, and also 301s in go1.8. It redirects by making a second request, changing the
// method to GET and removing the body. This produces very confusing error messages, so instead we
// set a redirect policy that always errors. This stops Go from executing the redirect.
//
// We have to be a little careful in case the user-provided http.Client has its own CheckRedirect
// policy - if so, we'll run through that policy first.
//
// Because this requires modifying the http.Client, we make a new copy of the client and return it.
func hijackHTTPClientRedirects(client *http.Client) *http.Client {
	copy := *client
	copy.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if client.CheckRedirect != nil {
			// Run the input's redirect if it exists, in case it has side effects, but ignore any error it
			// returns, since we want to use ErrUseLastResponse.
			err := client.CheckRedirect(req, via)

			_ = err // Silly, but this makes sure generated code passes errcheck -blank, which some people use.
		}
		return http.ErrUseLastResponse
	}

	return &copy
}

// newRequest makes an http.Request from a client, adding common headers.
func newRequest(ctx context.Context, absUrl, contentType string, payload io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("POST", absUrl, payload)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("X-Gogo-Version", "v1.0.0")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)

	return req, nil
}

// A errorJSON is used for parse gogo errors
type errorJSON struct {
	Code    string
	Message string
	Extra   string
	Error   []errorJSON `json:"Error,omitempty"`
}

// newErrorFromResponse builds a errors.RequestFailure from a non-200 HTTP response.
// If the response has a valid serialized errors, then it's returned.
// If not, the response status code is used to generate a similar errors error.
func newErrorFromResponse(resp *http.Response) errors.RequestFailure {
	statusCode := resp.StatusCode
	statusText := http.StatusText(statusCode)

	// is the response a redirect?
	if statusCode >= 300 && statusCode <= 399 {
		// Unexpected redirect: it must be an error from an intermediary.
		// Protocol clients don't follow redirects automatically. Cause protocol always issue a
		// POST requests, redirects should only happen on GET and HEAD requests.
		tmperr := errors.New("IntermediaryError", fmt.Sprintf(
			"unexpected HTTP response status code: %d %q, Location=%q", statusCode, statusText, resp.Header.Get("Location"),
		), nil)

		return errors.NewRequestFailure(tmperr, http.StatusNotAcceptable, "")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrIncompleteResponse.WithError(err)
	}

	var (
		jsonErr errorJSON
		origErr error
	)

	if err := json.Unmarshal(body, &jsonErr); err != nil {
		// Invalid JSON response; it must be an error from intermediary.
		tmperr := errors.New("IntermediaryError", fmt.Sprintf(
			"Error from intermediary with HTTP status code: %d %q", statusCode, statusText,
		), nil)

		return errors.NewRequestFailure(tmperr, http.StatusBadGateway, "")
	}

	if len(jsonErr.Error) > 0 {
		origErr = fmt.Errorf("%s", jsonErr.Error)
	}

	tmperr := errors.New(jsonErr.Code, jsonErr.Message, origErr)

	return errors.NewRequestFailure(tmperr, statusCode, jsonErr.Extra)
}
