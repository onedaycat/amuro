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
		return response, nil
	}

func main() {
	router := New()
  router.GET("/hello", HelloFunc)
  
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
  return response, nil
}

func main() {
	router := New()
  router.GET("/hello", HelloFunc)
  
  router.NotFound = CustomNotFound
  
	lambda.Start(router.MainHandler)
}
```