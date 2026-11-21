package entities

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
)

// These transaction-local indexes hold only source identities, relationships and
// normalized claims. PostgreSQL owns their work memory and spill. Deferred owner
// guards prevent committed scratch rows; no runtime DDL privileges are required.
// Every check is set-based over staged rows; the failure accumulator keeps the
// declared precedence independent of row order.
func validateEntityCrossRowsBeforeApply(ctx context.Context, tx pgx.Tx, prepared preparedEntityImport, input sourceport.ImportContext) error {
	if _, err := tx.Exec(ctx, `INSERT INTO public.entities_bundle_validation_binding VALUES ($1,$2,$3,$4)`, input.OperationID, input.IncidentID, input.ActorUserID, prepared.sourceDigest); err != nil {
		return err
	}
	store := input.Preparation
	hosts, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.hostRows, hostNodeCodec{})
	if err != nil {
		return err
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"public", "entities_bundle_validation_nodes"}, entityNodeColumns, hosts)
	if err := hosts.Finish(written, err); err != nil {
		return err
	}
	identities, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.identityRows, identityNodeCodec{})
	if err != nil {
		return err
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"public", "entities_bundle_validation_nodes"}, entityNodeColumns, identities)
	if err := identities.Finish(written, err); err != nil {
		return err
	}
	mentions, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.mentionRows, mentionNodeCodec{})
	if err != nil {
		return err
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"public", "entities_bundle_validation_mentions"}, []string{"id"}, mentions)
	if err := mentions.Finish(written, err); err != nil {
		return err
	}
	var failures entityFailureAccumulator
	claims := &entityClaimBatch{}
	checkNodes := func(kind string, targets, seeds []*uuid.UUID) error {
		var mergeTargets, seedIDs []uuid.UUID
		for _, target := range targets {
			if target != nil {
				mergeTargets = append(mergeTargets, *target)
			}
		}
		for _, seed := range seeds {
			if seed != nil {
				seedIDs = append(seedIDs, *seed)
			}
		}
		var missingTarget, missingSeed bool
		if err := tx.QueryRow(ctx, `SELECT
  EXISTS (SELECT 1 FROM unnest($2::uuid[]) AS t(id) WHERE NOT EXISTS (SELECT 1 FROM public.entities_bundle_validation_nodes n WHERE n.kind=$1 AND n.id=t.id AND n.state <> 'merged')),
  EXISTS (SELECT 1 FROM unnest($3::uuid[]) AS s(id) WHERE NOT EXISTS (SELECT 1 FROM public.entities_bundle_validation_mentions m WHERE m.id=s.id))`, kind, mergeTargets, seedIDs).Scan(&missingTarget, &missingSeed); err != nil {
			return err
		}
		if missingTarget {
			failures.add(entityResolutionMerge)
		}
		if missingSeed {
			failures.add(entitySourceIdentity)
		}
		return nil
	}
	if err := readEntityBatches(ctx, prepared, input, prepared.hostRows, hostCodec{}, func(batch []stagedEntityHost) error {
		ids := make([]uuid.UUID, len(batch))
		targets, seeds := make([]*uuid.UUID, len(batch)), make([]*uuid.UUID, len(batch))
		for index, host := range batch {
			ids[index], targets[index], seeds[index] = host.Row.RecordID, host.Row.MergedIntoRecordID, host.Row.SeedMentionID
		}
		envelopes, err := loadEntityPortableEnvelopes(ctx, tx, input.IncidentID, ids)
		if err != nil {
			return err
		}
		for _, host := range batch {
			envelope, present := envelopes[host.Row.RecordID]
			if !present || !hostMatchesPortableEnvelope(host.Row, envelope) {
				failures.add(entityEnvelopeTypeScope)
			}
			if present && envelope.deletedAt == nil && host.Row.State != "merged" {
				claims.add(&failures, host.Row.RecordID, "host", host.Claims...)
			}
		}
		if err := claims.flush(ctx, tx); err != nil {
			return err
		}
		return checkNodes("host", targets, seeds)
	}); err != nil {
		return err
	}
	if err := readEntityBatches(ctx, prepared, input, prepared.identityRows, identityCodec{}, func(batch []stagedEntityIdentity) error {
		ids := make([]uuid.UUID, len(batch))
		targets, seeds := make([]*uuid.UUID, len(batch)), make([]*uuid.UUID, len(batch))
		for index, identity := range batch {
			ids[index], targets[index], seeds[index] = identity.Row.RecordID, identity.Row.MergedIntoRecordID, identity.Row.SeedMentionID
		}
		envelopes, err := loadEntityPortableEnvelopes(ctx, tx, input.IncidentID, ids)
		if err != nil {
			return err
		}
		for _, identity := range batch {
			envelope, present := envelopes[identity.Row.RecordID]
			if !present || !identityMatchesPortableEnvelope(identity.Row, envelope) {
				failures.add(entityEnvelopeTypeScope)
			}
			if present && envelope.deletedAt == nil && identity.Row.State != "merged" {
				claims.add(&failures, identity.Row.RecordID, "identity", identity.Claims...)
			}
		}
		if err := claims.flush(ctx, tx); err != nil {
			return err
		}
		return checkNodes("identity", targets, seeds)
	}); err != nil {
		return err
	}
	if err := readEntityBatches(ctx, prepared, input, prepared.mentionRows, mentionCodec{}, func(batch []portableMentionRow) error {
		ids := make([]uuid.UUID, 0, 2*len(batch))
		for _, mention := range batch {
			ids = append(ids, mention.SourceRecordID)
			if mention.ResolvedRecordID != nil {
				ids = append(ids, *mention.ResolvedRecordID)
			}
		}
		envelopes, err := loadEntityPortableEnvelopes(ctx, tx, input.IncidentID, ids)
		if err != nil {
			return err
		}
		for _, mention := range batch {
			if _, present := envelopes[mention.SourceRecordID]; !present {
				failures.add(entityResolutionMerge)
			}
			if mention.ResolvedRecordID != nil {
				if target, present := envelopes[*mention.ResolvedRecordID]; !present || target.recordType != mention.EntityType {
					failures.add(entityResolutionMerge)
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	// owners reports, per association row, whether its owner node exists in the
	// staged collection and whether it is active in this incident.
	owners := func(kinds []string, ids []uuid.UUID) ([][2]bool, error) {
		rows, err := tx.Query(ctx, `
SELECT EXISTS (SELECT 1 FROM public.entities_bundle_validation_nodes WHERE kind=u.kind AND id=u.id),
       EXISTS (SELECT 1 FROM public.entities_bundle_validation_nodes n JOIN records r ON r.record_id=n.id WHERE n.kind=u.kind AND n.id=u.id AND n.state IN ('stub','canonical') AND r.deleted_at IS NULL AND r.incident_id=$3)
  FROM unnest($1::text[],$2::uuid[]) WITH ORDINALITY AS u(kind,id,ordinal)
 ORDER BY u.ordinal`, kinds, ids, input.IncidentID)
		if err != nil {
			return nil, err
		}
		result, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) ([2]bool, error) {
			var flags [2]bool
			return flags, row.Scan(&flags[0], &flags[1])
		})
		if err == nil && len(result) != len(kinds) {
			err = incidentportability.ErrPreparedStream
		}
		return result, err
	}
	associations := &entityAssociationBatch{}
	if err := readEntityBatches(ctx, prepared, input, prepared.preservedRows, preservedCodec{}, func(batch []stagedEntityPreserved) error {
		kinds, ids := make([]string, len(batch)), make([]uuid.UUID, len(batch))
		for index, preserved := range batch {
			kinds[index], ids[index] = preserved.Row.EntityType, preserved.Row.RecordID
		}
		flags, err := owners(kinds, ids)
		if err != nil {
			return err
		}
		for index, preserved := range batch {
			row := preserved.Row
			if !flags[index][0] {
				failures.add(entitySameIncident)
			}
			if row.DeletedAt == nil {
				associations.add("preserved", row.EntityType, row.RecordID, row.IdentifierType, row.NormalizedValue, row.Classification)
				if flags[index][1] && row.Classification == "exact_match_reuse" {
					claims.add(&failures, row.RecordID, row.EntityType, preserved.Claim)
				}
			}
		}
		if err := claims.flush(ctx, tx); err != nil {
			return err
		}
		return associations.flush(ctx, tx)
	}); err != nil {
		return err
	}
	if err := readEntityBatches(ctx, prepared, input, prepared.aliasRows, aliasCodec{}, func(batch []portableAliasRow) error {
		kinds, ids := make([]string, len(batch)), make([]uuid.UUID, len(batch))
		for index, alias := range batch {
			kinds[index], ids[index] = alias.EntityType, alias.RecordID
		}
		flags, err := owners(kinds, ids)
		if err != nil {
			return err
		}
		for index, alias := range batch {
			if !flags[index][0] {
				failures.add(entitySameIncident)
			}
			if alias.DeletedAt == nil {
				associations.add("alias", alias.EntityType, alias.RecordID, alias.NormalizedText, "", "")
			}
		}
		return associations.flush(ctx, tx)
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

// entityClaimBatch collects normalized claims; a claim whose normalization
// failed while preparing is reported once its owner is found active.
type entityClaimBatch struct {
	records              []uuid.UUID
	types, kinds, values []string
}

func (b *entityClaimBatch) add(failures *entityFailureAccumulator, recordID uuid.UUID, entityType string, claims ...entityClaimCandidate) {
	for _, claim := range claims {
		if claim.Identifier == "" {
			continue
		}
		if !claim.Valid {
			failures.add(entitySourceIdentity)
			continue
		}
		b.records = append(b.records, recordID)
		b.types = append(b.types, entityType)
		b.kinds = append(b.kinds, claim.Identifier)
		b.values = append(b.values, claim.Normalized)
	}
}

// flush writes the collected claims with one statement and empties the batch.
func (b *entityClaimBatch) flush(ctx context.Context, tx pgx.Tx) error {
	if len(b.records) == 0 {
		return nil
	}
	_, err := tx.Exec(ctx, `INSERT INTO public.entities_bundle_validation_claims (record_id,entity_type,identifier_type,normalized_value)
SELECT * FROM unnest($1::uuid[],$2::text[],$3::text[],$4::text[]) ON CONFLICT DO NOTHING`, b.records, b.types, b.kinds, b.values)
	b.records, b.types, b.kinds, b.values = b.records[:0], b.types[:0], b.kinds[:0], b.values[:0]
	return err
}

type entityAssociationBatch struct {
	rows [][]any
}

func (b *entityAssociationBatch) add(kind, entityType string, recordID uuid.UUID, key1, key2, key3 string) {
	b.rows = append(b.rows, []any{kind, entityType, recordID, key1, key2, key3})
}

// flush writes the collected associations with one COPY and empties the batch.
func (b *entityAssociationBatch) flush(ctx context.Context, tx pgx.Tx) error {
	if len(b.rows) == 0 {
		return nil
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"public", "entities_bundle_validation_associations"}, []string{"kind", "entity_type", "record_id", "key1", "key2", "key3"}, pgx.CopyFromRows(b.rows))
	if err == nil && written != int64(len(b.rows)) {
		err = incidentportability.ErrPreparedStream
	}
	b.rows = b.rows[:0]
	return err
}

func loadEntityPortableEnvelopes(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordIDs []uuid.UUID) (map[uuid.UUID]entityPortableEnvelope, error) {
	loaded, err := records.NewStore().LoadEnvelopesTx(ctx, tx, recordIDs, false)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]entityPortableEnvelope, len(loaded))
	for id, envelope := range loaded {
		if envelope.IncidentID != incidentID {
			continue
		}
		result[id] = entityPortableEnvelopeFrom(envelope)
	}
	return result, nil
}

func entityPortableEnvelopeFrom(envelope enveloperead.Envelope) entityPortableEnvelope {
	return entityPortableEnvelope{
		recordID: envelope.RecordID, incidentID: envelope.IncidentID, recordType: envelope.RecordType,
		rowVersion: envelope.RowVersion, createdAt: envelope.CreatedAt, updatedAt: envelope.UpdatedAt,
		createdByUserID: envelope.CreatedByUserID, updatedByUserID: envelope.UpdatedByUserID, deletedAt: envelope.DeletedAt,
	}
}

func validateStagedEntityClaims(ctx context.Context, tx pgx.Tx, prepared preparedEntityImport, input sourceport.ImportContext) error {
	var bindingMatches, claimsDiffer bool
	if err := tx.QueryRow(ctx, `SELECT count(*)=1 AND bool_and(operation_id=$1 AND incident_id=$2 AND actor_id=$3 AND source_digest=$4) FROM public.entities_bundle_validation_binding`, input.OperationID, input.IncidentID, input.ActorUserID, prepared.sourceDigest).Scan(&bindingMatches); err != nil {
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
