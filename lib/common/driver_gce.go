// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	compute "google.golang.org/api/compute/v1"
	impersonate "google.golang.org/api/impersonate"
	oauth2_svc "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	oslogin "google.golang.org/api/oslogin/v1"
	"google.golang.org/api/storage/v1"

	"github.com/hashicorp/packer-plugin-googlecompute/version"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer-plugin-sdk/useragent"
	vaultapi "github.com/hashicorp/vault/api"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	InstanceTerminationActionStop   = "STOP"
	InstanceTerminationActionDelete = "DELETE"
)

// driverGCE is a Driver implementation that actually talks to GCE.
// Create an instance using NewDriverGCE.
type driverGCE struct {
	projectId      string
	service        *compute.Service
	osLoginService *oslogin.Service
	oauth2Service  *oauth2_svc.Service
	storageService *storage.Service
	ui             packersdk.Ui
}

type GCEDriverConfig struct {
	Ui                            packersdk.Ui
	ProjectId                     string
	ImpersonateServiceAccountName string
	Scopes                        []string
	AccessToken                   string
	VaultOauthEngineName          string
	Credentials                   *google.Credentials
}

var DriverScopes = []string{
	"https://www.googleapis.com/auth/compute",
	"https://www.googleapis.com/auth/devstorage.full_control",
	"https://www.googleapis.com/auth/userinfo.email",
}

// Define a TokenSource that gets tokens from Vault
type OauthTokenSource struct {
	Path string
}

func (ots OauthTokenSource) Token() (*oauth2.Token, error) {
	log.Printf("Retrieving Oauth token from Vault...")
	vaultConfig := vaultapi.DefaultConfig()
	cli, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("%s\n", err)
	}
	resp, err := cli.Logical().Read(ots.Path)
	if err != nil {
		return nil, fmt.Errorf("Error reading vault resp: %s", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("Vault Oauth Engine does not exist at the given path.")
	}
	token, ok := resp.Data["token"]
	if !ok {
		return nil, fmt.Errorf("ERROR, token was not present in response body")
	}
	at := token.(string)

	log.Printf("Retrieved Oauth token from Vault")
	return &oauth2.Token{
		AccessToken: at,
		Expiry:      time.Now().Add(time.Minute * time.Duration(60)),
	}, nil

}

func NewClientOptionGoogle(vaultOauth string, impersonatesa string, accessToken string, credentials *google.Credentials, scopes []string) ([]option.ClientOption, error) {
	var err error

	var opts []option.ClientOption

	if vaultOauth != "" {
		// Auth with Vault Oauth
		log.Printf("Using Vault to generate Oauth token.")
		ts := OauthTokenSource{vaultOauth}
		opts = append(opts, option.WithTokenSource(ts))

	} else if impersonatesa != "" {
		log.Printf("[INFO] Using Google Cloud impersonation mechanism")
		ts, err := impersonate.CredentialsTokenSource(context.Background(), impersonate.CredentialsConfig{
			TargetPrincipal: impersonatesa,
			Scopes:          scopes,
		})
		if err != nil {
			return nil, err
		}
		opts = append(opts, option.WithTokenSource(ts))
	} else if accessToken != "" {
		// Auth with static access token
		log.Printf("[INFO] Using static Google Access Token")
		token := &oauth2.Token{AccessToken: accessToken}
		ts := oauth2.StaticTokenSource(token)
		opts = append(opts, option.WithTokenSource(ts))
	} else if credentials != nil {
		// Auth with Credentials if provided
		log.Printf("[INFO] Requesting Google token via credentials...")
		log.Printf("[INFO]   -- Scopes: %s", DriverScopes)

		opts = append(opts, option.WithCredentials(credentials))
	} else {
		log.Printf("[INFO] Requesting Google token via GCE API Default Client Token Source...")
		scopes := append(DriverScopes, "https://www.googleapis.com/auth/cloud-platform")
		ts, err := google.DefaultTokenSource(context.TODO(), scopes...)
		if err != nil {
			return nil, err
		}
		opts = append(opts, option.WithTokenSource(ts))
		// The DefaultClient uses the DefaultTokenSource of the google lib.
		// The DefaultTokenSource uses the "Application Default Credentials"
		// It looks for credentials in the following places, preferring the first location found:
		// 1. A JSON file whose path is specified by the
		//    GOOGLE_APPLICATION_CREDENTIALS environment variable.
		// 2. A JSON file in a location known to the gcloud command-line tool.
		//    On Windows, this is %APPDATA%/gcloud/application_default_credentials.json.
		//    On other systems, $HOME/.config/gcloud/application_default_credentials.json.
		// 3. On Google App Engine it uses the appengine.AccessToken function.
		// 4. On Google Compute Engine and Google App Engine Managed VMs, it fetches
		//    credentials from the metadata server.
		//    (In this final case any provided scopes are ignored.)
	}

	// Insert UserAgents
	opts = append(opts, option.WithUserAgent(useragent.String(version.PluginVersion.FormattedVersion())))

	if err != nil {
		return nil, err
	}

	return opts, nil
}

