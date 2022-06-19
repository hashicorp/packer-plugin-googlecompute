//go:generate packer-sdc mapstructure-to-hcl2 -type Config,DatasourceOutput
package secretmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"

	"github.com/zclconf/go-cty/cty"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type Config struct {
	MockOption []interface{} `mapstructure:"mock" cty:"mock" hcl:"mock"`

	Project string `mapstructure:"project"`
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type Datasource struct {
	config Config
}

type DatasourceOutput struct {
	Payload  string `mapstructure:"payload"`
	Checksum int64  `mapstructure:"checksum"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
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
	if d.config.Project == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a 'project' must be specified"))
	}
	if d.config.Name == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a 'name' must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	opts := make([]option.ClientOption, len(d.config.MockOption))
	for i, opt := range d.config.MockOption {
		opts[i] = opt.(option.ClientOption)
	}

	client, err := secretmanager.NewClient(context.Background(), opts...)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	secret, err := client.AccessSecretVersion(context.Background(), &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", d.config.Project, d.config.Name, d.config.Version),
	})
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	output := DatasourceOutput{
		Payload:  secret.GetPayload().String(),
		Checksum: *secret.Payload.DataCrc32C,
	}
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
