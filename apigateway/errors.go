package apigateway

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
)

var (
	ErrorUnmarshalJSON = errors.BadRequest("3000", "Unable unmarshal json")
	ErrorMarshalJSON   = errors.BadRequest("3001", "Unable marshal json")
)

func ErrorUnmarshalJSONResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: ErrorUnmarshalJSON.Status,
		Body:       ErrorUnmarshalJSON.Error(),
	}
}

func ErrorMarshalJSONResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: ErrorMarshalJSON.Status,
		Body:       ErrorMarshalJSON.Error(),
	}
}
