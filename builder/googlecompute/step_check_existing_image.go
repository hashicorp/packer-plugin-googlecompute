// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepCheckExistingImage represents a Packer build step that checks if the
// target image already exists, and aborts immediately if so.
type StepCheckExistingImage int

// Run executes the Packer build step that checks if the image already exists.
func (s *StepCheckExistingImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	d := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if c.UseMachineImage {
		ui.Say("Checking machine image does not exist...")
		c.machineImageAlreadyExists = d.MachineImageExists(c.ProjectId, c.MachineImageName)
		if !c.PackerForce && c.machineImageAlreadyExists {
			err := fmt.Errorf("Machine Image %s already exists in project %s.\n"+
				"Use the force flag to delete it prior to building.", c.MachineImageName, c.ProjectId)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		ui.Say("Checking image does not exist...")
		c.imageAlreadyExists = d.ImageExists(c.ImageProjectId, c.ImageName)
		if !c.PackerForce && c.imageAlreadyExists {
			err := fmt.Errorf("Image %s already exists in project %s.\n"+
				"Use the force flag to delete it prior to building.", c.ImageName, c.ImageProjectId)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}
	return multistep.ActionContinue
}

// Cleanup.
func (s *StepCheckExistingImage) Cleanup(state multistep.StateBag) {}
