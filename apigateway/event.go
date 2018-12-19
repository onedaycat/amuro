package apigateway

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type EventHandler func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)
type PreHandler func(ctx context.Context, request *events.APIGatewayProxyRequest)
type PostHandler func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse) (*events.APIGatewayProxyResponse, error)

type Option func(o *option)

type event struct {
	preHandlers  []PreHandler
	postHandlers []PostHandler
	eventHandler EventHandler
}
type option struct {
	preHandlers  []PreHandler
	postHandlers []PostHandler
}

func WithPreHandlers(preHandlers ...PreHandler) Option {
	return func(o *option) {
		o.preHandlers = preHandlers
	}
}

func WithPostHandlers(postHandlers ...PostHandler) Option {
	return func(o *option) {
		o.postHandlers = postHandlers
	}
}

func newOption(opts ...Option) *option {
	o := &option{}
	if opts == nil {
		return o
	}

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
