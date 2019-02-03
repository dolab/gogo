package named

import (
	"net/url"
	"strings"
)

// ToURL helps ensure that a url specifies a scheme.
// If the url specifies a scheme, use it. If not, default to
// http. If url.Parse fails on it, return it unchanged.
func ToURL(aurl string, isHTTPS ...bool) string {
	aurl = strings.TrimSpace(aurl)

	urlobj, err := url.Parse(aurl)
	if err != nil {
		return aurl
	}

	// default to http
	if len(urlobj.Scheme) == 0 {
		if len(isHTTPS) > 0 && isHTTPS[0] {
			urlobj.Scheme = "https"
		} else {
			urlobj.Scheme = "http"
		}
	}

	return urlobj.String()
}
