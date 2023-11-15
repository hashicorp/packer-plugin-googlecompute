// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type BlockDevice

package common

import (
	"fmt"
	"log"
	"regexp"

	"github.com/gofrs/uuid"
	compute "google.golang.org/api/compute/v1"
)

type BlockDeviceType string

const (
	LocalScratch  BlockDeviceType = "scratch"
	ZonalStandard                 = "pd-standard"
	ZonalBalanced                 = "pd-balanced"
	ZonalSSD                      = "pd-ssd"
	ZonalExtreme                  = "pd-extreme"
)

var diskNameRegex = regexp.MustCompile("^[a-z]([-a-z0-9]*[a-z0-9])?$")

// BlockDevice is a block device attachement/creation to an instance when building an image.
type BlockDevice struct {
	// How to attach the volume to the instance
	//
	// Can be either READ_ONLY or READ_WRITE (default).
	AttachmentMode string `mapstructure:"attachment_mode"`
	// If true, an image will be created for this disk, instead of the boot disk.
	//
	// This only applies to non-scratch disks, and can only be specified on one disk at a
	// time.
	CreateImage bool `mapstructure:"create_image"`
	// The device name as exposed to the OS in the /dev/disk/by-id/google-* directory
	//
	// If unspecified, the disk will have a default name in the form
	// persistent-disk-x with 'x' being a number assigned by GCE
	//
	// This field only applies to persistent disks, local SSDs will always
	// be exposed as /dev/disk/by-id/google-local-nvme-ssd-x.
	DeviceName string `mapstructure:"device_name"`
	// Disk encryption key to apply to the requested disk.
	//
	// Possible values:
	// * kmsKeyName -  The name of the encryption key that is stored in Google Cloud KMS.
	// * RawKey: - A 256-bit customer-supplied encryption key, encodes in RFC 4648 base64.
	DiskEncryptionKey CustomerEncryptionKey `mapstructure:"disk_encryption_key"`
	// Name of the disk to create.
	// This only applies to non-scratch disks. If the disk is persistent, and
	// not specified, Packer will generate a unique name for the disk.
	//
	// The name must be 1-63 characters long and comply to the regexp
	// '[a-z]([-a-z0-9]*[a-z0-9])?'
	DiskName string `mapstructure:"disk_name"`
	// The interface to use for attaching the disk.
	// Can be either NVME or SCSI. Defaults to SCSI.
	//
	// The available options depend on the type of disk, SEE: https://cloud.google.com/compute/docs/disks/persistent-disks#choose_an_interface
	InterfaceType string `mapstructure:"interface_type"`
	// The requested IOPS for the disk.
	//
	// This is only available for pd_extreme disks.
	IOPS int `mapstructure:"iops"`
	// Keep the device in the created disks after the instance is terminated.
	// By default, the builder will remove the disks at the end of the build.
	//
	// This cannot be used with 'scratch' volumes.
	KeepDevice bool `mapstructure:"keep_device"`
	// The list of extra zones to replicate the disk into
	//
	// The zone in which the instance is created will automatically be
	// added to the zones in which the disk is replicated.
	ReplicaZones []string `mapstructure:"replica_zones" required:"false"`
	// The URI of the volume to attach
	//
	// If this is specified, it won't be deleted after the instance is shut-down.
	SourceVolume string `mapstructure:"source_volume"`
	// Size of the volume to request, in gigabytes.
	//
	// The size specified must be in range of the sizes for the chosen volume type.
	VolumeSize int `mapstructure:"volume_size" required:"true"`
	// The volume type is the type of storage to reserve and attach to the instance being provisioned.
	//
	// The following values are supported by this builder:
	// * scratch: local SSD data, always 375 GiB (default)
	// * pd_standard: persistent, HDD-backed disk
	// * pd_balanced: persistent, SSD-backed disk
	// * pd_ssd: persistent, SSD-backed disk, with extra performance guarantees
	// * pd_extreme: persistent, fastest SSD-backed disk, with custom IOPS
	//
	// For details on the different types, refer to: https://cloud.google.com/compute/docs/disks#disk-types
	VolumeType BlockDeviceType `mapstructure:"volume_type" required:"true"`
	// Zone is the zone in which to create the disk in.
	//
	// It is not exposed since the parent config already specifies it
	// and it will be set for the block device when preparing it.
	Zone string `mapstructure:"_"`
}

