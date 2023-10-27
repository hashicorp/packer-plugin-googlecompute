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

	conf, err := google.JWTConfigFromJSON(data, DriverScopes...)

	if err != nil {
		return nil, fmt.Errorf("Error parsing account_file: %v", err)
	}

	return &ServiceAccount{
		jsonKey: data,
		jwt:     conf,
	}, nil
}
