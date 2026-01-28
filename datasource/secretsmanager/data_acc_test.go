// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:embed test-fixtures/template.pkr.hcl
var testTemplate string
var projectID = os.Getenv("GOOGLE_PROJECT_ID")

func TestAccGCPSecretManager(t *testing.T) {
	if projectID == "" {
		t.Skip("GOOGLE_PROJECT_ID must be set for acceptance tests")
	}

	cases := []struct {
		name       string
		secret     *GoogleSecret
		expect     string
		wantErr    bool
		skipCreate bool
	}{
		{
			name: "valid json value",
			secret: &GoogleSecret{
				Name:  "packer-secret-valid",
				Value: `{"foo":"bar"}`,
			},
			expect: "secret value: bar",
		},
		{
			name: "non-json value",
			secret: &GoogleSecret{
				Name:  "packer-secret-nonjson",
				Value: "some random string",
			},
			expect: "secret payload: some random string",
		},
		{
			name: "empty value",
			secret: &GoogleSecret{
				Name:  "packer-secret-empty",
				Value: `{"foo":""}`,
			},
			expect: "secret value:",
		},
		{
			name: "missing secret",
			secret: &GoogleSecret{
				Name: "packer-secret-does-not-exist",
			},
			wantErr:    true,
			expect:     "Secret Payload cannot be empty",
			skipCreate: true,
		},
	}

	for _, tc := range cases {

		extraArgs := []string{
			"-var", fmt.Sprintf("project_id=%s", projectID),
			"-var", fmt.Sprintf("secret_name=%s", tc.secret.Name),
		}

		// if value is not json, we need to set the key to empty string
		if tc.secret.Value != "" && !json.Valid([]byte(tc.secret.Value)) {
			extraArgs = append(extraArgs, "-var", "key=")
		}

		t.Run(tc.name, func(t *testing.T) {
			if !tc.wantErr {
				if err := tc.secret.Create(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				defer tc.secret.Delete()
			}

			testCase := &acctest.PluginTestCase{
				Name: "gcp_secretsmanager_" + tc.name,
				Setup: func() error {
					if tc.skipCreate {
						return nil
					}
					return tc.secret.Create()
				},
				Teardown: func() error {
					if tc.skipCreate {
						return nil
					}
					return tc.secret.Delete()
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
						if matched := regexp.MustCompile(fmt.Sprintf(`%s/versions/latest" not found`, tc.secret.Name)).MatchString(logsString); !matched {
							t.Errorf("Expected failure not found in logs")
						}
						return nil
					}

					if !regexp.MustCompile(regexp.QuoteMeta(tc.expect)).MatchString(logsString) {
						t.Errorf("Expected log not found: %s\nLogs: %s", tc.expect, logsString)
					}
					return nil
				},
			}

			acctest.TestPlugin(t, testCase)
		})
	}
}

// GoogleSecret represents a secret in Google Secret Manager.
type GoogleSecret struct {
	Name  string `mapstructure:"name" required:"true"`
	Value string `mapstructure:"value" required:"true"`
}

func (s *GoogleSecret) Create() error {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	// Try creating the secret
	secret, err := client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		SecretId: s.Name,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	})
	if err != nil && !isAlreadyExists(err) {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	secretName := fmt.Sprintf("projects/%s/secrets/%s", projectID, s.Name)
	if secret != nil {
		secretName = secret.Name
	}

	// Add secret version with payload
	_, err = client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(s.Value),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add secret version: %w", err)
	}

	return nil
}

func (s *GoogleSecret) Delete() error {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx, option.WithGRPCDialOption(grpc.WithDisableRetry()))
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	err = client.DeleteSecret(ctx, &secretmanagerpb.DeleteSecretRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s", projectID, s.Name),
	})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	log.Printf("Secret %q deleted successfully", s.Name)

	return nil
}

// Helpers

func isAlreadyExists(err error) bool {
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.AlreadyExists
}

func isNotFound(err error) bool {
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.NotFound
}
