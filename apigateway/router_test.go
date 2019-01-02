package apigateway

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
	"github.com/stretchr/testify/assert"
)

var handlerFunc = func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return nil, nil
}

func newRequest(method, path string) *events.APIGatewayProxyRequest {
	return &events.APIGatewayProxyRequest{HTTPMethod: method, Path: path}
}

func TestParams(t *testing.T) {
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		if val := ps.ByName(ps[i].Key); val != ps[i].Value {
			t.Errorf("Wrong value for %s: Got %s; Want %s", ps[i].Key, val, ps[i].Value)
		}
	}

	assert.Empty(t, ps.ByName("noKey"))
}

func TestMainHandlerWithAPIGatewayEvent(t *testing.T) {
	helloFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		response := NewResponse()
		response.StatusCode = http.StatusOK
		response.Body = request.QueryStringParameters["test"]
		return response, nil
	}

	router := New()
	router.GET("/hello", helloFunction)

	req := events.APIGatewayProxyRequest{
		Path:       "/hello",
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"test": "bar",
		},
	}

	res, _ := router.MainHandler(context.Background(), req)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "bar", res.Body)
}

func TestErrorHandler(t *testing.T) {
	helloFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		response := NewResponse()
		response.StatusCode = http.StatusNotFound
		response.Body = request.QueryStringParameters["test"]
		return response, errors.InternalError("test_error", "trigger error handle with error")
	}

	router := New()
	router.GET("/hello", helloFunction)

	req := events.APIGatewayProxyRequest{
		Path:       "/hello",
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"test": "bar",
		},
	}

	res, _ := router.MainHandler(context.Background(), req)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, "bar", res.Body)

	routed := false
	router.OnError = func(ctx context.Context, request *events.APIGatewayProxyRequest, response events.APIGatewayProxyResponse, err error) {
		assert.Error(t, err)
		routed = true
	}
	res, _ = router.MainHandler(context.Background(), req)
	assert.True(t, routed)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, "bar", res.Body)

}

func TestRouter(t *testing.T) {
	routed := false
	userFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		routed = true
		assert.Equal(t, "gopher", request.PathParameters["name"])

		return nil, nil
	}

	router := New()
	router.Handle("GET", "/user/:name", userFunction)

	req := &events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		PathParameters: map[string]string{"name": "gopher"},
		Path:           "/user/gopher",
	}

	router.ServeEvent(context.Background(), req)
	assert.True(t, routed)
}

type handlerStruct struct {
	handled *bool
}

func (h handlerStruct) ServeEvent(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	*h.handled = true
	return nil, nil
}

func TestRouterAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete bool

	getFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		get = true
		return nil, nil
	}

	headFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		head = true
		return nil, nil
	}

	optionFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		options = true
		return nil, nil
	}

	postFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		post = true
		return nil, nil
	}

	putFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		put = true
		return nil, nil
	}

	patchFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		patch = true
		return nil, nil
	}

	deleteFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		delete = true
		return nil, nil
	}

	router := New()
	router.GET("/GET", getFunction)
	router.HEAD("/GET", headFunction)
	router.OPTIONS("/GET", optionFunction)
	router.POST("/POST", postFunction)
	router.PUT("/PUT", putFunction)
	router.PATCH("/PATCH", patchFunction)
	router.DELETE("/DELETE", deleteFunction)

	router.ServeEvent(context.Background(), newRequest("GET", "/GET"))
	assert.True(t, get)

	router.ServeEvent(context.Background(), newRequest("HEAD", "/GET"))
	assert.True(t, head)

	router.ServeEvent(context.Background(), newRequest("OPTIONS", "/GET"))
	assert.True(t, options)

	router.ServeEvent(context.Background(), newRequest("POST", "/POST"))
	assert.True(t, post)

	router.ServeEvent(context.Background(), newRequest("PUT", "/PUT"))
	assert.True(t, put)

	router.ServeEvent(context.Background(), newRequest("PATCH", "/PATCH"))
	assert.True(t, patch)

	router.ServeEvent(context.Background(), newRequest("DELETE", "/DELETE"))
	assert.True(t, delete)
}

