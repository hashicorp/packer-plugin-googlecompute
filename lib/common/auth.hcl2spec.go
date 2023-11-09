// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package common

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatAuthentication is an auto-generated flat version of Authentication.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatAuthentication struct {
	AccessToken               *string `mapstructure:"access_token" required:"false" cty:"access_token" hcl:"access_token"`
	AccountFile               *string `mapstructure:"account_file" required:"false" cty:"account_file" hcl:"account_file"`
	CredentialsFile           *string `mapstructure:"credentials_file" required:"false" cty:"credentials_file" hcl:"credentials_file"`
	CredentialsJSON           *string `mapstructure:"credentials_json" required:"false" cty:"credentials_json" hcl:"credentials_json"`
	ImpersonateServiceAccount *string `mapstructure:"impersonate_service_account" required:"false" cty:"impersonate_service_account" hcl:"impersonate_service_account"`
	VaultGCPOauthEngine       *string `mapstructure:"vault_gcp_oauth_engine" cty:"vault_gcp_oauth_engine" hcl:"vault_gcp_oauth_engine"`
}

// FlatMapstructure returns a new FlatAuthentication.
// FlatAuthentication is an auto-generated flat version of Authentication.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Authentication) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatAuthentication)
}

// HCL2Spec returns the hcl spec of a Authentication.
// This spec is used by HCL to read the fields of Authentication.
// The decoded values from this spec will then be applied to a FlatAuthentication.
func (*FlatAuthentication) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"access_token":                &hcldec.AttrSpec{Name: "access_token", Type: cty.String, Required: false},
		"account_file":                &hcldec.AttrSpec{Name: "account_file", Type: cty.String, Required: false},
		"credentials_file":            &hcldec.AttrSpec{Name: "credentials_file", Type: cty.String, Required: false},
		"credentials_json":            &hcldec.AttrSpec{Name: "credentials_json", Type: cty.String, Required: false},
		"impersonate_service_account": &hcldec.AttrSpec{Name: "impersonate_service_account", Type: cty.String, Required: false},
		"vault_gcp_oauth_engine":      &hcldec.AttrSpec{Name: "vault_gcp_oauth_engine", Type: cty.String, Required: false},
	}
	return s
}
