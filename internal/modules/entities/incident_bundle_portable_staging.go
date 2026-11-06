package entities

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

// These transaction-local indexes hold only source identities, relationships and
// normalized claims. PostgreSQL owns their work memory and spill. Deferred owner
// guards prevent committed scratch rows; no runtime DDL privileges are required.
func validateEntityCrossRowsBeforeApply(ctx context.Context, tx pgx.Tx, prepared preparedEntityImport, input sourceport.ImportContext) error {
	if _, err := tx.Exec(ctx, `INSERT INTO public.entities_bundle_validation_binding VALUES ($1,$2,$3,$4)`, input.OperationID, input.IncidentID, input.ActorUserID, entityPreparedDigest(prepared)); err != nil {
		return err
	}
	if err := prepared.hosts(ctx, input, func(row portableHostRow) error {
		_, err := tx.Exec(ctx, `INSERT INTO public.entities_bundle_validation_nodes VALUES ('host',$1,$2)`, row.RecordID, row.State)
		return err
	}); err != nil {
		return err
	}
	if err := prepared.identities(ctx, input, func(row portableIdentityRow) error {
		_, err := tx.Exec(ctx, `INSERT INTO public.entities_bundle_validation_nodes VALUES ('identity',$1,$2)`, row.RecordID, row.State)
		return err
	}); err != nil {
		return err
	}
	if err := prepared.mentions(ctx, input, func(row portableMentionRow) error {
		_, err := tx.Exec(ctx, `INSERT INTO public.entities_bundle_validation_mentions VALUES ($1)`, row.MentionID)
		return err
	}); err != nil {
		return err
	}
	var failures entityFailureAccumulator
	addClaim := func(recordID uuid.UUID, kind, identifier string, raw *string) error {
		if raw == nil {
			return nil
		}
		normalized, ok := fieldnorm.NormalizeIdentifier(identifier, *raw)
		if !ok {
			failures.add(entitySourceIdentity)
			return nil
		}
		_, err := tx.Exec(ctx, `INSERT INTO public.entities_bundle_validation_claims VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`, recordID, kind, identifier, normalized)
		return err
	}
	checkNode := func(kind string, target, seed *uuid.UUID) error {
		if target != nil {
			var valid bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM public.entities_bundle_validation_nodes WHERE kind=$1 AND id=$2 AND state <> 'merged')`, kind, target).Scan(&valid); err != nil {
				return err
			}
			if !valid {
				failures.add(entityResolutionMerge)
			}
		}
		if seed != nil {
			var present bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM public.entities_bundle_validation_mentions WHERE id=$1)`, seed).Scan(&present); err != nil {
				return err
			}
			if !present {
				failures.add(entitySourceIdentity)
			}
		}
		return nil
	}
	if err := prepared.hosts(ctx, input, func(row portableHostRow) error {
		envelope, present, err := loadEntityPortableEnvelope(ctx, tx, input.IncidentID, row.RecordID)
		if err != nil {
			return err
		}
		if !present || !hostMatchesPortableEnvelope(row, envelope) {
			failures.add(entityEnvelopeTypeScope)
		}
		if err := checkNode("host", row.MergedIntoRecordID, row.SeedMentionID); err != nil {
			return err
		}
		if present && envelope.deletedAt == nil && row.State != "merged" {
			for _, claim := range []struct {
				kind  string
				value *string
			}{{"aad_device_id", row.AADDeviceID}, {"fqdn", row.FQDN}, {"hostname", row.Hostname}} {
				if err := addClaim(row.RecordID, "host", claim.kind, claim.value); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := prepared.identities(ctx, input, func(row portableIdentityRow) error {
		envelope, present, err := loadEntityPortableEnvelope(ctx, tx, input.IncidentID, row.RecordID)
		if err != nil {
			return err
		}
		if !present || !identityMatchesPortableEnvelope(row, envelope) {
			failures.add(entityEnvelopeTypeScope)
		}
		if err := checkNode("identity", row.MergedIntoRecordID, row.SeedMentionID); err != nil {
			return err
		}
		if present && envelope.deletedAt == nil && row.State != "merged" {
			for _, claim := range []struct {
				kind  string
				value *string
			}{{"aad_object_id", row.AADObjectID}, {"sid", row.SID}, {"upn", row.UPN}, {"email", row.Email}, {"sam_account_name", row.SAMAccountName}} {
				if err := addClaim(row.RecordID, "identity", claim.kind, claim.value); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := prepared.mentions(ctx, input, func(row portableMentionRow) error {
		_, present, err := loadEntityPortableEnvelope(ctx, tx, input.IncidentID, row.SourceRecordID)
		if err != nil {
			return err
		}
		if !present {
			failures.add(entityResolutionMerge)
		}
		if row.ResolvedRecordID != nil {
			target, present, err := loadEntityPortableEnvelope(ctx, tx, input.IncidentID, *row.ResolvedRecordID)
			if err != nil {
				return err
			}
			if !present || target.recordType != row.EntityType {
				failures.add(entityResolutionMerge)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	owner := func(kind string, recordID uuid.UUID) (bool, error) {
		var present, active bool
		err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM public.entities_bundle_validation_nodes WHERE kind=$1 AND id=$2), EXISTS (SELECT 1 FROM public.entities_bundle_validation_nodes n JOIN records r ON r.record_id=n.id WHERE n.kind=$1 AND n.id=$2 AND n.state IN ('stub','canonical') AND r.deleted_at IS NULL AND r.incident_id=$3)`, kind, recordID, input.IncidentID).Scan(&present, &active)
		if !present {
			failures.add(entitySameIncident)
		}
		return active, err
	}
	if err := prepared.preserved(ctx, input, func(row portablePreservedIdentifierRow) error {
		active, err := owner(row.EntityType, row.RecordID)
		if err != nil {
			return err
		}
		if row.DeletedAt == nil {
			if _, err := tx.Exec(ctx, `INSERT INTO public.entities_bundle_validation_associations VALUES ('preserved',$1,$2,$3,$4,$5)`, row.EntityType, row.RecordID, row.IdentifierType, row.NormalizedValue, row.Classification); err != nil {
				return err
			}
			if active && row.Classification == "exact_match_reuse" {
				return addClaim(row.RecordID, row.EntityType, row.IdentifierType, &row.RawValue)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := prepared.aliases(ctx, input, func(row portableAliasRow) error {
		if _, err := owner(row.EntityType, row.RecordID); err != nil {
			return err
		}
		if row.DeletedAt == nil {
			_, err := tx.Exec(ctx, `INSERT INTO public.entities_bundle_validation_associations VALUES ('alias',$1,$2,$3,'','')`, row.EntityType, row.RecordID, row.NormalizedText)
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	var conflict bool
	if err := tx.QueryRow(ctx, `SELECT
  EXISTS (SELECT 1 FROM public.entities_bundle_validation_associations GROUP BY kind,entity_type,record_id,key1,key2,key3 HAVING count(*) > 1)
  OR EXISTS (SELECT 1 FROM public.entities_bundle_validation_claims GROUP BY entity_type,identifier_type,normalized_value HAVING count(*) > 1)
  OR EXISTS (SELECT 1 FROM entity_active_identifier_claims WHERE incident_id=$1)`, input.IncidentID).Scan(&conflict); err != nil {
		return err
	}
	if conflict {
		failures.add(entityUnique)
	}
	if err := failures.err(); err != nil {
		return err
	}
	for _, statement := range []string{`DELETE FROM public.entities_bundle_validation_nodes WHERE transaction_id=pg_current_xact_id()`, `DELETE FROM public.entities_bundle_validation_mentions WHERE transaction_id=pg_current_xact_id()`, `DELETE FROM public.entities_bundle_validation_associations WHERE transaction_id=pg_current_xact_id()`} {
		if _, err := tx.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func entityPreparedDigest(prepared preparedEntityImport) string {
	raw, _ := json.Marshal(prepared.members.Members())
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:])
}

func validateStagedEntityClaims(ctx context.Context, tx pgx.Tx, prepared preparedEntityImport, input sourceport.ImportContext) error {
	var bindingMatches, claimsDiffer bool
	if err := tx.QueryRow(ctx, `SELECT count(*)=1 AND bool_and(operation_id=$1 AND incident_id=$2 AND actor_id=$3 AND source_digest=$4) FROM public.entities_bundle_validation_binding`, input.OperationID, input.IncidentID, input.ActorUserID, entityPreparedDigest(prepared)).Scan(&bindingMatches); err != nil {
		return err
	}
	if !bindingMatches {
		return entitySourceFailure(entitySourceIdentity)
	}
	if err := tx.QueryRow(ctx, `WITH actual AS (SELECT record_id,entity_type,identifier_type,normalized_value FROM entity_active_identifier_claims WHERE incident_id=$1), expected AS (SELECT record_id,entity_type,identifier_type,normalized_value FROM public.entities_bundle_validation_claims)
 SELECT EXISTS ((SELECT * FROM actual EXCEPT SELECT * FROM expected) UNION ALL (SELECT * FROM expected EXCEPT SELECT * FROM actual))`, input.IncidentID).Scan(&claimsDiffer); err != nil {
		return err
	}
	if claimsDiffer {
		return entitySourceFailure(entityUnique)
	}
	return nil
}
