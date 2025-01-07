package vcr

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/strcase"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// QueryMatcherIgnore is a list of query parameters that should be ignored when matching cassettes requests
var QueryMatcherIgnore = []string{
	"project_id",
	"project",
}

const UpdateCassettesEnvVariable = "PACKER_UPDATE_CASSETTES"

var UpdateCassettes = os.Getenv(UpdateCassettesEnvVariable) == "true"

// getTestFilePath returns a valid filename path based on the go test name and suffix. (Take care of non fs friendly char)
func getTestFilePath(t *testing.T, pkgFolder string, suffix string) string {
	t.Helper()
	specialChars := regexp.MustCompile(`[\\?%*:|"<>. ]`)

	// Replace nested tests separators.
	fileName := strings.ReplaceAll(t.Name(), "/", "-")

	fileName = strcase.ToBashArg(fileName)

	// Replace special characters.
	fileName = specialChars.ReplaceAllLiteralString(fileName, "") + suffix

	// Remove prefix to simplify
	fileName = strings.TrimPrefix(fileName, "test-")

	return filepath.Join(pkgFolder, "testdata", fileName)
}

func GetTestFilePath(t *testing.T, pkgFolder string) string {
	return getTestFilePath(t, pkgFolder, ".cassette")
}

// recorderAuthHook is a hook that will clean authorization tokens from cassette during record.
func recorderAuthHook(i *cassette.Interaction) error {
	i.Request.Headers = i.Request.Headers.Clone()
	delete(i.Request.Headers, "x-auth-token")
	delete(i.Request.Headers, "X-Auth-Token")
	delete(i.Request.Headers, "Authorization")

	return nil
}

func requestMatcher(actualRequest *http.Request, cassetteRequest cassette.Request) bool {
	cassetteURL, _ := url.Parse(cassetteRequest.URL)
	actualURL := actualRequest.URL
	cassetteQueryValues := cassetteURL.Query()
	actualQueryValues := actualURL.Query()
	for _, query := range QueryMatcherIgnore {
		actualQueryValues.Del(query)
		cassetteQueryValues.Del(query)
	}
	actualURL.RawQuery = actualQueryValues.Encode()
	cassetteURL.RawQuery = cassetteQueryValues.Encode()

	return actualRequest.Method == cassetteRequest.Method &&
		actualURL.String() == cassetteURL.String()
}

// GetHTTPRecorder creates a new httpClient that records all HTTP requests in a cassette.
// This cassette is then replayed whenever tests are executed again. This means that once the
// requests are recorded in the cassette, no more real HTTP requests must be made to run the tests.
//
// It is important to add a `defer cleanup()` so the given cassette files are correctly
// closed and saved after the requests.
func GetHTTPRecorder(cassetteFilePath string, update bool) (client *http.Client, cleanup func(), err error) {
	recorderMode := recorder.ModeReplayOnly
	if update {
		recorderMode = recorder.ModeRecordOnly
	}

	_, errorCassette := os.Stat(cassetteFilePath + ".yaml")

	// If in record mode we check that the cassette exists
	if recorderMode == recorder.ModeReplayOnly && errorCassette != nil {
		return nil, nil, fmt.Errorf("cannot stat file %s.yaml while in replay mode", cassetteFilePath)
	}

	// Setup recorder and scw client
	r, err := recorder.New(cassetteFilePath,
		recorder.WithMode(recorderMode),
		recorder.WithSkipRequestLatency(true),
		// Add a filter which removes Authorization headers from all requests:
		recorder.WithHook(recorderAuthHook, recorder.BeforeSaveHook),
		recorder.WithMatcher(requestMatcher))
	if err != nil {
		return nil, nil, err
	}
	defer func(r *recorder.Recorder) {
		_ = r.Stop()
	}(r)

	return &http.Client{Transport: r}, func() {
		r.Stop() // Make sure recorder is stopped once done with it
	}, nil
}
