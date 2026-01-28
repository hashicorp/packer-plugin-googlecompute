// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package image

import (
	"testing"
)

func TestDatasourceConfigure_EmptyProjectID(t *testing.T) {
	d := &Datasource{
		config: Config{
			Filters: "name=ubuntu",
		},
	}
	err := d.Configure()
	if err == nil {
		t.Fatal("expected error when project_id is missing")
	}
}

func TestDatasourceConfigure_EmptyFilters(t *testing.T) {
	d := &Datasource{
		config: Config{
			ProjectID: "test-project",
		},
	}
	err := d.Configure()
	if err == nil {
		t.Fatal("expected error when filters are missing")
	}
}

func TestDatasourceConfigure_ValidConfig(t *testing.T) {
	d := &Datasource{
		config: Config{
			ProjectID: "test-project",
			Filters:   "name=ubuntu",
		},
	}
	err := d.Configure()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
