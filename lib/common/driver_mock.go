// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"io"

	compute "google.golang.org/api/compute/v1"
	oauth2_svc "google.golang.org/api/oauth2/v2"
	oslogin "google.golang.org/api/oslogin/v1"
)

// DriverMock is a Driver implementation that is a mocked out so that
// it can be used for tests.
type DriverMock struct {
	CreateDiskConfig   BlockDevice
	CreateDiskResultCh <-chan *compute.Disk
	CreateDiskErrCh    <-chan error

	CreateImageProjectId      string
	CreateImageSpec           *compute.Image
	CreateImageReturnDiskSize int64
	CreateImageReturnSelfLink string
	CreateImageErrCh          <-chan error
	CreateImageResultCh       <-chan *Image

	DeleteProjectId  string
	DeleteImageName  string
	DeleteImageErrCh <-chan error

	DeleteInstanceZone  string
	DeleteInstanceName  string
	DeleteInstanceErrCh <-chan error
	DeleteInstanceErr   error

	DeleteDiskZone  string
	DeleteDiskName  string
	DeleteDiskErrCh chan error
	DeleteDiskErr   error

	DeleteFromBucketBucket     string
	DeleteFromBucketObjectName string
	DeleteFromBucketErr        error

	GetDiskName   string
	GetDiskZone   string
	GetDiskResult *compute.Disk
	GetDiskErr    error

	GetImageName           string
	GetImageSourceProjects []string
	GetImageFromFamily     bool
	GetImageResult         *Image
	GetImageErr            error

	GetImageFromProjectProject    string
	GetImageFromProjectName       string
	GetImageFromProjectFromFamily bool
	GetImageFromProjectResult     *Image
	GetImageFromProjectErr        error

	GetInstanceMetadataZone   string
	GetInstanceMetadataName   string
	GetInstanceMetadataKey    string
	GetInstanceMetadataResult string
	GetInstanceMetadataErr    error

	GetTokenInfoResult *oauth2_svc.Tokeninfo
	GetTokenInfoErr    error

	GetNatIPZone   string
	GetNatIPName   string
	GetNatIPResult string
	GetNatIPErr    error

	GetInternalIPZone   string
	GetInternalIPName   string
	GetInternalIPResult string
	GetInternalIPErr    error

	GetSerialPortOutputZone   string
	GetSerialPortOutputName   string
	GetSerialPortOutputResult string
	GetSerialPortOutputErr    error

	ImageExistsProjectId string
	ImageExistsName      string
	ImageExistsResult    bool

	RunInstanceConfig *InstanceConfig
	RunInstanceErrCh  <-chan error
	RunInstanceErr    error

	CreateOrResetWindowsPasswordZone     string
	CreateOrResetWindowsPasswordInstance string
	CreateOrResetWindowsPasswordConfig   *WindowsPasswordConfig
	CreateOrResetWindowsPasswordErr      error
	CreateOrResetWindowsPasswordErrCh    <-chan error

	WaitForInstanceState string
	WaitForInstanceZone  string
	WaitForInstanceName  string
	WaitForInstanceErrCh <-chan error

	AddToInstanceMetadataZone    string
	AddToInstanceMetadataName    string
	AddToInstanceMetadataKVPairs map[string]string
	AddToInstanceMetadataErrCh   <-chan error
	AddToInstanceMetadataErr     error

	UploadToBucketBucket     string
	UploadToBucketObjectName string
	UploadToBucketData       io.Reader
	UploadToBucketResult     string
	UploadToBucketError      error
}

