package scaleway

import (
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"

	_ "unsafe"
)

//go:linkname createServer github.com/scaleway/scaleway-sdk-go/api/instance/v1.(*API).createServer
func createServer(*instance.API, *instance.CreateServerRequest, ...scw.RequestOption) (*instance.CreateServerResponse, error)
