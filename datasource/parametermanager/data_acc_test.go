// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package parametermanager

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"

	parametermanager "cloud.google.com/go/parametermanager/apiv1"
	parametermanagerpb "cloud.google.com/go/parametermanager/apiv1/parametermanagerpb"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

//go:embed test-fixtures/template.pkr.hcl
var testTemplate string
var projectID = os.Getenv("GOOGLE_PROJECT_ID")

func TestAccGCPParameterManager(t *testing.T) {
	if projectID == "" {
		t.Skip("GOOGLE_PROJECT_ID must be set for acceptance tests")
	}

	cases := []struct {
		name            string
		param           *GoogleParam
		expectedOutputs []string
		wantErr         bool
		parsePayload    bool
		KeyName         string
	}{
		{
			name: "access value based on key when type is json",
			param: &GoogleParam{
				ParameterId: "packer-test-parameter",
				Format:      parametermanagerpb.ParameterFormat_JSON,
				Payload:     `{"project_id": "test_id_1","foo":"bar"}`,
				Version:     "1",
				Locations:   []string{"global", "us-east1"},
			},
			expectedOutputs: []string{"Parameter value using key is bar", "Project id parsed from JSON is test_id_1"},
			parsePayload:    true,
			KeyName:         "foo",
		},
		{
			name: "access value based on key when type is yaml",
			param: &GoogleParam{
				ParameterId: "packer-test-parameter",
				Format:      parametermanagerpb.ParameterFormat_YAML,
				Payload: `project_id: test_id_1
foo: bar`,
				Version:   "1",
				Locations: []string{"global", "us-east1"},
			},
			expectedOutputs: []string{"Parameter value using key is bar", "Project id parsed from YAML is test_id_1"},
			parsePayload:    true,
			KeyName:         "foo",
		},
		{
			name: "access value based on key when type is unformatted",
			param: &GoogleParam{
				ParameterId: "packer-test-parameter",
				Format:      parametermanagerpb.ParameterFormat_UNFORMATTED,
				Payload:     `"project_id":"test_id_1"`,
				Version:     "1",
				Locations:   []string{"global", "us-east1"},
			},
			wantErr: true,
			KeyName: "foo",
		},
		{
			name: "access unformatted payload",
			param: &GoogleParam{
				ParameterId: "packer-test-parameter",
				Format:      parametermanagerpb.ParameterFormat_UNFORMATTED,
				Payload:     `"project_id":"test_id_1"`,
				Version:     "1",
				Locations:   []string{"global", "us-east1"},
			},
			KeyName:         "",
			expectedOutputs: []string{"Unformatted Payload is \"project_id\":\"test_id_1\""},
		},
	}

	for _, tc := range cases {

		for _, location := range tc.param.Locations {
			extraArgs := []string{
				"-var", fmt.Sprintf("project_id=%s", projectID),
				"-var", fmt.Sprintf("format=%s", tc.param.Format.String()),
				"-var", fmt.Sprintf("parse_payload=%t", tc.parsePayload),
				"-var", fmt.Sprintf("key=%s", tc.KeyName),
				"-var", fmt.Sprintf("location=%s", location),
			}

			tc.name = fmt.Sprintf("%s_%s", tc.name, location)
			t.Run(tc.name, func(t *testing.T) {

				testCase := &acctest.PluginTestCase{
					Name: "gcp_paramsmanager_" + tc.name,
					Setup: func() error {
						return tc.param.Create(location)
					},
					Teardown: func() error {
						return tc.param.Delete(location)
					},
					Template:       testTemplate,
					BuildExtraArgs: extraArgs,
					Check: func(cmd *exec.Cmd, logFile string) error {
						logs, err := os.ReadFile(logFile)
						if err != nil {
							return fmt.Errorf("failed to read log file: %w", err)
						}
						logsString := string(logs)

						if tc.wantErr {
							if matched := regexp.MustCompile(`key extraction is supported only for JSON and YAML payload`).MatchString(logsString); !matched {
								t.Errorf("Expected failure not found in logs")
							}
							return nil
						}

						for _, expected := range tc.expectedOutputs {
							if !regexp.MustCompile(regexp.QuoteMeta(expected)).MatchString(logsString) {
								t.Errorf("Expected log not found: %s\nLogs: %s", expected, logsString)
							}
						}
						return nil
					},
				}

				acctest.TestPlugin(t, testCase)
			})
		}

	}
}

// GoogleSecret represents a secret in Google Secret Manager.
type GoogleParam struct {
	ParameterId string                             `mapstructure:"parameterid" required:"true"`
	Format      parametermanagerpb.ParameterFormat `mapstructure:"format" required:"true"`
	Payload     string                             `mapstructure:"payload" required:"true"`
	Version     string                             `mapstructure:"version" required:"true"`
	Locations   []string                           `mapstructure:"location" required:"true"`
}

