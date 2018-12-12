package apigateway

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type eventHandler func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse
type preHandler func(ctx context.Context, request *events.APIGatewayProxyRequest)
type postHandler func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse

type Event func(e *event)

type event struct {
	preHandlers  []preHandler
	postHandlers []postHandler
	eventHandler eventHandler
}

func WithPreHandlers(preHandlers ...preHandler) Event {
	return func(e *event) {
		e.preHandlers = preHandlers
	}
}

func WithPostHandlers(postHandlers ...postHandler) Event {
	return func(e *event) {
		e.postHandlers = postHandlers
	}
}

func WithEventHandler(eventHandler eventHandler) Event {
	return func(e *event) {
		e.eventHandler = eventHandler
	}
}

func NewEvent(events ...Event) *event {
	e := &event{}
	for _, event := range events {
		event(e)
	}

	return e
}

func NewResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		Headers: map[string]string{},
	}
}
