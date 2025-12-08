// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package image

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

var projectID = os.Getenv("GOOGLE_PROJECT_ID")

func TestAccGCPImageDatasource(t *testing.T) {
	if projectID == "" {
		t.Skip("GOOGLE_PROJECT_ID must be set")
	}

	imageName := "debian-12-bookworm-v202"

	tmpl := loadTemplate(t)

	family := "debian-12"

	projectID = "debian-cloud"

	tc := &acctest.PluginTestCase{
		Name:     "gcp_image_datasource",
		Template: tmpl,
		BuildExtraArgs: []string{
			"-var", fmt.Sprintf("project_id=%s", projectID),
			"-var", fmt.Sprintf("family=%s", family),
		},
		Check: func(cmd *exec.Cmd, logfile string) error {
			out, err := os.ReadFile(logfile)
			if err != nil {
				return err
			}
			output := string(out)
			if !regexp.MustCompile(regexp.QuoteMeta(imageName)).MatchString(output) {
				t.Errorf("expected image name %q in logs:\n%s", imageName, output)
			}
			return nil
		},
	}

	acctest.TestPlugin(t, tc)
}

func loadTemplate(t *testing.T) string {
	content, err := os.ReadFile("test-fixtures/template.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to read test template: %v", err)
	}
	return string(content)
}
