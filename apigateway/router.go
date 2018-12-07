package apigateway

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Make sure the Router conforms with the http.Handler interface
var _ Handler = New()

type Handler interface {
	ServeHTTP(res *CustomResponse, req *CustomRequest)
}

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

type Router struct {
	trees map[string]*node

	RedirectTrailingSlash  bool
	RedirectFixedPath      bool
	HandleMethodNotAllowed bool
	HandleOPTIONS          bool
	NotFound               Handler
	MethodNotAllowed       Handler
	PanicHandler           func(*CustomResponse, *CustomRequest, interface{})
}

func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

func (r *Router) GET(path string, handle CustomHandle) {
	r.Handle("GET", path, handle)
}

func (r *Router) HEAD(path string, handle CustomHandle) {
	r.Handle("HEAD", path, handle)
}

func (r *Router) OPTIONS(path string, handle CustomHandle) {
	r.Handle("OPTIONS", path, handle)
}

func (r *Router) POST(path string, handle CustomHandle) {
	r.Handle("POST", path, handle)
}

func (r *Router) PUT(path string, handle CustomHandle) {
	r.Handle("PUT", path, handle)
}

func (r *Router) PATCH(path string, handle CustomHandle) {
	r.Handle("PATCH", path, handle)
}

func (r *Router) DELETE(path string, handle CustomHandle) {
	r.Handle("DELETE", path, handle)
}

func (r *Router) Handle(method, path string, handle CustomHandle) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	root.addRoute(path, handle)
}

func (r *Router) HandlerFunc(method, path string, handler CustomHandle) {
	r.Handle(method, path, handler)
}

func (r *Router) recv(res *CustomResponse, req *CustomRequest) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(res, req, rcv)
	}
}

func (r *Router) Lookup(method, path string) (CustomHandle, Params, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

func (r *Router) allowed(path, reqMethod string) (allow string) {
	if path == "*" {
		for method := range r.trees {
			if method == "OPTIONS" {
				continue
			}

			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else {
		for method := range r.trees {
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handle, _, _ := r.trees[method].getValue(path)
			if handle != nil {
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

// Support Lambda Function
func (r *Router) MainHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req := NewCustomResquestFromEvent(request)
	res := NewCustomResponse()

	r.ServeHTTP(res, req)

	return res.ToAPIGatewayResponse()
}

func (r *Router) ServeHTTP(res *CustomResponse, req *CustomRequest) {
	if r.PanicHandler != nil {
		defer r.recv(res, req)
	}

	path := req.Path
	if root := r.trees[req.HTTPMethod]; root != nil {
		if handle, ps, tsr := root.getValue(path); handle != nil {
			handle(res, req, ps)
			return
		} else if req.HTTPMethod != "CONNECT" && path != "/" {
			code := http.StatusMovedPermanently
			if req.HTTPMethod != "GET" {
				code = http.StatusPermanentRedirect
			}

			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.Path = path[:len(path)-1]
				} else {
					req.Path = path + "/"
				}
				Redirect(res, req, req.URLString(), code)
				return
			}

			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					req.Path = string(fixedPath)
					Redirect(res, req, req.URLString(), code)
					return
				}
			}
		}
	}

	if req.HTTPMethod == "OPTIONS" && r.HandleOPTIONS {
		if allow := r.allowed(path, req.HTTPMethod); len(allow) > 0 {
			res.Headers.Set("Allow", allow)
			res.SetStatusCode(http.StatusOK)
			return
		}
	} else {
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, req.HTTPMethod); len(allow) > 0 {
				res.Headers.Set("Allow", allow)
				if r.MethodNotAllowed != nil {
					r.MethodNotAllowed.ServeHTTP(res, req)
				} else {
					HTTPError(res,
						"Method Not Allowed",
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	if r.NotFound != nil {
		r.NotFound.ServeHTTP(res, req)
	} else {
		NotFound(res, req)
	}
}
