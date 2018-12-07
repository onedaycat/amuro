package apigateway

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func Redirect(res *CustomResponse, req *CustomRequest, urlEndpoint string, code int) {
	// parseURL is just url.Parse (url is shadowed for godoc).
	if u, err := url.Parse(urlEndpoint); err == nil {
		// If url was relative, make its path absolute by
		// combining with request path.
		// The client would probably do this for us,
		// but doing it ourselves is more reliable.
		// See RFC 7231, section 7.1.2
		if u.Scheme == "" && u.Host == "" {
			oldpath := req.Path
			if oldpath == "" { // should not happen, but avoid a crash if it does
				oldpath = "/"
			}

			// no leading http://server
			if urlEndpoint == "" || urlEndpoint[0] != '/' {
				// make relative path absolute
				olddir, _ := path.Split(oldpath)
				urlEndpoint = olddir + urlEndpoint
			}

			var query string
			if i := strings.Index(urlEndpoint, "?"); i != defaultStatusCode {
				urlEndpoint, query = urlEndpoint[:i], urlEndpoint[i:]
			}

			// clean up but preserve trailing slash
			trailing := strings.HasSuffix(urlEndpoint, "/")
			urlEndpoint = path.Clean(urlEndpoint)
			if trailing && !strings.HasSuffix(urlEndpoint, "/") {
				urlEndpoint += "/"
			}
			urlEndpoint += query
		}
	}

	h := res.Headers

	// RFC 7231 notes that a short HTML body is usually included in
	// the response because older user agents may not understand 301/307.
	// Do it only if the request didn't already have a Content-Type header.
	_, hadCT := h["Content-Type"]

	res.Headers.Set("Location", hexEscapeNonASCII(urlEndpoint))
	if !hadCT && (req.HTTPMethod == "GET" || req.HTTPMethod == "HEAD") {
		res.Headers.Set("Content-Type", "text/html; charset=utf-8")
	}
	res.SetStatusCode(code)

	// Shouldn't send the body for POST or HEAD; that leaves GET.
	if !hadCT && req.HTTPMethod == "GET" {
		body := "<a href=\"" + htmlEscape(urlEndpoint) + "\">" + string(code) + "</a>.\n"
		fmt.Fprintln(res, body)
	}
}

func MethodNotAllowed(resp *CustomResponse, req *CustomRequest, url string, code int) error {
	Error(resp, "Method Not Allowed", 405)
	return nil
}

func Error(w *CustomResponse, errorMessage string, code int) {
	w.Headers.Set("Content-Type", "text/plain; charset=utf-8")
	w.Headers.Set("X-Content-Type-Options", "nosniff")
	w.SetStatusCode(code)
	w.Write([]byte(errorMessage))
}

func NotFound(resp *CustomResponse, req *CustomRequest) {
	Error(resp, "404 page not found", http.StatusNotFound)
}

func HTTPError(resp *CustomResponse, errorMessage string, code int) {
	Error(resp, errorMessage, code)
}
