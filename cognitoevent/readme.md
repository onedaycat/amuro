# Amuro/CognitoEvent
cognitoevent is lambda handler receive cognito event (PostConfirmation, PreSignup)

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
	eventManager.RegisterPostConfirmationHandlers(handler)

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
	eventManager := cognitoevent.NewEventManager()
	eventManager.RegisterPostConfirmationHandlers(
		handler,
		cognitoevent.WithPostConfirmationPreHandlers(preHandler),
		cognitoevent.WithPostConfirmationPostHandlers(postHandler),
	)
	lambda.Start(eventManager.MainHandler)
}

```


### Use Middleware in Global Level
cognitoevent support global middleware (PreHandler, PostHandler)

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
	eventManager := cognitoevent.NewEventManager()
	eventManager.UsePreHandler(preHandler)
	eventManager.UsePostHandler(postHandler)
	eventManager.RegisterPostConfirmationHandlers(handler)

	lambda.Start(eventManager.MainHandler)
}

```



## Custom Handler
cognitoevent has support custom error (ErrorHandler)

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
	eventManager.RegisterPostConfirmationHandlers(handler)

	lambda.Start(eventManager.MainHandler)
}

```
