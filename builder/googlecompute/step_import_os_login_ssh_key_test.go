// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	configsdk "github.com/hashicorp/packer-plugin-sdk/template/config"
	"google.golang.org/api/oauth2/v2"
)

func TestStepImportOSLoginSSHKey_impl(t *testing.T) {
	var _ multistep.Step = new(StepImportOSLoginSSHKey)
}

func TestStepImportOSLoginSSHKey(t *testing.T) {
	tt := []struct {
		Name           string
		UseOSLogin     configsdk.Trilean
		ExpectedEmail  string
		ExpectedAction multistep.StepAction
		PubKeyExpected bool
	}{
		{
			Name:           "UseOSLoginDisabled",
			ExpectedAction: multistep.ActionContinue,
		},
		{
			Name:           "UseOSLoginWithAccountFile",
			UseOSLogin:     configsdk.TriTrue,
			ExpectedAction: multistep.ActionContinue,
			ExpectedEmail:  "raffi-compute@developer.gserviceaccount.com",
			PubKeyExpected: true,
		},
	}

	for _, tc := range tt {
		tc := tc
		state := testState(t)
		fakeAccountEmail := "raffi-compute@developer.gserviceaccount.com"
		step := &StepImportOSLoginSSHKey{
			TokeninfoFunc: func() (*oauth2.Tokeninfo, error) {
				return &oauth2.Tokeninfo{Email: fakeAccountEmail}, nil
			},
		}
		defer step.Cleanup(state)

		config := state.Get("config").(*Config)
		config.UseOSLogin = tc.UseOSLogin

		if tc.PubKeyExpected {
			config.Comm.SSHPublicKey = []byte{'k', 'e', 'y'}
		}

		if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
			t.Fatalf("bad action: %#v", action)
		}

		if step.accountEmail != tc.ExpectedEmail {
			t.Fatalf("expected accountEmail to be %q but got %q", tc.ExpectedEmail, step.accountEmail)
		}

		if _, ok := state.GetOk("ssh_key_public_sha256"); !ok && tc.PubKeyExpected {
			t.Fatal("expected to see a public key")
		}
	}
}

func TestStepImportOSLoginSSHKey_withAccountFile(t *testing.T) {
	// default teststate contains an account file
	state := testState(t)
	fakeAccountEmail := "raffi-compute@developer.gserviceaccount.com"
	step := &StepImportOSLoginSSHKey{
		TokeninfoFunc: func() (*oauth2.Tokeninfo, error) {
			return &oauth2.Tokeninfo{Email: fakeAccountEmail}, nil
		},
	}
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	config.UseOSLogin = configsdk.TriTrue
	config.Comm.SSHPublicKey = []byte{'k', 'e', 'y'}

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if step.accountEmail != fakeAccountEmail {
		t.Fatalf("expected accountEmail to be %q but got %q", fakeAccountEmail, step.accountEmail)
	}

	pubKey, ok := state.GetOk("ssh_key_public_sha256")
	if !ok {
		t.Fatal("expected to see a public key")
	}

	sha256sum := sha256.Sum256(config.Comm.SSHPublicKey)
	if pubKey != hex.EncodeToString(sha256sum[:]) {
		t.Errorf("expected to see a matching public key, but got %q", pubKey)
	}
}

func TestStepImportOSLoginSSHKey_withNoAccountFile(t *testing.T) {
	state := testState(t)
	fakeAccountEmail := "testing@packer.io"
	step := &StepImportOSLoginSSHKey{
		TokeninfoFunc: func() (*oauth2.Tokeninfo, error) {
			return &oauth2.Tokeninfo{Email: fakeAccountEmail}, nil
		},
	}
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	config.UseOSLogin = configsdk.TriTrue
	config.Comm.SSHPublicKey = []byte{'k', 'e', 'y'}

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if step.accountEmail != fakeAccountEmail {
		t.Fatalf("expected accountEmail to be %q but got %q", fakeAccountEmail, step.accountEmail)
	}

	pubKey, ok := state.GetOk("ssh_key_public_sha256")
	if !ok {
		t.Fatal("expected to see a public key")
	}

	sha256sum := sha256.Sum256(config.Comm.SSHPublicKey)
	if pubKey != hex.EncodeToString(sha256sum[:]) {
		t.Errorf("expected to see a matching public key, but got %q", pubKey)
	}
}

