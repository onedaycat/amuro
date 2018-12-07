package apigateway

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var handlerFunc = func(_ *CustomResponse, _ *CustomRequest) {}

func newRequest(method, path string) *CustomRequest {
	return &CustomRequest{HTTPMethod: method, Path: path}
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
	if val := ps.ByName("noKey"); val != "" {
		t.Errorf("Expected empty string for not found key; got: %s", val)
	}
}

func TestRouterWithAPIGatewayEvent(t *testing.T) {
	router := New()
	router.GET("/hello", func(res *CustomResponse, req *CustomRequest) {
		res.SetStatusCode(200)
		res.Write([]byte(req.QueryStringParameters["test"]))
	})

	req := events.APIGatewayProxyRequest{
		Path:       "/hello",
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"test": "bar",
		},
	}
	res, err := router.MainHandler(req)
	if err != nil {
		t.Errorf("Expected empty error; got: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP Status 200; got: %d", res.StatusCode)

	}

	if res.Body != "bar" {
		t.Errorf("Expected body bar string; got: %s", res.Body)
	}
}

func TestRouter(t *testing.T) {
	router := New()

	routed := false
	router.Handle("GET", "/user/:name", func(res *CustomResponse, req *CustomRequest) {
		routed = true
		want := Params{Param{"name", "gopher"}}
		if "gopher" != want.ByName("name") {
			t.Fatalf("wrong wildcard values: want ghoper, got %v", want.ByName("name"))
		}
	})

	res := NewCustomResponse()
	req := newRequest("GET", "/user/gopher")
	router.ServeHTTP(res, req)

	if !routed {
		t.Fatal("routing failed")
	}
}

type handlerStruct struct {
	handled *bool
}

func (h handlerStruct) ServeHTTP(w *CustomResponse, r *CustomRequest) {
	*h.handled = true
}

func TestRouterAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete bool

	router := New()
	router.GET("/GET", func(w *CustomResponse, r *CustomRequest) {
		get = true
	})
	router.HEAD("/GET", func(w *CustomResponse, r *CustomRequest) {
		head = true
	})
	router.OPTIONS("/GET", func(w *CustomResponse, r *CustomRequest) {
		options = true
	})
	router.POST("/POST", func(w *CustomResponse, r *CustomRequest) {
		post = true
	})
	router.PUT("/PUT", func(w *CustomResponse, r *CustomRequest) {
		put = true
	})
	router.PATCH("/PATCH", func(w *CustomResponse, r *CustomRequest) {
		patch = true
	})
	router.DELETE("/DELETE", func(w *CustomResponse, r *CustomRequest) {
		delete = true
	})

	res := NewCustomResponse()

	router.ServeHTTP(res, newRequest("GET", "/GET"))
	if !get {
		t.Error("routing GET failed")
	}

	router.ServeHTTP(res, newRequest("HEAD", "/GET"))
	if !head {
		t.Error("routing HEAD failed")
	}

	router.ServeHTTP(res, newRequest("OPTIONS", "/GET"))
	if !options {
		t.Error("routing OPTIONS failed")
	}

	router.ServeHTTP(res, newRequest("POST", "/POST"))
	if !post {
		t.Error("routing POST failed")
	}

	router.ServeHTTP(res, newRequest("PUT", "/PUT"))
	if !put {
		t.Error("routing PUT failed")
	}

	router.ServeHTTP(res, newRequest("PATCH", "/PATCH"))
	if !patch {
		t.Error("routing PATCH failed")
	}

	router.ServeHTTP(res, newRequest("DELETE", "/DELETE"))
	if !delete {
		t.Error("routing DELETE failed")
	}
}

func TestRouterRoot(t *testing.T) {
	router := New()
	recv := catchPanic(func() {
		router.GET("noSlashRoot", nil)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}
}

func TestRouterOPTIONS(t *testing.T) {

	router := New()
	router.POST("/path", handlerFunc)

	// test not allowed
	// * (server)
	r := newRequest("OPTIONS", "*")
	w := NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	} else if allow := w.Headers.Get("Allow"); allow != "POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	r = newRequest("OPTIONS", "/path")
	w = NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	} else if allow := w.Headers.Get("Allow"); allow != "POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	r = newRequest("OPTIONS", "/doesnotexist")
	w = NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusNotFound) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	}

	// add another method
	router.GET("/path", handlerFunc)

	// test again
	// * (server)
	r = newRequest("OPTIONS", "*")
	w = NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	} else if allow := w.Headers.Get("Allow"); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	r = newRequest("OPTIONS", "/path")
	w = NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	} else if allow := w.Headers.Get("Allow"); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// custom handler
	var custom bool
	router.OPTIONS("/path", func(w *CustomResponse, r *CustomRequest) {
		w.SetStatusCode(http.StatusOK)
		custom = true
	})

	// test again
	// * (server)
	r = newRequest("OPTIONS", "*")
	w = NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	} else if allow := w.Headers.Get("Allow"); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}
	if custom {
		t.Error("custom handler called on *")
	}

	// path
	r = newRequest("OPTIONS", "/path")
	w = NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	}
	if !custom {
		t.Error("custom handler not called")
	}
}

