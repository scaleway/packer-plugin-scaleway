package e2e

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/strcase"
)

// getTestFilePath returns a valid filename path based on the go test name and suffix. (Take care of non fs friendly char)
func getTestFilePath(t *testing.T, suffix string) string {
	specialChars := regexp.MustCompile(`[\\?%*:|"<>. ]`)

	// Replace nested tests separators.
	fileName := strings.Replace(t.Name(), "/", "-", -1)

	fileName = strcase.ToBashArg(fileName)

	// Replace special characters.
	fileName = specialChars.ReplaceAllLiteralString(fileName, "") + suffix

	// Remove prefix to simplify
	fileName = strings.TrimPrefix(fileName, "test-")

	return filepath.Join(".", "testdata", fileName)
}