func NewDriverGCE(config GCEDriverConfig) (Driver, error) {

	opts, err := NewClientOptionGoogle(config.VaultOauthEngineName, config.ImpersonateServiceAccountName, config.AccessToken, config.Credentials, config.Scopes)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Instantiating GCE client...")
	service, err := compute.NewService(context.TODO(), opts...)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Instantiating OS Login client...")
	osLoginService, err := oslogin.NewService(context.TODO(), opts...)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Instantiating Oauth2 client...")
	oauth2Service, err := oauth2_svc.NewService(context.TODO(), opts...)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Instantiating storage client...")
	storageService, err := storage.NewService(context.TODO(), opts...)
	if err != nil {
		return nil, err
	}

	return &driverGCE{
		projectId:      config.ProjectId,
		service:        service,
		osLoginService: osLoginService,
		oauth2Service:  oauth2Service,
		storageService: storageService,
		ui:             config.Ui,
	}, nil
}

func (d *driverGCE) CreateImage(project string, imageSpec *compute.Image) (<-chan *Image, <-chan error) {
	imageCh := make(chan *Image, 1)
	errCh := make(chan error, 1)
	op, err := d.service.Images.Insert(project, imageSpec).Do()
	if err != nil {
		errCh <- err
	} else {
		go func() {
			err = waitForState(errCh, "DONE", d.refreshGlobalOp(project, op))
			if err != nil {
				close(imageCh)
				errCh <- err
				return
			}
			var image *Image
			image, err = d.GetImageFromProject(project, imageSpec.Name, false)
			if err != nil {
				close(imageCh)
				errCh <- err
				return
			}
			imageCh <- image
			close(imageCh)
		}()
	}

	return imageCh, errCh
}

func (d *driverGCE) SetImageDeprecationStatus(project, name string, deprecationStatus *compute.DeprecationStatus) error {
	if deprecationStatus == nil {
		return errors.New("deprecationStatus cannot be nil")
	}
	_, err := d.service.Images.Deprecate(project, name, deprecationStatus).Do()
	return err
}

func (d *driverGCE) DeleteImage(project, name string) <-chan error {
	errCh := make(chan error, 1)
	op, err := d.service.Images.Delete(project, name).Do()
	if err != nil {
		errCh <- err
	} else {
		go func() {
			_ = waitForState(errCh, "DONE", d.refreshGlobalOp(project, op))
		}()

	}

	return errCh
}

func (d *driverGCE) DeleteInstance(zone, name string) (<-chan error, error) {
	op, err := d.service.Instances.Delete(d.projectId, zone, name).Do()
	if err != nil {
		return nil, err
	}

	errCh := make(chan error, 1)
	go func() {
		_ = waitForState(errCh, "DONE", d.refreshZoneOp(zone, op))
	}()
	return errCh, nil
}