func TestRouterRoot(t *testing.T) {
	router := New()
	recv := catchPanic(func() {
		router.GET("noSlashRoot", nil)
	})

	assert.NotNil(t, recv, "registering path not beginning with '/' did not panic")
}

func TestMiddlewareRouter(t *testing.T) {
	mainPreHandler := false
	mainPreHandler2 := false
	mainHanlder := false
	mainPostHandler := false
	mainPostHandler2 := false

	routerPreHandler := false
	routerPreHandler2 := false
	routerPostHandler := false
	routerPostHandler2 := false

	preHandlers := []PreHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { mainPreHandler = true },
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { mainPreHandler2 = true },
	}

	postHandlers := []PostHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, err error) *events.APIGatewayProxyResponse {
			mainPostHandler = true
			return response
		},
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, err error) *events.APIGatewayProxyResponse {
			mainPostHandler2 = true
			return response
		},
	}

	eventHandler := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		response := NewResponse()
		response.StatusCode = http.StatusOK
		mainHanlder = true
		return response, nil
	}

	mainPreHandlers := []PreHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { routerPreHandler = true },
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { routerPreHandler2 = true },
	}

	mainPostHandlers := []PostHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, err error) *events.APIGatewayProxyResponse {
			routerPostHandler = true
			return response
		},
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, err error) *events.APIGatewayProxyResponse {
			routerPostHandler2 = true
			return response
		},
	}

	mainRouter := New()
	mainRouter.UsePreHandler(mainPreHandlers...)
	mainRouter.UsePostHandler(mainPostHandlers...)

	mainRouter.POST("/foo", eventHandler,
		WithPreHandlers(preHandlers...),
		WithPostHandlers(postHandlers...),
	)

	req := newRequest("POST", "/foo")
	res, err := mainRouter.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, mainPreHandler)
	assert.True(t, mainPreHandler2)
	assert.True(t, mainHanlder)
	assert.True(t, mainPostHandler)
	assert.True(t, mainPostHandler2)

	assert.True(t, routerPreHandler)
	assert.True(t, routerPreHandler2)
	assert.True(t, routerPostHandler)
	assert.True(t, routerPostHandler2)

	if !(res.StatusCode == http.StatusOK) {
		t.Errorf("Regular routing failed with router chaining.")
		t.FailNow()
	}
}

func TestRouterOPTIONS(t *testing.T) {
	testHandler := handlerFunc

	router := New()
	router.POST("/path", testHandler)

	// test not allowed
	// * (server)
	req := newRequest("OPTIONS", "*")
	res, err := router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	assert.Equal(t, "POST, OPTIONS", res.Headers["Allow"], "unexpected Allow header value: %v", res.Headers["Allow"])

	// path
	req = newRequest("OPTIONS", "/path")
	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	assert.Equal(t, "POST, OPTIONS", res.Headers["Allow"], "unexpected Allow header value: %v", res.Headers["Allow"])

	req = newRequest("OPTIONS", "/doesnotexist")
	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)

	// add another method
	router.GET("/path", testHandler)

	// test again
	// * (server)
	req = newRequest("OPTIONS", "*")
	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	req = newRequest("OPTIONS", "/path")
	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// custom handler
	var custom bool
	customHandler := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		custom = true

		response := NewResponse()
		response.StatusCode = http.StatusOK
		return response, nil
	}

	router.OPTIONS("/path", customHandler)

	// test again
	// * (server)
	req = newRequest("OPTIONS", "*")
	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, custom)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	req = newRequest("OPTIONS", "/path")
	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, custom)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
}

func TestRouterNotAllowed(t *testing.T) {
	testHandler := handlerFunc

	router := New()
	router.POST("/path", testHandler)

	// test not allowed
	req := newRequest("GET", "/path")
	res, err := router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode, "NotAllowed handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// add another method
	router.DELETE("/path", testHandler)
	router.OPTIONS("/path", testHandler) // must be ignored

	// test again
	req = newRequest("GET", "/path")
	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode, "NotAllowed handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// test custom handler
	responseText := "custom method"
	router.MethodNotAllowed = func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		response := NewResponse()
		response.StatusCode = http.StatusTeapot
		response.Body = responseText
		return response, nil
	}

	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "custom method", res.Body)
	assert.Equal(t, http.StatusTeapot, res.StatusCode)

	if allow := res.Headers["Allow"]; allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}
}

