package ownerfacade

import "testing"

func TestCharacterizationKnownNonConformanceUsesInternalUseNullToken(t *testing.T) {
	t.Parallel()

	value, include, err := NormalizeImportScalar(
		"cartulary.view.hosts.v1",
		"host.location",
		"",
		"use_null",
	)
	if err != nil || !include || value.Kind != "null" {
		t.Fatalf("current use_null behavior changed: value=%#v include=%v err=%v", value, include, err)
	}

	if _, _, err := NormalizeImportScalar(
		"cartulary.view.hosts.v1",
		"host.location",
		"",
		"write_null",
	); err == nil {
		t.Fatal("current owner facade unexpectedly accepts write_null before RS-08")
	}
}
