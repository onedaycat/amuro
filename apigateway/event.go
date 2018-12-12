package apigateway

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type eventHandler func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse
type preHandler func(ctx context.Context, request *events.APIGatewayProxyRequest)
type postHandler func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) *events.APIGatewayProxyResponse

type Option func(o *option)

type option struct {
	preHandlers  []preHandler
	postHandlers []postHandler
	eventHandler eventHandler
}

func WithPreHandlers(preHandlers ...preHandler) Option {
	return func(o *option) {
		o.preHandlers = preHandlers
	}
}

func WithPostHandlers(postHandlers ...postHandler) Option {
	return func(o *option) {
		o.postHandlers = postHandlers
	}
}

func WithEventHandler(eventHandler eventHandler) Option {
	return func(o *option) {
		o.eventHandler = eventHandler
	}
}

func NewOption(opts ...Option) *option {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}

	return o
}

func NewResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		Headers: map[string]string{},
	}
}
