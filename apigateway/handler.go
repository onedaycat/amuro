package apigateway

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"unicode/utf8"

	"github.com/aws/aws-lambda-go/events"
)

const defaultStatusCode = -1
const contentTypeHeaderKey = "Content-Type"

type CustomHandle func(*CustomResponse, *CustomRequest, Params)

type CustomRequest struct {
	Resource              string            `json:"resource"`
	Path                  string            `json:"path"`
	HTTPMethod            string            `json:"httpMethod"`
	Headers               map[string]string `json:"headers"`
	QueryStringParameters map[string]string `json:"queryStringParameters"`
	PathParameters        map[string]string `json:"pathParameters"`
	StageVariables        map[string]string `json:"stageVariables"`
	Body                  string            `json:"body"`
	IsBase64Encoded       bool              `json:"isBase64Encoded,omitempty"`
}

type CustomResponse struct {
	Headers    http.Header
	Body       bytes.Buffer
	StatusCode int
}

func (req *CustomRequest) URLString() string {
	return req.Path
}

func (res *CustomResponse) SetStatusCode(code int) {
	res.StatusCode = code
}

func (res *CustomResponse) ToAPIGatewayResponse() (events.APIGatewayProxyResponse, error) {
	if res.StatusCode == defaultStatusCode {
		return events.APIGatewayProxyResponse{}, errors.New("Status code not set on response")
	}

	var output string
	isBase64 := false

	bb := (&res.Body).Bytes()

	if utf8.Valid(bb) {
		output = string(bb)
	} else {
		output = base64.StdEncoding.EncodeToString(bb)
		isBase64 = true
	}

	proxyHeaders := make(map[string]string)

	for h := range res.Headers {
		proxyHeaders[h] = res.Headers.Get(h)
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      res.StatusCode,
		Headers:         proxyHeaders,
		Body:            output,
		IsBase64Encoded: isBase64,
	}, nil
}

func (res *CustomResponse) Write(body []byte) (int, error) {
	if res.StatusCode == -1 {
		res.StatusCode = http.StatusOK
	}

	// if the content type header is not set when we write the body we try to
	// detect one and set it by default. If the content type cannot be detected
	// it is automatically set to "application/octet-stream" by the
	// DetectContentType method
	if res.Headers.Get(contentTypeHeaderKey) == "" {
		res.Headers.Add(contentTypeHeaderKey, http.DetectContentType(body))
	}

	return (&res.Body).Write(body)
}

func NewCustomResponse() *CustomResponse {
	return &CustomResponse{
		StatusCode: defaultStatusCode,
		Headers:    make(http.Header),
	}
}

func NewCustomResquestFromEvent(event events.APIGatewayProxyRequest) *CustomRequest {
	return &CustomRequest{
		Resource:              event.Resource,
		Path:                  event.Path,
		HTTPMethod:            event.HTTPMethod,
		Headers:               event.Headers,
		QueryStringParameters: event.QueryStringParameters,
		PathParameters:        event.PathParameters,
		StageVariables:        event.StageVariables,
		Body:                  event.Body,
		IsBase64Encoded:       event.IsBase64Encoded,
	}
}
