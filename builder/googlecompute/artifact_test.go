// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"testing"

	"github.com/hashicorp/packer-plugin-googlecompute/lib/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/mitchellh/mapstructure"
)

func TestArtifact_impl(t *testing.T) {
	var _ packersdk.Artifact = new(Artifact)
}

func TestArtifactState_StateData(t *testing.T) {
	expectedData := "this is the data"
	artifact := &Artifact{
		StateData: map[string]interface{}{"state_data": expectedData},
	}

	// Valid state
	result := artifact.State("state_data")
	if result != expectedData {
		t.Fatalf("Bad: State data was %s instead of %s", result, expectedData)
	}

	// Invalid state
	result = artifact.State("invalid_key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for invalid state data name")
	}

	// Nil StateData should not fail and should return nil
	artifact = &Artifact{}
	result = artifact.State("key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for nil StateData")
	}
}

func TestArtifactState_RegistryImageMetadata(t *testing.T) {
	artifact := &Artifact{
		config: &Config{Zone: "us1"},
		image:  &common.Image{Name: "test-image", ProjectId: "5678"},
	}

	// Valid state
	result := artifact.State(registryimage.ArtifactStateURI)
	if result == nil {
		t.Fatalf("Bad: HCP Packer registry image data was nil")
	}

	var image registryimage.Image
	err := mapstructure.Decode(result, &image)
	if err != nil {
		t.Errorf("Bad: unexpected error when trying to decode state into registryimage.Image %v", err)
	}

	if image.ImageID != artifact.image.Name {
		t.Errorf("Bad: unexpected value for ImageID %q, expected %q", image.ImageID, artifact.image.Name)
	}

	if image.ProviderRegion != artifact.State("BuildZone").(string) {
		t.Errorf("Bad: unexpected value for ImageID %q, expected %q", image.ProviderRegion, artifact.State("BuildZone").(string))
	}

	if image.Labels["project_id"] != artifact.image.ProjectId {
		t.Errorf("Bad: unexpected value for Labels %q, expected %q", image.Labels["project_id"], artifact.image.ProjectId)
	}

}
