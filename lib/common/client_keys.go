// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CustomerEncryptionKey

package common

import compute "google.golang.org/api/compute/v1"

type CustomerEncryptionKey struct {
	// KmsKeyName: The name of the encryption key that is stored in Google
	// Cloud KMS.
	KmsKeyName string `mapstructure:"kmsKeyName" json:"kmsKeyName,omitempty"`

	// RawKey: Specifies a 256-bit customer-supplied encryption key, encoded
	// in RFC 4648 base64 to either encrypt or decrypt this resource.
	RawKey string `mapstructure:"rawKey" json:"rawKey,omitempty"`
}

func (k *CustomerEncryptionKey) ComputeType() *compute.CustomerEncryptionKey {
	if k == nil {
		return nil
	}
	return &compute.CustomerEncryptionKey{
		KmsKeyName: k.KmsKeyName,
		RawKey:     k.RawKey,
	}
}
