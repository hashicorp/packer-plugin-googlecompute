// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ReservationAffinity

package common

import compute "google.golang.org/api/compute/v1"

// ReservationAffinity is the configuration structure for instance reservation
// affinity. It allows you to consume a specific reservation.
type ReservationAffinity struct {
	// ConsumeReservationType: Specifies the type of reservation from which this
	// instance can consume resources.
	// See https://cloud.google.com/compute/docs/instances/consuming-reserved-instances for examples.
	ConsumeReservationType string `mapstructure:"consume_reservation_type"`

	// Key: Corresponds to the label key of a reservation resource. To target a
	// SPECIFIC_RESERVATION by name, specify `compute.googleapis.com/reservation-name`
	// as the key.
	Key string `mapstructure:"key"`

	// Values: Corresponds to the label values of a reservation resource.
	Values []string `mapstructure:"values"`
}

// ComputeType converts the Packer-specific ReservationAffinity struct to the
// type required by the Google Cloud API client library.
func (r *ReservationAffinity) ComputeType() *compute.ReservationAffinity {
	if r == nil {
		return nil
	}
	return &compute.ReservationAffinity{
		ConsumeReservationType: r.ConsumeReservationType,
		Key:                    r.Key,
		Values:                 r.Values,
	}
}

