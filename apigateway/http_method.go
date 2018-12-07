package apigateway

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func Redirect(res *CustomResponse, req *CustomRequest, urlEndpoint string, code int) {
	if u, err := url.Parse(urlEndpoint); err == nil {
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

func MethodNotAllowed(res *CustomResponse, req *CustomRequest, url string, code int) error {
	Error(res, "Method Not Allowed", 405)
	return nil
}

func Error(res *CustomResponse, errorMessage string, code int) {
	res.Headers.Set("Content-Type", "text/plain; charset=utf-8")
	res.Headers.Set("X-Content-Type-Options", "nosniff")
	res.SetStatusCode(code)
	res.Write([]byte(errorMessage))
}

func NotFound(res *CustomResponse, req *CustomRequest) {
	Error(res, "404 page not found", http.StatusNotFound)
}

func HTTPError(res *CustomResponse, errorMessage string, code int) {
	Error(res, errorMessage, code)
}
