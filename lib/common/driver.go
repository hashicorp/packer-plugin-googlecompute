// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"crypto/rsa"
	"io"
	"time"

	compute "google.golang.org/api/compute/v1"
	oauth2_svc "google.golang.org/api/oauth2/v2"
	oslogin "google.golang.org/api/oslogin/v1"
)

// Driver is the interface that has to be implemented to communicate
// with GCE. The Driver interface exists mostly to allow a mock implementation
// to be used to test the steps.
type Driver interface {
	// CreateDisk creates a persistent disk from the specified config.
	CreateDisk(diskConfig BlockDevice) (<-chan *compute.Disk, <-chan error)

	// CreateImage creates an image from the given disk in Google Compute
	// Engine.
	CreateImage(project string, imageSpec *compute.Image) (<-chan *Image, <-chan error)

	// SetImageDeprecationStatus sets the deprecation, obsolete and deletion date
	// for the image with the given name.
	SetImageDeprecationStatus(project, name string, deprecationStatus *compute.DeprecationStatus) error

	// DeleteImage deletes the image with the given name.
	DeleteImage(project, name string) <-chan error

	// DeleteInstance deletes the given instance, keeping the boot disk.
	DeleteInstance(zone, name string) (<-chan error, error)

	// DeleteDisk deletes the disk with the given name.
	DeleteDisk(zone, name string) <-chan error

	// GetDisk gets the disk with the given name in a zone/region.
	GetDisk(zone, name string) (*compute.Disk, error)

	// GetImage gets an image; tries the default and public projects. If
	// fromFamily is true, name designates an image family instead of a
	// particular image.
	GetImage(name string, fromFamily bool) (*Image, error)

	// GetImageFromProject gets an image from a specific projects.
	// Returns the image from the first project in slice it can find one
	// If fromFamily is true, name designates an image family instead of a particular image.
	GetImageFromProjects(project []string, name string, fromFamily bool) (*Image, error)

	// GetImageFromProject gets an image from a specific project. If fromFamily
	// is true, name designates an image family instead of a particular image.
	GetImageFromProject(project, name string, fromFamily bool) (*Image, error)

	// GetProjectMetadata gets a metadata variable for the project.
	GetProjectMetadata(zone, key string) (string, error)

	// GetInstanceMetadata gets a metadata variable for the instance, name.
	GetInstanceMetadata(zone, name, key string) (string, error)

	// GetInternalIP gets the GCE-internal IP address for the instance.
	GetInternalIP(zone, name string) (string, error)

	// GetNatIP gets the NAT IP address for the instance.
	GetNatIP(zone, name string) (string, error)

	// GetSerialPortOutput gets the Serial Port contents for the instance.
	GetSerialPortOutput(zone, name string) (string, error)

	// GetTokenInfo gets the information about the token used for authentication
	GetTokenInfo() (*oauth2_svc.Tokeninfo, error)

	// ImageExists returns true if the specified image exists. If an error
	// occurs calling the API, this method returns false.
	ImageExists(project, name string) bool

	// RunInstance takes the given config and launches an instance.
	RunInstance(*InstanceConfig) (<-chan error, error)

	// WaitForInstance waits for an instance to reach the given state.
	WaitForInstance(state, zone, name string) <-chan error

	// CreateOrResetWindowsPassword creates or resets the password for a user on an Windows instance.
	CreateOrResetWindowsPassword(zone, name string, config *WindowsPasswordConfig) (<-chan error, error)

	// ImportOSLoginSSHKey imports SSH public key for OSLogin.
	ImportOSLoginSSHKey(user, sshPublicKey string) (*oslogin.LoginProfile, error)

	// DeleteOSLoginSSHKey deletes the SSH public key for OSLogin with the given key.
	DeleteOSLoginSSHKey(user, fingerprint string) error

	// Add to the instance metadata for the existing instance
	AddToInstanceMetadata(zone string, name string, metadata map[string]string) error

	// UploadToBucket uploads an artifact to a bucket on GCS.
	UploadToBucket(bucket, objectName string, data io.Reader) (string, error)

	// DeleteFromBucket deletes an object from a bucket on GCS.
	DeleteFromBucket(bucket, objectName string) error
}

// WindowsPasswordConfig is the data structure that GCE needs to encrypt the created
// windows password.
type WindowsPasswordConfig struct {
	Key                    *rsa.PrivateKey
	Password               string
	UserName               string        `json:"userName"`
	Modulus                string        `json:"modulus"`
	Exponent               string        `json:"exponent"`
	Email                  string        `json:"email"`
	ExpireOn               time.Time     `json:"expireOn"`
	WindowsPasswordTimeout time.Duration `json:"timeout"`
}

type windowsPasswordResponse struct {
	UserName          string `json:"userName"`
	PasswordFound     bool   `json:"passwordFound"`
	EncryptedPassword string `json:"encryptedPassword"`
	Modulus           string `json:"modulus"`
	Exponent          string `json:"exponent"`
	ErrorMessage      string `json:"errorMessage"`
}
