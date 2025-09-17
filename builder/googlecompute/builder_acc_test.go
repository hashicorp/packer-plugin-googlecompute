// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-googlecompute/lib/common"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed testdata
var testDataFs embed.FS

func TestAccBuilder_Basic(t *testing.T) {
	t.Parallel()

	tmpl, err := testDataFs.ReadFile("testdata/basic.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file %s", err)
	}
	testCase := &acctest.PluginTestCase{
		Name:     "googlecompute-packer-basic",
		Template: string(tmpl),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_DefaultTokenSource(t *testing.T) {
	t.Parallel()

	tmpl, err := testDataFs.ReadFile("testdata/oslogin/default-token.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file %s", err)
	}
	testCase := &acctest.PluginTestCase{
		Name:     "googlecompute-packer-default-ts",
		Template: string(tmpl),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

// generateSSHPrivateKey generates a PEM encoded ssh private key file
//
// The file's deletion is the responsibility of the caller.
func generateSSHPrivateKey() (string, error) {
	outFile := fmt.Sprintf("%s/temp_key", os.TempDir())

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", fmt.Errorf("failed to generate SSH key: %s", err)
	}

	x509key := x509.MarshalPKCS1PrivateKey(priv)

	pemKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509key,
	})

	err = os.WriteFile(outFile, pemKey, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to write private key to %q: %s", outFile, err)
	}

	return outFile, nil
}

