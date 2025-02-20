package scaleway

import (
	"encoding/json"
	"errors"
	_ "unsafe" // Import required for link

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

//go:linkname createServer github.com/scaleway/scaleway-sdk-go/api/instance/v1.(*API).createServer
func createServer(*instance.API, *instance.CreateServerRequest, ...scw.RequestOption) (*instance.CreateServerResponse, error)

func newResponseErrorFromBody(rawBody []byte) error {
	responseError := scw.ResponseError{}
	_ = json.Unmarshal(rawBody, &responseError)

	return &responseError
}

// formatNonStandardError provides a way to format non-standard errors returned by instance's API.
// If error is not detected as non-standard, format will be no-op.
func formatInstanceError(err error) error {
	preconditionFailedError := &scw.PreconditionFailedError{}
	if errors.As(err, &preconditionFailedError) && preconditionFailedError.Precondition == "" {
		return newResponseErrorFromBody(preconditionFailedError.RawBody)
	}

	return err
}
