// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecomputeexport

import (
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/mitchellh/mapstructure"
)

func TestArtifact_impl(t *testing.T) {
	var _ packersdk.Artifact = new(Artifact)
}

func TestArtifactState_RegistryImageMetadata(t *testing.T) {
	artifact := &Artifact{
		paths: []string{"gs://testbucket/packer/file.gz"},
	}

	// Valid state
	result := artifact.State(registryimage.ArtifactStateURI)
	if result == nil {
		t.Fatalf("Bad: HCP Packer registry image data was nil")
	}

	var images []registryimage.Image
	err := mapstructure.Decode(result, &images)
	if err != nil {
		t.Errorf("Bad: unexpected error when trying to decode state into registryimage.Image %v", err)
	}

	if len(images) != 1 {
		t.Errorf("Bad: we should have one image for this test Artifact but we got %d", len(images))
	}

	image := images[0]
	for _, p := range artifact.Files() {
		pathParts := strings.SplitN(p, "/", 4)
		if image.ImageID != p {
			t.Errorf("Bad: unexpected value for ImageID %q, expected %q", image.ImageID, p)
		}

		if image.ProviderRegion != pathParts[2] {
			t.Errorf("Bad: unexpected value for Region %q, expected %q", image.ProviderRegion, pathParts[2])
		}
	}

}