func (d *driverGCE) CreateDisk(diskConfig BlockDevice) (<-chan *compute.Disk, <-chan error) {
	if len(diskConfig.ReplicaZones) != 0 {
		return d.createRegionalDisk(diskConfig)
	}

	return d.createZonalDisk(diskConfig)
}

func (d *driverGCE) createRegionalDisk(diskConfig BlockDevice) (<-chan *compute.Disk, <-chan error) {
	diskChan := make(chan *compute.Disk, 1)
	errChan := make(chan error, 1)

	computePayload, err := diskConfig.GenerateComputeDiskPayload()
	if err != nil {
		errChan <- err
		close(diskChan)
		close(errChan)
		return diskChan, errChan
	}

	region, _ := GetRegionFromZone(diskConfig.Zone)
	op, err := d.service.RegionDisks.Insert(d.projectId, region, computePayload).Do()
	if err != nil {
		errChan <- err
		close(diskChan)
		close(errChan)
		return diskChan, errChan
	}

	go func() {
		defer func() {
			close(errChan)
			close(diskChan)
		}()

		err := waitForState(errChan, "DONE", d.refreshRegionOp(region, op))
		if err != nil {
			errChan <- err
			return
		}
		disk, err := d.service.Disks.Get(d.projectId, region, diskConfig.DiskName).Do()
		if err != nil {
			errChan <- err
			return
		}

		diskChan <- disk
	}()
	return diskChan, errChan
}

func (d *driverGCE) createZonalDisk(diskConfig BlockDevice) (<-chan *compute.Disk, <-chan error) {
	diskChan := make(chan *compute.Disk, 1)
	errChan := make(chan error, 1)

	zone := diskConfig.Zone

	var op *compute.Operation
	var err error

	computePayload, err := diskConfig.GenerateComputeDiskPayload()
	if err != nil {
		errChan <- err
		close(diskChan)
		close(errChan)
		return diskChan, errChan
	}

	op, err = d.service.Disks.Insert(d.projectId, zone, computePayload).Do()
	if err != nil {
		errChan <- err
		close(diskChan)
		close(errChan)
		return diskChan, errChan
	}

	go func() {
		defer func() {
			close(errChan)
			close(diskChan)
		}()

		err := waitForState(errChan, "DONE", d.refreshZoneOp(zone, op))
		if err != nil {
			errChan <- err
			return
		}
		disk, err := d.service.Disks.Get(d.projectId, zone, diskConfig.DiskName).Do()
		if err != nil {
			errChan <- err
			return
		}

		diskChan <- disk
	}()
	return diskChan, errChan
}

func (d *driverGCE) DeleteDisk(zoneOrRegion, name string) <-chan error {
	if IsZoneARegion(zoneOrRegion) {
		return d.deleteRegionalDisk(zoneOrRegion, name)
	}

	return d.deleteZonalDisk(zoneOrRegion, name)
}

func (d *driverGCE) deleteZonalDisk(zone, name string) <-chan error {
	errCh := make(chan error, 1)

	op, err := d.service.Disks.Delete(d.projectId, zone, name).Do()
	if err != nil {
		errCh <- err
		close(errCh)
		return errCh
	}

	go func() {
		_ = waitForState(errCh, "DONE", d.refreshZoneOp(zone, op))
		close(errCh)
	}()
	return errCh
}

func (d *driverGCE) deleteRegionalDisk(region, name string) <-chan error {
	errCh := make(chan error, 1)

	op, err := d.service.RegionDisks.Delete(d.projectId, region, name).Do()
	if err != nil {
		errCh <- err
		close(errCh)
		return errCh
	}

	go func() {
		_ = waitForState(errCh, "DONE", d.refreshRegionOp(region, op))
		close(errCh)
	}()
	return errCh
}

func (d *driverGCE) GetDisk(zoneOrRegion, name string) (*compute.Disk, error) {
	if IsZoneARegion(zoneOrRegion) {
		return d.service.RegionDisks.Get(d.projectId, zoneOrRegion, name).Do()
	}

	return d.service.Disks.Get(d.projectId, zoneOrRegion, name).Do()
}

