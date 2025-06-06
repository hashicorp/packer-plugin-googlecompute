// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package secretsmanager

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatConfig is an auto-generated flat version of Config.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConfig struct {
	ProjectId *string `mapstructure:"project_id" required:"true" cty:"project_id" hcl:"project_id"`
	Name      *string `mapstructure:"name" required:"true" cty:"name" hcl:"name"`
	Key       *string `mapstructure:"key" cty:"key" hcl:"key"`
	Version   *string `mapstructure:"version" cty:"version" hcl:"version"`
}

// FlatMapstructure returns a new FlatConfig.
// FlatConfig is an auto-generated flat version of Config.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Config) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfig)
}

// HCL2Spec returns the hcl spec of a Config.
// This spec is used by HCL to read the fields of Config.
// The decoded values from this spec will then be applied to a FlatConfig.
func (*FlatConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"project_id": &hcldec.AttrSpec{Name: "project_id", Type: cty.String, Required: false},
		"name":       &hcldec.AttrSpec{Name: "name", Type: cty.String, Required: false},
		"key":        &hcldec.AttrSpec{Name: "key", Type: cty.String, Required: false},
		"version":    &hcldec.AttrSpec{Name: "version", Type: cty.String, Required: false},
	}
	return s
}

// FlatDatasourceOutput is an auto-generated flat version of DatasourceOutput.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatDatasourceOutput struct {
	Payload  *string `mapstructure:"payload" cty:"payload" hcl:"payload"`
	Value    *string `mapstructure:"value" cty:"value" hcl:"value"`
	Checksum *int64  `mapstructure:"checksum" cty:"checksum" hcl:"checksum"`
}

// FlatMapstructure returns a new FlatDatasourceOutput.
// FlatDatasourceOutput is an auto-generated flat version of DatasourceOutput.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*DatasourceOutput) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatDatasourceOutput)
}

// HCL2Spec returns the hcl spec of a DatasourceOutput.
// This spec is used by HCL to read the fields of DatasourceOutput.
// The decoded values from this spec will then be applied to a FlatDatasourceOutput.
func (*FlatDatasourceOutput) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"payload":  &hcldec.AttrSpec{Name: "payload", Type: cty.String, Required: false},
		"value":    &hcldec.AttrSpec{Name: "value", Type: cty.String, Required: false},
		"checksum": &hcldec.AttrSpec{Name: "checksum", Type: cty.Number, Required: false},
	}
	return s
}
