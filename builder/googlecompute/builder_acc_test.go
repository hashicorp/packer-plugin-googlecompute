/*
deregister the test image with
aws ec2 deregister-image --image-id $(aws ec2 describe-images --output text --filters "Name=name,Values=packer-test-packer-test-dereg" --query 'Images[*].{ID:ImageId}')
*/
//nolint:unparam
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