func (p *GoogleParam) Create(location string) error {

	ctx := context.Background()

	// Create a client
	var client *parametermanager.Client
	var err error

	if location == "global" {
		client, err = parametermanager.NewClient(ctx)
	} else {
		endpoint := fmt.Sprintf("parametermanager.%s.rep.googleapis.com:443", location)
		client, err = parametermanager.NewClient(ctx, option.WithEndpoint(endpoint))
	}

	if err != nil {
		return fmt.Errorf("failed to create Parameter Manager client: %w", err)
	}
	defer client.Close()

	// Construct the name of the create parameter.
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)

	// Build the request to create a new parameter with the specified format.
	req := &parametermanagerpb.CreateParameterRequest{
		Parent:      parent,
		ParameterId: p.ParameterId,
		Parameter: &parametermanagerpb.Parameter{
			Format: p.Format,
		},
	}

	// Call the API to create the parameter
	parameter, err := client.CreateParameter(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create parameter: %w", err)
	}

	fmt.Printf("Created parameter %s with format %s\n", parameter.Name, parameter.Format.String())

	// Construct the name of the create parameter version.
	paramparent := fmt.Sprintf("projects/%s/locations/%s/parameters/%s", projectID, location, p.ParameterId)

	// Build the request to create a new parameter version with the unformatted payload.
	paramreq := &parametermanagerpb.CreateParameterVersionRequest{
		Parent:             paramparent,
		ParameterVersionId: p.Version,
		ParameterVersion: &parametermanagerpb.ParameterVersion{
			Payload: &parametermanagerpb.ParameterVersionPayload{
				Data: []byte(p.Payload),
			},
		},
	}

	// Call the API to create the parameter version.
	version, err := client.CreateParameterVersion(ctx, paramreq)
	if err != nil {
		return fmt.Errorf("failed to create parameter version: %w", err)
	}

	fmt.Printf("Created parameter version: %s\n", version.Name)
	return nil
}

func (p *GoogleParam) DeleteParam(location string) error {
	ctx := context.Background()
	// Create a client
	var client *parametermanager.Client
	var err error

	if location == "global" {
		client, err = parametermanager.NewClient(ctx)
	} else {
		endpoint := fmt.Sprintf("parametermanager.%s.rep.googleapis.com:443", location)
		client, err = parametermanager.NewClient(ctx, option.WithEndpoint(endpoint))
	}

	if err != nil {
		return fmt.Errorf("failed to create Parameter Manager client: %w", err)
	}
	defer client.Close()

	// Construct the name of the parameter to delete.
	name := fmt.Sprintf("projects/%s/locations/%s/parameters/%s", projectID, location, p.ParameterId)

	// Build the request to delete the parameter.
	req := &parametermanagerpb.DeleteParameterRequest{
		Name: name,
	}

	// Call the API to delete the parameter.
	err = client.DeleteParameter(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete parameter: %w", err)
	}

	fmt.Printf("Deleted parameter: %s\n", name)
	return nil
}

func (p *GoogleParam) Delete(location string) error {
	// Create a context and a Parameter Manager client.
	ctx := context.Background()
	// Create a client
	var client *parametermanager.Client
	var err error

	if location == "global" {
		client, err = parametermanager.NewClient(ctx)
	} else {
		endpoint := fmt.Sprintf("parametermanager.%s.rep.googleapis.com:443", location)
		client, err = parametermanager.NewClient(ctx, option.WithEndpoint(endpoint))
	}

	if err != nil {
		return fmt.Errorf("failed to create Parameter Manager client: %w", err)
	}
	defer client.Close()

	// Construct the name of the list parameter.
	parent := fmt.Sprintf("projects/%s/locations/%s/parameters/%s", projectID, location, p.ParameterId)

	// Build the request to list parameter versions.
	req := &parametermanagerpb.ListParameterVersionsRequest{
		Parent: parent,
	}

	// Call the API to list parameter versions.
	parameterVersions := client.ListParameterVersions(ctx, req)
	for {
		version, err := parameterVersions.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list parameter versions: %w", err)
		}

		fmt.Printf("Found parameter version %s with disabled state in %v\n", version.Name, version.Disabled)
		deletereq := &parametermanagerpb.DeleteParameterVersionRequest{
			Name: version.Name,
		}
		if err := client.DeleteParameterVersion(ctx, deletereq); err != nil {
			return fmt.Errorf("failed to delete parameter version: %w", err)
		}
		fmt.Printf("Deleted parameter version: %s\n", version.Name)
	}

	return p.DeleteParam(location)
}
