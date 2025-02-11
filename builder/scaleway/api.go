package scaleway

import (
	_ "unsafe" // Import required for link

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

//go:linkname createServer github.com/scaleway/scaleway-sdk-go/api/instance/v1.(*API).createServer
func createServer(*instance.API, *instance.CreateServerRequest, ...scw.RequestOption) (*instance.CreateServerResponse, error)
