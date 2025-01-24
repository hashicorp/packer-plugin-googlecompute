// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	compute "google.golang.org/api/compute/v1"
)

func TestBlockDevice_Prepare(t *testing.T) {
	testcases := []struct {
		name      string
		config    *BlockDevice
		expectErr bool
	}{
		{
			name: "OK - minimum scratch device",
			config: &BlockDevice{
				VolumeType: "scratch",
				VolumeSize: 375,
			},
			expectErr: false,
		},
		{
			name: "OK - minimum persistent device",
			config: &BlockDevice{
				VolumeType: "pd-standard",
				VolumeSize: 25,
			},
			expectErr: false,
		},
		{
			name: "OK - minimum persistent pd-balanced device",
			config: &BlockDevice{
				VolumeType: "pd-balanced",
				VolumeSize: 25,
				DiskName:   "test-disk",
			},
			expectErr: false,
		},
		{
			name: "OK - minimum persistent pd-ssd device",
			config: &BlockDevice{
				VolumeType: "pd-ssd",
				VolumeSize: 25,
				DiskName:   "test-disk",
			},
			expectErr: false,
		},
		{
			name: "OK - minimum persistent pd-extreme device",
			config: &BlockDevice{
				VolumeType: "pd-extreme",
				VolumeSize: 25,
				DiskName:   "test-disk",
			},
			expectErr: false,
		},
		{
			name: "Fail - minimum persistent device with invalid volume type",
			config: &BlockDevice{
				VolumeType: "pd-invalid",
				VolumeSize: 25,
				DiskName:   "test-disk",
			},
			expectErr: true,
		},
		{
			name: "OK - minimum persistent device with disk_name",
			config: &BlockDevice{
				VolumeType: "pd-standard",
				VolumeSize: 25,
				DiskName:   "test-disk",
			},
			expectErr: false,
		},
		{
			name: "OK - minimum persistent device with long disk_name",
			config: &BlockDevice{
				VolumeType: "pd-standard",
				VolumeSize: 25,
				DiskName:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			expectErr: false,
		},
		{
			name: "Fail - minimum persistent device with too long disk_name",
			config: &BlockDevice{
				VolumeType: "pd-standard",
				VolumeSize: 25,
				DiskName:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			expectErr: true,
		},
		{
			name: "Fail - minimum persistent device with too non-compliant disk_name",
			config: &BlockDevice{
				VolumeType: "pd-standard",
				VolumeSize: 25,
				DiskName:   "z_",
			},
			expectErr: true,
		},
		{
			name: "Fail - minimum scratch device with disk_name",
			config: &BlockDevice{
				VolumeType: "scratch",
				VolumeSize: 25,
				DiskName:   "test-disk",
			},
			expectErr: true,
		},
		{
			name: "Fail - minimum scratch device with device_name",
			config: &BlockDevice{
				VolumeType: "scratch",
				VolumeSize: 25,
				DeviceName: "test-disk",
			},
			expectErr: true,
		},
		{
			name: "OK - minimum scratch device with READ_ONLY attachment_mode",
			config: &BlockDevice{
				VolumeType:     "scratch",
				VolumeSize:     25,
				AttachmentMode: "READ_ONLY",
			},
			expectErr: false,
		},
		{
			name: "OK - minimum scratch device with READ_WRITE attachment_mode",
			config: &BlockDevice{
				VolumeType:     "scratch",
				VolumeSize:     25,
				AttachmentMode: "READ_WRITE",
			},
			expectErr: false,
		},
		{
			name: "Fail - minimum scratch device with invalid attachment_mode",
			config: &BlockDevice{
				VolumeType:     "scratch",
				VolumeSize:     25,
				AttachmentMode: "invalid",
			},
			expectErr: true,
		},
		{
			name: "OK - minimum scratch device with SCSI interface_type",
			config: &BlockDevice{
				VolumeType:    "scratch",
				VolumeSize:    25,
				InterfaceType: "SCSI",
			},
			expectErr: false,
		},
		{
			name: "OK - minimum scratch device with NVME interface_type",
			config: &BlockDevice{
				VolumeType:    "scratch",
				VolumeSize:    25,
				InterfaceType: "NVME",
			},
			expectErr: false,
		},
		{
			name: "Fail - minimum scratch device with invalid interface_type",
			config: &BlockDevice{
				VolumeType:    "scratch",
				VolumeSize:    25,
				InterfaceType: "SATA",
			},
			expectErr: true,
		},
		{
			name: "OK - IOPS in top range on pd_extreme",
			config: &BlockDevice{
				VolumeType: "pd-extreme",
				VolumeSize: 125,
				IOPS:       120000,
			},
			expectErr: false,
		},
		{
			name: "OK - IOPS in bottom range on pd_extreme",
			config: &BlockDevice{
				VolumeType: "pd-extreme",
				VolumeSize: 125,
				IOPS:       10000,
			},
			expectErr: false,
		},
		{
			name: "Fail - IOPS too low",
			config: &BlockDevice{
				VolumeType: "pd-extreme",
				VolumeSize: 125,
				IOPS:       9999,
			},
			expectErr: true,
		},
		{
			name: "Fail - IOPS too high",
			config: &BlockDevice{
				VolumeType: "pd-extreme",
				VolumeSize: 125,
				IOPS:       120001,
			},
			expectErr: true,
		},
		{
			name: "Fail - IOPS set on non-compatible volume type",
			config: &BlockDevice{
				VolumeType: "pd-standard",
				VolumeSize: 125,
				IOPS:       100000,
			},
			expectErr: true,
		},
		{
			name: "OK - keep_device set for persistent volume",
			config: &BlockDevice{
				VolumeType: "pd-standard",
				VolumeSize: 125,
				KeepDevice: true,
			},
			expectErr: false,
		},
		{
			name: "Fail - keep_device set for scratch volume",
			config: &BlockDevice{
				VolumeType: "scratch",
				VolumeSize: 125,
				KeepDevice: true,
			},
			expectErr: true,
		},
		{
			name: "OK - source volume set",
			config: &BlockDevice{
				SourceVolume: "zones/us-central1-a/disks/source-disk",
			},
			expectErr: false,
		},
		{
			name: "fail - source volume set along with volume_type",
			config: &BlockDevice{
				SourceVolume: "zones/us-central1-a/disks/source-disk",
				VolumeType:   "scratch",
			},
			expectErr: true,
		},
		{
			name: "fail - source volume set along with volume_size",
			config: &BlockDevice{
				SourceVolume: "zones/us-central1-a/disks/source-disk",
				VolumeSize:   20,
			},
			expectErr: true,
		},
		{
			name: "fail - source volume set along with iops",
			config: &BlockDevice{
				SourceVolume: "zones/us-central1-a/disks/source-disk",
				VolumeSize:   20,
			},
			expectErr: true,
		},
		{
			name: "fail - source volume set along with disk_name",
			config: &BlockDevice{
				SourceVolume: "zones/us-central1-a/disks/source-disk",
				DiskName:     "abcde",
			},
			expectErr: true,
		},
		{
			name: "fail - source volume set along with keep_device",
			config: &BlockDevice{
				SourceVolume: "zones/us-central1-a/disks/source-disk",
				KeepDevice:   true,
			},
			expectErr: true,
		},
		{
			name: "fail - source volume set along with source image",
			config: &BlockDevice{
				SourceImage:  "projects/p/global/images/family/f",
				SourceVolume: "zones/us-central1-a/disks/source-disk",
				KeepDevice:   true,
			},
			expectErr: true,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Prepare()
			if (len(errs) != 0) != tt.expectErr {
				t.Errorf("errors mismatch, expected %t errrors, got %d", tt.expectErr, len(errs))
			}

			for _, err := range errs {
				t.Logf("%s", err)
			}
		})
	}
}

func TestGenerateDiskAttachment(t *testing.T) {
	testcases := []struct {
		name      string
		config    BlockDevice
		expectval *compute.AttachedDisk
	}{
		{
			name: "basic scratch disk",
			config: BlockDevice{
				VolumeSize: 375,
				VolumeType: "scratch",
				Zone:       "us-central1-a",
			},
			expectval: &compute.AttachedDisk{
				AutoDelete:        true,
				Boot:              false,
				DiskEncryptionKey: &compute.CustomerEncryptionKey{},
				DiskSizeGb:        375,
				Interface:         "SCSI",
				Mode:              "READ_WRITE",
				Type:              "SCRATCH",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskType: "zones/us-central1-a/diskTypes/local-ssd",
				},
			},
		},
		{
			name: "basic persistent disk",
			config: BlockDevice{
				VolumeSize:     25,
				VolumeType:     "pd-standard",
				AttachmentMode: "READ_ONLY",
				InterfaceType:  "NVME",
				Zone:           "us-central1-a",
			},
			expectval: &compute.AttachedDisk{
				AutoDelete:        true,
				Boot:              false,
				DiskEncryptionKey: &compute.CustomerEncryptionKey{},
				Interface:         "NVME",
				Mode:              "READ_ONLY",
				Type:              "PERSISTENT",
			},
		},
		{
			name: "basic persistent disk with device name",
			config: BlockDevice{
				VolumeSize:     25,
				VolumeType:     "pd-standard",
				AttachmentMode: "READ_ONLY",
				InterfaceType:  "NVME",
				DeviceName:     "packer-test",
				Zone:           "us-central1-a",
			},
			expectval: &compute.AttachedDisk{
				AutoDelete:        true,
				Boot:              false,
				DiskEncryptionKey: &compute.CustomerEncryptionKey{},
				Interface:         "NVME",
				Mode:              "READ_ONLY",
				Type:              "PERSISTENT",
				DeviceName:        "packer-test",
			},
		},
		{
			name: "basic persistent disk from source",
			config: BlockDevice{
				AttachmentMode: "READ_ONLY",
				InterfaceType:  "NVME",
				Zone:           "us-central1-a",
				SourceVolume:   "dummy_source",
			},
			expectval: &compute.AttachedDisk{
				AutoDelete:        false,
				Boot:              false,
				DiskEncryptionKey: &compute.CustomerEncryptionKey{},
				Interface:         "NVME",
				Mode:              "READ_ONLY",
				Type:              "PERSISTENT",
				Source:            "dummy_source",
			},
		},
		{
			name: "basic persistent disk with keep device",
			config: BlockDevice{
				VolumeSize:     25,
				VolumeType:     "pd-standard",
				AttachmentMode: "READ_ONLY",
				InterfaceType:  "NVME",
				DeviceName:     "packer-test",
				Zone:           "us-central1-a",
				KeepDevice:     true,
			},
			expectval: &compute.AttachedDisk{
				AutoDelete:        false,
				Boot:              false,
				DiskEncryptionKey: &compute.CustomerEncryptionKey{},
				Interface:         "NVME",
				Mode:              "READ_ONLY",
				Type:              "PERSISTENT",
				DeviceName:        "packer-test",
			},
		},
		{
			name: "basic persistent disk with source image",
			config: BlockDevice{
				SourceImage: "projects/p/global/images/family/f",
				VolumeType:  "pd-standard",
				DeviceName:  "packer-test",
				DiskName:    "packer-test-disk",
			},
			expectval: &compute.AttachedDisk{
				AutoDelete:        true,
				Boot:              false,
				DiskEncryptionKey: &compute.CustomerEncryptionKey{},
				Interface:         "SCSI",
				Mode:              "READ_WRITE",
				Type:              "PERSISTENT",
				DeviceName:        "packer-test",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    "packer-test-disk",
					SourceImage: "projects/p/global/images/family/f",
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Prepare()
			if err != nil {
				t.Fatalf("failed to prepare config: %s", err)
			}

			att := tt.config.GenerateDiskAttachment()
			diff := cmp.Diff(att, tt.expectval)
			if diff != "" {
				t.Errorf("found differences in generated disk attachment: %s", diff)
			}
		})
	}
}

