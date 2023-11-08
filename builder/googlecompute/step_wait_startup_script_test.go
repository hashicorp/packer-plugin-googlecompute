// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-googlecompute/lib/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/stretchr/testify/assert"
)

func TestStepWaitStartupScript(t *testing.T) {
	state := testState(t)
	step := new(StepWaitStartupScript)
	c := state.Get("config").(*Config)
	d := state.Get("driver").(*common.DriverMock)

	testZone := "test-zone"
	testInstanceName := "test-instance-name"

	c.Zone = testZone
	state.Put("instance_name", testInstanceName)

	// This step stops when it gets Done back from the metadata.
	d.GetInstanceMetadataResult = StartupScriptStatusDone

	// Run the step.
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have passed and continued.")

	// Check that GetInstanceMetadata was called properly.
	assert.Equal(t, d.GetInstanceMetadataZone, testZone, "Incorrect zone passed to GetInstanceMetadata.")
	assert.Equal(t, d.GetInstanceMetadataName, testInstanceName, "Incorrect instance name passed to GetInstanceMetadata.")
}

func TestStepWaitStartupScript_withWrapStartupScript(t *testing.T) {
	tt := []struct {
		Name                               string
		WrapStartup                        config.Trilean
		MetadataResult, Zone, MetadataName string
		StepResult                         multistep.StepAction //Zero value for StepAction is StepContinue; this is expected for all passing test cases.
	}{
		{Name: "no- wrapped startup script", WrapStartup: config.TriFalse},
		{Name: "good - wrapped startup script", WrapStartup: config.TriTrue, MetadataResult: StartupScriptStatusDone, Zone: "test-zone", MetadataName: "test-instance-name"},
		{
			Name:           "failed - wrapped startup script",
			WrapStartup:    config.TriTrue,
			MetadataResult: StartupScriptStatusError,
			Zone:           "test-zone",
			MetadataName:   "failed-instance-name",
			StepResult:     multistep.ActionHalt,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			state := testState(t)
			step := new(StepWaitStartupScript)
			c := state.Get("config").(*Config)
			d := state.Get("driver").(*common.DriverMock)

			c.StartupScriptFile = "startup.sh"
			c.WrapStartupScriptFile = tc.WrapStartup
			c.Zone = tc.Zone
			state.Put("instance_name", tc.MetadataName)

			// This step stops when it gets Done back from the metadata.
			d.GetInstanceMetadataResult = tc.MetadataResult

			// Run the step.
			assert.Equal(t, step.Run(context.Background(), state), tc.StepResult, "Step should have continued.")

			assert.Equal(t, d.GetInstanceMetadataResult, tc.MetadataResult, "MetadataResult was not the expected value.")
			assert.Equal(t, d.GetInstanceMetadataZone, tc.Zone, "Zone was not the expected value.")
			assert.Equal(t, d.GetInstanceMetadataName, tc.MetadataName, "Instance name was not the expected value.")
		})
	}
}
