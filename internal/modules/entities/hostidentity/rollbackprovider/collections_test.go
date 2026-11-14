package rollbackprovider

import (
	"context"
	"encoding/base64"
	"errors"
	"maps"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestParseCollectionTargetAcceptsCanonicalOwnerIdentities(t *testing.T) {
	t.Parallel()
	incidentID, recordID := uuid.New(), uuid.MustParse("00000000-0000-4000-8000-000000000101")
	tests := []rollbackcontract.NonRowTarget{
		canonicalAliasTargetFixture(incidentID, recordID),
		canonicalPreservedTargetFixture(incidentID, recordID, "exact_match_reuse"),
	}
	for index := range tests {
		identity, err := parseCollectionTarget(tests[index])
		if err != nil || identity.recordID != recordID || identity.incidentID != incidentID || identity.entityType != "host" {
			t.Fatalf("parse canonical target = %#v, %v", identity, err)
		}
	}
}

func TestParseCollectionTargetRejectsMalformedOrMismatchedIdentity(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		mutate func(*rollbackcontract.NonRowTarget)
	}{
		{"partial", func(target *rollbackcontract.NonRowTarget) { delete(target.AfterValue, "created_at") }},
		{"legacy target", func(target *rollbackcontract.NonRowTarget) {
			target.TargetID = "entity_alias:" + base64.RawURLEncoding.EncodeToString([]byte(target.AfterValue["record_id"].(string))) + ":aG9zdA:TG9zZXIgQWxpYXM"
		}},
		{"wrong physical row", func(target *rollbackcontract.NonRowTarget) { target.TargetID = "entity_alias:" + uuid.NewString() }},
		{"empty target", func(target *rollbackcontract.NonRowTarget) { target.TargetID = "" }},
		{"wrong target", func(target *rollbackcontract.NonRowTarget) { target.TargetID = "entity_alias:wrong" }},
		{"wrong normalization", func(target *rollbackcontract.NonRowTarget) { target.AfterValue["normalized_text"] = "not-canonical" }},
		{"wrong classification", func(target *rollbackcontract.NonRowTarget) { target.AfterValue["classification"] = "exact_match_reuse" }},
		{"wrong entity type", func(target *rollbackcontract.NonRowTarget) { target.AfterValue["entity_type"] = "evidence" }},
		{"invalid incident", func(target *rollbackcontract.NonRowTarget) { target.AfterValue["incident_id"] = "invalid" }},
		{"invalid record", func(target *rollbackcontract.NonRowTarget) { target.AfterValue["record_id"] = "invalid" }},
		{"missing value", func(target *rollbackcontract.NonRowTarget) { target.AfterValue = nil }},
		{"unsupported operation", func(target *rollbackcontract.NonRowTarget) { target.OperationKind = "patch" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := canonicalAliasTargetFixture(uuid.New(), uuid.New())
			// Each case starts with an admitted target and changes one predicate only.
			if _, err := parseCollectionTarget(target); err != nil {
				t.Fatalf("valid control: %v", err)
			}
			test.mutate(&target)
			if _, err := parseCollectionTarget(target); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
				t.Fatalf("parse malformed target error = %v", err)
			}
		})
	}
}

func TestParseCollectionTargetAcceptsPhysicalAliasDeleteIdentity(t *testing.T) {
	t.Parallel()
	incidentID, recordID := uuid.New(), uuid.New()
	target := canonicalAliasTargetFixture(incidentID, recordID)
	aliasID := uuid.MustParse(target.AfterValue["entity_alias_id"].(string))
	target.OperationKind = "delete"
	target.AfterValue["entity_type"] = "identity"
	target.AfterValue["raw_text"] = "Analyst Alias"
	target.AfterValue["normalized_text"] = "Analyst Alias"
	target.BeforeValue = maps.Clone(target.AfterValue)
	target.AfterValue["deleted_at"] = "2026-11-11T12:01:00Z"

	identity, err := parseCollectionTarget(target)
	if err != nil || identity.rowID != aliasID || identity.recordID != recordID {
		t.Fatalf("parse physical alias delete target = %#v, %v", identity, err)
	}
}

