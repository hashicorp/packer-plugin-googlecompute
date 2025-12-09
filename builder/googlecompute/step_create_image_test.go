// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/packer-plugin-googlecompute/lib/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/assert"
)

func TestStepCreateImage_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateImage)
}

func TestStepCreateImage(t *testing.T) {
	state := testState(t)
	step := new(StepCreateImage)
	defer step.Cleanup(state)

	c := state.Get("config").(*Config)
	d := state.Get("driver").(*common.DriverMock)

	d.CreateImageReturnSelfLink = "https://selflink/compute/v1/test"
	d.CreateImageReturnDiskSize = 420

	// run the step
	action := step.Run(context.Background(), state)
	assert.Equal(t, action, multistep.ActionContinue, "Step did not pass.")

	uncastImage, ok := state.GetOk("image")
	assert.True(t, ok, "State does not have resulting image.")
	image, ok := uncastImage.(*common.Image)
	assert.True(t, ok, "Image in state is not an Image.")

	// Verify created Image results.
	assert.Equal(t, c.ImageName, image.Name, "Created image does not match config name.")
	assert.Equal(t, len(c.ImageGuestOsFeatures), len(image.GuestOsFeatures), "Created image features does not match config.")
	assert.Equal(t, c.ImageLabels, image.Labels, "Created image labels does not match config.")
	assert.Equal(t, c.ImageLicenses, image.Licenses, "Created image licenses does not match config.")
	assert.Equal(t, c.ProjectId, image.ProjectId, "Created image project ID does not match config.")
	assert.Equal(t, d.CreateImageReturnSelfLink, image.SelfLink, "Created image selflink does not match config")
	assert.Equal(t, d.CreateImageReturnDiskSize, image.SizeGb, "Created image disk size does not match config")

	// Verify proper args passed to driver.CreateImage.
	assert.Equal(t, c.ProjectId, d.CreateImageProjectId, "Incorrect project ID passed to driver.")
}

func TestStepCreateImage_errorOnChannel(t *testing.T) {
	state := testState(t)
	step := new(StepCreateImage)
	defer step.Cleanup(state)

	errCh := make(chan error, 1)
	errCh <- errors.New("error")

	driver := state.Get("driver").(*common.DriverMock)
	driver.CreateImageErrCh = errCh

	// run the step
	action := step.Run(context.Background(), state)
	assert.Equal(t, action, multistep.ActionHalt, "Step should not have passed.")
	_, ok := state.GetOk("error")
	assert.True(t, ok, "State should have an error.")
	_, ok = state.GetOk("image_name")
	assert.False(t, ok, "State should not have a resulting image.")
}

func TestStepCreateImage_setsDeprecationFields(t *testing.T) {
	state := testState(t)
	step := new(StepCreateImage)
	defer step.Cleanup(state)

	c := state.Get("config").(*Config)
	d := state.Get("driver").(*common.DriverMock)

	// Set deprecation timestamps in config
	c.DeprecateAt = "2125-06-01T00:00:00Z"
	c.ObsoleteAt = "2125-07-01T00:00:00Z"
	c.DeleteAt = "2125-08-01T00:00:00Z"

	// Run step
	action := step.Run(context.Background(), state)
	assert.Equal(t, multistep.ActionContinue, action, "Step did not pass.")

	assert.Equal(t, c.DeprecateAt, d.DeprecatedImageStatus.Deprecated, "DeprecateAt mismatch")
	assert.Equal(t, c.ObsoleteAt, d.DeprecatedImageStatus.Obsolete, "ObsoleteAt mismatch")
	assert.Equal(t, c.DeleteAt, d.DeprecatedImageStatus.Deleted, "DeleteAt mismatch")
	assert.Equal(t, "DEPRECATED", d.DeprecatedImageStatus.State, "State should be DEPRECATED")
}