func (d *DriverMock) CreateImage(project string, imageSpec *compute.Image) (<-chan *Image, <-chan error) {
	d.CreateImageProjectId = project
	d.CreateImageSpec = imageSpec
	resultCh := d.CreateImageResultCh
	if resultCh == nil {
		ch := make(chan *Image, 1)

		selfLink := d.CreateImageReturnSelfLink
		if selfLink == "" {
			selfLink = fmt.Sprintf("http://content.googleapis.com/compute/v1/%s/global/licenses/test", d.CreateImageProjectId)
		}

		diskSizeGb := d.CreateImageReturnDiskSize
		if diskSizeGb == 0 {
			diskSizeGb = 25
		}

		ch <- &Image{
			Architecture:    imageSpec.Architecture,
			GuestOsFeatures: imageSpec.GuestOsFeatures,
			Labels:          imageSpec.Labels,
			Licenses:        imageSpec.Licenses,
			Name:            imageSpec.Name,
			ProjectId:       d.CreateImageProjectId,
			SelfLink:        selfLink,
			SizeGb:          diskSizeGb,
		}
		close(ch)
		resultCh = ch
	}

	errCh := d.CreateImageErrCh
	if errCh == nil {
		ch := make(chan error)
		close(ch)
		errCh = ch
	}

	return resultCh, errCh
}

// CreateImageFromRaw is very similar to CreateImage, so we'll merge the two together in a later commit.
//
// Let's not spend time mocking it now, we'll make it mockable after merging the two functions.
func (d *DriverMock) CreateImageFromRaw(
	project string,
	rawImageURL string,
	imageName string,
	imageDescription string,
	imageFamily string,
	imageLabels map[string]string,
	imageGuestOsFeatures []string,
	shieldedVMStateConfig *compute.InitialStateConfig,
	imageStorageLocations []string,
	imageArchitecture string,
) (<-chan *Image, <-chan error) {
	return nil, nil
}

func (d *DriverMock) DeleteImage(project, name string) <-chan error {
	d.DeleteProjectId = project
	d.DeleteImageName = name

	resultCh := d.DeleteImageErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh
}

func (d *DriverMock) DeleteInstance(zone, name string) (<-chan error, error) {
	d.DeleteInstanceZone = zone
	d.DeleteInstanceName = name

	resultCh := d.DeleteInstanceErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh, d.DeleteInstanceErr
}

func (d *DriverMock) DeleteFromBucket(bucket, objectName string) error {
	d.DeleteFromBucketBucket = bucket
	d.DeleteFromBucketObjectName = objectName

	return d.DeleteFromBucketErr
}

func (d *DriverMock) CreateDisk(diskConfig BlockDevice) (<-chan *compute.Disk, <-chan error) {
	d.CreateDiskConfig = diskConfig

	resultCh := d.CreateDiskResultCh
	if resultCh == nil {
		ch := make(chan *compute.Disk)
		close(ch)
		resultCh = ch
	}

	errCh := d.CreateDiskErrCh
	if errCh != nil {
		ch := make(chan error)
		close(ch)
		errCh = ch
	}

	return resultCh, errCh
}

func (d *DriverMock) DeleteDisk(zone, name string) <-chan error {
	d.DeleteDiskZone = zone
	d.DeleteDiskName = name

	resultCh := d.DeleteDiskErrCh
	if resultCh == nil {
		ch := make(chan error)
		resultCh = ch
	}

	if d.DeleteDiskErr != nil {
		resultCh <- d.DeleteDiskErr
	}

	close(resultCh)

	return resultCh
}

func (d *DriverMock) GetDisk(zoneOrRegion, name string) (*compute.Disk, error) {
	d.GetDiskZone = zoneOrRegion
	d.GetDiskName = name

	return d.GetDiskResult, d.GetDiskErr
}

func (d *DriverMock) GetImage(name string, fromFamily bool) (*Image, error) {
	d.GetImageName = name
	d.GetImageFromFamily = fromFamily
	return d.GetImageResult, d.GetImageErr
}
func (d *DriverMock) GetImageFromProjects(projects []string, name string, fromFamily bool) (*Image, error) {
	d.GetImageSourceProjects = projects
	d.GetImageFromProjectName = name
	d.GetImageFromProjectFromFamily = fromFamily
	return d.GetImageFromProjectResult, d.GetImageFromProjectErr
}

