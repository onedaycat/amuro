package apigateway

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func Redirect(ctx context.Context, req *events.APIGatewayProxyRequest, urlEndpoint string, code int) *HttpResponseAndError {
	res := NewResponse()
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
			if i := strings.Index(urlEndpoint, "?"); i != -1 {
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

	_, hadCT := res.Headers["Content-Type"]

	res.Headers["Location"] = hexEscapeNonASCII(urlEndpoint)
	if !hadCT && (req.HTTPMethod == "GET" || req.HTTPMethod == "HEAD") {
		res.Headers["Content-Type"] = "text/html; charset=utf-8"
	}
	res.StatusCode = code

	// Shouldn't send the body for POST or HEAD; that leaves GET.
	if !hadCT && req.HTTPMethod == "GET" {
		res.Body = "<a href=\"" + htmlEscape(urlEndpoint) + "\">" + string(code) + "</a>.\n"
	}

	return &HttpResponseAndError{HttpResponse: res, Error: nil}
}

func MethodNotAllowed(ctx context.Context, url string, code int) *events.APIGatewayProxyResponse {
	return NewError(ctx, "Method Not Allowed", 405)
}

func NewError(ctx context.Context, errorMessage string, code int) *events.APIGatewayProxyResponse {
	res := NewResponse()
	res.StatusCode = code
	res.Headers["Content-Type"] = "text/plain; charset=utf-8"
	res.Headers["X-Content-Type-Options"] = "nosniff"
	res.Body = errorMessage

	return res
}

func NotFound(ctx context.Context) *events.APIGatewayProxyResponse {
	return NewError(ctx, "404 page not found", http.StatusNotFound)
}

func HTTPError(ctx context.Context, errorMessage string, code int) *events.APIGatewayProxyResponse {
	return NewError(ctx, errorMessage, code)
}
