package apigateway

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestBodyJSON struct {
	Name            string   `json:"name"`
	ListProductName []string `json:"listProductName"`
}

func TestNewSuccessResponse(t *testing.T) {
	testBodyJSON := &TestBodyJSON{
		Name:            "john",
		ListProductName: []string{"product1", "product2"},
	}

	expectedBodyJSONString := `{"name":"john","listProductName":["product1","product2"]}`

	resp, err := NewSuccessResponse(testBodyJSON)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedBodyJSONString, resp.Body)
}
