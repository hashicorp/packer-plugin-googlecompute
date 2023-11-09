// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// The googlecompute package contains a packersdk.Builder implementation that
// builds images for Google Compute Engine.
package googlecompute

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-googlecompute/lib/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

// The unique ID for this builder.
const BuilderId = "packer.googlecompute"

// Builder represents a Packer Builder.
type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}
	generatedDataKeys := []string{
		// This will be set with the source image name even if the config
		// uses source image family instead of source image id.
		"SourceImageName",
	}

	return generatedDataKeys, warnings, nil
}

// Run executes a googlecompute Packer build and returns a packersdk.Artifact
// representing a GCE machine image.
func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	cfg := &common.GCEDriverConfig{
		Ui:        ui,
		ProjectId: b.config.ProjectId,
	}
	b.config.Authentication.ApplyDriverConfig(cfg)

	driver, err := common.NewDriverGCE(*cfg)
	if err != nil {
		return nil, err
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)
	generatedData := &packerbuilderdata.GeneratedData{State: state}

	// Build the steps.
	steps := []multistep.Step{
		new(StepCheckExistingImage),
		&communicator.StepSSHKeyGen{
			CommConf:            &b.config.Comm,
			SSHTemporaryKeyPair: b.config.Comm.SSH.SSHTemporaryKeyPair,
		},
		multistep.If(b.config.PackerDebug && b.config.Comm.SSHPrivateKeyFile == "",
			&communicator.StepDumpSSHKey{
				Path: fmt.Sprintf("gce_%s.pem", b.config.PackerBuildName),
				SSH:  &b.config.Comm.SSH,
			},
		),
		&StepCreateDisks{
			DiskConfiguration: b.config.ExtraBlockDevices,
		},
		&StepImportOSLoginSSHKey{
			Debug: b.config.PackerDebug,
		},
		&StepCreateInstance{
			Debug:         b.config.PackerDebug,
			GeneratedData: generatedData,
		},
		&StepCreateWindowsPassword{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("gce_windows_%s.pem", b.config.PackerBuildName),
		},
		&StepInstanceInfo{
			Debug: b.config.PackerDebug,
		},
		&StepStartTunnel{
			IAPConf:            &b.config.IAPConfig,
			CommConf:           &b.config.Comm,
			AccountFile:        b.config.AccountFile,
			ImpersonateAccount: b.config.ImpersonateServiceAccount,
			ProjectId:          b.config.ProjectId,
		},
		&communicator.StepConnect{
			Config:      &b.config.Comm,
			Host:        communicator.CommHost(b.config.Comm.Host(), "instance_ip"),
			SSHConfig:   b.config.Comm.SSHConfigFunc(),
			WinRMConfig: winrmConfig,
		},
		new(commonsteps.StepProvision),
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
	}
	if _, exists := b.config.Metadata[StartupScriptKey]; exists || b.config.StartupScriptFile != "" {
		steps = append(steps, new(StepWaitStartupScript))
	}
	steps = append(steps, new(StepTeardownInstance), new(StepCreateImage))

	// Run the steps.
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}
	if _, ok := state.GetOk("image"); !ok {
		log.Println("Failed to find image in state. Bug?")
		return nil, nil
	}

	artifact := &Artifact{
		image:     state.Get("image").(*common.Image),
		driver:    driver,
		config:    &b.config,
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	return artifact, nil
}
