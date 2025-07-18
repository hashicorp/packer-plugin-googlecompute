// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/packer-plugin-googlecompute/lib/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

// StepCreateInstance represents a Packer build step that creates GCE instances.
type StepCreateInstance struct {
	Debug         bool
	GeneratedData *packerbuilderdata.GeneratedData
}

func (c *Config) createInstanceMetadata(sourceImage *common.Image, sshPublicKey string) (map[string]string, map[string]string, error) {

	instanceMetadataNoSSHKeys := make(map[string]string)
	instanceMetadataSSHKeys := make(map[string]string)

	sshMetaKey := "ssh-keys"

	var err error
	var errs *packersdk.MultiError

	// Copy metadata from config.
	for k, v := range c.Metadata {
		if k == sshMetaKey {
			instanceMetadataSSHKeys[k] = v
		} else {
			instanceMetadataNoSSHKeys[k] = v
		}
	}

	if c.Comm.SSHPrivateKeyFile == "" && sshPublicKey != "" {
		sshMetaKey := "ssh-keys"
		sshPublicKey = strings.TrimSuffix(sshPublicKey, "\n")
		sshKeys := fmt.Sprintf("%s:%s %s", c.loginProfileUsername, sshPublicKey, c.loginProfileUsername)
		if confSSHKeys, exists := instanceMetadataSSHKeys[sshMetaKey]; exists {
			sshKeys = fmt.Sprintf("%s\n%s", sshKeys, confSSHKeys)
		}
		instanceMetadataSSHKeys[sshMetaKey] = sshKeys
	}

	startupScript := instanceMetadataNoSSHKeys[StartupScriptKey]
	if c.StartupScriptFile != "" {
		var content []byte
		content, err = os.ReadFile(c.StartupScriptFile)
		if err != nil {
			return nil, instanceMetadataNoSSHKeys, err
		}
		startupScript = string(content)
	}
	instanceMetadataNoSSHKeys[StartupScriptKey] = startupScript

	// Wrap any found startup script with our own startup script wrapper.
	if startupScript != "" && c.WrapStartupScriptFile.True() {
		instanceMetadataNoSSHKeys[StartupScriptKey] = StartupScriptLinux
		instanceMetadataNoSSHKeys[StartupWrappedScriptKey] = startupScript
		instanceMetadataNoSSHKeys[StartupScriptStatusKey] = StartupScriptStatusNotDone
	}

	if sourceImage.IsWindows() {
		// Windows startup script support is not yet implemented so clear any script data and set status to done
		instanceMetadataNoSSHKeys[StartupScriptKey] = StartupScriptWindows
		instanceMetadataNoSSHKeys[StartupScriptStatusKey] = StartupScriptStatusDone
	}

	// If UseOSLogin is true, force `enable-oslogin` in metadata
	// In the event that `enable-oslogin` is not enabled at project level
	if c.UseOSLogin.True() {
		instanceMetadataNoSSHKeys[EnableOSLoginKey] = "TRUE"
	}

	if c.UseOSLogin.False() {
		instanceMetadataNoSSHKeys[EnableOSLoginKey] = "FALSE"
	}

	for key, value := range c.MetadataFiles {
		var content []byte
		content, err = os.ReadFile(value)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
		instanceMetadataNoSSHKeys[key] = string(content)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return instanceMetadataNoSSHKeys, instanceMetadataSSHKeys, errs
	}
	return instanceMetadataNoSSHKeys, instanceMetadataSSHKeys, nil
}

func getImage(c *Config, d common.Driver) (*common.Image, error) {
	name := c.SourceImageFamily
	fromFamily := true
	if c.SourceImage != "" {
		name = c.SourceImage
		fromFamily = false
	}
	if len(c.SourceImageProjectId) == 0 {
		return d.GetImage(name, fromFamily)
	} else {
		return d.GetImageFromProjects(c.SourceImageProjectId, name, fromFamily)
	}
}

