// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-googlecompute/lib/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
)

// ErrStartupScriptMetadata means that the user provided startup script resulted in
// setting the set-startup-script metadata status to error.
var ErrStartupScriptMetadata = errors.New("Startup script exited with error.")

// StepWaitStartupScript is a trivial implementation of a Packer multistep
// It can be used for tracking the set-startup-script metadata status.
type StepWaitStartupScript int

// Run reads the instance metadata and looks for the log entry
// indicating the startup script finished.
func (s *StepWaitStartupScript) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(common.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	instanceName := state.Get("instance_name").(string)

	if config.WrapStartupScriptFile.False() {
		return multistep.ActionContinue
	}

	ui.Say("Waiting for any running startup script to finish...")
	// Keep checking the serial port output to see if the startup script is done.
	err := retry.Config{
		ShouldRetry: func(err error) bool {
			if errors.Is(err, ErrStartupScriptMetadata) {
				return false
			}
			return true
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 10 * time.Second, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		status, err := driver.GetInstanceMetadata(config.Zone,
			instanceName, StartupScriptStatusKey)

		if err != nil {
			ui.Message(fmt.Sprintf("Metadata %s on instance %s not available. Waiting...", StartupScriptStatusKey, instanceName))
			err := fmt.Errorf("Error getting startup script status: %s", err)
			return err
		}

		switch status {
		case StartupScriptStatusError:
			ui.Message("Startup script in error. Exiting...")
			return ErrStartupScriptMetadata

		case StartupScriptStatusDone:
			ui.Message("Startup script successfully finished.")
			return nil

		default:
			ui.Message("Startup script not finished yet. Waiting...")
			return errors.New("Startup script not done.")
		}
	})

	if err != nil {
		err := fmt.Errorf("Error waiting for startup script to finish: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say("Startup script, if any, has finished running.")
	return multistep.ActionContinue
}

// Cleanup.
func (s *StepWaitStartupScript) Cleanup(state multistep.StateBag) {}
