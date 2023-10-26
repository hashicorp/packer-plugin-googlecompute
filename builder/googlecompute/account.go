// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

type ServiceAccount struct {
	jsonKey []byte
	// Used for JWT-based authentication (service account JSON file)
	jwt *jwt.Config
	// Used for other auth methods (client_credentials.json, service account key file,
	// gcloud user credentials file, JSON config file for workload identity federation)
	credentials *google.Credentials
}

// ProcessAccountFile will return a ServiceAccount for the JSON account file stored in text.
// Otherwise it will return an error if text does not look or reference a valid account file.
func ProcessAccountFile(text string) (*ServiceAccount, error) {
	var data []byte = []byte(text)
	var err error

	if !strings.HasPrefix(text, "{") {
		// If text was not JSON, assume it is a file path instead
		if _, err := os.Stat(text); os.IsNotExist(err) {
			return nil, fmt.Errorf("account_file path does not exist: %s", text)
		}

		data, err = os.ReadFile(text)
		if err != nil {
			return nil, fmt.Errorf("Error reading account_file from path '%s': %s", text, err)
		}
	}

	var jwtConf *jwt.Config
	var jwtErr error
	jwtConf, jwtErr = google.JWTConfigFromJSON(data, DriverScopes...)

	var credentials *google.Credentials
	var credentialsErr error
	// If JWT format failed, try alternate format. We don't want to load a given file as both. JWT is the
	// more restricted of the two.
	if jwtErr != nil {
		credentials, credentialsErr = google.CredentialsFromJSON(context.Background(), data, DriverScopes...)
	}

	if jwtConf != nil || credentials != nil {
		return &ServiceAccount{
			jsonKey:     data,
			jwt:         jwtConf,
			credentials: credentials,
		}, nil
	}
	return nil, fmt.Errorf("Error parsing account_file. Neither JWT format nor alternate format succeeded.\nJWT format error: %s\nAlternate format error: %s", jwtErr, credentialsErr)
}
