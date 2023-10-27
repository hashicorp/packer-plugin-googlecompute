// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	OIDC_JSON = []byte(`{
		"type":"external_account",
	"audience":"//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/poolName/providers/some-provider-name",
	"subject_token_type":"urn:ietf:params:oauth:token-type:jwt","token_url":"https://sts.googleapis.com/v1/token",
	"service_account_impersonation_url":"https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/some-service-account@some-gcp-project.iam.gserviceaccount.com:generateAccessToken",
	"credential_source":{"url":"https://pipelines.actions.githubusercontent.com/blahblahblah",
	"headers":{"Authorization":"***"},
	"format":{"type":"json","subject_token_field_name":"value"}}}
	`)
)

func TestProcessCredentialsFile_OIDCtext(t *testing.T) {
	account, err := ProcessCredentialsFile(string(JWT_JSON))
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotNil(t, account.credentials)
}

func TestProcessCredentialsFile_OIDCFile(t *testing.T) {
	workdir := t.TempDir()
	credsFile := path.Join(workdir, "jwt-file")
	err := os.WriteFile(credsFile, JWT_JSON, 0777)
	assert.NoError(t, err)
	account, err := ProcessCredentialsFile(credsFile)
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotNil(t, account.credentials)
}

func TestProcessCredentialsFile_invalidContent(t *testing.T) {
	credentials, err := ProcessCredentialsFile("{}")
	assert.Nil(t, credentials)
	assert.Error(t, err)
}