func TestStepImportOSLoginSSHKey_withGCEAndNoAccount(t *testing.T) {
	state := testState(t)
	fakeGCEEmail := "testing@packer.io"
	step := &StepImportOSLoginSSHKey{
		GCEUserFunc: func() string {
			return fakeGCEEmail
		},
	}
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	config.UseOSLogin = configsdk.TriTrue
	config.Comm.SSHPublicKey = []byte{'k', 'e', 'y'}

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if step.accountEmail != fakeGCEEmail {
		t.Fatalf("expected accountEmail to be %q but got %q", fakeGCEEmail, step.accountEmail)
	}

	pubKey, ok := state.GetOk("ssh_key_public_sha256")
	if !ok {
		t.Fatal("expected to see a public key")
	}

	sha256sum := sha256.Sum256(config.Comm.SSHPublicKey)
	if pubKey != hex.EncodeToString(sha256sum[:]) {
		t.Errorf("expected to see a matching public key, but got %q", pubKey)
	}
}

func TestStepImportOSLoginSSHKey_withGCEAndAccount(t *testing.T) {
	state := testState(t)
	fakeGCEEmail := "testing@packer.io"
	fakeAccountEmail := "raffi-compute@developer.gserviceaccount.com"
	step := &StepImportOSLoginSSHKey{
		TokeninfoFunc: func() (*oauth2.Tokeninfo, error) {
			return &oauth2.Tokeninfo{Email: fakeAccountEmail}, nil
		},
		GCEUserFunc: func() string {
			return fakeGCEEmail
		},
	}
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	config.UseOSLogin = configsdk.TriTrue
	config.Comm.SSHPublicKey = []byte{'k', 'e', 'y'}

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if step.accountEmail != fakeAccountEmail {
		t.Fatalf("expected accountEmail to be %q but got %q", fakeAccountEmail, step.accountEmail)
	}

	pubKey, ok := state.GetOk("ssh_key_public_sha256")
	if !ok {
		t.Fatal("expected to see a public key")
	}

	sha256sum := sha256.Sum256(config.Comm.SSHPublicKey)
	if pubKey != hex.EncodeToString(sha256sum[:]) {
		t.Errorf("expected to see a matching public key, but got %q", pubKey)
	}
}

func TestStepImportOSLoginSSHKey_withPrivateSSHKey(t *testing.T) {
	// default teststate contains an account file
	state := testState(t)
	step := new(StepImportOSLoginSSHKey)
	defer step.Cleanup(state)

	pkey, err := generateSSHPrivateKey()
	if err != nil {
		t.Fatalf("failed to generate SSH key: %s", err)
	}
	defer os.Remove(pkey)

	config := state.Get("config").(*Config)
	config.UseOSLogin = configsdk.TriTrue
	config.Comm.SSHPrivateKeyFile = pkey

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if step.accountEmail != "" {
		t.Fatalf("expected accountEmail to be unset but got %q", step.accountEmail)
	}

	pubKey, ok := state.GetOk("ssh_key_public_sha256")
	if ok {
		t.Errorf("expected to not see a public key when using a dedicated private key, but got %q", pubKey)
	}
}

func TestGetOSLoginUsername(t *testing.T) {
	tests := []struct {
		name             string
		usernameFromAPI  string
		configUsername   string
		expectedUsername string
		shouldPrependExt bool
	}{
		{
			name:             "Auto mode returns API username",
			usernameFromAPI:  "apiuser",
			configUsername:   "__auto__",
			expectedUsername: "apiuser",
		},
		{
			name:             "Empty config returns API username",
			usernameFromAPI:  "apiuser",
			configUsername:   "",
			expectedUsername: "apiuser",
		},
		{
			name:             "External mode prepends ext_",
			usernameFromAPI:  "userabc",
			configUsername:   "__external__",
			shouldPrependExt: true,
			expectedUsername: "ext_userabc",
		},
		{
			name:             "External mode truncates overlength name",
			usernameFromAPI:  "averyveryverylongusernamethatexceedsthemax",
			configUsername:   "__external__",
			shouldPrependExt: true,
			expectedUsername: "ext_averyveryverylongusernametha",
		},
		{
			name:             "Custom username returned directly",
			usernameFromAPI:  "ignored",
			configUsername:   "my_custom_user",
			expectedUsername: "my_custom_user",
		},
		{
			name:             "Custom username is truncated",
			usernameFromAPI:  "ignored",
			configUsername:   "averyveryveryveryverylongcustomusername",
			expectedUsername: "averyveryveryveryverylongcustomu",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			username := getUsername(tc.usernameFromAPI, tc.configUsername)
			if username != tc.expectedUsername {
				t.Errorf("Expected username %q, got %q", tc.expectedUsername, username)
			}
			if tc.shouldPrependExt {
				if len(username) > 0 && username[:4] != "ext_" {
					t.Errorf("Expected username to start with 'ext_', got %q", username)
				}
			} else {
				if len(username) > 0 && username[:4] == "ext_" {
					t.Errorf("Expected username to not start with 'ext_', got %q", username)
				}
			}

			if len(username) > 32 {
				t.Errorf("Expected username to be truncated to 32 characters, got %q", username)
			}
		})
	}
}
