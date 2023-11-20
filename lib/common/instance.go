// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

type InstanceConfig struct {
	AcceleratorType              string
	AcceleratorCount             int64
	Address                      string
	Description                  string
	DisableDefaultServiceAccount bool
	DiskName                     string
	DiskSizeGb                   int64
	DiskType                     string
	DiskEncryptionKey            *CustomerEncryptionKey
	EnableNestedVirtualization   bool
	EnableSecureBoot             bool
	EnableVtpm                   bool
	EnableIntegrityMonitoring    bool
	ExtraBlockDevices            []BlockDevice
	Image                        *Image
	Labels                       map[string]string
	MachineType                  string
	Metadata                     map[string]string
	MinCpuPlatform               string
	Name                         string
	Network                      string
	NetworkProjectId             string
	OmitExternalIP               bool
	OnHostMaintenance            string
	Preemptible                  bool
	NodeAffinities               []NodeAffinity
	Region                       string
	ServiceAccountEmail          string
	Scopes                       []string
	Subnetwork                   string
	Tags                         []string
	Zone                         string
}