func TestAccBuilder_DefaultTokenSourceWithPrivateKey(t *testing.T) {
	t.Parallel()

	keyFile, err := generateSSHPrivateKey()
	if err != nil {
		t.Fatalf("failed to generate SSH private key: %s", err)
	}

	defer os.Remove(keyFile)

	tmpl, err := testDataFs.ReadFile("testdata/oslogin/default-token-and-pkey.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file %s", err)
	}

	testCase := &acctest.PluginTestCase{
		Name:     "googlecompute-packer-default-ts-with-pkey",
		Template: fmt.Sprintf(string(tmpl), keyFile),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() == 0 {
					return fmt.Errorf("Packer build should have failed because of the unknown SSH key for the target instance, but succeeded. Logfile: %s", logfile)
				}
			}

			rawLogs, err := os.ReadFile(logfile)
			if err != nil {
				return fmt.Errorf("failed to read logfile %q: %s", logfile, err)
			}

			logs := string(rawLogs)

			if !strings.Contains(logs, "Private key file specified, won't import SSH key for OSLogin") {
				return fmt.Errorf("did not find message stating that a private key file was specified")
			}

			if strings.Contains(logs, "Deleting SSH public key for OSLogin...") {
				return fmt.Errorf("found a message about deleting OSLogin SSH public key, shouldn't have")
			}

			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_WrappedStartupScriptSuccess(t *testing.T) {
	t.Parallel()

	tmpl, err := testDataFs.ReadFile("testdata/wrapped-startup-scripts/successful.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file %s", err)
	}
	testCase := &acctest.PluginTestCase{
		Name:     "googlecompute-packer-good-startup-metadata",
		Template: string(tmpl),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_WrappedStartupScriptError(t *testing.T) {
	t.Parallel()

	tmpl, err := testDataFs.ReadFile("testdata/wrapped-startup-scripts/errored.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file %s", err)
	}
	testCase := &acctest.PluginTestCase{
		Name:     "googlecompute-packer-bad-startup-metadata",
		Template: string(tmpl),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 1 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_WithExtraScratchDisk(t *testing.T) {
	t.Parallel()

	tmpl, err := testDataFs.ReadFile("testdata/extra_scratch_disk.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file %s", err)
	}

	testCase := &acctest.PluginTestCase{
		Name:     "googlecompute-packer-extra-scratch-device",
		Template: string(tmpl),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_WithExtraPersistentDisk(t *testing.T) {
	t.Parallel()

	tmpl, err := testDataFs.ReadFile("testdata/extra_persistent_disk.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file %s", err)
	}

	testCase := &acctest.PluginTestCase{
		Name:     "googlecompute-packer-extra-persistent-device",
		Template: string(tmpl),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.ReadFile(logfile)
			if err != nil {
				t.Fatalf("failed to open logfile %q: %s", logfile, err)
			}

			if strings.Contains(string(logs), "Deleting persistent disk") {
				t.Errorf("extra persistent disk should be automatically deleted on instance tear-down, but was deleted during the cleanup for the step_extra_disks")
			}

			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_WithExtraPersistentDiskAndRegions(t *testing.T) {
	t.Parallel()

	tmpl, err := testDataFs.ReadFile("testdata/extra_persistent_disk_and_regions.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file %s", err)
	}

	testCase := &acctest.PluginTestCase{
		Name:     "googlecompute-packer-extra-persistent-device-and-regions",
		Template: string(tmpl),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.ReadFile(logfile)
			if err != nil {
				t.Fatalf("failed to open logfile %q: %s", logfile, err)
			}

			if strings.Contains(string(logs), "Deleting persistent disk") {
				t.Errorf("extra persistent disk should be automatically deleted on instance tear-down, but was deleted during the cleanup for the step_extra_disks")
			}

			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_WithMultipleDisks(t *testing.T) {
	t.Parallel()

	tmpl, err := testDataFs.ReadFile("testdata/multiple_disks.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file: %s", err)
	}

	testCase := &acctest.PluginTestCase{
		Name:     "googlecompute-packer-with-multiple-extra-disks",
		Template: string(tmpl),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.ReadFile(logfile)
			if err != nil {
				t.Fatalf("failed to open logfile %q: %s", logfile, err)
			}

			if strings.Contains(string(logs), "Deleting persistent disk") {
				t.Errorf("extra persistent disk should be automatically deleted on instance tear-down, but was deleted during the cleanup for the step_extra_disks")
			}

			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_ImageArchVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		baseImageFamily string
		arch            string
		instanceType    string
		expectedArch    string
	}{
		{
			name:            "googlecompute-packer-with-x86-64-explicit-image-arch",
			baseImageFamily: "fedora-cloud-38",
			arch:            "x86_64",
			instanceType:    "e2-standard-2",
			expectedArch:    "x86_64",
		},
		{
			name:            "googlecompute-packer-with-x86-64-implicit-image-arch",
			baseImageFamily: "fedora-cloud-38",
			arch:            "",
			instanceType:    "e2-standard-2",
			// Empty for this case as the source image has an empty
			// architecture, so this gets forwarded as-is
			expectedArch: "",
		},
		{
			name:            "googlecompute-packer-with-arm64-explicit-image-arch",
			baseImageFamily: "ubuntu-minimal-2204-lts-arm64",
			arch:            "arm64",
			instanceType:    "t2a-standard-2",
			expectedArch:    "arm64",
		},
		{
			name:            "googlecompute-packer-with-arm64-implicit-image-arch",
			baseImageFamily: "ubuntu-minimal-2204-lts-arm64",
			arch:            "",
			instanceType:    "t2a-standard-2",
			expectedArch:    "arm64",
		},
	}

	for _, tt := range tests {
		testRun := tt
		t.Run(testRun.name, func(t *testing.T) {
			t.Parallel()

			imageName := fmt.Sprintf("%s-%d", testRun.name, time.Now().UTC().Unix())

			tmpl, err := testDataFs.ReadFile("testdata/image_arch_builds.pkr.hcl")
			if err != nil {
				t.Fatalf("failed to read testdata file: %s", err)
			}

			rawTemplate := fmt.Sprintf(string(tmpl), imageName, testRun.baseImageFamily, testRun.arch, testRun.instanceType)

			testCase := &acctest.PluginTestCase{
				Name:     testRun.name,
				Template: rawTemplate,
				Teardown: func() error {
					driver, err := common.NewDriverGCE(common.GCEDriverConfig{})
					if err != nil {
						return fmt.Errorf("failed to create GCE driver: %s", err)
					}

					chErr := driver.DeleteImage(os.Getenv("GOOGLE_PROJECT_ID"), imageName)
					for err := range chErr {
						return err
					}
					return nil
				},
				Check: func(buildCommand *exec.Cmd, logfile string) error {
					if buildCommand.ProcessState != nil {
						if buildCommand.ProcessState.ExitCode() != 0 {
							return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
						}
					}

					driver, err := common.NewDriverGCE(common.GCEDriverConfig{})
					if err != nil {
						return fmt.Errorf("failed to create GCE driver: %s", err)
					}

					img, err := driver.GetImageFromProject(os.Getenv("GOOGLE_PROJECT_ID"), imageName, false)
					if err != nil {
						return fmt.Errorf("failed to get image: %s", err)
					}

					// Manually uppercase the arch as it cannot be used
					// for the image name otherwise
					upperArch := strings.ToUpper(testRun.expectedArch)
					if img.Architecture != upperArch {
						return fmt.Errorf("image architecture mismatch, expected %q, got %q", upperArch, img.Architecture)
					}

					return nil
				},
			}
			acctest.TestPlugin(t, testCase)
		})
	}
}

func TestAccBuilder_NetworkIP(t *testing.T) {
	t.Parallel()
	state := testState(t)

	zone := "asia-east1-a"
	networkIP := "10.140.0.13"
	imageName := fmt.Sprintf("%s-%d", "packer-network-ip", time.Now().UTC().Unix())

	extraArgs := []string{
		"-var", fmt.Sprintf("zone=%s", zone),
		"-var", fmt.Sprintf("network_ip=%s", networkIP),
		"-var", fmt.Sprintf("image_name=%s", imageName),
	}

	tmpl, err := testDataFs.ReadFile("testdata/network_ip.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read testdata file %s", err)
	}
	testCase := &acctest.PluginTestCase{
		Name:           "googlecompute-packer-network-ip",
		Template:       string(tmpl),
		BuildExtraArgs: extraArgs,
		Setup: func() error {
			step := new(StepCreateImage)
			defer step.Cleanup(state)

			return nil
		},
		Teardown: func() error {
			driver, err := common.NewDriverGCE(common.GCEDriverConfig{})
			if err != nil {
				return fmt.Errorf("failed to create GCE driver: %s", err)
			}

			chErr := driver.DeleteImage(os.Getenv("GOOGLE_PROJECT_ID"), imageName)
			for err := range chErr {
				return err
			}
			return nil
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			rawLogs, err := os.ReadFile(logfile)
			if err != nil {
				return fmt.Errorf("failed to read logfile %q: %s", logfile, err)
			}

			logs := string(rawLogs)

			if !strings.Contains(logs, networkIP) {
				return fmt.Errorf("did not find message stating that a network IP was specified")
			}

			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_CustomEndpointsAndUniverse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		universeDomain  string
		customEndpoints string // HCL block as a string for easy templating
		expectSuccess   bool
	}{
		{
			name:           "googlecompute-with-custom-endpoint-beta",
			universeDomain: "",
			// We use the real, public beta endpoint for the compute API.
			// A successful build proves the builder correctly used this endpoint.
			customEndpoints: "https://compute.googleapis.com/compute/beta/",
			expectSuccess:   true,
		},
		{
			name: "googlecompute-with-invalid-universe",
			// We use a fake domain that does not exist.
			// The build MUST fail. A failure proves the builder tried to use
			// this domain instead of ignoring it and using the default.
			universeDomain:  "packer.example.com",
			customEndpoints: "",
			expectSuccess:   false,
		},
	}

	for _, tt := range tests {
		testRun := tt
		t.Run(testRun.name, func(t *testing.T) {
			t.Parallel()

			imageName := fmt.Sprintf("%s-%d", testRun.name, time.Now().UTC().Unix())

			tmpl, err := testDataFs.ReadFile("testdata/custom_endpoint.pkr.hcl")
			if err != nil {
				t.Fatalf("failed to read testdata file: %s", err)
			}

			extraArgs := []string{
				"-var", fmt.Sprintf("image_name=%s", imageName),
				"-var", fmt.Sprintf("compute_custom_endpoints=%s", tt.customEndpoints),
				"-var", fmt.Sprintf("universe_domain=%s", tt.universeDomain),
			}

			driverConfig := common.GCEDriverConfig{
				ProjectId:       os.Getenv("GOOGLE_PROJECT_ID"),
				UniverseDomain:  tt.universeDomain,
				CustomEndpoints: map[string]string{"compute": tt.customEndpoints},
			}

			testCase := &acctest.PluginTestCase{
				Name:           testRun.name,
				Template:       string(tmpl),
				BuildExtraArgs: extraArgs,
				Teardown: func() error {
					if !testRun.expectSuccess {
						return nil
					}
					// Always attempt to delete the image, even on failure, in case it was partially created.
					driver, err := common.NewDriverGCE(driverConfig)
					if err != nil {
						// Don't fail teardown if the driver can't be created (e.g., bad credentials).
						t.Logf("Failed to create GCE driver for teardown: %s", err)
						return nil
					}

					chErr := driver.DeleteImage(os.Getenv("GOOGLE_PROJECT_ID"), imageName)
					err = <-chErr
					if err != nil {
						t.Logf("Error during image cleanup for %s: %s", imageName, err)
					}
					return nil
				},
				Check: func(buildCommand *exec.Cmd, logfile string) error {
					exitCode := buildCommand.ProcessState.ExitCode()

					// Case 1: Build was expected to succeed, but it failed.
					if testRun.expectSuccess && exitCode != 0 {
						return fmt.Errorf("build was expected to succeed but failed with exit code %d. Logfile: %s", exitCode, logfile)
					}

					// Case 2: Build was expected to fail, but it succeeded.
					if !testRun.expectSuccess && exitCode == 0 {
						return fmt.Errorf("build was expected to fail but it succeeded. This indicates the invalid configuration was ignored.")
					}

					// Case 3: Build succeeded as expected. Verify the artifact.
					if testRun.expectSuccess {
						driver, err := common.NewDriverGCE(driverConfig)
						if err != nil {
							return fmt.Errorf("failed to create GCE driver for verification: %s", err)
						}

						_, err = driver.GetImageFromProject(os.Getenv("GOOGLE_PROJECT_ID"), imageName, false)
						if err != nil {
							return fmt.Errorf("failed to get the created image '%s' for verification: %s", imageName, err)
						}
						t.Logf("Successfully verified that image '%s' was created.", imageName)
					}

					// Case 4: Build failed as expected. The test passes.
					if !testRun.expectSuccess {
						t.Logf("Build failed as expected with a non-zero exit code.")
					}

					return nil
				},
			}
			acctest.TestPlugin(t, testCase)
		})
	}
}
