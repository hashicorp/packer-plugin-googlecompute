package googlecomputeimport

import (
	"fmt"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
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

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
