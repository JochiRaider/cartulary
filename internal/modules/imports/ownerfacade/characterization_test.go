package ownerfacade

import (
	"errors"
	"testing"
)

func TestNormalizeImportScalarUsesCanonicalEmptyValuePolicy(t *testing.T) {
	t.Parallel()

	value, include, err := NormalizeImportScalar(
		"cartulary.view.hosts.v1",
		"host.location",
		"",
		"write_null",
	)
	if err != nil || !include || value.Kind != "null" {
		t.Fatalf("write_null = value %#v include=%v err=%v", value, include, err)
	}

	value, include, err = NormalizeImportScalar(
		"cartulary.view.hosts.v1",
		"host.location",
		"",
		"omit_field",
	)
	if err != nil || include || value.Kind != "" {
		t.Fatalf("omit_field = value %#v include=%v err=%v", value, include, err)
	}

	if _, _, err := NormalizeImportScalar(
		"cartulary.view.hosts.v1",
		"host.location",
		"",
		"use_null",
	); err == nil {
		t.Fatal("legacy use_null token unexpectedly accepted")
	}
}

func TestNormalizeImportScalarRejectsNullForNonNullableCreateField(t *testing.T) {
	t.Parallel()

	_, _, err := NormalizeImportScalar(
		"cartulary.view.indicators.v1",
		"indicator.display_value",
		"",
		"write_null",
	)
	var ownerErr *ImportOwnerCreateError
	if !errors.As(err, &ownerErr) ||
		ownerErr.OwnerCode != ImportOwnerCreateValidationFailed ||
		ownerErr.ReasonCode != "field_not_nullable" ||
		ownerErr.Field != "indicator.display_value" ||
		ownerErr.Guard != "clearable" ||
		ownerErr.Retryable {
		t.Fatalf("non-nullable write_null error = %#v", err)
	}
	detail, ok := ImportOwnerCreateErrorDetail(err)
	if !ok {
		t.Fatal("typed owner error did not produce safe detail")
	}
	safeDetails, _ := detail["safe_details"].(map[string]any)
	if safeDetails["reason_code"] != "field_not_nullable" ||
		safeDetails["field"] != "indicator.display_value" ||
		safeDetails["guard"] != "clearable" {
		t.Fatalf("safe detail = %#v", detail)
	}
}

func TestPartyMatchConflictDetailIsClosedSortedAndValueFree(t *testing.T) {
	t.Parallel()

	err := NewPartyMatchConflictError(
		"cross_key_exact_match",
		[]string{"party.external_ref", "party.primary_email"},
		errors.New("private source value and storage diagnostic"),
	)
	detail, ok := ImportOwnerCreateErrorDetail(err)
	if !ok {
		t.Fatal("valid Party match conflict did not produce safe detail")
	}
	safeDetails, _ := detail["safe_details"].(map[string]any)
	fields, _ := safeDetails["conflicting_field_keys"].([]string)
	if detail["owner_code"] != ImportOwnerCreateValidationFailed ||
		detail["retryable"] != false ||
		safeDetails["reason_code"] != "party_match_conflict" ||
		safeDetails["party_reason_code"] != "cross_key_exact_match" ||
		len(fields) != 2 || fields[0] != "party.external_ref" || fields[1] != "party.primary_email" {
		t.Fatalf("Party safe detail = %#v", detail)
	}
	if _, leaked := safeDetails["cause"]; leaked {
		t.Fatalf("Party safe detail leaked cause: %#v", detail)
	}

	for _, invalid := range []*ImportOwnerCreateError{
		NewPartyMatchConflictError("unknown", []string{"party.primary_email"}, nil),
		NewPartyMatchConflictError("exact_match_key_claimed", []string{"party.primary_email", "party.external_ref"}, nil),
		NewPartyMatchConflictError("exact_match_key_claimed", []string{"party.primary_email", "party.primary_email"}, nil),
	} {
		if detail, ok := ImportOwnerCreateErrorDetail(invalid); ok || detail != nil {
			t.Fatalf("invalid Party match conflict produced detail: %#v", detail)
		}
	}
}
