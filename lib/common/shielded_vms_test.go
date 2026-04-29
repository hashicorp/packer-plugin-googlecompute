// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// When no Secure Boot signature inputs are configured, the helper must return
// a nil *InitialStateConfig so the caller does not attach an empty
// shieldedInstanceInitialState to the image insert request. Attaching an
// empty value overrides the PK/KEKs/db/dbx that would otherwise be inherited
// from the source disk and breaks Secure Boot on the resulting image.
func TestCreateShieldedVMStateConfig_NoInputsReturnsNil(t *testing.T) {
	cfg, err := CreateShieldedVMStateConfig("", nil, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, cfg, "expected nil config when no signature inputs are configured")

	cfg, err = CreateShieldedVMStateConfig("", []string{}, []string{}, []string{})
	assert.NoError(t, err)
	assert.Nil(t, cfg, "expected nil config when signature inputs are empty slices")
}

func TestCreateShieldedVMStateConfig_PopulatesFieldsWhenInputsProvided(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "fake-key")
	if err := os.WriteFile(keyPath, []byte("fake key data"), 0600); err != nil {
		t.Fatalf("failed to write fake key: %v", err)
	}

	tests := []struct {
		name string
		pk   string
		keks []string
		dbs  []string
		dbxs []string
	}{
		{name: "platform key only", pk: keyPath},
		{name: "kek only", keks: []string{keyPath}},
		{name: "db only", dbs: []string{keyPath}},
		{name: "dbx only", dbxs: []string{keyPath}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := CreateShieldedVMStateConfig(tc.pk, tc.keks, tc.dbs, tc.dbxs)
			assert.NoError(t, err)
			assert.NotNil(t, cfg, "expected non-nil config when at least one signature input is provided")
		})
	}
}
