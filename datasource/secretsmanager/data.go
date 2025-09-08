// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,DatasourceOutput

package secretsmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"

	"github.com/zclconf/go-cty/cty"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {

	// The Google Cloud project ID where the secret is stored.
	ProjectId string `mapstructure:"project_id" required:"true"`

	// The name of the secret in the secret manager.
	Name string `mapstructure:"name" required:"true"`

	// The key to extract from the secret payload.
	// If not provided, the entire payload will be returned.
	Key string `mapstructure:"key"`

	// The version of the secret to access. Defaults to "latest" if not specified.
	Version string `mapstructure:"version"`

	// Specify the GCP universe to deploy in. The default is "googleapis.com".
	UniverseDomain string `mapstructure:"universe_domain"`
	// Custom service endpoints, typically used to configure the Google provider to
	// communicate with GCP-like APIs such as the Cloud Functions emulator.
	//  Supported keys are `secretmanager`.
	//
	// Example:
	//   custom_endpoints = {
	//     secretmanager = "https://{your-endpoint}/"
	//   }
	//
	CustomEndpoints map[string]string `mapstructure:"custom_endpoints"`
}

type Datasource struct {
	config Config
}

type DatasourceOutput struct {
	// The raw string payload of the secret version.
	Payload string `mapstructure:"payload"`

	// The value extracted using the 'key', if provided.
	Value string `mapstructure:"value"`

	// The crc32c checksum for the payload.
	Checksum int64 `mapstructure:"checksum"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError
	if d.config.Version == "" {
		d.config.Version = "latest"
	}
	if d.config.ProjectId == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a 'project_id' must be specified"))
	}
	if d.config.Name == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a 'name' must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (d *Datasource) Execute() (cty.Value, error) {
	ctx := context.Background()

	var opts []option.ClientOption
	if d.config.UniverseDomain != "" {
		opts = append(opts, option.WithUniverseDomain(d.config.UniverseDomain))
	}
	if len(d.config.CustomEndpoints) > 0 {
		if endpoint, ok := d.config.CustomEndpoints["secretmanager"]; ok {
			opts = append(opts, option.WithEndpoint(endpoint))
		}
	}

	client, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to create secret manager client: %w", err)
	}
	defer client.Close()

	secretName := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", d.config.ProjectId, d.config.Name, d.config.Version)

	secret, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("secret %q not found", secretName)
		}
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error accessing secret: %w", err)
	}

	payload := secret.GetPayload()
	checksum := int64(0)
	if secret.Payload.DataCrc32C != nil {
		checksum = *payload.DataCrc32C
	}
	computedChecksum := crc32.Checksum(payload.Data, crc32.MakeTable(crc32.Castagnoli))

	if payload.DataCrc32C != nil && int64(computedChecksum) != checksum {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("data integrity check failed: expected crc32c %d but got %d", *payload.DataCrc32C, computedChecksum)
	}

	var value string
	if d.config.Key != "" {
		var payloadMap map[string]interface{}
		if err := json.Unmarshal(payload.GetData(), &payloadMap); err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to parse JSON payload for key extraction: %w", err)
		}

		val, ok := payloadMap[d.config.Key]
		if !ok {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("key %q not found in secret payload", d.config.Key)
		}

		value = fmt.Sprintf("%v", val)
	}

	output := DatasourceOutput{
		Payload:  string(payload.GetData()),
		Value:    value,
		Checksum: checksum,
	}
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
