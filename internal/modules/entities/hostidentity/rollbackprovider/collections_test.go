package rollbackprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestParseCollectionTargetAcceptsCanonicalOwnerIdentities(t *testing.T) {
	t.Parallel()
	incidentID, recordID := uuid.New(), uuid.New()
	tests := []rollbackcontract.NonRowTarget{
		{
			TargetKind: "entity_alias", OperationKind: "create",
			AfterValue: map[string]any{
				"incident_id": incidentID.String(), "record_id": recordID.String(), "entity_type": "host",
				"raw_text": "Loser Alias", "normalized_text": "Loser Alias", "classification": "suggestion_only",
			},
		},
		{
			TargetKind: "entity_preserved_identifier", OperationKind: "create",
			AfterValue: map[string]any{
				"incident_id": incidentID.String(), "record_id": recordID.String(), "entity_type": "host",
				"identifier_type": "fqdn", "raw_value": "LOSER.EXAMPLE.TEST", "normalized_value": "loser.example.test", "classification": "exact_match_reuse",
			},
		},
	}
	for index := range tests {
		identity, err := parseCollectionTargetWithCanonicalID(tests[index])
		if err != nil || identity.recordID != recordID {
			t.Fatalf("parse canonical target = %#v, %v", identity, err)
		}
	}
}

func TestParseCollectionTargetRejectsMalformedOrMismatchedIdentity(t *testing.T) {
	t.Parallel()
	incidentID, recordID := uuid.New(), uuid.New()
	target := rollbackcontract.NonRowTarget{
		TargetKind: "entity_alias", TargetID: "entity_alias:wrong", OperationKind: "create",
		AfterValue: map[string]any{
			"incident_id": incidentID.String(), "record_id": recordID.String(), "entity_type": "host",
			"raw_text": " Alias ", "normalized_text": "not-canonical", "classification": "suggestion_only",
		},
	}
	if _, err := parseCollectionTarget(target); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("parse malformed target error = %v", err)
	}
}

func TestParseCollectionTargetAcceptsPhysicalAliasDeleteIdentity(t *testing.T) {
	t.Parallel()
	incidentID, recordID, aliasID := uuid.New(), uuid.New(), uuid.New()
	target := rollbackcontract.NonRowTarget{
		TargetKind: "entity_alias", TargetID: "entity_alias:" + aliasID.String(), OperationKind: "delete",
		BeforeValue: map[string]any{
			"entity_alias_id": aliasID.String(), "incident_id": incidentID.String(), "record_id": recordID.String(), "entity_type": "identity",
			"raw_text": "Analyst Alias", "normalized_text": "Analyst Alias", "classification": "suggestion_only",
		},
		AfterValue: map[string]any{"deleted_at": "2026-07-11T00:00:00Z"},
	}
	identity, err := parseCollectionTarget(target)
	if err != nil || identity.rowID != aliasID || identity.recordID != recordID {
		t.Fatalf("parse physical alias delete target = %#v, %v", identity, err)
	}
}

func TestCollectionChangedFieldsRemainOwnerDefined(t *testing.T) {
	t.Parallel()
	alias := collectionIdentity{targetKind: "entity_alias", entityType: "identity"}
	if got := collectionChangedFieldKeys(alias); len(got) != 1 || got[0] != "identity.aliases" {
		t.Fatalf("alias changed keys = %v", got)
	}
	preserved := collectionIdentity{targetKind: "entity_preserved_identifier", entityType: "host", classification: "exact_match_reuse"}
	if got := collectionChangedFieldKeys(preserved); len(got) != 1 || got[0] != "host.reusable_identifiers" {
		t.Fatalf("identifier changed keys = %v", got)
	}
	t.Run("only exact preserved identifiers nominate claim preparation", testCollectionClaimPreparationNomination)
}

func testCollectionClaimPreparationNomination(t *testing.T) {
	incidentID, recordID := uuid.New(), uuid.New()
	provider := CollectionProvider{}
	exact := rollbackcontract.NonRowTarget{
		TargetKind: "entity_preserved_identifier", OperationKind: "create", IncidentID: incidentID,
		AfterValue: map[string]any{
			"incident_id": incidentID.String(), "record_id": recordID.String(), "entity_type": "host",
			"identifier_type": "fqdn", "raw_value": "Example.Test", "normalized_value": "example.test",
			"classification": "exact_match_reuse",
		},
	}
	identity, err := parseCollectionTargetWithoutTargetID(exact)
	if err != nil {
		t.Fatalf("build exact target: %v", err)
	}
	exact.TargetID = collectionTargetID(identity)
	got, err := provider.IdentifierClaimRecordTx(context.Background(), nil, exact)
	if err != nil || got != recordID {
		t.Fatalf("exact preserved identifier claim record = %s, %v; want %s", got, err, recordID)
	}

	suggestion := rollbackcontract.NonRowTarget{
		TargetKind: "entity_preserved_identifier", OperationKind: "create", IncidentID: incidentID,
		AfterValue: map[string]any{
			"incident_id": incidentID.String(), "record_id": recordID.String(), "entity_type": "host",
			"identifier_type": "fqdn", "raw_value": "Example.Test", "normalized_value": "example.test",
			"classification": "suggestion_only",
		},
	}
	identity, err = parseCollectionTargetWithoutTargetID(suggestion)
	if err != nil {
		t.Fatalf("build suggestion target: %v", err)
	}
	suggestion.TargetID = collectionTargetID(identity)
	got, err = provider.IdentifierClaimRecordTx(context.Background(), nil, suggestion)
	if err != nil || got != uuid.Nil {
		t.Fatalf("suggestion-only claim record = %s, %v; want nil", got, err)
	}
}

func parseCollectionTargetWithCanonicalID(target rollbackcontract.NonRowTarget) (collectionIdentity, error) {
	identity, err := parseCollectionTargetWithoutTargetID(target)
	if err != nil {
		return collectionIdentity{}, err
	}
	target.TargetID = collectionTargetID(identity)
	return parseCollectionTarget(target)
}

func parseCollectionTargetWithoutTargetID(target rollbackcontract.NonRowTarget) (collectionIdentity, error) {
	target.TargetID = "placeholder"
	identity := collectionIdentity{targetKind: target.TargetKind}
	var err error
	if identity.incidentID, err = requiredCollectionUUID(target.AfterValue, "incident_id"); err != nil {
		return collectionIdentity{}, err
	}
	if identity.recordID, err = requiredCollectionUUID(target.AfterValue, "record_id"); err != nil {
		return collectionIdentity{}, err
	}
	identity.entityType = requiredCollectionText(target.AfterValue, "entity_type")
	if target.TargetKind == "entity_alias" {
		identity.rawValue = rawCollectionText(target.AfterValue, "raw_text")
		identity.normalizedValue = requiredCollectionText(target.AfterValue, "normalized_text")
		identity.classification = requiredCollectionText(target.AfterValue, "classification")
	} else {
		identity.identifierType = requiredCollectionText(target.AfterValue, "identifier_type")
		identity.rawValue = rawCollectionText(target.AfterValue, "raw_value")
		identity.normalizedValue = requiredCollectionText(target.AfterValue, "normalized_value")
		identity.classification = requiredCollectionText(target.AfterValue, "classification")
	}
	return identity, nil
}
