package scaleway

import (
	"fmt"
	"log"

	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type ArtifactSnapshot struct {
	// The ID of the snapshot
	ID string
	// The name of the snapshot
	Name string
}

func (snap ArtifactSnapshot) String() string {
	return fmt.Sprintf("(%s: %s)", snap.Name, snap.ID)
}

type Artifact struct {
	// The name of the image
	ImageName string

	// The ID of the image
	ImageID string

	// Snapshots used by the generated image
	Snapshots []ArtifactSnapshot

	// The name of the zone
	ZoneName string

	// The Client for making API calls
	Client *scw.Client

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderID
}

func (*Artifact) Files() []string {
	// No files with Scaleway
	return nil
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s:%s", a.ZoneName, a.ImageID)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("An image was created: '%v' (ID: %v) in zone '%v' based on snapshots %v",
		a.ImageName, a.ImageID, a.ZoneName, a.Snapshots)
}

func (a *Artifact) State(name string) interface{} {
	if name == registryimage.ArtifactStateURI {
		img, err := registryimage.FromArtifact(a,
			registryimage.WithID(a.ImageID),
			registryimage.WithProvider("scaleway"),
			registryimage.WithRegion(a.ZoneName),
		)
		if err != nil {
			log.Printf("error when creating hcp registry image %v", err)

			return nil
		}

		return img
	}

	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	instanceAPI := instance.NewAPI(a.Client)

	log.Printf("Destroying image: %s (%s)", a.ImageID, a.ImageName)

	err := instanceAPI.DeleteImage(&instance.DeleteImageRequest{
		ImageID: a.ImageID,
	})
	if err != nil {
		return err
	}

	log.Printf("Destroying snapshots: %v", a.Snapshots)

	for _, snapshot := range a.Snapshots {
		err = instanceAPI.DeleteSnapshot(&instance.DeleteSnapshotRequest{
			SnapshotID: snapshot.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
