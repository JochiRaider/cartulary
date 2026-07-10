package rollbackprovider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

type CollectionProvider struct{}

var _ rollbackcontract.NonRowTargetProvider = CollectionProvider{}

func NewCollectionProvider() CollectionProvider { return CollectionProvider{} }

type collectionIdentity struct {
	targetKind      string
	incidentID      uuid.UUID
	recordID        uuid.UUID
	entityType      string
	rawValue        string
	normalizedValue string
	identifierType  string
	classification  string
}

func (p CollectionProvider) DescribeTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	identity, err := parseCollectionTarget(request.Target)
	if err != nil {
		return rollbackcontract.TargetDescriptor{}, err
	}
	descriptor := rollbackcontract.TargetDescriptor{
		AffectedRecordIDs:      []uuid.UUID{identity.recordID},
		RequiresWholeChangeSet: true,
	}
	if request.Target.IncidentID != identity.incidentID {
		return descriptor, rollbackcontract.ErrTargetNotFound
	}
	var incidentID uuid.UUID
	var recordType string
	if err := tx.QueryRow(ctx, `SELECT incident_id, record_type FROM records WHERE record_id = $1`, identity.recordID).Scan(&incidentID, &recordType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return descriptor, rollbackcontract.ErrTargetNotFound
		}
		return descriptor, err
	}
	if incidentID != identity.incidentID || recordType != identity.entityType {
		return descriptor, rollbackcontract.ErrTargetNotFound
	}
	if _, _, err := loadActiveCollectionValueTx(ctx, tx, identity); err != nil {
		return descriptor, err
	}
	return descriptor, nil
}

func (p CollectionProvider) ApplyInverseTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
	descriptor, err := p.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: request.Target})
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	identity, err := parseCollectionTarget(request.Target)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	rowID, before, err := loadActiveCollectionValueTx(ctx, tx, identity)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	if err := tombstoneCollectionTx(ctx, tx, identity.targetKind, rowID, request.Now); err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	after, err := loadCollectionValueByIDTx(ctx, tx, identity.targetKind, rowID)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	return rollbackcontract.ApplyInverseResult{
		AffectedRecordIDs: descriptor.AffectedRecordIDs,
		BeforeValue:       before,
		AfterValue:        after,
		ChangedFieldKeys: map[uuid.UUID][]string{
			identity.recordID: collectionChangedFieldKeys(identity),
		},
	}, nil
}

