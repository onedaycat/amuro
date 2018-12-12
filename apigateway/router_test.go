package apigateway

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

var handlerFunc = WithEventHandler(func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	return nil
})

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
	helloFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		response := NewResponse()
		response.StatusCode = http.StatusOK
		response.Body = request.QueryStringParameters["test"]
		return response
	}
	helloHandler := WithEventHandler(helloFunction)

	router := New()
	router.GET("/hello", helloHandler)

	req := events.APIGatewayProxyRequest{
		Path:       "/hello",
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"test": "bar",
		},
	}

	res := router.MainHandler(context.Background(), req)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "bar", res.Body)
}

func TestErrorHandler(t *testing.T) {
	helloFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		response := NewResponse()
		response.StatusCode = http.StatusNotFound
		response.Body = request.QueryStringParameters["test"]
		return response
	}
	helloHandler := WithEventHandler(helloFunction)

	router := New()
	router.GET("/hello", helloHandler)

	req := events.APIGatewayProxyRequest{
		Path:       "/hello",
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"test": "bar",
		},
	}

	res := router.MainHandler(context.Background(), req)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, "bar", res.Body)

	routed := false
	router.ErrorHandler = func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse {
		routed = true
		return response
	}
	res = router.MainHandler(context.Background(), req)
	assert.True(t, routed)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, "bar", res.Body)

}

func TestRouter(t *testing.T) {
	routed := false
	userFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		routed = true
		assert.Equal(t, "gopher", request.PathParameters["name"])

		return nil
	}
	userHandler := WithEventHandler(userFunction)

	router := New()
	router.Handle("GET", "/user/:name", userHandler)

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

func (h handlerStruct) ServeEvent(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	*h.handled = true
	return nil
}

func TestRouterAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete bool

	getFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		get = true
		return nil
	}

	headFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		head = true
		return nil
	}

	optionFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		options = true
		return nil
	}

	postFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		post = true
		return nil
	}

	putFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		put = true
		return nil
	}

	patchFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		patch = true
		return nil
	}

	deleteFunction := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		delete = true
		return nil
	}

	getHandler := WithEventHandler(getFunction)
	headHandler := WithEventHandler(headFunction)
	optionHandler := WithEventHandler(optionFunction)
	postHandler := WithEventHandler(postFunction)
	putHandler := WithEventHandler(putFunction)
	patchHandler := WithEventHandler(patchFunction)
	deleteHandler := WithEventHandler(deleteFunction)

	router := New()
	router.GET("/GET", getHandler)
	router.HEAD("/GET", headHandler)
	router.OPTIONS("/GET", optionHandler)
	router.POST("/POST", postHandler)
	router.PUT("/PUT", putHandler)
	router.PATCH("/PATCH", patchHandler)
	router.DELETE("/DELETE", deleteHandler)

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

	preHandlers := []preHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { mainPreHandler = true },
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { mainPreHandler2 = true },
	}

	postHandlers := []postHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse {
			mainPostHandler = true
			return response
		},
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse {
			mainPostHandler2 = true
			return response
		},
	}

	eventHandler := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		response := NewResponse()
		response.StatusCode = http.StatusOK
		mainHanlder = true
		return response
	}

	mainPreHandlers := []preHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { routerPreHandler = true },
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { routerPreHandler2 = true },
	}

	mainPostHandlers := []postHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse {
			routerPostHandler = true
			return response
		},
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse {
			routerPostHandler2 = true
			return response
		},
	}

	mainRouter := New()
	mainRouter.UsePreHandler(mainPreHandlers...)
	mainRouter.UsePostHandler(mainPostHandlers...)

	mainRouter.POST("/foo",
		WithPreHandlers(preHandlers...),
		WithPostHandlers(postHandlers...),
		WithEventHandler(eventHandler),
	)

	req := newRequest("POST", "/foo")
	res := mainRouter.ServeEvent(context.Background(), req)
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
	res := router.ServeEvent(context.Background(), req)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	assert.Equal(t, "POST, OPTIONS", res.Headers["Allow"], "unexpected Allow header value: %v", res.Headers["Allow"])

	// path
	req = newRequest("OPTIONS", "/path")
	res = router.ServeEvent(context.Background(), req)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	assert.Equal(t, "POST, OPTIONS", res.Headers["Allow"], "unexpected Allow header value: %v", res.Headers["Allow"])

	req = newRequest("OPTIONS", "/doesnotexist")
	res = router.ServeEvent(context.Background(), req)

	assert.Equal(t, http.StatusNotFound, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)

	// add another method
	router.GET("/path", testHandler)

	// test again
	// * (server)
	req = newRequest("OPTIONS", "*")
	res = router.ServeEvent(context.Background(), req)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	req = newRequest("OPTIONS", "/path")
	res = router.ServeEvent(context.Background(), req)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// custom handler
	var custom bool
	customHandler := WithEventHandler(func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		custom = true

		response := NewResponse()
		response.StatusCode = http.StatusOK
		return response
	})

	router.OPTIONS("/path", customHandler)

	// test again
	// * (server)
	req = newRequest("OPTIONS", "*")
	res = router.ServeEvent(context.Background(), req)
	assert.False(t, custom)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	req = newRequest("OPTIONS", "/path")
	res = router.ServeEvent(context.Background(), req)
	assert.True(t, custom)
	assert.Equal(t, http.StatusOK, res.StatusCode, "OPTIONS handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
}