func (d *driverGCE) GetImage(name string, fromFamily bool) (*Image, error) {

	projects := []string{
		d.projectId,
		// Public projects, drawn from
		// https://cloud.google.com/compute/docs/images
		"almalinux-cloud",
		"centos-cloud",
		"cloud-hpc-image-public",
		"cos-cloud",
		"coreos-cloud",
		"debian-cloud",
		"fedora-cloud",
		"fedora-coreos-cloud",
		"freebsd-org-cloud-dev",
		"rhel-cloud",
		"rhel-sap-cloud",
		"rocky-linux-cloud",
		"suse-cloud",
		"suse-sap-cloud",
		"suse-byos-cloud",
		"ubuntu-os-cloud",
		"windows-cloud",
		"windows-sql-cloud",
		"gce-nvme",
		// misc
		"google-containers",
		"opensuse-cloud",
		"ubuntu-os-pro-cloud",
		"ml-images",
	}
	return d.GetImageFromProjects(projects, name, fromFamily)
}
func (d *driverGCE) GetImageFromProjects(projects []string, name string, fromFamily bool) (*Image, error) {
	var errs error
	for _, project := range projects {
		image, err := d.GetImageFromProject(project, name, fromFamily)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
		if image != nil {
			return image, nil
		}
	}

	return nil, fmt.Errorf(
		"Could not find image, %s, in projects, %s: %s", name,
		projects, errs)
}

func (d *driverGCE) GetImageFromProject(project, name string, fromFamily bool) (*Image, error) {
	var (
		image *compute.Image
		err   error
	)

	if fromFamily {
		image, err = d.service.Images.GetFromFamily(project, name).Do()
	} else {
		image, err = d.service.Images.Get(project, name).Do()
	}

	if err != nil {
		return nil, err
	} else if image == nil || image.SelfLink == "" {
		return nil, fmt.Errorf("Image, %s, could not be found in project: %s", name, project)
	} else {
		return &Image{
			Architecture:    image.Architecture,
			GuestOsFeatures: image.GuestOsFeatures,
			Licenses:        image.Licenses,
			Name:            image.Name,
			ProjectId:       project,
			SelfLink:        image.SelfLink,
			SizeGb:          image.DiskSizeGb,
		}, nil
	}
}

func (d *driverGCE) GetProjectMetadata(zone, key string) (string, error) {
	project, err := d.service.Projects.Get(d.projectId).Do()
	if err != nil {
		return "", err
	}

	for _, item := range project.CommonInstanceMetadata.Items {
		if item.Key == key {
			return *item.Value, nil
		}
	}

	return "", fmt.Errorf("Project metadata key, %s, not found.", key)
}

func (d *driverGCE) GetInstanceMetadata(zone, name, key string) (string, error) {
	instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
	if err != nil {
		return "", err
	}

	for _, item := range instance.Metadata.Items {
		if item.Key == key {
			return *item.Value, nil
		}
	}

	return "", fmt.Errorf("Instance metadata key, %s, not found.", key)
}

func (d *driverGCE) GetNatIP(zone, name string) (string, error) {
	instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
	if err != nil {
		return "", err
	}

	for _, ni := range instance.NetworkInterfaces {
		if ni.AccessConfigs == nil {
			continue
		}
		for _, ac := range ni.AccessConfigs {
			if ac.NatIP != "" {
				return ac.NatIP, nil
			}
		}
	}

	return "", nil
}

func (d *driverGCE) GetInternalIP(zone, name string) (string, error) {
	instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
	if err != nil {
		return "", err
	}

	for _, ni := range instance.NetworkInterfaces {
		if ni.NetworkIP == "" {
			continue
		}
		return ni.NetworkIP, nil
	}

	return "", nil
}

