package rollbackprovider

import (
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