func TestRouterNotAllowed(t *testing.T) {
	testHandler := handlerFunc

	router := New()
	router.POST("/path", testHandler)

	// test not allowed
	req := newRequest("GET", "/path")
	res := router.ServeEvent(context.Background(), req)
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode, "NotAllowed handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// add another method
	router.DELETE("/path", testHandler)
	router.OPTIONS("/path", testHandler) // must be ignored

	// test again
	req = newRequest("GET", "/path")
	res = router.ServeEvent(context.Background(), req)
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode, "NotAllowed handling failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	if allow := res.Headers["Allow"]; allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// test custom handler
	responseText := "custom method"
	router.MethodNotAllowed = func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		response := NewResponse()
		response.StatusCode = http.StatusTeapot
		response.Body = responseText
		return response
	}

	res = router.ServeEvent(context.Background(), req)
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
		{"/path/", http.StatusMovedPermanently, "/path"},   // TSR -/
		{"/dir", http.StatusMovedPermanently, "/dir/"},     // TSR +/
		{"", http.StatusMovedPermanently, "/"},             // TSR +/
		{"/PATH", http.StatusMovedPermanently, "/path"},    // Fixed Case
		{"/DIR/", http.StatusMovedPermanently, "/dir/"},    // Fixed Case
		{"/PATH/", http.StatusMovedPermanently, "/path"},   // Fixed Case -/
		{"/DIR", http.StatusMovedPermanently, "/dir/"},     // Fixed Case +/
		{"/../path", http.StatusMovedPermanently, "/path"}, // CleanPath
		{"/nope", http.StatusNotFound, ""},                 // NotFound
	}
	for _, tr := range testRoutes {
		req := newRequest("GET", tr.route)
		res := router.ServeEvent(context.Background(), req)
		if !(res.StatusCode == tr.code && (res.StatusCode == http.StatusNotFound || res.Headers["Location"] == tr.location)) {
			t.Errorf("NotFound handling route %s failed: Code=%d, Header=%v", tr.route, res.StatusCode, res.Headers["Location"])
		}
	}

	// Test custom not found handler
	var notFound bool
	router.PathNotFound = eventHandler(func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		notFound = true

		response := NewResponse()
		response.StatusCode = http.StatusNotFound
		return response
	})

	req := newRequest("GET", "/nope")
	res := router.ServeEvent(context.Background(), req)
	if !(res.StatusCode == http.StatusNotFound && notFound == true) {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	}

	// Test other method than GET (want 307 instead of 301)
	router.PATCH("/path", handlerFunc)

	req = newRequest("PATCH", "/path/")
	res = router.ServeEvent(context.Background(), req)
	if !(res.StatusCode == 307 && fmt.Sprint(res.Headers) == "map[Location:/path]") {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", res.StatusCode, res.Headers)
	}

	// Test special case where no node for the prefix "/" exists
	router = New()
	router.GET("/a", handlerFunc)
	req = newRequest("GET", "/")
	res = router.ServeEvent(context.Background(), req)
	if !(res.StatusCode == http.StatusNotFound) {
		t.Errorf("NotFound handling route / failed: Code=%d", res.StatusCode)
	}
}

func TestRouterPanicHandler(t *testing.T) {
	panicHandled := false

	panicFunc := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		panic("oops!")
		return nil
	}
	panicHandler := WithEventHandler(panicFunc)

	router := New()
	router.PanicHandler = func(ctx context.Context, request *events.APIGatewayProxyRequest, p interface{}) {
		panicHandled = true
	}

	router.Handle("PUT", "/user/:name", panicHandler)

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
	wantedFunc := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		routed = true
		return nil
	}

	wantHandle := WithEventHandler(wantedFunc)
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
	router.GET("/user/:name", wantHandle)

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