func (d *driverGCE) GetSerialPortOutput(zone, name string) (string, error) {
	output, err := d.service.Instances.GetSerialPortOutput(d.projectId, zone, name).Do()
	if err != nil {
		return "", err
	}

	return output.Contents, nil
}

func (d *driverGCE) ImageExists(project, name string) bool {
	_, err := d.GetImageFromProject(project, name, false)
	// The API may return an error for reasons other than the image not
	// existing, but this heuristic is sufficient for now.
	return err == nil
}

func (d *driverGCE) RunInstance(c *InstanceConfig) (<-chan error, error) {
	// Get the zone
	d.ui.Message(fmt.Sprintf("Loading zone: %s", c.Zone))
	zone, err := d.service.Zones.Get(d.projectId, c.Zone).Do()
	if err != nil {
		return nil, err
	}

	// Get the machine type
	d.ui.Message(fmt.Sprintf("Loading machine type: %s", c.MachineType))
	machineType, err := d.service.MachineTypes.Get(
		d.projectId, zone.Name, c.MachineType).Do()
	if err != nil {
		return nil, err
	}
	// TODO(mitchellh): deprecation warnings

	networkId, subnetworkId, err := GetNetworking(c)
	if err != nil {
		return nil, err
	}

	var accessconfig *compute.AccessConfig
	// Use external IP if OmitExternalIP isn't set
	if !c.OmitExternalIP {
		accessconfig = &compute.AccessConfig{
			Name: "AccessConfig created by Packer",
			Type: "ONE_TO_ONE_NAT",
		}

		// If given a static IP, use it
		if c.Address != "" {
			region_url := strings.Split(zone.Region, "/")
			region := region_url[len(region_url)-1]
			address, err := d.service.Addresses.Get(d.projectId, region, c.Address).Do()
			if err != nil {
				return nil, err
			}
			accessconfig.NatIP = address.Address
		}
	}

	// Build up the metadata
	metadata := make([]*compute.MetadataItems, 0, len(c.Metadata))
	for k, v := range c.Metadata {
		vCopy := v
		metadata = append(metadata, &compute.MetadataItems{
			Key:   k,
			Value: &vCopy,
		})
	}

	var guestAccelerators []*compute.AcceleratorConfig
	if c.AcceleratorCount > 0 {
		ac := &compute.AcceleratorConfig{
			AcceleratorCount: c.AcceleratorCount,
			AcceleratorType:  c.AcceleratorType,
		}
		guestAccelerators = append(guestAccelerators, ac)
	}

	// Configure the instance's service account. If the user has set
	// disable_default_service_account, then the default service account
	// will not be used. If they also do not set service_account_email, then
	// the instance will be created with no service account or scopes.
	serviceAccount := &compute.ServiceAccount{}
	if !c.DisableDefaultServiceAccount {
		serviceAccount.Email = "default"
		serviceAccount.Scopes = c.Scopes
	}
	if c.ServiceAccountEmail != "" {
		serviceAccount.Email = c.ServiceAccountEmail
		serviceAccount.Scopes = c.Scopes
	}

	var diskEncryptionKey *compute.CustomerEncryptionKey
	if c.DiskEncryptionKey != nil {
		log.Printf("[DEBUG] using customer-managed encryption key for boot disk, KmsKeyName=%s, RawKey=%s",
			c.DiskEncryptionKey.KmsKeyName, c.DiskEncryptionKey.RawKey)
		diskEncryptionKey = c.DiskEncryptionKey.ComputeType()
	} else {
		log.Printf("[DEBUG] using google-managed encryption key for boot disk")
	}

	computeDisks := []*compute.AttachedDisk{
		{
			Type:              "PERSISTENT",
			Mode:              "READ_WRITE",
			Kind:              "compute#attachedDisk",
			Boot:              true,
			AutoDelete:        false,
			DiskEncryptionKey: diskEncryptionKey,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				SourceImage: c.Image.SelfLink,
				DiskName:    c.DiskName,
				DiskSizeGb:  c.DiskSizeGb,
				DiskType:    fmt.Sprintf("zones/%s/diskTypes/%s", zone.Name, c.DiskType),
			},
		},
	}

	for _, disk := range c.ExtraBlockDevices {
		computeDisks = append(computeDisks, disk.GenerateDiskAttachment())
	}

	// Create the instance information
	instance := compute.Instance{
		AdvancedMachineFeatures: &compute.AdvancedMachineFeatures{
			EnableNestedVirtualization: c.EnableNestedVirtualization,
		},
		Description:       c.Description,
		Disks:             computeDisks,
		GuestAccelerators: guestAccelerators,
		Labels:            c.Labels,
		MachineType:       machineType.SelfLink,
		Metadata: &compute.Metadata{
			Items: metadata,
		},
		MinCpuPlatform: c.MinCpuPlatform,
		Name:           c.Name,
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{accessconfig},
				Network:       networkId,
				Subnetwork:    subnetworkId,
				NetworkIP:     c.NetworkIP,
			},
		},
		Scheduling: &compute.Scheduling{
			OnHostMaintenance: c.OnHostMaintenance,
			Preemptible:       c.Preemptible,
		},
		ServiceAccounts: []*compute.ServiceAccount{
			serviceAccount,
		},
		Tags: &compute.Tags{
			Items: c.Tags,
		},
	}

	if c.MaxRunDurationInSeconds > 0 {
		log.Printf("[DEBUG] setting max run duration to %d seconds", c.MaxRunDurationInSeconds)
		instance.Scheduling.MaxRunDuration = &compute.Duration{
			Seconds: c.MaxRunDurationInSeconds,
		}
		log.Printf("[DEBUG] setting instance termination action to %s", c.InstanceTerminationAction)
		instance.Scheduling.InstanceTerminationAction = c.InstanceTerminationAction
	}

	// Shielded VMs configuration. If the user has set at least one of the
	// options, the shielded VM configuration will reflect that. If they
	// don't set any of the options the settings will default to the ones
	// of the source compute image which is used for creating the virtual
	// machine.
	shieldedInstanceConfig := &compute.ShieldedInstanceConfig{
		EnableSecureBoot:          c.EnableSecureBoot,
		EnableVtpm:                c.EnableVtpm,
		EnableIntegrityMonitoring: c.EnableIntegrityMonitoring,
	}
	shieldedUiMessage := ""
	if c.EnableSecureBoot || c.EnableVtpm || c.EnableIntegrityMonitoring {
		instance.ShieldedInstanceConfig = shieldedInstanceConfig
		shieldedUiMessage = " Shielded VM"
	}

	// Node affinity configuration. For example, if you want to build on sole
	// tenancy nodes.
	if len(c.NodeAffinities) > 0 {
		instance.Scheduling.NodeAffinities = make([]*compute.SchedulingNodeAffinity, 0, len(c.NodeAffinities))
		for _, nodeAffinity := range c.NodeAffinities {
			instance.Scheduling.NodeAffinities = append(instance.Scheduling.NodeAffinities, nodeAffinity.ComputeType())
		}
	}

	d.ui.Message(fmt.Sprintf("Requesting%s instance creation...", shieldedUiMessage))
	op, err := d.service.Instances.Insert(d.projectId, zone.Name, &instance).Do()
	if err != nil {
		return nil, err
	}

	errCh := make(chan error, 1)
	go func() {
		_ = waitForState(errCh, "DONE", d.refreshZoneOp(zone.Name, op))
	}()
	return errCh, nil
}

