package apigateway

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type EventHandler func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)

func (f EventHandler) ServeEvent(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return f(ctx, request)
}

func setHeader(res *events.APIGatewayProxyResponse, field, value string) {
	if res.Headers == nil {
		res.Headers = map[string]string{}
	}
	res.Headers[field] = value
}

func NewResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		Headers: map[string]string{},
	}
}
