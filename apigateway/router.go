package apigateway

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
)

type PanicHandlerFunc func(context.Context, *events.APIGatewayProxyRequest, interface{})
type ErrorHandlerFunc func(context.Context, *events.APIGatewayProxyRequest, events.APIGatewayProxyResponse, error)

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
	PathNotFound           EventHandler
	MethodNotAllowed       EventHandler
	OnPanic                PanicHandlerFunc
	OnError                ErrorHandlerFunc
	preHandlers            []PreHandler
	postHandlers           []PostHandler
}

func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

func (r *Router) GET(path string, handler EventHandler, options ...Option) {
	r.Handle("GET", path, handler, options...)
}

func (r *Router) HEAD(path string, handler EventHandler, options ...Option) {
	r.Handle("HEAD", path, handler, options...)
}

func (r *Router) OPTIONS(path string, handler EventHandler, options ...Option) {
	r.Handle("OPTIONS", path, handler, options...)
}

func (r *Router) POST(path string, handler EventHandler, options ...Option) {
	r.Handle("POST", path, handler, options...)
}

func (r *Router) PUT(path string, handler EventHandler, options ...Option) {
	r.Handle("PUT", path, handler, options...)
}

func (r *Router) PATCH(path string, handler EventHandler, options ...Option) {
	r.Handle("PATCH", path, handler, options...)
}

func (r *Router) DELETE(path string, handler EventHandler, options ...Option) {
	r.Handle("DELETE", path, handler, options...)
}

func (r *Router) Handle(method, path string, handler EventHandler, options ...Option) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if handler == nil {
		panic("handler should not nil")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	opts := newOption(options...)
	e := &event{
		eventHandler: handler,
	}

	if len(opts.preHandlers) > 0 {
		e.preHandlers = opts.preHandlers
	}

	if len(opts.postHandlers) > 0 {
		e.postHandlers = opts.postHandlers
	}

	root.addRoute(path, e)
}

func (r *Router) MainHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := r.ServeEvent(ctx, &request)
	if response == nil {
		httpInternalError := events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: "service not implment return data"}

		if r.OnError != nil {
			r.OnError(ctx, &request, httpInternalError, errors.InternalError("handler_error", "handler return nil response"))
		}

		return httpInternalError, nil
	}

	if response.Error != nil && r.OnError != nil {
		r.OnError(ctx, &request, *response.HttpResponse, response.Error)
	}

	return *response.HttpResponse, nil
}

func (r *Router) recv(ctx context.Context, request *events.APIGatewayProxyRequest) {
	if rcv := recover(); rcv != nil {
		r.OnPanic(ctx, request, rcv)
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

func (r *Router) UsePreHandler(handlers ...PreHandler) {
	if len(handlers) == 0 {
		return
	}

	r.preHandlers = handlers
}

func (r *Router) UsePostHandler(handlers ...PostHandler) {
	if len(handlers) == 0 {
		return
	}

	r.postHandlers = handlers
}

func (r *Router) runPreHandler(ctx context.Context, request *events.APIGatewayProxyRequest, handlers []PreHandler) {
	for _, handler := range handlers {
		handler(ctx, request)
	}
}

func (r *Router) runPostHandler(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, handlers []PostHandler) {
	for _, handler := range handlers {
		handler(ctx, request, response)
	}
}

func (r *Router) Run(ctx context.Context, request *events.APIGatewayProxyRequest, option *event) *HttpResponseAndError {
	if option != nil {
		r.runPreHandler(ctx, request, r.preHandlers)
		r.runPreHandler(ctx, request, option.preHandlers)

		response := option.eventHandler(ctx, request)

		r.runPostHandler(ctx, request, response.HttpResponse, option.postHandlers)
		r.runPostHandler(ctx, request, response.HttpResponse, r.postHandlers)

		return response
	}

	err := errors.InternalErrorf("HANDLE_NOT_FOUND", "Not found handle on path %s", request.Path)
	response := NewResponse()
	response.StatusCode = http.StatusNotFound
	response.Body = err.Error()

	return &HttpResponseAndError{HttpResponse: response, Error: err}
}

func (r *Router) ServeEvent(ctx context.Context, request *events.APIGatewayProxyRequest) *HttpResponseAndError {
	if r.OnPanic != nil {
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
			return &HttpResponseAndError{HttpResponse: response, Error: nil}
		}
	} else {
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, request.HTTPMethod); len(allow) > 0 {
				if r.MethodNotAllowed != nil {
					response := r.MethodNotAllowed(ctx, request)
					response.HttpResponse.Headers["Allow"] = allow
					return response
				}

				response := HTTPError(ctx, "Method Not Allowed", http.StatusMethodNotAllowed)
				response.Headers["Allow"] = allow

				return &HttpResponseAndError{HttpResponse: response, Error: nil}
			}
		}
	}

	if r.PathNotFound != nil {
		return r.PathNotFound(ctx, request)
	}

	return &HttpResponseAndError{HttpResponse: NotFound(ctx), Error: nil}
}