func TestGenerateComputeDisk(t *testing.T) {
	testcases := []struct {
		name      string
		config    BlockDevice
		expectval *compute.Disk
	}{
		{
			name: "with a source volume, return nil",
			config: BlockDevice{
				SourceVolume: "abcd",
			},
			expectval: nil,
		},
		{
			name: "simple, without replica zones",
			config: BlockDevice{
				VolumeType: "pd-ssd",
				VolumeSize: 250,
				DiskName:   "packer-test",
				Zone:       "us-central1-a",
			},
			expectval: &compute.Disk{
				Description:       "created by Packer",
				SizeGb:            250,
				Name:              "packer-test",
				DiskEncryptionKey: &compute.CustomerEncryptionKey{},
				Type:              "zones/us-central1-a/diskTypes/pd-ssd",
			},
		},
		{
			name: "with custom IOPS set",
			config: BlockDevice{
				VolumeType: "pd-extreme",
				VolumeSize: 250,
				DiskName:   "packer-test",
				IOPS:       110000,
				Zone:       "us-central1-a",
			},
			expectval: &compute.Disk{
				Description:       "created by Packer",
				SizeGb:            250,
				Name:              "packer-test",
				DiskEncryptionKey: &compute.CustomerEncryptionKey{},
				ProvisionedIops:   110000,
				Type:              "zones/us-central1-a/diskTypes/pd-extreme",
			},
		},
		{
			name: "with extra zones set",
			config: BlockDevice{
				VolumeType:   "pd-extreme",
				VolumeSize:   250,
				DiskName:     "packer-test",
				IOPS:         110000,
				ReplicaZones: []string{"us-central1-b", "us-central1-c"},
				Zone:         "us-central1-a",
			},
			expectval: &compute.Disk{
				Description:       "created by Packer",
				SizeGb:            250,
				Name:              "packer-test",
				DiskEncryptionKey: &compute.CustomerEncryptionKey{},
				ProvisionedIops:   110000,
				ReplicaZones: []string{
					"zones/us-central1-a",
					"zones/us-central1-b",
					"zones/us-central1-c",
				},
				Type: "regions/us-central1/diskTypes/pd-extreme",
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Prepare()
			if errs != nil {
				t.Fatalf("failed to prepare config: %#v", errs)
			}

			att, err := tt.config.GenerateComputeDiskPayload()
			if err != nil {
				t.Fatalf("failed to generate compute disk payload: %s", err)
			}
			diff := cmp.Diff(att, tt.expectval)
			if diff != "" {
				t.Errorf("found differences in generated compute disk: %s", diff)
			}
		})
	}
}

func TestGetRegionFromZone(t *testing.T) {
	zone := "us-central1-a"
	region, err := GetRegionFromZone(zone)
	if err != nil {
		t.Fatalf("region extraction failed: %s", err)
	}
	assert.Equal(t, "us-central1", region)
}

func TestIsRegion(t *testing.T) {
	zone := "us-central1-a"
	isRegion := IsZoneARegion(zone)
	if isRegion {
		t.Errorf("expected zone %q not to be a region, but isZoneARegion returned %t", zone, isRegion)
	}

	region := "us-central1"
	isRegion = IsZoneARegion(region)
	if !isRegion {
		t.Errorf("expected region %q to be a region, but isZoneARegion returned %t", region, isRegion)
	}
}
