package apigateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
)

type EventHandler func(ctx context.Context, request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)
type PreHandler func(ctx context.Context, request *events.APIGatewayProxyRequest)
type PostHandler func(ctx context.Context, request *events.APIGatewayProxyRequest, response *events.APIGatewayProxyResponse, err error) *events.APIGatewayProxyResponse

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

func NewSuccessResponse(body interface{}) (*events.APIGatewayProxyResponse, error) {
	var jsonString string
	if body != nil {
		jsonByte, err := json.Marshal(body)
		if err != nil {
			return ErrorMarshalJSONResponse(), err
		}

		jsonString = string(jsonByte)
	}

	return &events.APIGatewayProxyResponse{
		Headers:    map[string]string{},
		StatusCode: http.StatusOK,
		Body:       jsonString,
	}, nil
}

func NewErrorResponse(err error) *events.APIGatewayProxyResponse {
	appError, ok := err.(*errors.AppError)
	if !ok {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       transfrormErrorToJsonResponse(err),
		}
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: appError.Status,
		Body:       transfrormErrorToJsonResponse(appError),
	}
}
