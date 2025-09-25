//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,DatasourceOutput

package parametermanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/stretchr/testify/assert/yaml"

	"github.com/zclconf/go-cty/cty"

	parametermanager "cloud.google.com/go/parametermanager/apiv1"
	parametermanagerpb "cloud.google.com/go/parametermanager/apiv1/parametermanagerpb"

	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config is the configuration structure for the data source.
type Config struct {
	// The Google Cloud project ID where the parameter is stored.
	ProjectId string `mapstructure:"project_id" required:"true"`

	// The location in which parameter is stored. Defaults to "global" if not specified.
	Location string `mapstructure:"location"`

	// The name of the parameter within the Parameter Manager.
	Name string `mapstructure:"name" required:"true"`

	// The version of the parameter within the Parameter Manager.
	Version string `mapstructure:"version" required:"true"`

	// A specific key to extract from the parameter payload if it's a JSON or YAML object.
	Key string `mapstructure:"key"`
}

// Datasource is the data source implementation.
type Datasource struct {
	config Config
}

// DatasourceOutput contains the data returned by the data source.
type DatasourceOutput struct {
	// The raw string payload of the parameter version.
	Payload string `mapstructure:"payload"`

	// The value extracted using the 'key', if provided.
	Value string `mapstructure:"value"`
}

// ConfigSpec returns the HCL object spec for the data source configuration.
func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

// OutputSpec returns the HCL object spec for the data source output.
func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

// Configure sets up the data source configuration.
func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError

	if d.config.ProjectId == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a 'project_id' must be specified"))
	}

	if d.config.Location == "" {
		d.config.Location = "global"
	}

	if d.config.Name == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a 'name' must be specified"))
	}

	if d.config.Version == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a 'version' must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

// Execute fetches the parameter from Google Cloud Parameter Manager.
func (d *Datasource) Execute() (cty.Value, error) {
	ctx := context.Background()

	// Create a client
	var client *parametermanager.Client
	var err error

	if d.config.Location == "global" {
		client, err = parametermanager.NewClient(ctx)
	} else {
		endpoint := fmt.Sprintf("parametermanager.%s.rep.googleapis.com:443", d.config.Location)
		client, err = parametermanager.NewClient(ctx, option.WithEndpoint(endpoint))
	}

	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to create parameter manager client: %w", err)
	}
	defer client.Close()

	// Build the request to get parameter format
	parameterFormatURL := fmt.Sprintf("projects/%s/locations/%s/parameters/%s", d.config.ProjectId, d.config.Location, d.config.Name)

	// Build the request to get parameter format
	parameterFormatReq := &parametermanagerpb.GetParameterRequest{
		Name: parameterFormatURL,
	}

	// Call the API to get parameter.
	paramFormatResp, err := client.GetParameter(ctx, parameterFormatReq)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to identify parameter format %w", err)
	}

	paramFormat := paramFormatResp.Format

	parameterPayloadURL := fmt.Sprintf("projects/%s/locations/%s/parameters/%s/versions/%s", d.config.ProjectId, d.config.Location, d.config.Name, d.config.Version)

	// Build the request to get parameter payload
	parameterPayloadReq := &parametermanagerpb.RenderParameterVersionRequest{
		Name: parameterPayloadURL,
	}

	// Call the API to render a parameter version.
	parameterPayloadResp, err := client.RenderParameterVersion(ctx, parameterPayloadReq)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("value %q not found", parameterPayloadResp)
		}
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error rendering parameter: %w", err)
	}

	payload := parameterPayloadResp.RenderedPayload

	// Extract value from payload if key is provided
	var value string
	var payloadMap map[string]interface{}
	if d.config.Key != "" {
		switch paramFormat {
		case parametermanagerpb.ParameterFormat_JSON:
			jsonErr := json.Unmarshal(payload, &payloadMap)
			if jsonErr != nil {
				return cty.NullVal(cty.EmptyObject), fmt.Errorf("error unmarshaling JSON payload: %w", jsonErr)
			}

		case parametermanagerpb.ParameterFormat_YAML:
			yamlErr := yaml.Unmarshal(payload, &payloadMap)
			if yamlErr != nil {
				return cty.NullVal(cty.EmptyObject), fmt.Errorf("error unmarshaling YAML payload: %w", yamlErr)
			}

		default:
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("key extraction is supported only for JSON and YAML payload")
		}

		val, ok := payloadMap[d.config.Key]
		if !ok {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("key %q not found in payload", d.config.Key)
		}
		value = fmt.Sprintf("%v", val)
	}

	output := DatasourceOutput{
		Payload: string(payload),
		Value:   value,
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
