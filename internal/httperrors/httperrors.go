package httperrors

import (
	"errors"
	"net/http"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

// IsHTTPCodeError returns true if err is an http error with code statusCode
func IsHTTPCodeError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	responseError := &scw.ResponseError{}
	if errors.As(err, &responseError) && responseError.StatusCode == statusCode {
		return true
	}
	return false
}

// Is404 returns true if err is an HTTP 404 error
func Is404(err error) bool {
	notFoundError := &scw.ResourceNotFoundError{}
	return IsHTTPCodeError(err, http.StatusNotFound) || errors.As(err, &notFoundError)
}
