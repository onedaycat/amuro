package apigateway

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Make sure the Router conforms with the http.Handler interface
var _ Handler = New()

type Handler interface {
	ServeEvent(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)
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
	PanicHandler           func(context.Context, *events.APIGatewayProxyRequest, interface{})
	ErrorHandler           func(context.Context, *events.APIGatewayProxyRequest, *events.APIGatewayProxyResponse, error) (events.APIGatewayProxyResponse, error)
}

func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

func (r *Router) GET(path string, handler EventHandler) {
	r.Handle("GET", path, handler)
}

func (r *Router) HEAD(path string, handler EventHandler) {
	r.Handle("HEAD", path, handler)
}

func (r *Router) OPTIONS(path string, handler EventHandler) {
	r.Handle("OPTIONS", path, handler)
}

func (r *Router) POST(path string, handler EventHandler) {
	r.Handle("POST", path, handler)
}

func (r *Router) PUT(path string, handler EventHandler) {
	r.Handle("PUT", path, handler)
}

func (r *Router) PATCH(path string, handler EventHandler) {
	r.Handle("PATCH", path, handler)
}

func (r *Router) DELETE(path string, handler EventHandler) {
	r.Handle("DELETE", path, handler)
}

func (r *Router) Handle(method, path string, handler EventHandler) {
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

	root.addRoute(path, handler)
}

func (r *Router) MainHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response, err := r.ServeEvent(ctx, &request)
	if (response.StatusCode >= 400 || err != nil) && r.ErrorHandler != nil {
		return r.ErrorHandler(ctx, &request, response, err)
	}

	return *response, nil
}

func (r *Router) recv(ctx context.Context, request *events.APIGatewayProxyRequest) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(ctx, request, rcv)
	}
}

func (r *Router) Lookup(method, path string) (EventHandler, Params, bool) {
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

func (r *Router) ServeEvent(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if r.PanicHandler != nil {
		defer r.recv(ctx, request)
	}

	path := request.Path
	if root := r.trees[request.HTTPMethod]; root != nil {
		if handle, _, tsr := root.getValue(path); handle != nil {
			return handle(ctx, request)
		} else if request.HTTPMethod != "CONNECT" && path != "/" {
			code := http.StatusMovedPermanently
			if request.HTTPMethod != "GET" {
				code = 307
			}

			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					request.Path = path[:len(path)-1]
				} else {
					request.Path = path + "/"
				}
				return Redirect(ctx, request, request.Path, code), nil
			}

			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					cleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					request.Path = string(fixedPath)
					return Redirect(ctx, request, request.Path, code), nil

				}
			}
		}
	}

	response := NewResponse()
	if request.HTTPMethod == "OPTIONS" && r.HandleOPTIONS {
		if allow := r.allowed(path, request.HTTPMethod); len(allow) > 0 {
			response.Headers["Allow"] = allow
			response.StatusCode = http.StatusOK
			return response, nil
		}
	} else {
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, request.HTTPMethod); len(allow) > 0 {
				if r.MethodNotAllowed != nil {
					response, err := r.MethodNotAllowed.ServeEvent(ctx, request)
					response.Headers["Allow"] = allow
					return response, err
				}

				response := HTTPError(ctx, "Method Not Allowed", http.StatusMethodNotAllowed)
				response.Headers["Allow"] = allow

				return response, nil
			}
		}
	}

	if r.NotFound != nil {
		return r.NotFound.ServeEvent(ctx, request)
	}

	return NotFound(ctx), nil
}
