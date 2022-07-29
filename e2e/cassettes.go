package e2e

import (
	"flag"
	"net/http"
	"os"
	"testing"

	"github.com/dnaeon/go-vcr/v2/cassette"
	"github.com/dnaeon/go-vcr/v2/recorder"
	"github.com/stretchr/testify/assert"
)

// UpdateCassettes will update all cassettes of a given test
var updateCassettes = flag.Bool("cassettes", os.Getenv("PACKER_UPDATE_CASSETTES") == "true", "Record Cassettes")

// getHTTPRecoder creates a new httpClient that records all HTTP requests in a cassette.
// This cassette is then replayed whenever tests are executed again. This means that once the
// requests are recorded in the cassette, no more real HTTP requests must be made to run the tests.
//
// It is important to add a `defer cleanup()` so the given cassette files are correctly
// closed and saved after the requests.
func getHTTPRecoder(t *testing.T, update bool) (client *http.Client, cleanup func(), err error) {
	recorderMode := recorder.ModeReplaying
	if update {
		recorderMode = recorder.ModeRecording
	}

	r, err := recorder.NewAsMode(getTestFilePath(t, ".cassette"), recorderMode, nil)
	if err != nil {
		return nil, nil, err
	}

	// Add a filter which removes Authorization headers from all requests:
	r.AddFilter(func(i *cassette.Interaction) error {
		i.Request.Headers = i.Request.Headers.Clone()
		delete(i.Request.Headers, "x-auth-token")
		delete(i.Request.Headers, "X-Auth-Token")
		return nil
	})

	return &http.Client{
			Transport: &retryableHTTPTransport{transport: r},
		},
		func() {
			assert.NoError(t, r.Stop()) // Make sure recorder is stopped once done with it
		},
		nil
}