func (d *driverGCE) CreateOrResetWindowsPassword(instance, zone string, c *WindowsPasswordConfig) (<-chan error, error) {

	errCh := make(chan error, 1)
	go d.createWindowsPassword(errCh, instance, zone, c)

	return errCh, nil
}

func (d *driverGCE) createWindowsPassword(errCh chan<- error, name, zone string, c *WindowsPasswordConfig) {

	data, err := json.Marshal(c)

	if err != nil {
		errCh <- err
		return
	}
	dCopy := string(data)

	instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
	if err != nil {
		errCh <- err
		return
	}
	instance.Metadata.Items = append(instance.Metadata.Items, &compute.MetadataItems{Key: "windows-keys", Value: &dCopy})

	op, err := d.service.Instances.SetMetadata(d.projectId, zone, name, &compute.Metadata{
		Fingerprint: instance.Metadata.Fingerprint,
		Items:       instance.Metadata.Items,
	}).Do()

	if err != nil {
		errCh <- err
		return
	}

	newErrCh := make(chan error, 1)
	go func() {
		_ = waitForState(newErrCh, "DONE", d.refreshZoneOp(zone, op))
	}()

	select {
	case err = <-newErrCh:
	case <-time.After(time.Second * 30):
		err = errors.New("time out while waiting for instance to create")
	}

	if err != nil {
		errCh <- err
		return
	}

	timeout := time.Now().Add(c.WindowsPasswordTimeout)
	hash := sha1.New()
	random := rand.Reader

	for time.Now().Before(timeout) {
		if passwordResponses, err := d.getPasswordResponses(zone, name); err == nil {
			for _, response := range passwordResponses {
				if response.Modulus == c.Modulus {

					decodedPassword, err := base64.StdEncoding.DecodeString(response.EncryptedPassword)

					if err != nil {
						errCh <- err
						return
					}
					password, err := rsa.DecryptOAEP(hash, random, c.Key, decodedPassword, nil)

					if err != nil {
						errCh <- err
						return
					}

					c.Password = string(password)
					errCh <- nil
					return
				}
			}
		}

		time.Sleep(2 * time.Second)
	}
	err = errors.New("Could not retrieve password. Timed out.")

	errCh <- err
}