func (d *DriverMock) GetImageFromProject(project, name string, fromFamily bool) (*Image, error) {
	d.GetImageFromProjectProject = project
	d.GetImageFromProjectName = name
	d.GetImageFromProjectFromFamily = fromFamily
	return d.GetImageFromProjectResult, d.GetImageFromProjectErr
}

func (d *DriverMock) GetInstanceMetadata(zone, name, key string) (string, error) {
	d.GetInstanceMetadataZone = zone
	d.GetInstanceMetadataName = name
	d.GetInstanceMetadataKey = key
	return d.GetInstanceMetadataResult, d.GetInstanceMetadataErr
}

func (d *DriverMock) GetNatIP(zone, name string) (string, error) {
	d.GetNatIPZone = zone
	d.GetNatIPName = name
	return d.GetNatIPResult, d.GetNatIPErr
}

func (d *DriverMock) GetInternalIP(zone, name string) (string, error) {
	d.GetInternalIPZone = zone
	d.GetInternalIPName = name
	return d.GetInternalIPResult, d.GetInternalIPErr
}

func (d *DriverMock) GetSerialPortOutput(zone, name string) (string, error) {
	d.GetSerialPortOutputZone = zone
	d.GetSerialPortOutputName = name
	return d.GetSerialPortOutputResult, d.GetSerialPortOutputErr
}

func (d *DriverMock) ImageExists(project, name string) bool {
	d.ImageExistsProjectId = project
	d.ImageExistsName = name
	return d.ImageExistsResult
}

func (d *DriverMock) RunInstance(c *InstanceConfig) (<-chan error, error) {
	d.RunInstanceConfig = c

	resultCh := d.RunInstanceErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh, d.RunInstanceErr
}

func (d *DriverMock) WaitForInstance(state, zone, name string) <-chan error {
	d.WaitForInstanceState = state
	d.WaitForInstanceZone = zone
	d.WaitForInstanceName = name

	resultCh := d.WaitForInstanceErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh
}

func (d *DriverMock) GetWindowsPassword() (string, error) {
	return "", nil
}

func (d *DriverMock) CreateOrResetWindowsPassword(instance, zone string, c *WindowsPasswordConfig) (<-chan error, error) {

	d.CreateOrResetWindowsPasswordInstance = instance
	d.CreateOrResetWindowsPasswordZone = zone
	d.CreateOrResetWindowsPasswordConfig = c

	c.Password = "MOCK_PASSWORD"

	resultCh := d.CreateOrResetWindowsPasswordErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh, d.CreateOrResetWindowsPasswordErr
}

func (d *DriverMock) ImportOSLoginSSHKey(user, key string) (*oslogin.LoginProfile, error) {
	account := oslogin.PosixAccount{Primary: true, Username: "testing_packer_io"}
	profile := oslogin.LoginProfile{
		PosixAccounts: []*oslogin.PosixAccount{&account},
	}
	return &profile, nil
}

func (d *DriverMock) DeleteOSLoginSSHKey(user, fingerprint string) error {
	return nil
}

func (d *DriverMock) AddToInstanceMetadata(zone string, name string, metadata map[string]string) error {
	d.AddToInstanceMetadataZone = zone
	d.AddToInstanceMetadataName = name
	d.AddToInstanceMetadataKVPairs = metadata

	resultCh := d.AddToInstanceMetadataErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
	}

	return nil
}

func (d *DriverMock) GetTokenInfo() (*oauth2_svc.Tokeninfo, error) {
	if d.GetTokenInfoResult == nil {
		d.GetTokenInfoErr = fmt.Errorf("no token found")
	}

	return d.GetTokenInfoResult, d.GetTokenInfoErr
}

func (d *DriverMock) UploadToBucket(bucket, object string, data io.Reader) (string, error) {
	d.UploadToBucketBucket = bucket
	d.UploadToBucketObjectName = object
	d.UploadToBucketData = data

	return d.UploadToBucketResult, d.UploadToBucketError
}
