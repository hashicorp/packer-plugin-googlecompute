// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecomputeimport

import (
	"fmt"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

const BuilderId = "packer.post-processor.googlecompute-import"

type Artifact struct {
	paths []string
}

var _ packersdk.Artifact = new(Artifact)

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Id() string {
	return ""
}

func (a *Artifact) Files() []string {
	pathsCopy := make([]string, len(a.paths))
	copy(pathsCopy, a.paths)
	return pathsCopy
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Exported artifacts in: %s", a.paths)
}

func (a *Artifact) State(name string) interface{} {
	if name == registryimage.ArtifactStateURI {
		return a.hcpPackerRegistryMetadata()
	}
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}

func (a *Artifact) hcpPackerRegistryMetadata() []*registryimage.Image {

	var images []*registryimage.Image
	for _, exportedPath := range a.Files() {
		ep := exportedPath
		pathParts := strings.SplitN(exportedPath, "/", 4)
		img, _ := registryimage.FromArtifact(a,
			registryimage.WithID(ep),
			registryimage.WithRegion(pathParts[2]))

		images = append(images, img)
	}

	return images
}