func (d *driverGCE) getPasswordResponses(zone, instance string) ([]windowsPasswordResponse, error) {
	output, err := d.service.Instances.GetSerialPortOutput(d.projectId, zone, instance).Port(4).Do()

	if err != nil {
		return nil, err
	}

	responses := strings.Split(output.Contents, "\n")

	passwordResponses := make([]windowsPasswordResponse, 0, len(responses))

	for _, response := range responses {
		var passwordResponse windowsPasswordResponse
		if err := json.Unmarshal([]byte(response), &passwordResponse); err != nil {
			continue
		}

		passwordResponses = append(passwordResponses, passwordResponse)
	}

	return passwordResponses, nil
}

func (d *driverGCE) ImportOSLoginSSHKey(user, sshPublicKey string) (*oslogin.LoginProfile, error) {
	parent := fmt.Sprintf("users/%s", user)

	resp, err := d.osLoginService.Users.ImportSshPublicKey(parent, &oslogin.SshPublicKey{
		Key: sshPublicKey,
	}).Do()
	if err != nil {
		return nil, err
	}

	return resp.LoginProfile, nil
}

func (d *driverGCE) DeleteOSLoginSSHKey(user, fingerprint string) error {
	name := fmt.Sprintf("users/%s/sshPublicKeys/%s", user, fingerprint)
	_, err := d.osLoginService.Users.SshPublicKeys.Delete(name).Do()
	if err != nil {
		return err
	}

	return nil
}

func (d *driverGCE) WaitForInstance(state, zone, name string) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		_ = waitForState(errCh, state, d.refreshInstanceState(zone, name))
	}()
	return errCh
}

func (d *driverGCE) refreshInstanceState(zone, name string) stateRefreshFunc {
	return func() (string, error) {
		instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
		if err != nil {
			return "", err
		}
		return instance.Status, nil
	}
}

func (d *driverGCE) refreshGlobalOp(project string, op *compute.Operation) stateRefreshFunc {
	return func() (string, error) {
		newOp, err := d.service.GlobalOperations.Get(project, op.Name).Do()
		if err != nil {
			return "", err
		}

		// If the op is done, check for errors
		err = nil
		if newOp.Status == "DONE" {
			if newOp.Error != nil {
				for _, e := range newOp.Error.Errors {
					err = packersdk.MultiErrorAppend(err, errors.New(e.Message))
				}
			}
		}

		return newOp.Status, err
	}
}

