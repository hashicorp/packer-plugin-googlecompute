// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer-plugin-googlecompute/lib/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"google.golang.org/api/compute/v1"
)

// StepCreateImage represents a Packer build step that creates GCE machine
// images.
type StepCreateImage int

// Run executes the Packer build step that creates a GCE machine image.
//
// The image is created from the persistent disk used by the instance. The
// instance must be deleted and the disk retained before doing this step.
func (s *StepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(common.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if config.SkipCreateImage {
		ui.Say("Skipping image creation...")
		return multistep.ActionContinue
	}

	if config.PackerForce && config.imageAlreadyExists {
		ui.Say("Deleting previous image...")

		errCh := driver.DeleteImage(config.ImageProjectId, config.ImageName)
		err := <-errCh
		if err != nil {
			err := fmt.Errorf("Error deleting image: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Creating image...")

	sourceDiskURI := fmt.Sprintf("/compute/v1/projects/%s/zones/%s/disks/%s", config.ProjectId, config.Zone, config.imageSourceDisk)

	imageFeatures := make([]*compute.GuestOsFeature, 0, len(config.ImageGuestOsFeatures))
	for _, v := range config.ImageGuestOsFeatures {
		imageFeatures = append(imageFeatures, &compute.GuestOsFeature{
			Type: v,
		})
	}
	imagePayload := &compute.Image{
		Architecture:       config.ImageArchitecture,
		Description:        config.ImageDescription,
		Name:               config.ImageName,
		Family:             config.ImageFamily,
		Labels:             config.ImageLabels,
		Licenses:           config.ImageLicenses,
		GuestOsFeatures:    imageFeatures,
		ImageEncryptionKey: config.ImageEncryptionKey.ComputeType(),
		SourceDisk:         sourceDiskURI,
		SourceType:         "RAW",
		StorageLocations:   config.ImageStorageLocations,
	}
	imageCh, errCh := driver.CreateImage(config.ImageProjectId, imagePayload)
	var err error
	select {
	case err = <-errCh:
	case <-time.After(config.StateTimeout):
		err = errors.New("time out while waiting for image to register")
	}

	if err != nil {
		err := fmt.Errorf("Error waiting for image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("image", <-imageCh)

	deprecationStatus, err := s.getDeprecationStatus(config)
	if err != nil {
		err := fmt.Errorf("Error getting deprecation status: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = driver.SetImageDeprecationStatus(config.ImageProjectId, config.ImageName, deprecationStatus)
	if err != nil {
		err := fmt.Errorf("Error setting image deprecation status: %s", err)
		state.Put("error", err.Error())
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if config.DeprecateAt != "" || config.ObsoleteAt != "" || config.DeleteAt != "" {
		ui.Say("Image deprecation status set")
	}

	return multistep.ActionContinue
}

func (s *StepCreateImage) getDeprecationStatus(config *Config) (*compute.DeprecationStatus, error) {
	var errs error
	deprecation := &compute.DeprecationStatus{}

	if config.DeprecateAt != "" || config.ObsoleteAt != "" || config.DeleteAt != "" {
		deprecation.State = "DEPRECATED"

		now := time.Now().UTC()

		if config.DeprecateAt != "" {
			t, err := time.Parse(time.RFC3339, config.DeprecateAt)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("invalid deprecate_at format (RFC3339 expected): %w", err))
			} else if t.Before(now) {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("deprecate_at must be a future time"))
			} else {
				deprecation.Deprecated = config.DeprecateAt
			}
		}

		if config.ObsoleteAt != "" {
			t, err := time.Parse(time.RFC3339, config.ObsoleteAt)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("invalid obsolete_at format (RFC3339 expected): %w", err))
			} else if t.Before(now) {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("obsolete_at must be a future time"))
			} else {
				deprecation.Obsolete = config.ObsoleteAt
			}

		}

		if config.DeleteAt != "" {
			t, err := time.Parse(time.RFC3339, config.DeleteAt)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("invalid delete_at format (RFC3339 expected): %w", err))
			} else if t.Before(now) {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("delete_at must be a future time"))
			} else {
				deprecation.Deleted = config.DeleteAt
			}
		}

	}
	log.Printf("[DEBUG] deprecate_at: %s", config.DeprecateAt)
	log.Printf("[DEBUG] obsolete_at: %s", config.ObsoleteAt)
	log.Printf("[DEBUG] delete_at: %s", config.DeleteAt)
	return deprecation, errs
}

// Cleanup.
func (s *StepCreateImage) Cleanup(state multistep.StateBag) {}
