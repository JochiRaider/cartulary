package merge

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
)

func (s *Store) planHostMergeCarryForwardTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, survivor hostidentity.HostRecord, loser hostidentity.HostRecord, survivorIdentifiers []mergePreservedIdentifierRecord, loserIdentifiers []mergePreservedIdentifierRecord, survivorAliases []mergeAliasRecord, loserAliases []mergeAliasRecord, actorUserID uuid.UUID, now time.Time) (mergeCarryPlan, error) {
	precedence := s.hostIdentity.HostExactMatchPrecedence()
	summary, candidates := buildMergeClassSummary(precedence, hostCanonicalCandidates(s.hostIdentity, loser), loserIdentifiers)
	next := survivor
	existingValues, canonicalFilled := hostExistingIdentifierState(s.hostIdentity, survivor, survivorIdentifiers)
	plan, err := s.applyCarryPlanTx(ctx, tx, incidentID, "host", survivor.RecordID, loser.RecordID, precedence, existingValues, canonicalFilled, candidates, summary, actorUserID, now)
	if err != nil {
		return mergeCarryPlan{}, err
	}
	for _, candidate := range plan.IdentifierInserts {
		switch candidate.Seed.IdentifierType {
		case "aad_device_id":
			if next.AADDeviceID == nil && candidate.MutationTag == "promoted" {
				next.AADDeviceID = stringPointer(candidate.Seed.RawValue)
			}
		case "fqdn":
			if next.FQDN == nil && candidate.MutationTag == "promoted" {
				next.FQDN = stringPointer(candidate.Seed.RawValue)
			}
		case "hostname":
			if next.Hostname == nil && candidate.MutationTag == "promoted" {
				next.Hostname = stringPointer(candidate.Seed.RawValue)
			}
		}
	}
	plan.SurvivorHost = next
	plan.AliasActions = aliasActionsFromRecords(loserAliases)
	_ = actorUserID
	_ = now
	return plan, nil
}

func (s *Store) planIdentityMergeCarryForwardTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, survivor hostidentity.IdentityRecord, loser hostidentity.IdentityRecord, survivorIdentifiers []mergePreservedIdentifierRecord, loserIdentifiers []mergePreservedIdentifierRecord, survivorAliases []mergeAliasRecord, loserAliases []mergeAliasRecord, actorUserID uuid.UUID, now time.Time) (mergeCarryPlan, error) {
	precedence := s.hostIdentity.IdentityExactMatchPrecedence()
	summary, candidates := buildMergeClassSummary(precedence, identityCanonicalCandidates(s.hostIdentity, loser), loserIdentifiers)
	next := survivor
	existingValues, canonicalFilled := identityExistingIdentifierState(s.hostIdentity, survivor, survivorIdentifiers)
	plan, err := s.applyCarryPlanTx(ctx, tx, incidentID, "identity", survivor.RecordID, loser.RecordID, precedence, existingValues, canonicalFilled, candidates, summary, actorUserID, now)
	if err != nil {
		return mergeCarryPlan{}, err
	}
	for _, candidate := range plan.IdentifierInserts {
		switch candidate.Seed.IdentifierType {
		case "aad_object_id":
			if next.AADObjectID == nil && candidate.MutationTag == "promoted" {
				next.AADObjectID = stringPointer(candidate.Seed.RawValue)
			}
		case "sid":
			if next.SID == nil && candidate.MutationTag == "promoted" {
				next.SID = stringPointer(candidate.Seed.RawValue)
			}
		case "upn":
			if next.UPN == nil && candidate.MutationTag == "promoted" {
				next.UPN = stringPointer(candidate.Seed.RawValue)
			}
		case "email":
			if next.Email == nil && candidate.MutationTag == "promoted" {
				next.Email = stringPointer(candidate.Seed.RawValue)
			}
		case "sam_account_name":
			if next.SamAccountName == nil && candidate.MutationTag == "promoted" {
				next.SamAccountName = stringPointer(candidate.Seed.RawValue)
			}
		}
	}
	plan.SurvivorIdentity = next
	plan.AliasActions = aliasActionsFromRecords(loserAliases)
	_ = actorUserID
	_ = now
	return plan, nil
}

func (s *Store) applyMergeCarryForwardTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	entityType string,
	survivorRecordID uuid.UUID,
	actorUserID uuid.UUID,
	now time.Time,
	plan *mergeCarryPlan,
) error {
	for _, candidate := range plan.IdentifierInserts {
		created, err := s.hostIdentity.SyncPreservedIdentifierTx(
			ctx,
			tx,
			incidentID,
			survivorRecordID,
			entityType,
			candidate.Seed.IdentifierType,
			candidate.Seed.RawValue,
			candidate.Seed.Classification,
			actorUserID,
			now,
		)
		if err != nil {
			return err
		}
		if created {
			plan.IdentifierMutations = append(plan.IdentifierMutations, mergeIdentifierMutation{
				After: buildMergePreservedIdentifierValueFromSeed(incidentID, survivorRecordID, entityType, candidate.Seed),
			})
		}
	}
	if len(plan.AliasActions) == 0 {
		return nil
	}
	syncResult, err := s.hostIdentity.SyncAliasesTx(
		ctx,
		tx,
		incidentID,
		survivorRecordID,
		entityType,
		plan.AliasActions,
		actorUserID,
		now,
	)
	if err != nil {
		return err
	}
	plan.SuggestionAliasesCopiedCount += len(syncResult.Added)
	plan.SuggestionAliasDuplicateNoop += syncResult.DuplicateNoopCount
	for _, alias := range syncResult.Added {
		plan.AliasMutations = append(plan.AliasMutations, mergeAliasMutation{After: alias.MutationValue()})
	}
	return nil
}

func (s *Store) applyCarryPlanTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, precedence []string, survivorExisting map[string]map[string]struct{}, canonicalFilled map[string]bool, candidates map[string][]mergeExactMatchCandidate, summary map[string]MergeExactMatchClassSummary, actorUserID uuid.UUID, now time.Time) (mergeCarryPlan, error) {
	plan := mergeCarryPlan{
		IdentifierInserts: make([]mergeIdentifierInsert, 0),
		ExactMatchClasses: make([]MergeExactMatchClassSummary, 0, len(precedence)),
	}
	for _, identifierClass := range precedence {
		classSummary := summary[identifierClass]
		currentSet := survivorExisting[identifierClass]
		if currentSet == nil {
			currentSet = make(map[string]struct{})
			survivorExisting[identifierClass] = currentSet
		}
		promoted := false
		for _, candidate := range candidates[identifierClass] {
			if _, ok := currentSet[candidate.NormalizedValue]; ok {
				classSummary.DuplicateNoopCount++
				continue
			}
			conflictingRecordID, found, err := s.findThirdPartyExactMatchConflictTx(ctx, tx, incidentID, entityType, identifierClass, candidate.NormalizedValue, survivorRecordID, loserRecordID)
			if err != nil {
				return mergeCarryPlan{}, err
			}
			if found {
				classSummary.BlockedConflict++
				return mergeCarryPlan{}, &MergePreconditionError{
					ReasonCode: "carry_forward_identifier_collision",
					Details: map[string]any{
						"record_type":        entityType,
						"identifier_class":   identifierClass,
						"normalized_value":   candidate.NormalizedValue,
						"blocking_record_id": conflictingRecordID.String(),
					},
				}
			}

			insert := mergeIdentifierInsert{
				Seed: mergeIdentifierSeed{
					IdentifierType:  identifierClass,
					RawValue:        candidate.RawValue,
					NormalizedValue: candidate.NormalizedValue,
					Classification:  "exact_match_reuse",
				},
				MutationTag: "carried",
			}
			if !promoted && !canonicalFilled[identifierClass] {
				insert.MutationTag = "promoted"
				classSummary.PromotedCount++
				promoted = true
				canonicalFilled[identifierClass] = true
			} else {
				classSummary.CarriedCount++
			}
			plan.IdentifierInserts = append(plan.IdentifierInserts, insert)
			currentSet[candidate.NormalizedValue] = struct{}{}
		}
		plan.ExactMatchClasses = append(plan.ExactMatchClasses, classSummary)
	}
	_ = actorUserID
	_ = now
	return plan, nil
}

func loadMergeAliasesTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) ([]mergeAliasRecord, error) {
	rows, err := tx.Query(ctx, `
SELECT
    entity_alias_id,
    incident_id,
    record_id,
    entity_type,
    raw_text,
    normalized_text,
    classification,
    created_at,
    deleted_at
  FROM entity_aliases
 WHERE record_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY normalized_text ASC, created_at ASC, entity_alias_id ASC
 FOR UPDATE
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("load merge aliases: %w", err)
	}
	defer rows.Close()

	records := make([]mergeAliasRecord, 0)
	for rows.Next() {
		record, err := scanMergeAliasRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate merge aliases: %w", err)
	}
	return records, nil
}

func loadMergePreservedIdentifiersTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) ([]mergePreservedIdentifierRecord, error) {
	rows, err := tx.Query(ctx, `
SELECT
    entity_preserved_identifier_id,
    incident_id,
    record_id,
    entity_type,
    identifier_type,
    raw_value,
    normalized_value,
    classification,
    created_at,
    deleted_at
  FROM entity_preserved_identifiers
 WHERE record_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY identifier_type ASC, normalized_value ASC, created_at ASC, entity_preserved_identifier_id ASC
 FOR UPDATE
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("load merge preserved identifiers: %w", err)
	}
	defer rows.Close()

	records := make([]mergePreservedIdentifierRecord, 0)
	for rows.Next() {
		record, err := scanMergePreservedIdentifierRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate merge preserved identifiers: %w", err)
	}
	return records, nil
}

