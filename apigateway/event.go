package apigateway

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type EventHandler func(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse

func (f EventHandler) ServeEvent(ctx context.Context, request *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	return f(ctx, request)
}

func NewResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		Headers: map[string]string{},
	}
}