func (d *driverGCE) refreshZoneOp(zone string, op *compute.Operation) stateRefreshFunc {
	return func() (string, error) {
		newOp, err := d.service.ZoneOperations.Get(d.projectId, zone, op.Name).Do()
		if err != nil {
			return "", err
		}

		// If the op is done, check for errors
		err = nil
		if newOp.Status == "DONE" {
			if newOp.Error != nil {
				for _, e := range newOp.Error.Errors {
					err = packersdk.MultiErrorAppend(err, errors.New(e.Message))
				}
			}
		}

		return newOp.Status, err
	}
}

func (d *driverGCE) refreshRegionOp(region string, op *compute.Operation) stateRefreshFunc {
	return func() (string, error) {
		newOp, err := d.service.RegionOperations.Get(d.projectId, region, op.Name).Do()
		if err != nil {
			return "", err
		}

		// If the op is done, check for errors
		err = nil
		if newOp.Status == "DONE" {
			if newOp.Error != nil {
				for _, e := range newOp.Error.Errors {
					err = packersdk.MultiErrorAppend(err, errors.New(e.Message))
				}
			}
		}

		return newOp.Status, err
	}
}

// used in conjunction with waitForState.
type stateRefreshFunc func() (string, error)

// waitForState will spin in a loop forever waiting for state to
// reach a certain target.
func waitForState(errCh chan<- error, target string, refresh stateRefreshFunc) error {
	ctx := context.TODO()
	err := retry.Config{
		RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 2 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		state, err := refresh()
		if err != nil {
			// If there is an error, but the state is done, the function will spin
			// until timeout. Instead, emit the error immediately and return.
			if state == target {
				errCh <- err
				return nil
			}
			return err
		}
		if state == target {
			return nil
		}
		return fmt.Errorf("retrying for state %s, got %s", target, state)
	})
	errCh <- err
	return err
}

func (d *driverGCE) AddToInstanceMetadata(zone string, name string, metadata map[string]string) error {

	instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
	if err != nil {
		return err
	}

	// Build up the metadata
	metadataForInstance := make([]*compute.MetadataItems, 0, len(metadata))
	for k, v := range metadata {
		vCopy := v
		metadataForInstance = append(metadataForInstance, &compute.MetadataItems{
			Key:   k,
			Value: &vCopy,
		})
	}

	instance.Metadata.Items = append(instance.Metadata.Items, metadataForInstance...)

	op, err := d.service.Instances.SetMetadata(d.projectId, zone, name, &compute.Metadata{
		Fingerprint: instance.Metadata.Fingerprint,
		Items:       instance.Metadata.Items,
	}).Do()

	if err != nil {
		return err
	}

	newErrCh := make(chan error, 1)

	go func() {
		err = waitForState(newErrCh, "DONE", d.refreshZoneOp(zone, op))

		select {
		case err = <-newErrCh:
		case <-time.After(time.Second * 30):
			err = errors.New("time out while waiting for instance to create")
		}
	}()

	if err != nil {
		newErrCh <- err
		return err
	}

	return nil
}

// GetTokenInfo gets the information about the token used for authentication
func (d *driverGCE) GetTokenInfo() (*oauth2_svc.Tokeninfo, error) {
	return d.oauth2Service.Tokeninfo().Do()
}

func (d *driverGCE) UploadToBucket(bucket, objectName string, data io.Reader) (string, error) {
	storageObject, err := d.storageService.Objects.Insert(bucket, &storage.Object{Name: objectName}).Media(data).Do()
	if err != nil {
		return "", err
	}

	return storageObject.SelfLink, nil
}

func (d *driverGCE) DeleteFromBucket(bucket, objectName string) error {
	return d.storageService.Objects.Delete(bucket, objectName).Do()
}