func volumeTypeError() string {
	return fmt.Sprintf("valid volume types are: %s, %s, %s, %s and %s",
		LocalScratch,
		ZonalStandard,
		ZonalBalanced,
		ZonalSSD,
		ZonalExtreme)
}

func (bd *BlockDevice) Prepare() []error {
	var errs []error

	err := bd.prepareDiskCreate()
	if err != nil {
		errs = append(errs, err...)
	}

	switch bd.InterfaceType {
	case "SCSI", "NVME":
	case "":
		bd.InterfaceType = "SCSI"
	default:
		errs = append(errs, fmt.Errorf("Invalid interface_type: %q", bd.InterfaceType))
		errs = append(errs, fmt.Errorf("Valid values are SCSI or NVME"))
	}

	switch bd.AttachmentMode {
	case "READ_ONLY", "READ_WRITE":
	case "":
		bd.AttachmentMode = "READ_WRITE"
	default:
		errs = append(errs, fmt.Errorf("Invalid attachment_mode: %q", bd.AttachmentMode))
		errs = append(errs, fmt.Errorf("Valid values are READ_ONLY or READ_WRITE"))
	}

	if bd.DeviceName != "" && bd.VolumeType == LocalScratch {
		errs = append(errs, fmt.Errorf("Scratch volumes may not have a device_name attached to them"))
	}

	if bd.CreateImage && bd.VolumeType == LocalScratch {
		errs = append(errs, fmt.Errorf("Scratch volumes may not have create_image enabled"))
	}

	if bd.SourceVolume != "" {
		bd.KeepDevice = true
	}

	return errs
}

func (bd BlockDevice) hasDiskCreationArgs() bool {
	return bd.VolumeSize != 0 ||
		bd.VolumeType != "" ||
		bd.DiskName != "" ||
		bd.IOPS != 0 ||
		bd.KeepDevice
}

func (bd *BlockDevice) prepareDiskCreate() []error {
	if bd.SourceVolume != "" && bd.hasDiskCreationArgs() {
		return []error{
			fmt.Errorf(`when specifying a source_volume, the following configuration arguments cannot be used:
* disk_name
* volume_type
* volume_size
* iops
* keep_device`),
		}
	}

	if bd.SourceVolume != "" {
		return nil
	}

	var errs []error

	switch bd.VolumeType {
	case LocalScratch,
		ZonalStandard, ZonalBalanced, ZonalSSD, ZonalExtreme:
	default:
		errs = append(errs, fmt.Errorf("A valid volume type was not specified %q", bd.VolumeType))
		errs = append(errs, fmt.Errorf("%s", volumeTypeError()))
		return errs
	}

	if (bd.IOPS != 0) && (bd.VolumeType != ZonalExtreme) {
		errs = append(errs, fmt.Errorf("IOPS may only be specified for %q volumes", ZonalExtreme))
	}

	if bd.IOPS != 0 && (bd.IOPS < 10000 || bd.IOPS > 120000) {
		errs = append(errs, fmt.Errorf("Requested IOPS must be >= 10000 and <= 120000"))
	}

	if bd.VolumeType == LocalScratch && bd.KeepDevice {
		errs = append(errs, fmt.Errorf("Scratch volumes cannot be kept after the instance is shutdown"))
	}

	if bd.VolumeType == LocalScratch && bd.DiskName != "" {
		errs = append(errs, fmt.Errorf("Scratch volumes cannot have a name specified."))
	}

	if bd.VolumeSize == 0 {
		errs = append(errs, fmt.Errorf("volume_size must be specified"))
	}

	// No need to continue checking for LocalScratch types
	if bd.VolumeType == LocalScratch {
		return errs
	}

	if bd.DiskName != "" && (!diskNameRegex.MatchString(bd.DiskName) || len(bd.DiskName) > 63) {
		errs = append(errs, fmt.Errorf("Disk name %q is non-compliant.", bd.DiskName))
		errs = append(errs, fmt.Errorf("The name must be 1-63 characters long and comply to the regexp '[a-z]([-a-z0-9]*[a-z0-9])?'"))
	}

	// If we've gotten here then we know VolumeType is not LocalScratch
	if bd.DiskName == "" {
		uuid, err := uuid.NewV4()
		if err != nil {
			errs = append(errs, fmt.Errorf("error creating the disk name: %s", err))
			return errs
		}
		bd.DiskName = fmt.Sprintf("packer-%s", uuid.String())
		log.Printf("[TRACE] - Set disk name as %q", bd.DiskName)
	}

	return errs
}

