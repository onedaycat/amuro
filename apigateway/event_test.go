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

func TestNewSuccessResponseWithNilData(t *testing.T) {
	expectedBodyJSONString := ``

	resp, err := NewSuccessResponse(nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedBodyJSONString, resp.Body)
}

func TestNewSuccessResponseError(t *testing.T) {
	resp, err := NewSuccessResponse(func() {})
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "3001: Unable marshal json", resp.Body)
}
