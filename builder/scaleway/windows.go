package scaleway

import "strings"

func isWindowsCommercialType(commercialType string) bool {
	return strings.HasSuffix(commercialType, "-WIN")
}
