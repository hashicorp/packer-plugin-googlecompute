// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecomputemachineimage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-googlecompute/lib/common"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepCreateImage represents a Packer build step that creates GCE machine
// images.
type StepCreateMachineImage int

// Run executes the Packer build step that creates a GCE machine image.
//
// The image is created from the persistent disk used by the instance. The
// instance must be deleted and the disk retained before doing this step.
func (s *StepCreateMachineImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(common.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	// if config.SkipCreateImage {
	// 	ui.Say("Skipping image creation...")
	// 	return multistep.ActionContinue
	// }

	if config.PackerForce && config.machineImageAlreadyExists {
		ui.Say("Deleting previous machine image...")

		errCh := driver.DeleteMachineImage(config.ProjectId, config.MachineImageName)
		err := <-errCh
		if err != nil {
			err := fmt.Errorf("Error deleting image: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Creating machine image...")

	imageCh, errCh := driver.CreateMachineImage(
		config.ProjectId, config.MachineImageName, config.MachineImageDesc, config.MachineImageSourceInstance, config.MachineImageSourceInstanceZone)
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
	return multistep.ActionContinue
}

// Cleanup.
func (s *StepCreateMachineImage) Cleanup(state multistep.StateBag) {}
