// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package parametermanager

import (
	"testing"
)

func TestDatasourceConfigure_EmptyProjectId(t *testing.T) {
	d := &Datasource{
		config: Config{
			Name:    "test-secret",
			Version: "1",
		},
	}
	err := d.Configure()
	if err == nil {
		t.Fatal("expected error when project_id is missing")
	}
}

func TestDatasourceConfigure_EmptyName(t *testing.T) {
	d := &Datasource{
		config: Config{
			ProjectId: "test-project",
			Version:   "1",
		},
	}
	err := d.Configure()
	if err == nil {
		t.Fatal("expected error when name is missing")
	}
}

func TestDatasourceConfigure_EmptyVersion(t *testing.T) {
	d := &Datasource{
		config: Config{
			Name:      "test-secret",
			ProjectId: "test-project",
		},
	}
	err := d.Configure()
	if err == nil {
		t.Fatal("expected error when version is missing")
	}
}

func TestDatasourceConfigure_Defaults(t *testing.T) {
	d := &Datasource{
		config: Config{
			Name:      "test-secret",
			ProjectId: "test-project",
			Version:   "1",
		},
	}
	err := d.Configure()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
