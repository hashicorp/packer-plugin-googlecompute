// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
)

type AccountCredentials struct {
	jsonKey []byte
	// Used for other auth methods (client_credentials.json, service account key file,
	// gcloud user credentials file, JSON config file for workload identity federation)
	credentials *google.Credentials
}

func ProcessCredentialsFile(text string) (*AccountCredentials, error) {
	var data []byte = []byte(text)
	var err error

	if !strings.HasPrefix(text, "{") {
		// If text was not JSON, assume it is a file path instead
		if _, err := os.Stat(text); os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials_file path does not exist: %v", text)
		}

		data, err = os.ReadFile(text)
		if err != nil {
			return nil, fmt.Errorf("error reading account_file from path '%v': %v", text, err)
		}
	}

	var credentials *google.Credentials
	if len(data) == 0 {
		credentials, err = google.FindDefaultCredentials(context.Background(), DriverScopes...)
		if err != nil {
			return nil, fmt.Errorf("error finding default credentials: %v", err)
		}
	} else if strings.HasPrefix(string(data), "{") {
		credentials, err = google.CredentialsFromJSON(context.Background(), data, DriverScopes...)
		if err != nil {
			return nil, fmt.Errorf("error parsing credentials JSON: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unknown credentials format - please provide a JSON file or JSON string, or an empty string to use the Application Default Credentials")
	}

	return &AccountCredentials{
		jsonKey:     data,
		credentials: credentials,
	}, nil
}
