package googlecompute

import (
	"embed"
	"fmt"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed testdata
var testDataFs embed.FS

func TestAccBuilder_Basic(t *testing.T) {
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

func TestAccBuilder_WrappedStartupScriptSuccess(t *testing.T) {
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
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccBuilder_WithExtraPersistentDiskAndRegions(t *testing.T) {
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
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
