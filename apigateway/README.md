# Amuro
amuro is router for lambda (apigateway event) inspire from [httprouter](https://github.com/julienschmidt/httprouter) 

amuro have features same [httprouter](https://github.com/julienschmidt/httprouter) except serverFile

## Usage
```
package main

import (
    "fmt"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)


func HelloFunc(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		response := NewResponse()
		response.StatusCode = http.StatusOK
		response.Body = request.QueryStringParameters["name"]
		return response
	}

func main() {
	router := New()
  router.GET("/hello", NewEventFlowHandler(HelloFunc))
  
	lambda.Start(router.MainHandler)
}
```

### Use Middleware in Handle Level

```
package main

import (
    "fmt"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

func main() {
	router := New()
  router.GET("/hello", &EventFlowHandler{
		preHandlers: []PreEventHandler{
			func(ctx context.Context, request *events.APIGatewayProxyRequest) { 
        // do something        
      },
		},
		handler: func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
			response := NewResponse()
			response.StatusCode = http.StatusOK
			mainHanlder = true
			return response
		},
		postHandlers: []PostEventHandler{
			func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse {
        // do something
				return response
			},
			func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse {
        // do something
				return response
			},
		},
	})
  
	lambda.Start(router.MainHandler)
}
```


### Use Middleware in Router Level

```
package main

import (
    "fmt"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)



func main() {
	helloFunc := func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
		response := NewResponse()
		response.StatusCode = http.StatusOK
		mainHanlder = true
		return response
	}
	helloHandler := NewEvent(WithEvenHandler(helloFunc))

	router := New()
  router.GET("/hello", helloHandler)


	mainPreHandlers := []PreEventHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { 
			// do something      
    },
		func(ctx context.Context, request *events.APIGatewayProxyRequest) { routerPreHandler2 = true
			// do something
    },
	}

	mainPostHandlers := []PostEventHandler{
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse {
      // do something			
			return response
		},
		func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse {
      // do something
			return response
		},
	}

  mainRouter.UsePreHandler(mainPreHandlers...)
	mainRouter.UsePostHandler(mainPostHandlers...)

  
	lambda.Start(router.MainHandler)
}
```



## Custom Handler
amuro has support custom handler (NotFound, MethodNotAllowed, PanicHandler, ErrorHandler)

```
func CustomNotFound(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
  response := NewResponse()
  response.StatusCode = http.StatusNotFound
  response.Body = "custom_notfound"
  return response
}

func main() {
	router := New()
  router.GET("/hello", HelloFunc)
  
  router.NotFound = CustomNotFound
  
	lambda.Start(router.MainHandler)
}
```