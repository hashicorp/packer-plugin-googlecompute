// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"os"

	"golang.org/x/oauth2/google"
)

// ProcessCredentialsFile reads a valid google.Credentials JSON from the file at `path`
func ProcessCredentialsFile(path string) (*google.Credentials, error) {
	cnts, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ProcessCredentials(cnts)
}

// ProcessCredentials essentially proxies the google lib's function to read
// credentials from raw JSON
func ProcessCredentials(text []byte) (*google.Credentials, error) {
	return google.CredentialsFromJSON(context.Background(), text, DriverScopes...)
}
