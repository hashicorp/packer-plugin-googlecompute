// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"testing"
)

func TestDatasourceConfigure_EmptyProjectId(t *testing.T) {
	d := &Datasource{
		config: Config{
			Name: "test-secret",
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
		},
	}
	err := d.Configure()
	if err == nil {
		t.Fatal("expected error when name is missing")
	}
}

func TestDatasourceConfigure_Defaults(t *testing.T) {
	d := &Datasource{
		config: Config{
			Name:      "test-secret",
			ProjectId: "test-project",
		},
	}
	err := d.Configure()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if d.config.Version != "latest" {
		t.Fatalf("expected version to default to 'latest', got %s", d.config.Version)
	}
}
