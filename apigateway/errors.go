package apigateway

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
)

type ErrorJsonResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	ErrorUnmarshalJSON = errors.BadRequest("3000", "Unable unmarshal json")
	ErrorMarshalJSON   = errors.InternalError("3001", "Unable marshal json")
)

func TransfrormErrorToJsonResponse(err error) string {
	if err == nil {
		return ""
	}

	var errorResponse *ErrorJsonResponse

	appErr, ok := err.(*errors.AppError)
	if !ok {
		errorResponse = &ErrorJsonResponse{
			Code:    "UNKNOWN_CODE",
			Message: err.Error(),
		}
	} else {
		errorResponse = &ErrorJsonResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
		}
	}

	jsonError, _ := json.Marshal(errorResponse)
	return string(jsonError)
}

func ErrorUnmarshalJSONResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: ErrorUnmarshalJSON.Status,
		Body:       TransfrormErrorToJsonResponse(ErrorUnmarshalJSON),
	}
}

func ErrorMarshalJSONResponse() *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: ErrorMarshalJSON.Status,
		Body:       TransfrormErrorToJsonResponse(ErrorMarshalJSON),
	}
}
