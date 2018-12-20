package apigateway

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
)

var (
	ErrorUnmarshalJSON = errors.BadRequest("3000", "Unable unmarshal json")
)

func ErrorUnmarshalJSONResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: ErrorUnmarshalJSON.Status,
		Body:       ErrorUnmarshalJSON.Error(),
	}
}
