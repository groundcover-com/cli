package api

import (
	"fmt"
	"net/http"
)

type ResponseError struct {
	error
	statusCode int
}

func NewResponseError(response *http.Response) error {
	return ResponseError{
		error:      fmt.Errorf("%s - %s", response.Request.URL, response.Status),
		statusCode: response.StatusCode,
	}
}
