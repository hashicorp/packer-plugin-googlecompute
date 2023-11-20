// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Authentication

package common

import (
	"context"
	"fmt"
	"os"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"golang.org/x/oauth2/google"
)

type Authentication struct {
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
}

func (a *Authentication) Prepare() ([]string, error) {
	var warnings []string
	var errs error

	var authTypes []string

	if a.AccessToken != "" {
		authTypes = append(authTypes, "access_token")
	}

	if a.AccountFile != "" {
		authTypes = append(authTypes, "account_file")
	}

	if a.CredentialsFile != "" {
		authTypes = append(authTypes, "credentials_file")
	}

	if a.CredentialsJSON != "" {
		authTypes = append(authTypes, "credentials_json")
	}

	if a.ImpersonateServiceAccount != "" {
		authTypes = append(authTypes, "impersonate_service_account")
	}

	if a.VaultGCPOauthEngine != "" {
		authTypes = append(authTypes, "vault_gcp_oauth_engine")
	}

	if len(authTypes) > 1 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("too many authentication methods specified (%s), choose only one", strings.Join(authTypes, ", ")))
	}

	// Authenticating via an account file
	if a.AccountFile != "" {
		warnings = append(warnings, "account_file is deprecated, please use either credentials_json or credentials_file instead")
		// Heuristic, but should be good enough to discriminate between
		// the two somewhat reliably.
		if strings.HasPrefix(strings.TrimSpace(a.AccountFile), "{") {
			a.CredentialsJSON = a.AccountFile
		} else {
			a.CredentialsFile = a.AccountFile
		}
	}

	if a.CredentialsFile != "" {
		cnts, err := os.ReadFile(a.CredentialsFile)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("failed to read credentials from file: %s", err))
		} else {
			a.CredentialsJSON = string(cnts)
		}
	}

	if a.CredentialsJSON != "" {
		cfg, err := google.CredentialsFromJSON(context.Background(), []byte(a.CredentialsJSON), DriverScopes...)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
		a.credentials = cfg
	}

	return warnings, errs
}

// ApplyDriverConfig applies the authentication configuration to the config for the GCE Driver
func (a Authentication) ApplyDriverConfig(cfg *GCEDriverConfig) {
	cfg.AccessToken = a.AccessToken
	cfg.ImpersonateServiceAccountName = a.ImpersonateServiceAccount
	cfg.VaultOauthEngineName = a.VaultGCPOauthEngine
	cfg.Credentials = a.credentials
}
