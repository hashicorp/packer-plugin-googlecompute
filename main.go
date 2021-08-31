package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-googlecompute/version"
	"github.com/hashicorp/packer-plugin-sdk/plugin"

	googlecompute "github.com/hashicorp/packer-plugin-googlecompute/builder/googlecompute"
	googlecomputeexport "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-export"
	googlecomputeimport "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-import"
)

var (
	// Version is the main version number that is being run at the moment.
	Version = "1.0.2"

	// VersionPrerelease is A pre-release marker for the Version. If this is ""
	// (empty string) then it means that it is a final release. Otherwise, this
	// is a pre-release such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease = ""
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(googlecompute.Builder))
	pps.RegisterPostProcessor("import", new(googlecomputeimport.PostProcessor))
	pps.RegisterPostProcessor("export", new(googlecomputeexport.PostProcessor))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