func buildMergeClassSummary(precedence []string, canonicalCandidates map[string][]mergeExactMatchCandidate, preservedIdentifiers []mergePreservedIdentifierRecord) (map[string]MergeExactMatchClassSummary, map[string][]mergeExactMatchCandidate) {
	summary := make(map[string]MergeExactMatchClassSummary, len(precedence))
	candidates := make(map[string][]mergeExactMatchCandidate, len(precedence))
	for _, identifierClass := range precedence {
		summary[identifierClass] = MergeExactMatchClassSummary{IdentifierClass: identifierClass}
		seen := make(map[string]struct{})
		for _, candidate := range canonicalCandidates[identifierClass] {
			if _, ok := seen[candidate.NormalizedValue]; ok {
				continue
			}
			candidates[identifierClass] = append(candidates[identifierClass], candidate)
			seen[candidate.NormalizedValue] = struct{}{}
		}
		for _, identifier := range preservedIdentifiers {
			if identifier.IdentifierType != identifierClass {
				continue
			}
			current := summary[identifierClass]
			switch identifier.Classification {
			case "exact_match_reuse":
				if _, ok := seen[identifier.NormalizedValue]; ok {
					summary[identifierClass] = current
					continue
				}
				candidates[identifierClass] = append(candidates[identifierClass], mergeExactMatchCandidate{
					IdentifierClass: identifierClass,
					RawValue:        identifier.RawValue,
					NormalizedValue: identifier.NormalizedValue,
				})
				seen[identifier.NormalizedValue] = struct{}{}
			case "provenance_only":
				current.ProvenanceOnly++
			case "suggestion_only":
				current.SuggestionOnly++
			}
			summary[identifierClass] = current
		}
	}
	return summary, candidates
}