func TestRouterNotAllowed(t *testing.T) {
	router := New()
	router.POST("/path", handlerFunc)

	// test not allowed
	r := newRequest("GET", "/path")
	w := NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	} else if allow := w.Headers.Get("Allow"); allow != "POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// add another method
	router.DELETE("/path", handlerFunc)
	router.OPTIONS("/path", handlerFunc) // must be ignored

	// test again
	r = newRequest("GET", "/path")
	w = NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	} else if allow := w.Headers.Get("Allow"); allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// test custom handler
	w = NewCustomResponse()
	responseText := "custom method"
	router.MethodNotAllowed = CustomHandler(func(w *CustomResponse, r *CustomRequest) {
		w.SetStatusCode(http.StatusTeapot)
		w.Write([]byte(responseText))
	})

	router.ServeHTTP(w, r)
	if got := w.Body; !(got.String() == responseText) {
		t.Errorf("unexpected response got %q want %q", got, responseText)
	}
	if w.StatusCode != http.StatusTeapot {
		t.Errorf("unexpected response code %d want %d", w.StatusCode, http.StatusTeapot)
	}
	if allow := w.Headers.Get("Allow"); allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
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
		r := newRequest("GET", tr.route)
		w := NewCustomResponse()
		router.ServeHTTP(w, r)
		if !(w.StatusCode == tr.code && (w.StatusCode == http.StatusNotFound || fmt.Sprint(w.Headers.Get("Location")) == tr.location)) {
			t.Errorf("NotFound handling route %s failed: Code=%d, Header=%v", tr.route, w.StatusCode, w.Headers["Location"])
		}
	}

	// Test custom not found handler
	var notFound bool
	router.NotFound = CustomHandler(func(res *CustomResponse, req *CustomRequest) {
		res.SetStatusCode(http.StatusNotFound)
		notFound = true
	})

	r := newRequest("GET", "/nope")
	w := NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusNotFound && notFound == true) {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	}

	// Test other method than GET (want 307 instead of 301)
	router.PATCH("/path", handlerFunc)
	r = newRequest("PATCH", "/path/")
	w = NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusPermanentRedirect && fmt.Sprint(w.Headers) == "map[Location:[/path]]") {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.StatusCode, w.Headers)
	}

	// Test special case where no node for the prefix "/" exists
	router = New()
	router.GET("/a", handlerFunc)
	r = newRequest("GET", "/")
	w = NewCustomResponse()
	router.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusNotFound) {
		t.Errorf("NotFound handling route / failed: Code=%d", w.StatusCode)
	}
}

func TestRouterPanicHandler(t *testing.T) {
	router := New()
	panicHandled := false

	router.PanicHandler = func(res *CustomResponse, req *CustomRequest, p interface{}) {
		panicHandled = true
	}

	router.Handle("PUT", "/user/:name", func(_ *CustomResponse, _ *CustomRequest) {
		panic("oops!")
	})

	r := newRequest("PUT", "/user/gopher")
	w := NewCustomResponse()

	defer func() {
		if rcv := recover(); rcv != nil {
			t.Fatal("handling panic failed")
		}
	}()

	router.ServeHTTP(w, r)

	if !panicHandled {
		t.Fatal("simulating failed")
	}
}

func TestRouterChaining(t *testing.T) {
	router1 := New()
	router2 := New()
	router1.NotFound = router2

	fooHit := false
	router1.POST("/foo", func(w *CustomResponse, req *CustomRequest) {
		fooHit = true
		w.SetStatusCode(http.StatusOK)
	})

	barHit := false
	router2.POST("/bar", func(w *CustomResponse, req *CustomRequest) {
		barHit = true
		w.SetStatusCode(http.StatusOK)
	})

	r := newRequest("POST", "/foo")
	w := NewCustomResponse()
	router1.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusOK && fooHit) {
		t.Errorf("Regular routing failed with router chaining.")
		t.FailNow()
	}

	r = newRequest("POST", "/bar")
	w = NewCustomResponse()
	router1.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusOK && barHit) {
		t.Errorf("Chained routing failed with router chaining.")
		t.FailNow()
	}

	r = newRequest("POST", "/qax")
	w = NewCustomResponse()
	router1.ServeHTTP(w, r)
	if !(w.StatusCode == http.StatusNotFound) {
		t.Errorf("NotFound behavior failed with router chaining.")
		t.FailNow()
	}
}

func TestRouterLookup(t *testing.T) {
	routed := false
	wantHandle := func(_ *CustomResponse, _ *CustomRequest) {
		routed = true
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
	router.GET("/user/:name", wantHandle)

	handle, params, tsr := router.Lookup("GET", "/user/gopher")
	if handle == nil {
		t.Fatal("Got no handle!")
	} else {
		handle(nil, nil)
		if !routed {
			t.Fatal("Routing failed!")
		}
	}

	if !reflect.DeepEqual(params, wantParams) {
		t.Fatalf("Wrong parameter values: want %v, got %v", wantParams, params)
	}

	handle, _, tsr = router.Lookup("GET", "/user/gopher/")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if !tsr {
		t.Error("Got no TSR recommendation!")
	}

	handle, _, tsr = router.Lookup("GET", "/nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}
}

type mockFileSystem struct {
	opened bool
}

func (mfs *mockFileSystem) Open(name string) (http.File, error) {
	mfs.opened = true
	return nil, errors.New("this is just a mock")
}