func parseCollectionTarget(target rollbackcontract.NonRowTarget) (collectionIdentity, error) {
	if target.OperationKind != "create" || target.AfterValue == nil {
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	identity := collectionIdentity{targetKind: target.TargetKind}
	var err error
	if identity.incidentID, err = requiredCollectionUUID(target.AfterValue, "incident_id"); err != nil {
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	if identity.recordID, err = requiredCollectionUUID(target.AfterValue, "record_id"); err != nil {
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	identity.entityType = requiredCollectionText(target.AfterValue, "entity_type")
	if identity.entityType != "host" && identity.entityType != "identity" {
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	switch target.TargetKind {
	case "entity_alias":
		identity.rawValue = rawCollectionText(target.AfterValue, "raw_text")
		identity.normalizedValue = requiredCollectionText(target.AfterValue, "normalized_text")
		identity.classification = requiredCollectionText(target.AfterValue, "classification")
		normalized, valid := fieldnorm.NormalizeLine(identity.rawValue)
		if !valid || normalized != identity.normalizedValue || identity.classification != "suggestion_only" {
			return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
	case "entity_preserved_identifier":
		identity.identifierType = requiredCollectionText(target.AfterValue, "identifier_type")
		identity.rawValue = rawCollectionText(target.AfterValue, "raw_value")
		identity.normalizedValue = requiredCollectionText(target.AfterValue, "normalized_value")
		identity.classification = requiredCollectionText(target.AfterValue, "classification")
		normalized, valid := fieldnorm.NormalizeIdentifier(identity.identifierType, identity.rawValue)
		if !valid || normalized != identity.normalizedValue || !validIdentifierType(identity.entityType, identity.identifierType) || !validIdentifierClassification(identity.classification) {
			return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
		}
	default:
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	if target.TargetID != collectionTargetID(identity) {
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	return identity, nil
}

func loadActiveCollectionValueTx(ctx context.Context, tx pgx.Tx, identity collectionIdentity) (uuid.UUID, map[string]any, error) {
	var rows pgx.Rows
	var err error
	switch identity.targetKind {
	case "entity_alias":
		rows, err = tx.Query(ctx, `
SELECT entity_alias_id,
       raw_text,
       jsonb_build_object(
           'entity_alias_id', entity_alias_id::text,
           'incident_id', incident_id::text,
           'record_id', record_id::text,
           'entity_type', entity_type,
           'raw_text', raw_text,
           'normalized_text', normalized_text,
           'classification', classification,
           'created_by_user_id', created_by_user_id::text,
           'created_at', created_at,
           'deleted_at', deleted_at
       )
  FROM entity_aliases
 WHERE incident_id = $1
   AND record_id = $2
   AND entity_type = $3
   AND normalized_text = $4
   AND classification = $5
   AND deleted_at IS NULL
 ORDER BY entity_alias_id
 LIMIT 2
`, identity.incidentID, identity.recordID, identity.entityType, identity.normalizedValue, identity.classification)
	case "entity_preserved_identifier":
		rows, err = tx.Query(ctx, `
SELECT entity_preserved_identifier_id,
       raw_value,
       jsonb_build_object(
           'entity_preserved_identifier_id', entity_preserved_identifier_id::text,
           'incident_id', incident_id::text,
           'record_id', record_id::text,
           'entity_type', entity_type,
           'identifier_type', identifier_type,
           'raw_value', raw_value,
           'normalized_value', normalized_value,
           'classification', classification,
           'created_by_user_id', created_by_user_id::text,
           'created_at', created_at,
           'deleted_at', deleted_at
       )
  FROM entity_preserved_identifiers
 WHERE incident_id = $1
   AND record_id = $2
   AND entity_type = $3
   AND identifier_type = $4
   AND normalized_value = $5
   AND classification = $6
   AND deleted_at IS NULL
 ORDER BY entity_preserved_identifier_id
 LIMIT 2
`, identity.incidentID, identity.recordID, identity.entityType, identity.identifierType, identity.normalizedValue, identity.classification)
	default:
		return uuid.Nil, nil, rollbackcontract.ErrTargetNotReversible
	}
	if err != nil {
		return uuid.Nil, nil, err
	}
	defer rows.Close()
	var rowID uuid.UUID
	var rawValue string
	var rawJSON []byte
	count := 0
	for rows.Next() {
		count++
		if err := rows.Scan(&rowID, &rawValue, &rawJSON); err != nil {
			return uuid.Nil, nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return uuid.Nil, nil, err
	}
	if count > 1 {
		return uuid.Nil, nil, rollbackcontract.ErrTargetNotReversible
	}
	if count == 0 {
		exists, err := collectionTargetExistsTx(ctx, tx, identity)
		if err != nil {
			return uuid.Nil, nil, err
		}
		if exists {
			return uuid.Nil, nil, rollbackcontract.ErrStaleTarget
		}
		return uuid.Nil, nil, rollbackcontract.ErrTargetNotFound
	}
	if rawValue != identity.rawValue {
		return uuid.Nil, nil, rollbackcontract.ErrStaleTarget
	}
	var value map[string]any
	if err := json.Unmarshal(rawJSON, &value); err != nil {
		return uuid.Nil, nil, err
	}
	return rowID, value, nil
}

func collectionTargetExistsTx(ctx context.Context, tx pgx.Tx, identity collectionIdentity) (bool, error) {
	var exists bool
	switch identity.targetKind {
	case "entity_alias":
		err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM entity_aliases
     WHERE incident_id = $1 AND record_id = $2 AND entity_type = $3
       AND normalized_text = $4 AND classification = $5
)
`, identity.incidentID, identity.recordID, identity.entityType, identity.normalizedValue, identity.classification).Scan(&exists)
		return exists, err
	case "entity_preserved_identifier":
		err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM entity_preserved_identifiers
     WHERE incident_id = $1 AND record_id = $2 AND entity_type = $3
       AND identifier_type = $4 AND normalized_value = $5 AND classification = $6
)
`, identity.incidentID, identity.recordID, identity.entityType, identity.identifierType, identity.normalizedValue, identity.classification).Scan(&exists)
		return exists, err
	default:
		return false, rollbackcontract.ErrTargetNotReversible
	}
}

func tombstoneCollectionTx(ctx context.Context, tx pgx.Tx, targetKind string, rowID uuid.UUID, now time.Time) error {
	var rowsAffected int64
	switch targetKind {
	case "entity_alias":
		tag, err := tx.Exec(ctx, `UPDATE entity_aliases SET deleted_at = $2 WHERE entity_alias_id = $1 AND deleted_at IS NULL`, rowID, now.UTC())
		if err != nil {
			return err
		}
		rowsAffected = tag.RowsAffected()
	case "entity_preserved_identifier":
		tag, err := tx.Exec(ctx, `UPDATE entity_preserved_identifiers SET deleted_at = $2 WHERE entity_preserved_identifier_id = $1 AND deleted_at IS NULL`, rowID, now.UTC())
		if err != nil {
			return err
		}
		rowsAffected = tag.RowsAffected()
	default:
		return rollbackcontract.ErrTargetNotReversible
	}
	if rowsAffected != 1 {
		return rollbackcontract.ErrStaleTarget
	}
	return nil
}

func loadCollectionValueByIDTx(ctx context.Context, tx pgx.Tx, targetKind string, rowID uuid.UUID) (map[string]any, error) {
	var rawJSON []byte
	var err error
	switch targetKind {
	case "entity_alias":
		err = tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'entity_alias_id', entity_alias_id::text,
    'incident_id', incident_id::text,
    'record_id', record_id::text,
    'entity_type', entity_type,
    'raw_text', raw_text,
    'normalized_text', normalized_text,
    'classification', classification,
    'created_by_user_id', created_by_user_id::text,
    'created_at', created_at,
    'deleted_at', deleted_at
)
  FROM entity_aliases
 WHERE entity_alias_id = $1
`, rowID).Scan(&rawJSON)
	case "entity_preserved_identifier":
		err = tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'entity_preserved_identifier_id', entity_preserved_identifier_id::text,
    'incident_id', incident_id::text,
    'record_id', record_id::text,
    'entity_type', entity_type,
    'identifier_type', identifier_type,
    'raw_value', raw_value,
    'normalized_value', normalized_value,
    'classification', classification,
    'created_by_user_id', created_by_user_id::text,
    'created_at', created_at,
    'deleted_at', deleted_at
)
  FROM entity_preserved_identifiers
 WHERE entity_preserved_identifier_id = $1
`, rowID).Scan(&rawJSON)
	default:
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, rollbackcontract.ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(rawJSON, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func collectionChangedFieldKeys(identity collectionIdentity) []string {
	switch identity.targetKind {
	case "entity_alias":
		return []string{identity.entityType + ".aliases"}
	case "entity_preserved_identifier":
		if identity.classification == "exact_match_reuse" {
			return []string{identity.entityType + ".reusable_identifiers"}
		}
	}
	return nil
}

func collectionTargetID(identity collectionIdentity) string {
	components := []string{identity.targetKind, identity.recordID.String(), identity.entityType}
	if identity.targetKind == "entity_preserved_identifier" {
		components = append(components, identity.identifierType, identity.normalizedValue, identity.classification)
	} else {
		components = append(components, identity.normalizedValue)
	}
	for index := 1; index < len(components); index++ {
		components[index] = base64.RawURLEncoding.EncodeToString([]byte(components[index]))
	}
	return strings.Join(components, ":")
}

func validIdentifierType(entityType string, identifierType string) bool {
	allowed := map[string][]string{
		"host":     {"aad_device_id", "fqdn", "hostname"},
		"identity": {"aad_object_id", "sid", "upn", "email", "sam_account_name"},
	}
	for _, value := range allowed[entityType] {
		if value == identifierType {
			return true
		}
	}
	return false
}

func validIdentifierClassification(value string) bool {
	switch value {
	case "exact_match_reuse", "suggestion_only", "provenance_only":
		return true
	default:
		return false
	}
}

func requiredCollectionUUID(value map[string]any, key string) (uuid.UUID, error) {
	raw := requiredCollectionText(value, key)
	if raw == "" {
		return uuid.Nil, errors.New("missing uuid")
	}
	return uuid.Parse(raw)
}

func requiredCollectionText(value map[string]any, key string) string {
	raw, _ := value[key].(string)
	return strings.TrimSpace(raw)
}

func rawCollectionText(value map[string]any, key string) string {
	raw, _ := value[key].(string)
	return raw
}
