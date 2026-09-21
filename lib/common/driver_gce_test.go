// Copyright IBM Corp. 2013, 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
)

func TestIsRetryableAPIError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "conflicting concurrent operation", err: &googleapi.Error{Code: 409}, want: true},
		{name: "rate limited", err: &googleapi.Error{Code: 429}, want: true},
		{name: "backend error", err: &googleapi.Error{Code: 500}, want: true},
		{name: "bad gateway", err: &googleapi.Error{Code: 502}, want: true},
		{name: "service unavailable", err: &googleapi.Error{Code: 503}, want: true},
		{name: "gateway timeout", err: &googleapi.Error{Code: 504}, want: true},
		{name: "not found", err: &googleapi.Error{Code: 404}, want: false},
		{name: "forbidden", err: &googleapi.Error{Code: 403}, want: false},
		{name: "bad request", err: &googleapi.Error{Code: 400}, want: false},
		{name: "wrapped retryable api error", err: fmt.Errorf("deprecate: %w", &googleapi.Error{Code: 503}), want: true},
		{name: "unexpected EOF", err: io.ErrUnexpectedEOF, want: true},
		{name: "context canceled", err: context.Canceled, want: false},
		{name: "context deadline exceeded", err: context.DeadlineExceeded, want: false},
		{name: "wrapped context cancellation", err: fmt.Errorf("deprecate: %w", context.Canceled), want: false},
		{name: "generic transport failure", err: errors.New("connection reset by peer"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isRetryableAPIError(tt.err))
		})
	}
}

func TestSetImageDeprecationStatus_nilStatus(t *testing.T) {
	d := &driverGCE{}
	assert.Error(t, d.SetImageDeprecationStatus("project", "image", nil))
}
