package apigateway

import (
	"github.com/onedaycat/errors"
)

var (
	ErrorUnmarshalJSON = errors.BadRequest("3000", "Unable unmarshal json")
)
