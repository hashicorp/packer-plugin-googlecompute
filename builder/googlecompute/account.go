// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

type ServiceAccount struct {
	jsonKey []byte
	jwt     *jwt.Config
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

	var conf *jwt.Config

	conf, err = google.JWTConfigFromJSON(data, DriverScopes...)
	// If we fail to load the file as JWTConfig, we're actually probably OK.
	// The only thing we really use this for setting email address defaults
	// in a few places. We squelch the error here so that non-JWT credentials
	// can be used, but we don't actually load data analogous to the JWT data.
	if err != nil && !strings.Contains(err.Error(), "google: read JWT from JSON credentials: 'type' field is") {
		return nil, fmt.Errorf("Error parsing account_file: %s", err)
	}

	return &ServiceAccount{
		jsonKey: data,
		// Conf can be nil if we failed to load the file as JWTConfig.
		jwt: conf,
	}, nil
}
