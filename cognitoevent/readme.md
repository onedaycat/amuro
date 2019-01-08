# Amuro/CognitoEvent
amuro/cognitoevent is lambda handler receive cognito event

## Usage
```
package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/onedaycat/amuro/cognitoevent"
)

func handler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	fmt.Printf("PostConfirmation for user: %s\n", event.UserName)
	return event, nil
}

func main() {
	eventManager := cognitoevent.NewEventManager()
	eventManager.RegisterPostConfirmationHandlers(handler, nil, nil)

	lambda.Start(eventManager.MainHandler)
}

```

### Use Middleware in Handle Level

```
package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/onedaycat/amuro/cognitoevent"
)

func handler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	fmt.Printf("PostConfirmation for user: %s\n", event.UserName)
	return event, nil
}

func preHandler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) {
	fmt.Printf("PreHandler: PostConfirmation for user: %s\n", event.UserName)

}

func postHandler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation, err error) {
	fmt.Printf("PostHandler: PostConfirmation for user: %s\n", event.UserName)
}

func main() {
	preHandlers := []cognitoevent.CognitoPostConfirmationPreHandler{preHandler}
	postHandlers := []cognitoevent.CognitoPostConfirmationPostHandler{postHandler}

	eventManager := cognitoevent.NewEventManager()
	eventManager.RegisterPostConfirmationHandlers(handler, preHandlers, postHandlers)

	lambda.Start(eventManager.MainHandler)
}

```


### Use Middleware in Global Level
amuro support global middleware (PreHandler, PostHandler)

```
package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/onedaycat/amuro/cognitoevent"
)

func handler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	fmt.Printf("PostConfirmation for user: %s\n", event.UserName)
	return event, nil
}

func preHandler(ctx context.Context, event interface{}) {
	fmt.Printf("PreHandler: PostConfirmation for event: %v\n", event)

}

func postHandler(ctx context.Context, event interface{}, err error) {
	fmt.Printf("PostHandler: PostConfirmation for event: %v\n", event)
}

func main() {
	eventManager := cognitoevent.NewEventManager(cognitoevent.WithPostHandlers(postHandler), cognitoevent.WithPreHandlers(preHandler))
	eventManager.RegisterPostConfirmationHandlers(handler, nil, nil)

	lambda.Start(eventManager.MainHandler)
}

```



## Custom Handler
amuro has support custom error (ErrorHandler)

```
package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/onedaycat/amuro/cognitoevent"
)

func handler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	fmt.Printf("PostConfirmation for user: %s\n", event.UserName)
	return event, nil
}

func customError(ctx context.Context, event interface{}, err error) {
	fmt.Println(err)
}

func main() {
	eventManager := cognitoevent.NewEventManager()
	eventManager.OnError = customError
	eventManager.RegisterPostConfirmationHandlers(handler, nil, nil)

	lambda.Start(eventManager.MainHandler)
}

```