// Run executes the Packer build step that creates a GCE instance.
func (s *StepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	d := state.Get("driver").(common.Driver)

	ui := state.Get("ui").(packersdk.Ui)

	sourceImage, err := getImage(c, d)
	if err != nil {
		err := fmt.Errorf("Error getting source image for instance creation: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// The export post processor uses the `StepCreateInstance`, but with this we
	// avoid overwritting the SourceImageName.
	if s.GeneratedData != nil {
		// Store source image name for use in PARtifact.
		s.GeneratedData.Put("SourceImageName", sourceImage.Name)
	}

	if c.EnableSecureBoot && !sourceImage.IsSecureBootCompatible() {
		err := fmt.Errorf("Image: %s is not secure boot compatible. Please set 'enable_secure_boot' to false or choose another source image.", sourceImage.Name)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Using image: %s", sourceImage.Name))

	if sourceImage.IsWindows() && c.Comm.Type == "winrm" && c.Comm.WinRMPassword == "" {
		state.Put("create_windows_password", true)
	}

	ui.Say("Creating instance...")
	name := c.InstanceName

	var errCh <-chan error
	var metadataNoSSHKeys map[string]string
	var metadataSSHKeys map[string]string
	metadataForInstance := make(map[string]string)

	metadataNoSSHKeys, metadataSSHKeys, errs := c.createInstanceMetadata(sourceImage, string(c.Comm.SSHPublicKey))
	if errs != nil {
		state.Put("error", errs.Error())
		ui.Error(errs.Error())
		return multistep.ActionHalt
	}

	if c.WaitToAddSSHKeys > 0 {
		log.Printf("[DEBUG] Adding metadata during instance creation, but not SSH keys...")
		metadataForInstance = metadataNoSSHKeys
	} else {
		log.Printf("[DEBUG] Adding metadata during instance creation...")

		// Union of both non-SSH key meta data and SSH key meta data
		addmap(metadataForInstance, metadataSSHKeys)
		addmap(metadataForInstance, metadataNoSSHKeys)
	}

	errCh, err = d.RunInstance(&common.InstanceConfig{
		AcceleratorType:              c.AcceleratorType,
		AcceleratorCount:             c.AcceleratorCount,
		Address:                      c.Address,
		Description:                  "New instance created by Packer",
		DisableDefaultServiceAccount: c.DisableDefaultServiceAccount,
		DiskName:                     c.DiskName,
		DiskSizeGb:                   c.DiskSizeGb,
		DiskType:                     c.DiskType,
		DiskEncryptionKey:            c.DiskEncryptionKey,
		EnableNestedVirtualization:   c.EnableNestedVirtualization,
		EnableSecureBoot:             c.EnableSecureBoot,
		EnableVtpm:                   c.EnableVtpm,
		EnableIntegrityMonitoring:    c.EnableIntegrityMonitoring,
		ExtraBlockDevices:            c.ExtraBlockDevices,
		Image:                        sourceImage,
		Labels:                       c.Labels,
		MachineType:                  c.MachineType,
		Metadata:                     metadataForInstance,
		MinCpuPlatform:               c.MinCpuPlatform,
		Name:                         name,
		Network:                      c.Network,
		NetworkProjectId:             c.NetworkProjectId,
		OmitExternalIP:               c.OmitExternalIP,
		OnHostMaintenance:            c.OnHostMaintenance,
		MaxRunDurationInSeconds:      c.MaxRunDurationInSeconds,
		InstanceTerminationAction:    c.InstanceTerminationAction,
		Preemptible:                  c.Preemptible,
		NodeAffinities:               c.NodeAffinities,
		Region:                       c.Region,
		ServiceAccountEmail:          c.ServiceAccountEmail,
		Scopes:                       c.Scopes,
		Subnetwork:                   c.Subnetwork,
		Tags:                         c.Tags,
		Zone:                         c.Zone,
		NetworkIP:                    c.NetworkIP,
	})

	if err == nil {
		ui.Message("Waiting for creation operation to complete...")
		select {
		case err = <-errCh:
		case <-time.After(c.StateTimeout):
			err = errors.New("time out while waiting for instance to create")
		}
	}

	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message("Instance has been created!")

	if s.Debug {
		if name != "" {
			ui.Message(fmt.Sprintf("Instance: %s started in %s", name, c.Zone))
		}
	}

	// Things succeeded, store the name so we can remove it later
	state.Put("instance_name", name)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", name)

	if c.WaitToAddSSHKeys > 0 {
		ui.Message(fmt.Sprintf("Waiting %s before adding SSH keys...",
			c.WaitToAddSSHKeys.String()))
		cancelled := s.waitForBoot(ctx, c.WaitToAddSSHKeys)
		if cancelled {
			return multistep.ActionHalt
		}

		log.Printf("[DEBUG] %s wait is over. Adding SSH keys to existing instance...",
			c.WaitToAddSSHKeys.String())
		err = d.AddToInstanceMetadata(c.Zone, name, metadataSSHKeys)

		if err != nil {
			err := fmt.Errorf("Error adding SSH keys to existing instance: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateInstance) waitForBoot(ctx context.Context, waitLen time.Duration) bool {
	// Use a select to determine if we get cancelled during the wait
	select {
	case <-ctx.Done():
		return true
	case <-time.After(waitLen):
	}

	return false
}

// Cleanup destroys the GCE instance created during the image creation process.
func (s *StepCreateInstance) Cleanup(state multistep.StateBag) {
	nameRaw, ok := state.GetOk("instance_name")
	if !ok {
		return
	}
	name := nameRaw.(string)
	if name == "" {
		return
	}

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(common.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting instance...")
	errCh, err := driver.DeleteInstance(config.Zone, name)
	if err == nil {
		select {
		case err = <-errCh:
		case <-time.After(config.StateTimeout):
			err = errors.New("time out while waiting for instance to delete")
		}
	}

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting instance. Please delete it manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", name, err))
	}

	ui.Message("Instance has been deleted!")
	state.Put("instance_name", "")

	// Deleting the instance does not remove the boot disk. This cleanup removes
	// the disk.
	ui.Say("Deleting disk...")
	errCh = driver.DeleteDisk(config.Zone, config.DiskName)
	select {
	case err = <-errCh:
	case <-time.After(config.StateTimeout):
		err = errors.New("time out while waiting for disk to delete")
	}

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting disk. Please delete it manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", config.InstanceName, err))
	}

	ui.Message("Disk has been deleted!")
}

func addmap(a map[string]string, b map[string]string) {

	for k, v := range b {
		a[k] = v
	}
}