var regionRegexp = regexp.MustCompile("^(.+)-[^-]$")

func GetRegionFromZone(zone string) (string, error) {
	matches := regionRegexp.FindStringSubmatch(zone)
	if len(matches) != 2 {
		return "", fmt.Errorf("failed to extract region from zone %q", zone)
	}
	return matches[1], nil
}

var zoneRegexp = regexp.MustCompile("^[a-z]+-[a-z]+[0-9]-[a-z]$")

func IsZoneARegion(zone string) bool {
	return !zoneRegexp.MatchString(zone)
}

func (bd BlockDevice) GenerateComputeDiskPayload() (*compute.Disk, error) {
	// We don't create a new disk if it is referenced
	if bd.SourceVolume != "" {
		return nil, nil
	}

	payload := &compute.Disk{
		Name:              bd.DiskName,
		DiskEncryptionKey: bd.DiskEncryptionKey.ComputeType(),
		SizeGb:            int64(bd.VolumeSize),
		Description:       "created by Packer",
	}

	if bd.IOPS != 0 {
		payload.ProvisionedIops = int64(bd.IOPS)
	}

	if len(bd.ReplicaZones) == 0 {
		payload.Type = fmt.Sprintf("zones/%s/diskTypes/%s", bd.Zone, bd.VolumeType)
	} else {
		region, err := GetRegionFromZone(bd.Zone)
		if err != nil {
			return nil, err
		}
		payload.Type = fmt.Sprintf("regions/%s/diskTypes/%s", region, bd.VolumeType)
		payload.ReplicaZones = []string{
			fmt.Sprintf("zones/%s", bd.Zone),
		}
	}

	for _, zone := range bd.ReplicaZones {
		log.Printf("setting extra replica zone %s", zone)
		payload.ReplicaZones = append(payload.ReplicaZones, fmt.Sprintf("zones/%s", zone))
	}

	log.Printf("payload type is %q", payload.Type)
	log.Printf("replica zones are: %#v", payload.ReplicaZones)

	return payload, nil
}

// shouldAutoDelete returns whether the disk should be automatically deleted after build is done.
func (bd BlockDevice) shouldAutoDelete() bool {
	if bd.VolumeType == LocalScratch {
		return true
	}

	if bd.CreateImage || bd.KeepDevice {
		return false
	}

	return true
}

func (bd BlockDevice) GenerateDiskAttachment() *compute.AttachedDisk {
	if bd.VolumeType == LocalScratch {
		return &compute.AttachedDisk{
			AutoDelete:        true,
			Boot:              false,
			DiskEncryptionKey: bd.DiskEncryptionKey.ComputeType(),
			DiskSizeGb:        int64(bd.VolumeSize),
			Interface:         bd.InterfaceType,
			Mode:              bd.AttachmentMode,
			Type:              "SCRATCH",
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskType: fmt.Sprintf("zones/%s/diskTypes/local-ssd", bd.Zone),
			},
		}
	}

	return &compute.AttachedDisk{
		AutoDelete:        bd.shouldAutoDelete(),
		Boot:              false,
		DeviceName:        bd.DeviceName,
		Interface:         bd.InterfaceType,
		Mode:              bd.AttachmentMode,
		DiskEncryptionKey: bd.DiskEncryptionKey.ComputeType(),
		Type:              "PERSISTENT",
		Source:            bd.SourceVolume,
	}
}
