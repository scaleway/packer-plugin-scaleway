package scaleway

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepBackup struct{}

func (s *stepBackup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*scw.Client)
	instanceAPI := instance.NewAPI(client)
	blockAPI := block.NewAPI(client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	server := state.Get("server").(*instance.Server)
	zone := scw.Zone(c.Zone)

	ui.Say(fmt.Sprintf("Backing up server to image: %v", c.ImageName))

	snapshots := make(map[string]any, len(server.Volumes))
	for index, volume := range server.Volumes {
		volumeIdentifier := volume.ID
		if volume.Name != nil {
			volumeIdentifier = *volume.Name
		}

		ui.Say("Creating snapshot for volume: " + volumeIdentifier)

		switch volume.VolumeType {
		case instance.VolumeServerVolumeTypeLSSD:
			req := &instance.CreateSnapshotRequest{
				Zone:       zone,
				Name:       c.RootVolume.SnapshotName,
				VolumeID:   &volume.ID,
				VolumeType: instance.SnapshotVolumeTypeLSSD,
			}

			if len(c.Tags) > 0 {
				req.Tags = &c.Tags
			}

			snap, err := instanceAPI.CreateSnapshot(req, scw.WithContext(ctx))
			if err != nil {
				return putErrorAndHalt(state, ui, fmt.Errorf("failed to snapshot instance volume: %w", err))
			}

			snapshots[index] = snap.Snapshot
		case instance.VolumeServerVolumeTypeSbsVolume:
			snapshotName := c.RootVolume.SnapshotName
			if i, _ := strconv.Atoi(index); i != 0 {
				snapshotName = c.BlockVolumes[i-1].SnapshotName
			}

			snap, err := blockAPI.CreateSnapshot(&block.CreateSnapshotRequest{
				Zone:     zone,
				VolumeID: volume.ID,
				Name:     snapshotName,
				Tags:     c.Tags,
			}, scw.WithContext(ctx))
			if err != nil {
				return putErrorAndHalt(state, ui, fmt.Errorf("failed to snapshot block volume: %w", err))
			}

			availableStatus := block.SnapshotStatusAvailable

			_, err = blockAPI.WaitForSnapshot(&block.WaitForSnapshotRequest{
				SnapshotID:     snap.ID,
				Zone:           zone,
				TerminalStatus: &availableStatus,
			}, scw.WithContext(ctx))
			if err != nil {
				return putErrorAndHalt(state, ui, fmt.Errorf("failed to wait for block snapshot: %w", err))
			}

			snapshots[index] = snap
		default:
			return putErrorAndHalt(state, ui, fmt.Errorf("cannot snapshot unknown volume type %T", volume.VolumeType))
		}
	}

	// Build volume templates for image creation
	rootVolumeSnapID := ""
	extraVolumes := make(map[string]*instance.VolumeTemplate, len(snapshots)-1)

	for index, snapshot := range snapshots {
		if instanceSnap, ok := snapshot.(*instance.Snapshot); ok && index == "0" {
			rootVolumeSnapID = instanceSnap.ID
		} else if blockSnap, ok := snapshot.(*block.Snapshot); ok {
			if index == "0" {
				rootVolumeSnapID = blockSnap.ID
			} else {
				extraVolumes[index] = &instance.VolumeTemplate{ID: blockSnap.ID}
			}
		}
	}

	ui.Say(fmt.Sprintf("Creating image from snapshots: %v", c.ImageName))

	imageResp, err := instanceAPI.CreateImage(&instance.CreateImageRequest{
		Zone:         zone,
		Name:         c.ImageName,
		RootVolume:   rootVolumeSnapID,
		Arch:         server.Arch,
		ExtraVolumes: extraVolumes,
		Tags:         c.Tags,
		Public:       scw.BoolPtr(false),
	}, scw.WithContext(ctx))
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("failed to create image: %w", err))
	}

	image, err := instanceAPI.WaitForImage(&instance.WaitForImageRequest{
		ImageID: imageResp.Image.ID,
		Zone:    zone,
		Timeout: &c.ImageCreationTimeout,
	}, scw.WithContext(ctx))
	if err != nil {
		return putErrorAndHalt(state, ui, fmt.Errorf("failed to fetch generated image: %w", err))
	}

	artifactsSnapshots := artifactSnapshotFromImage(image)

	// Apply tags to volumes
	if len(c.Tags) != 0 {
		err = applyTags(ctx, instanceAPI, blockAPI, scw.Zone(c.Zone), server.Volumes, c.Tags)
		if err != nil {
			return putErrorAndHalt(state, ui, err)
		}
	}

	state.Put("snapshots", artifactsSnapshots)
	state.Put("image_id", image.ID)
	state.Put("image_name", c.ImageName)
	state.Put("region", c.Zone) // Deprecated
	state.Put("zone", c.Zone)

	return multistep.ActionContinue
}

func (s *stepBackup) Cleanup(_ multistep.StateBag) {
	// no cleanup
}

func artifactSnapshotFromImage(image *instance.Image) []ArtifactSnapshot {
	snapshots := make([]ArtifactSnapshot, 0, len(image.ExtraVolumes)+1)

	snapshots = append(snapshots, ArtifactSnapshot{
		ID:   image.RootVolume.ID,
		Name: image.RootVolume.Name,
	})
	for _, extraVolume := range image.ExtraVolumes {
		snapshots = append(snapshots, ArtifactSnapshot{
			ID:   extraVolume.ID,
			Name: extraVolume.Name,
		})
	}

	return snapshots
}

func applyTags(ctx context.Context, instanceAPI *instance.API, blockAPI *block.API, zone scw.Zone, volumes map[string]*instance.VolumeServer, tags []string) error {
	for _, volume := range volumes {
		if _, blockErr := blockAPI.UpdateVolume(&block.UpdateVolumeRequest{
			VolumeID: volume.ID,
			Zone:     zone,
			Tags:     &tags,
		}, scw.WithContext(ctx)); blockErr != nil {
			_, instanceErr := instanceAPI.UpdateVolume(&instance.UpdateVolumeRequest{
				Zone:     zone,
				VolumeID: volume.ID,
				Tags:     &tags,
			}, scw.WithContext(ctx))
			if instanceErr != nil {
				return fmt.Errorf("failed to set tags on the volume: %w", errors.Join(blockErr, instanceErr))
			}
		}
	}

	return nil
}
