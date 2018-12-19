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
  router.GET("/hello", HelloFunc)

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
  mainFunc := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
    response := NewResponse()
    response.StatusCode = http.StatusOK
    return response, nil
  }

  preHandlers := []PreHandler{
    func(ctx context.Context, request *events.APIGatewayProxyRequest) { 
      // do something        
    },
  }

  postHandlers := []PostHandler{
    func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, err error) *events.APIGatewayProxyResponse {
      // do something
      return response
    },
    func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, err error) *events.APIGatewayProxyResponse {
      // do something
      return response
    },
  }

  router := New()
  router.GET("/hello", mainFunc,
    WithPreHandlers(preHandler...),
    WithPostHandlers(postHandler...),
  )
  
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
  helloFunc := func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
    response := NewResponse()
    response.StatusCode = http.StatusOK
    mainHanlder = true
    return response, nil
  }

  router := New()
  router.GET("/hello", helloFunc)
  
  mainPreHandlers := []PreHandler{
    func(ctx context.Context, request *events.APIGatewayProxyRequest) { 
      // do something      
    },
    func(ctx context.Context, request *events.APIGatewayProxyRequest) {
      // do something
    },
  }

  mainPostHandlers := []PostHandler{
    func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, err error) *events.APIGatewayProxyResponse {
      // do something			
      return response
    },
    func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, err error) *events.APIGatewayProxyResponse {
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
func CustomPathNotFound(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
  response := NewResponse()
  response.StatusCode = http.StatusNotFound
  response.Body = "custom_notfound"
  return response, nil
}

func main() {
  router := New()
  router.GET("/hello", HelloFunc)
  
  router.PathNotFound = CustomPathNotFound
  
  lambda.Start(router.MainHandler)
}
```
