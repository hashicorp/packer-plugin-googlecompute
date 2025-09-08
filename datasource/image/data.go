// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,DatasourceOutput

package image

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type Datasource struct {
	config Config
}

type Config struct {
	// The Google Cloud project ID to search for images.
	ProjectID string `mapstructure:"project_id"`
	// The filter expression to narrow down the image search.
	// For example: "name=ubuntu" or "family=ubuntu-2004".
	// The exrpressions can be combined with AND/OR like this:
	// "name=ubuntu AND family=ubuntu-2004".
	// See https://cloud.google.com/sdk/gcloud/reference/topic/filters
	Filters string `mapstructure:"filters"`
	// If true, the most recent image will be returned.
	// If false, an error will be returned if more than one image matches the filters.
	MostRecent bool `mapstructure:"most_recent"`

	// Specify the GCP universe to deploy in. The default is "googleapis.com".
	UniverseDomain string `mapstructure:"universe_domain"`
	// Custom service endpoints, typically used to configure the Google provider to
	// communicate with GCP-like APIs such as the Cloud Functions emulator.
	//  Supported keys are `compute`.
	//
	// Example:
	//   custom_endpoints = {
	//     compute = "https://{your-endpoint}/"
	//   }
	//
	CustomEndpoints map[string]string `mapstructure:"custom_endpoints"`
}

type DatasourceOutput struct {
	ID           string            `mapstructure:"id"`
	Name         string            `mapstructure:"name"`
	CreationDate string            `mapstructure:"creation_date"`
	Labels       map[string]string `mapstructure:"labels"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	if err := config.Decode(&d.config, nil, raws...); err != nil {
		return err
	}
	if d.config.ProjectID == "" {
		return fmt.Errorf("project_id must be specified")
	}
	if d.config.Filters == "" {
		return fmt.Errorf("filters must be specified to narrow down image search")
	}
	return nil
}

func (d *Datasource) Execute() (cty.Value, error) {
	ctx := context.Background()

	var opts []option.ClientOption
	if d.config.UniverseDomain != "" {
		opts = append(opts, option.WithUniverseDomain(d.config.UniverseDomain))
	}
	if len(d.config.CustomEndpoints) > 0 {
		if endpoint, ok := d.config.CustomEndpoints["compute"]; ok {
			opts = append(opts, option.WithEndpoint(endpoint))
		}
	}

	service, err := compute.NewService(ctx, opts...)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to create compute client: %w", err)
	}

	images, err := service.Images.List(d.config.ProjectID).Filter(d.config.Filters).Do()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	if len(images.Items) == 0 {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("no images found with filter expression: %q", d.config.Filters)
	}

	if len(images.Items) > 1 && !d.config.MostRecent {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf(
			"Your query returned more than one result. Please try a more specific search, or set most_recent = true",
		)
	}
	// Sort by most recent first
	sort.Slice(images.Items, func(i, j int) bool {
		return images.Items[i].CreationTimestamp > images.Items[j].CreationTimestamp
	})

	matched := images.Items[0]
	out := DatasourceOutput{
		ID:           fmt.Sprintf("%d", matched.Id),
		Name:         matched.Name,
		CreationDate: matched.CreationTimestamp,
		Labels: func() map[string]string {
			if matched.Labels == nil {
				return map[string]string{}
			}
			return matched.Labels
		}(),
	}

	log.Printf("[DEBUG] - datasource: found image %q with ID %s", matched.Name, out.ID)

	return hcl2helper.HCL2ValueFromConfig(out, d.OutputSpec()), nil
}
