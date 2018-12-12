package apigateway

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
)

type PanicHandlerFunc func(context.Context, *events.APIGatewayProxyRequest, interface{})
type ErrorHandlerFunc func(context.Context, *events.APIGatewayProxyRequest, *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse

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
	PathNotFound           eventHandler
	MethodNotAllowed       eventHandler
	PanicHandler           PanicHandlerFunc
	ErrorHandler           ErrorHandlerFunc
	preHandlers            []preHandler
	postHandlers           []postHandler
}

func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

func (r *Router) GET(path string, handlers *event) {
	r.Handle("GET", path, handlers)
}

func (r *Router) HEAD(path string, handlers *event) {
	r.Handle("HEAD", path, handlers)
}

func (r *Router) OPTIONS(path string, handlers *event) {
	r.Handle("OPTIONS", path, handlers)
}

func (r *Router) POST(path string, handlers *event) {
	r.Handle("POST", path, handlers)
}

func (r *Router) PUT(path string, handlers *event) {
	r.Handle("PUT", path, handlers)
}

func (r *Router) PATCH(path string, handlers *event) {
	r.Handle("PATCH", path, handlers)
}

func (r *Router) DELETE(path string, handlers *event) {
	r.Handle("DELETE", path, handlers)
}

func (r *Router) Handle(method, path string, handlers *event) {
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

	root.addRoute(path, handlers)
}

func (r *Router) MainHandler(ctx context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	response := r.ServeEvent(ctx, &request)
	if response.StatusCode >= 400 && r.ErrorHandler != nil {
		response = r.ErrorHandler(ctx, &request, response)
	}

	return *response
}

func (r *Router) recv(ctx context.Context, request *events.APIGatewayProxyRequest) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(ctx, request, rcv)
	}
}

func (r *Router) Lookup(method, path string) (*event, Params, bool) {
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

func (r *Router) UsePreHandler(handlers ...preHandler) {
	if len(handlers) == 0 {
		return
	}

	r.preHandlers = handlers
}

func (r *Router) UsePostHandler(handlers ...postHandler) {
	if len(handlers) == 0 {
		return
	}

	r.postHandlers = handlers
}

func (r *Router) runPreHandler(ctx context.Context, request *events.APIGatewayProxyRequest, handlers []preHandler) {
	for _, handler := range handlers {
		handler(ctx, request)
	}
}

func (r *Router) runPostHandler(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, handlers []postHandler) {
	for _, handler := range handlers {
		handler(ctx, request, response)
	}
}

func (r *Router) Run(ctx context.Context, request *events.APIGatewayProxyRequest, eventFlowHandle *event) *events.APIGatewayProxyResponse {
	if eventFlowHandle != nil {
		r.runPreHandler(ctx, request, r.preHandlers)
		r.runPreHandler(ctx, request, eventFlowHandle.preHandlers)

		response := eventFlowHandle.eventHandler(ctx, request)

		r.runPostHandler(ctx, request, response, eventFlowHandle.postHandlers)
		r.runPostHandler(ctx, request, response, r.postHandlers)

		return response
	}

	response := NewResponse()
	response.StatusCode = http.StatusNotFound
	response.Body = errors.InternalErrorf("HANDLE_NOT_FOUND", "Not found handle on path %s", request.Path).Error()

	return response
}

func (r *Router) ServeEvent(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	if r.PanicHandler != nil {
		defer r.recv(ctx, request)
	}

	path := request.Path
	if root := r.trees[request.HTTPMethod]; root != nil {
		if eventFlowHandle, _, tsr := root.getValue(path); eventFlowHandle != nil {
			return r.Run(ctx, request, eventFlowHandle)
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
				return Redirect(ctx, request, request.Path, code)
			}

			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					cleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					request.Path = string(fixedPath)
					return Redirect(ctx, request, request.Path, code)

				}
			}
		}
	}

	response := NewResponse()
	if request.HTTPMethod == "OPTIONS" && r.HandleOPTIONS {
		if allow := r.allowed(path, request.HTTPMethod); len(allow) > 0 {
			response.Headers["Allow"] = allow
			response.StatusCode = http.StatusOK
			return response
		}
	} else {
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, request.HTTPMethod); len(allow) > 0 {
				if r.MethodNotAllowed != nil {
					response := r.MethodNotAllowed(ctx, request)
					response.Headers["Allow"] = allow
					return response
				}

				response := HTTPError(ctx, "Method Not Allowed", http.StatusMethodNotAllowed)
				response.Headers["Allow"] = allow

				return response
			}
		}
	}

	if r.PathNotFound != nil {
		return r.PathNotFound(ctx, request)
	}

	return NotFound(ctx)
}