func TestCollectionClaimPreparationNomination(t *testing.T) {
	t.Parallel()
	t.Run("only exact preserved identifiers nominate claim preparation", testCollectionClaimPreparationNomination)
}

func testCollectionClaimPreparationNomination(t *testing.T) {
	incidentID, recordID := uuid.New(), uuid.MustParse("00000000-0000-4000-8000-000000000101")
	provider := CollectionProvider{}
	exact := canonicalPreservedTargetFixture(incidentID, recordID, "exact_match_reuse")
	identity, err := parseCollectionTarget(exact)
	if err != nil || identity.rowID.String() != exact.AfterValue["entity_preserved_identifier_id"] {
		t.Fatalf("physical preserved identity=%#v error=%v", identity, err)
	}
	got, err := provider.IdentifierClaimRecordTx(context.Background(), nil, exact)
	if err != nil || got != recordID {
		t.Fatalf("exact preserved identifier claim record = %s, %v; want %s", got, err, recordID)
	}

	suggestion := canonicalPreservedTargetFixture(incidentID, recordID, "suggestion_only")
	got, err = provider.IdentifierClaimRecordTx(context.Background(), nil, suggestion)
	if err != nil || got != uuid.Nil {
		t.Fatalf("suggestion-only claim record = %s, %v; want nil", got, err)
	}
	for _, defect := range []string{"partial", "logical_target", "wrong_row"} {
		target := canonicalPreservedTargetFixture(incidentID, recordID, "exact_match_reuse")
		switch defect {
		case "partial":
			delete(target.AfterValue, "created_at")
		case "logical_target":
			target.TargetID = "entity_preserved_identifier:MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMTAx:aG9zdA:ZnFkbg:bG9zZXIuZXhhbXBsZS50ZXN0:ZXhhY3RfbWF0Y2hfcmV1c2U"
		case "wrong_row":
			target.TargetID = "entity_preserved_identifier:" + uuid.NewString()
		}
		if _, err := provider.IdentifierClaimRecordTx(context.Background(), nil, target); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("%s admitted: %v", defect, err)
		}
	}

}

func canonicalPreservedTargetFixture(incidentID, recordID uuid.UUID, classification string) rollbackcontract.NonRowTarget {
	rowID := uuid.New()
	return rollbackcontract.NonRowTarget{
		TargetKind: "entity_preserved_identifier", TargetID: "entity_preserved_identifier:" + rowID.String(), OperationKind: "create", IncidentID: incidentID,
		AfterValue: map[string]any{
			"entity_preserved_identifier_id": rowID.String(), "incident_id": incidentID.String(), "record_id": recordID.String(),
			"entity_type": "host", "identifier_type": "fqdn", "raw_value": "LOSER.EXAMPLE.TEST", "normalized_value": "loser.example.test",
			"classification": classification, "created_by_user_id": uuid.NewString(), "created_at": "2026-11-11T12:00:00Z", "deleted_at": nil,
		},
	}
}

func canonicalAliasTargetFixture(incidentID, recordID uuid.UUID) rollbackcontract.NonRowTarget {
	id := uuid.New()
	return rollbackcontract.NonRowTarget{TargetKind: "entity_alias", TargetID: "entity_alias:" + id.String(), OperationKind: "create", IncidentID: incidentID,
		AfterValue: map[string]any{"entity_alias_id": id.String(), "incident_id": incidentID.String(), "record_id": recordID.String(), "entity_type": "host",
			"raw_text": "Loser Alias", "normalized_text": "Loser Alias", "classification": "suggestion_only", "created_by_user_id": uuid.NewString(), "created_at": "2026-11-11T12:00:00Z", "deleted_at": nil}}
}
