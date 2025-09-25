// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-googlecompute/version"
	"github.com/hashicorp/packer-plugin-sdk/plugin"

	googlecompute "github.com/hashicorp/packer-plugin-googlecompute/builder/googlecompute"
	"github.com/hashicorp/packer-plugin-googlecompute/datasource/image"
	"github.com/hashicorp/packer-plugin-googlecompute/datasource/parametermanager"
	"github.com/hashicorp/packer-plugin-googlecompute/datasource/secretsmanager"
	googlecomputeexport "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-export"
	googlecomputeimport "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-import"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(googlecompute.Builder))
	pps.RegisterPostProcessor("import", new(googlecomputeimport.PostProcessor))
	pps.RegisterPostProcessor("export", new(googlecomputeexport.PostProcessor))
	pps.RegisterDatasource("secretsmanager", new(secretsmanager.Datasource))
	pps.RegisterDatasource("parametermanager", new(parametermanager.Datasource))
	pps.RegisterDatasource("image", new(image.Datasource))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
