// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package version

import "github.com/hashicorp/packer-plugin-sdk/version"

var (
	// Version is the main version number that is being run at the moment.
	Version = "1.2.2"

	// VersionPrerelease is A pre-release marker for the Version. If this is ""
	// (empty string) then it means that it is a final release. Otherwise, this
	// is a pre-release such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease = "dev"

	// VersionMetadata is extra information to add to the version string.
	//
	// It is ignored for comparison, and aims to add extra metadata for
	// documentation purposes.
	VersionMetadata = ""

	// PluginVersion is used by the plugin set to allow Packer to recognize
	// what version this plugin is.
	PluginVersion = version.NewPluginVersion(Version, VersionPrerelease, VersionMetadata)
)