func hostCanonicalCandidates(capability *hostidentity.MergeCapability, record hostidentity.HostRecord) map[string][]mergeExactMatchCandidate {
	candidates := make(map[string][]mergeExactMatchCandidate, len(capability.HostExactMatchPrecedence()))
	if normalized := capability.HostCanonicalNormalized(record, "aad_device_id"); normalized != "" {
		candidates["aad_device_id"] = append(candidates["aad_device_id"], mergeExactMatchCandidate{IdentifierClass: "aad_device_id", RawValue: *record.AADDeviceID, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := capability.HostCanonicalNormalized(record, "fqdn"); normalized != "" {
		candidates["fqdn"] = append(candidates["fqdn"], mergeExactMatchCandidate{IdentifierClass: "fqdn", RawValue: *record.FQDN, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := capability.HostCanonicalNormalized(record, "hostname"); normalized != "" {
		candidates["hostname"] = append(candidates["hostname"], mergeExactMatchCandidate{IdentifierClass: "hostname", RawValue: *record.Hostname, NormalizedValue: normalized, FromCanonical: true})
	}
	return candidates
}

func identityCanonicalCandidates(capability *hostidentity.MergeCapability, record hostidentity.IdentityRecord) map[string][]mergeExactMatchCandidate {
	candidates := make(map[string][]mergeExactMatchCandidate, len(capability.IdentityExactMatchPrecedence()))
	if normalized := capability.IdentityCanonicalNormalized(record, "aad_object_id"); normalized != "" {
		candidates["aad_object_id"] = append(candidates["aad_object_id"], mergeExactMatchCandidate{IdentifierClass: "aad_object_id", RawValue: *record.AADObjectID, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := capability.IdentityCanonicalNormalized(record, "sid"); normalized != "" {
		candidates["sid"] = append(candidates["sid"], mergeExactMatchCandidate{IdentifierClass: "sid", RawValue: *record.SID, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := capability.IdentityCanonicalNormalized(record, "upn"); normalized != "" {
		candidates["upn"] = append(candidates["upn"], mergeExactMatchCandidate{IdentifierClass: "upn", RawValue: *record.UPN, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := capability.IdentityCanonicalNormalized(record, "email"); normalized != "" {
		candidates["email"] = append(candidates["email"], mergeExactMatchCandidate{IdentifierClass: "email", RawValue: *record.Email, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := capability.IdentityCanonicalNormalized(record, "sam_account_name"); normalized != "" {
		candidates["sam_account_name"] = append(candidates["sam_account_name"], mergeExactMatchCandidate{IdentifierClass: "sam_account_name", RawValue: *record.SamAccountName, NormalizedValue: normalized, FromCanonical: true})
	}
	return candidates
}

func hostExistingIdentifierState(capability *hostidentity.MergeCapability, record hostidentity.HostRecord, preserved []mergePreservedIdentifierRecord) (map[string]map[string]struct{}, map[string]bool) {
	precedence := capability.HostExactMatchPrecedence()
	set := make(map[string]map[string]struct{}, len(precedence))
	filled := make(map[string]bool, len(precedence))
	for _, identifierClass := range precedence {
		set[identifierClass] = make(map[string]struct{})
	}
	if normalized := capability.HostCanonicalNormalized(record, "aad_device_id"); normalized != "" {
		set["aad_device_id"][normalized] = struct{}{}
		filled["aad_device_id"] = true
	}
	if normalized := capability.HostCanonicalNormalized(record, "fqdn"); normalized != "" {
		set["fqdn"][normalized] = struct{}{}
		filled["fqdn"] = true
	}
	if normalized := capability.HostCanonicalNormalized(record, "hostname"); normalized != "" {
		set["hostname"][normalized] = struct{}{}
		filled["hostname"] = true
	}
	for _, identifier := range preserved {
		if identifier.Classification != "exact_match_reuse" {
			continue
		}
		current := set[identifier.IdentifierType]
		if current == nil {
			current = make(map[string]struct{})
			set[identifier.IdentifierType] = current
		}
		current[identifier.NormalizedValue] = struct{}{}
	}
	return set, filled
}

func identityExistingIdentifierState(capability *hostidentity.MergeCapability, record hostidentity.IdentityRecord, preserved []mergePreservedIdentifierRecord) (map[string]map[string]struct{}, map[string]bool) {
	precedence := capability.IdentityExactMatchPrecedence()
	set := make(map[string]map[string]struct{}, len(precedence))
	filled := make(map[string]bool, len(precedence))
	for _, identifierClass := range precedence {
		set[identifierClass] = make(map[string]struct{})
	}
	if normalized := capability.IdentityCanonicalNormalized(record, "aad_object_id"); normalized != "" {
		set["aad_object_id"][normalized] = struct{}{}
		filled["aad_object_id"] = true
	}
	if normalized := capability.IdentityCanonicalNormalized(record, "sid"); normalized != "" {
		set["sid"][normalized] = struct{}{}
		filled["sid"] = true
	}
	if normalized := capability.IdentityCanonicalNormalized(record, "upn"); normalized != "" {
		set["upn"][normalized] = struct{}{}
		filled["upn"] = true
	}
	if normalized := capability.IdentityCanonicalNormalized(record, "email"); normalized != "" {
		set["email"][normalized] = struct{}{}
		filled["email"] = true
	}
	if normalized := capability.IdentityCanonicalNormalized(record, "sam_account_name"); normalized != "" {
		set["sam_account_name"][normalized] = struct{}{}
		filled["sam_account_name"] = true
	}
	for _, identifier := range preserved {
		if identifier.Classification != "exact_match_reuse" {
			continue
		}
		current := set[identifier.IdentifierType]
		if current == nil {
			current = make(map[string]struct{})
			set[identifier.IdentifierType] = current
		}
		current[identifier.NormalizedValue] = struct{}{}
	}
	return set, filled
}

func aliasValuesFromRecords(records []mergeAliasRecord) []hostidentity.AliasValue {
	values := make([]hostidentity.AliasValue, 0, len(records))
	for _, record := range records {
		values = append(values, hostidentity.AliasValue{EntityAliasID: record.EntityAliasID, AliasText: record.NormalizedText})
	}
	return values
}

func aliasActionsFromRecords(records []mergeAliasRecord) []hostidentity.CollectionAction {
	actions := make([]hostidentity.CollectionAction, 0, len(records))
	for _, record := range records {
		actions = append(actions, hostidentity.CollectionAction{
			Op:             "add_alias",
			RawText:        record.RawText,
			NormalizedText: record.NormalizedText,
		})
	}
	return actions
}

func countProvenanceOnlyIdentifiers(values []mergePreservedIdentifierRecord) int {
	count := 0
	for _, value := range values {
		if value.Classification == "provenance_only" && value.DeletedAt == nil {
			count++
		}
	}
	return count
}

func mergeScopeKey(survivorRecordID uuid.UUID, loserRecordID uuid.UUID) string {
	return survivorRecordID.String() + ":" + loserRecordID.String()
}

func buildMergePreservedIdentifierValueFromSeed(incidentID uuid.UUID, recordID uuid.UUID, entityType string, seed mergeIdentifierSeed) map[string]any {
	return map[string]any{
		"incident_id":      incidentID.String(),
		"record_id":        recordID.String(),
		"entity_type":      entityType,
		"identifier_type":  seed.IdentifierType,
		"raw_value":        seed.RawValue,
		"normalized_value": seed.NormalizedValue,
		"classification":   seed.Classification,
	}
}