func TestRouterNotFound(t *testing.T) {
	router := New()
	router.GET("/path", handlerFunc)
	router.GET("/dir/", handlerFunc)
	router.GET("/", handlerFunc)

	testRoutes := []struct {
		route    string
		code     int
		location string
	}{
		{"/path/", http.StatusFound, "/path"},   // TSR -/
		{"/dir", http.StatusFound, "/dir/"},     // TSR +/
		{"", http.StatusFound, "/"},             // TSR +/
		{"/PATH", http.StatusFound, "/path"},    // Fixed Case
		{"/DIR/", http.StatusFound, "/dir/"},    // Fixed Case
		{"/PATH/", http.StatusFound, "/path"},   // Fixed Case -/
		{"/DIR", http.StatusFound, "/dir/"},     // Fixed Case +/
		{"/../path", http.StatusFound, "/path"}, // CleanPath
		{"/nope", http.StatusNotFound, ""},      // NotFound
	}
	for _, tr := range testRoutes {
		req := newRequest("GET", tr.route)
		res, err := router.ServeEvent(context.Background(), req)
		assert.NoError(t, err)
		if !(res.StatusCode == tr.code && (res.StatusCode == http.StatusNotFound || res.Headers["Location"] == tr.location)) {
			t.Errorf("NotFound handling route %s failed: Code=%d, Header=%v", tr.route, res.StatusCode, res.Headers["Location"])
		}
	}

	// Test custom not found handler
	var notFound bool
	router.PathNotFound = func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		notFound = true

		response := NewResponse()
		response.StatusCode = http.StatusNotFound
		return response, nil
	}

	req := newRequest("GET", "/nope")
	res, err := router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	if !(res.StatusCode == http.StatusNotFound && notFound == true) {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	}

	// Test other method than GET (want 307 instead of 301)
	router.PATCH("/path", handlerFunc)

	req = newRequest("PATCH", "/path/")
	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	if !(res.StatusCode == 307 && fmt.Sprint(res.Headers) == "map[Location:/path]") {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	}

	// Test special case where no node for the prefix "/" exists
	router = New()
	router.GET("/a", handlerFunc)
	req = newRequest("GET", "/")
	res, err = router.ServeEvent(context.Background(), req)
	assert.NoError(t, err)
	if !(res.StatusCode == http.StatusNotFound) {
		t.Errorf("NotFound handling route / failed: Code=%d", res.StatusCode)
	}
}

func TestRouterPanicHandler(t *testing.T) {
	panicHandled := false

	panicFunc := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		panic("oops!")
		return nil, nil
	}

	router := New()
	router.OnPanic = func(ctx context.Context, request *events.APIGatewayProxyRequest, p interface{}) {
		panicHandled = true
	}

	router.Handle("PUT", "/user/:name", panicFunc)

	req := newRequest("PUT", "/user/gopher")

	defer func() {
		if rcv := recover(); rcv != nil {
			t.Fatal("handling panic failed")
		}
	}()

	router.ServeEvent(context.Background(), req)

	if !panicHandled {
		t.Fatal("simulating failed")
	}
}

func TestRouterLookup(t *testing.T) {
	routed := false
	wantedFunc := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		routed = true
		return nil, nil
	}

	wantParams := Params{Param{"name", "gopher"}}

	router := New()

	// try empty router first
	handle, _, tsr := router.Lookup("GET", "/nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}

	// insert route and try again
	router.GET("/user/:name", wantedFunc)

	event, params, tsr := router.Lookup("GET", "/user/gopher")
	if event == nil {
		t.Fatal("Got no handle!")
	} else {
		event.eventHandler(context.Background(), nil)
		if !routed {
			t.Fatal("Routing failed!")
		}
	}

	if !reflect.DeepEqual(params, wantParams) {
		t.Fatalf("Wrong parameter values: want %v, got %v", wantParams, params)
	}

	event, _, tsr = router.Lookup("GET", "/user/gopher/")
	if event != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", event)
	}
	if !tsr {
		t.Error("Got no TSR recommendation!")
	}

	event, _, tsr = router.Lookup("GET", "/nope")
	if event != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", event)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}
}
