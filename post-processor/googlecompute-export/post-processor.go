// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package googlecomputeexport

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-googlecompute/builder/googlecompute"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Authentication methods

	// A temporary [OAuth 2.0 access token](https://developers.google.com/identity/protocols/oauth2)
	// obtained from the Google Authorization server, i.e. the `Authorization: Bearer` token used to
	// authenticate HTTP requests to GCP APIs.
	// This is an alternative to `account_file`, and ignores the `scopes` field.
	// If both are specified, `access_token` will be used over the `account_file` field.
	//
	// These access tokens cannot be renewed by Packer and thus will only work until they expire.
	// If you anticipate Packer needing access for longer than a token's lifetime (default `1 hour`),
	// please use a service account key with `account_file` instead.
	AccessToken string `mapstructure:"access_token" required:"false"`
	// The JSON file containing your account credentials. Not required if you
	// run Packer on a GCE instance with a service account. Instructions for
	// creating the file or using service accounts are above.
	AccountFile string `mapstructure:"account_file" required:"false"`
	// The JSON file containing your account credentials.
	//
	// The file's contents may be anything supported by the Google Go client, i.e.:
	//
	// * Service account JSON
	// * OIDC-provided token for federation
	// * Gcloud user credentials file (refresh-token JSON)
	// * A Google Developers Console client_credentials.json
	CredentialsFile string `mapstructure:"credentials_file" required:"false"`
	// The raw JSON payload for credentials.
	//
	// The accepted data formats are same as those described under
	// [credentials_file](#credentials_file).
	CredentialsJSON string `mapstructure:"credentials_json" required:"false"`
	// This allows service account impersonation as per the [docs](https://cloud.google.com/iam/docs/impersonating-service-accounts).
	ImpersonateServiceAccount string `mapstructure:"impersonate_service_account" required:"false"`
	// Can be set instead of account_file. If set, this builder will use
	// HashiCorp Vault to generate an Oauth token for authenticating against
	// Google Cloud. The value should be the path of the token generator
	// within vault.
	// For information on how to configure your Vault + GCP engine to produce
	// Oauth tokens, see https://www.vaultproject.io/docs/auth/gcp
	// You must have the environment variables VAULT_ADDR and VAULT_TOKEN set,
	// along with any other relevant variables for accessing your vault
	// instance. For more information, see the Vault docs:
	// https://www.vaultproject.io/docs/commands/#environment-variables
	// Example:`"vault_gcp_oauth_engine": "gcp/token/my-project-editor",`
	VaultGCPOauthEngine string `mapstructure:"vault_gcp_oauth_engine"`
	credentials         *google.Credentials

	// The service account scopes for launched exporter post-processor instance.
	// Defaults to:
	//
	// ```json
	// [
	//   "https://www.googleapis.com/auth/cloud-platform"
	// ]
	// ```
	Scopes []string `mapstructure:"scopes" required:"false"`
	//The size of the export instances disk.
	//The disk is unused for the export but a larger size will increase `pd-ssd` read speed.
	//This defaults to `200`, which is 200GB.
	DiskSizeGb int64 `mapstructure:"disk_size"`
	//Type of disk used to back the export instance, like
	//`pd-ssd` or `pd-standard`. Defaults to `pd-ssd`.
	DiskType string `mapstructure:"disk_type"`
	//The export instance machine type. Defaults to `"n1-highcpu-4"`.
	MachineType string `mapstructure:"machine_type"`
	//The Google Compute network id or URL to use for the export instance.
	//Defaults to `"default"`. If the value is not a URL, it
	//will be interpolated to `projects/((builder_project_id))/global/networks/((network))`.
	//This value is not required if a `subnet` is specified.
	Network string `mapstructure:"network"`
	//A list of GCS paths where the image will be exported.
	//For example `'gs://mybucket/path/to/file.tar.gz'`
	Paths []string `mapstructure:"paths" required:"true"`
	//The Google Compute subnetwork id or URL to use for
	//the export instance. Only required if the `network` has been created with
	//custom subnetting. Note, the region of the subnetwork must match the
	//`zone` in which the VM is launched. If the value is not a URL,
	//it will be interpolated to
	//`projects/((builder_project_id))/regions/((region))/subnetworks/((subnetwork))`
	Subnetwork string `mapstructure:"subnetwork"`
	//The zone in which to launch the export instance. Defaults
	//to `googlecompute` builder zone. Example: `"us-central1-a"`
	Zone                string `mapstructure:"zone"`
	IAP                 bool   `mapstructure-to-hcl2:",skip"`
	ServiceAccountEmail string `mapstructure:"service_account_email"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
	runner multistep.Runner
}

func (p *PostProcessor) CheckAuth() error {
	return googlecompute.CheckAuth(
		p.config.AccessToken,
		p.config.AccountFile,
		p.config.CredentialsFile,
		p.config.CredentialsJSON,
		p.config.ImpersonateServiceAccount,
		p.config.VaultGCPOauthEngine,
	)
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
	}, raws...)
	if err != nil {
		return err
	}

	errs := new(packersdk.MultiError)

	if len(p.config.Paths) == 0 {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("paths must be specified"))
	}

	// Set defaults.
	if p.config.DiskSizeGb == 0 {
		p.config.DiskSizeGb = 200
	}

	if p.config.DiskType == "" {
		p.config.DiskType = "pd-ssd"
	}

	if p.config.MachineType == "" {
		p.config.MachineType = "n1-highcpu-4"
	}

	if p.config.Network == "" && p.config.Subnetwork == "" {
		p.config.Network = "default"
	}

	err = p.CheckAuth()
	if err != nil {
		errs = packersdk.MultiErrorAppend(errs, err)
	}

	// Authenticating via an account file
	if p.config.AccountFile != "" {
		log.Printf("account_file is deprecated, please use either credentials_json or credentials_file instead")
		// Heuristic, but should be good enough to discriminate between
		// the two somewhat reliably.
		if strings.HasPrefix(strings.TrimSpace(p.config.AccountFile), "{") {
			p.config.CredentialsJSON = p.config.AccountFile
		} else {
			p.config.CredentialsFile = p.config.AccountFile
		}
	}

	if p.config.CredentialsFile != "" {
		cfg, err := googlecompute.ProcessCredentialsFile(p.config.CredentialsFile)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
		p.config.credentials = cfg
	}

	if p.config.CredentialsJSON != "" {
		cfg, err := googlecompute.ProcessCredentials([]byte(p.config.CredentialsJSON))
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
		p.config.credentials = cfg
	}

	if len(p.config.Scopes) == 0 {
		p.config.Scopes = []string{
			storage.CloudPlatformScope,
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	switch artifact.BuilderId() {
	// TODO: uncomment when Packer core stops importing this plugin.
	// case googlecompute.BuilderId, artifice.BuilderId:
	case googlecompute.BuilderId, "packer.post-processor.artifice":
		break
	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only export from Google Compute Engine builder and Artifice post-processor artifacts.",
			artifact.BuilderId())
		return nil, false, false, err
	}

	builderImageName := artifact.State("ImageName").(string)
	builderProjectId := artifact.State("ProjectId").(string)
	builderZone := artifact.State("BuildZone").(string)

	ui.Say(fmt.Sprintf("Exporting image %v to destination: %v", builderImageName, p.config.Paths))

	if p.config.Zone == "" {
		p.config.Zone = builderZone
	}

	// Set up exporter instance configuration.
	exporterName := fmt.Sprintf("%s-exporter", artifact.Id())
	exporterMetadata := map[string]string{
		"image_name":     builderImageName,
		"name":           exporterName,
		"paths":          strings.Join(p.config.Paths, " "),
		"startup-script": StartupScript,
		"zone":           p.config.Zone,
		// Pre-fill the startup script status with "notdone" status
		googlecompute.StartupScriptStatusKey: googlecompute.StartupScriptStatusNotDone,
	}

	exporterConfig := googlecompute.Config{
		DiskName:             exporterName,
		DiskSizeGb:           p.config.DiskSizeGb,
		DiskType:             p.config.DiskType,
		InstanceName:         exporterName,
		MachineType:          p.config.MachineType,
		Metadata:             exporterMetadata,
		Network:              p.config.Network,
		NetworkProjectId:     builderProjectId,
		StateTimeout:         5 * time.Minute,
		SourceImageFamily:    "debian-9-worker",
		SourceImageProjectId: []string{"compute-image-tools"},
		Subnetwork:           p.config.Subnetwork,
		Zone:                 p.config.Zone,
		Scopes: []string{
			"https://www.googleapis.com/auth/compute",
			"https://www.googleapis.com/auth/devstorage.full_control",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/logging.write",
		},
	}
	if p.config.ServiceAccountEmail != "" {
		exporterConfig.ServiceAccountEmail = p.config.ServiceAccountEmail
	}
	cfg := googlecompute.GCEDriverConfig{
		Ui:                            ui,
		ProjectId:                     builderProjectId,
		AccessToken:                   p.config.AccessToken,
		ImpersonateServiceAccountName: p.config.ImpersonateServiceAccount,
		Scopes:                        p.config.Scopes,
		VaultOauthEngineName:          p.config.VaultGCPOauthEngine,
		Credentials:                   p.config.credentials,
	}

	driver, err := googlecompute.NewDriverGCE(cfg)
	if err != nil {
		ui.Error(fmt.Sprintf("Error creating GCE driver: %s", err.Error()))
		return nil, false, false, err
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &exporterConfig)
	state.Put("driver", driver)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		&communicator.StepSSHKeyGen{
			CommConf: &exporterConfig.Comm,
		},
		multistep.If(p.config.PackerDebug,
			&communicator.StepDumpSSHKey{
				Path: fmt.Sprintf("gce_%s.pem", p.config.PackerBuildName),
			},
		),
		&googlecompute.StepCreateInstance{
			Debug: p.config.PackerDebug,
		},
		new(googlecompute.StepWaitStartupScript),
		new(googlecompute.StepTeardownInstance),
	}

	// Run the steps.
	p.runner = commonsteps.NewRunner(steps, p.config.PackerConfig, ui)
	p.runner.Run(ctx, state)

	result := &Artifact{
		paths:     p.config.Paths,
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}

	return result, false, false, nil
}
