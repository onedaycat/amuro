package apigateway

import (
	"github.com/aws/aws-lambda-go/events"
)

type Handler interface {
	ServeHTTP(w *CustomResponse, req *CustomRequest)
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
	NotFound               CustomHandle
	MethodNotAllowed       CustomHandle
	PanicHandler           func(*CustomResponse, *CustomRequest, interface{})
}

// Make sure the Router conforms with the http.Handler interface
var _ Handler = New()

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

func (r *Router) recv(w *CustomResponse, req *CustomRequest) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req, rcv)
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

func (r *Router) ServeHTTP(w *CustomResponse, req *CustomRequest) {
	if r.PanicHandler != nil {
		defer r.recv(w, req)
	}

	path := req.Path

	if root := r.trees[req.HTTPMethod]; root != nil {
		if handle, ps, tsr := root.getValue(path); handle != nil {
			handle(w, req, ps)
			return
		} else if req.HTTPMethod != "CONNECT" && path != "/" {
			code := 301
			if req.HTTPMethod != "GET" {
				code = 307
			}

			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.Path = path[:len(path)-1]
				} else {
					req.Path = path + "/"
				}
				Redirect(w, req, req.URLString(), code)
				return
			}

			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					req.Path = string(fixedPath)
					Redirect(w, req, req.URLString(), code)
					return
				}
			}
		}
	}

	if req.HTTPMethod == "OPTIONS" && r.HandleOPTIONS {
		if allow := r.allowed(path, req.HTTPMethod); len(allow) > 0 {
			w.Headers.Set("Allow", allow)
			w.SetStatusCode(200)
			return
		}
	} else {
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, req.HTTPMethod); len(allow) > 0 {
				w.Headers.Set("Allow", allow)
				if r.MethodNotAllowed != nil {
					r.MethodNotAllowed(w, req, nil)
				} else {
					HTTPError(w,
						"Method Not Allowed",
						405,
					)
				}
				return
			}
		}
	}

	if r.NotFound != nil {
		r.NotFound(w, req, nil)
	} else {
		NotFound(w, req)
	}
}
