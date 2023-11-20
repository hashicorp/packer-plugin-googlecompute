// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type NodeAffinity

package common

import compute "google.golang.org/api/compute/v1"

// Node affinity label configuration
type NodeAffinity struct {
	// Key: Corresponds to the label key of Node resource.
	Key string `mapstructure:"key" json:"key"`

	// Operator: Defines the operation of node selection. Valid operators are IN for affinity and
	// NOT_IN for anti-affinity.
	Operator string `mapstructure:"operator" json:"operator"`

	// Values: Corresponds to the label values of Node resource.
	Values []string `mapstructure:"values" json:"values"`
}

func (a *NodeAffinity) ComputeType() *compute.SchedulingNodeAffinity {
	if a == nil {
		return nil
	}
	return &compute.SchedulingNodeAffinity{
		Key:      a.Key,
		Operator: a.Operator,
		Values:   a.Values,
	}
}
