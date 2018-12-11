package apigateway

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type EventHandler func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse
type PreEventHandler func(ctx context.Context, request *events.APIGatewayProxyRequest)
type PostEventHandler func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse

type EventFlowHandler struct {
	preHandlers  []PreEventHandler
	handler      EventHandler
	postHandlers []PostEventHandler
}

func (f EventHandler) ServeEvent(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	return f(ctx, request)
}

func NewEventFlowHandler(eventHandler EventHandler) *EventFlowHandler {
	return &EventFlowHandler{
		handler: eventHandler,
	}
}

func NewResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		Headers: map[string]string{},
	}
}
