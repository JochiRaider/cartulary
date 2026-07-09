package revisions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	linkrevisionprovider "github.com/JochiRaider/cartulary/internal/modules/links/revisionprovider"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var ErrRecordNotFound = errors.New("revisions: record not found")

type Store struct {
	db                          postgres.DB
	incidentAccess              incidents.Access
	importedAttributionResolver ImportedAttributionResolver
	linkRollbackTargets         LinkRollbackTargetProvider
	tagRollbackTargets          TagRollbackTargetProvider
	projectionRebuilder         ProjectionRebuilder
}

type ImportedAttributionResolver interface {
	ResolveImportedSourceActorsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, sourceTable string, sourceColumn string, sourceRowIDs []string) (map[string]string, error)
}

type noopImportedAttributionResolver struct{}

func (noopImportedAttributionResolver) ResolveImportedSourceActorsTx(context.Context, pgx.Tx, uuid.UUID, string, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

type StoreOptions struct {
	ImportedAttributionResolver ImportedAttributionResolver
	LinkRollbackTargetProvider  LinkRollbackTargetProvider
	TagRollbackTargetProvider   TagRollbackTargetProvider
	ProjectionRebuilder         ProjectionRebuilder
}

type LinkRollbackTargetProvider interface {
	ValidateRecordLinkValue(value map[string]any) error
	LoadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error)
	TombstoneRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, actorUserID uuid.UUID, now time.Time) error
	RestoreRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, value map[string]any, actorUserID uuid.UUID, now time.Time) error
}

type TagRollbackTargetProvider interface {
	ParseRecordTagIdentity(value map[string]any) (linkrevisionprovider.RecordTagIdentity, error)
	LoadRecordTagValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (map[string]any, error)
	RestoreRecordTagTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, value map[string]any, now time.Time) error
	TombstoneRecordTagTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, actorUserID uuid.UUID, now time.Time) error
}

type ChangeSetParams struct {
	ChangeSetID *uuid.UUID
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Source      string
	Reason      *string
	ClientTxnID *string
	RequestID   *string
	CreatedAt   time.Time
}

type MutationParams struct {
	ChangeSetID     uuid.UUID
	SequenceNo      int
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type RecordRevisionParams struct {
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeValue any
	AfterValue  any
}

type RecordHistoryRecord struct {
	IncidentID  uuid.UUID
	RecordID    uuid.UUID
	RecordType  string
	RowVersion  int64
	Deleted     bool
	DeletedAt   *time.Time
	DeletedByID *uuid.UUID
}

type RecordHistoryItem struct {
	ActorUserID              uuid.UUID
	SourceActorID            *string
	CommittedAt              time.Time
	HistoryItemRef           string
	Operation                string
	DiffSummary              map[string]any
	ChangeSetID              uuid.UUID
	Reversible               bool
	AvailableRollbackActions []string
	HistoryEntryRef          *string
	RevisionNo               *int64

	createdAt      time.Time
	changeSetID    uuid.UUID
	sequenceNo     int
	syntheticRank  int
	targetKey      string
	hasTargetEntry bool
}

func NewStore(db ...postgres.DB) *Store {
	var handle postgres.DB
	if len(db) > 0 {
		handle = db[0]
	}
	return NewStoreWithOptions(handle, StoreOptions{})
}

func NewStoreWithOptions(db postgres.DB, options StoreOptions) *Store {
	resolver := options.ImportedAttributionResolver
	if resolver == nil {
		resolver = noopImportedAttributionResolver{}
	}
	linkTargets := options.LinkRollbackTargetProvider
	if linkTargets == nil {
		linkTargets = linkrevisionprovider.NewProvider()
	}
	tagTargets := options.TagRollbackTargetProvider
	if tagTargets == nil {
		tagTargets = linkrevisionprovider.NewProvider()
	}
	projectionRebuilder := options.ProjectionRebuilder
	if projectionRebuilder == nil {
		projectionRebuilder = defaultProjectionRebuilder{}
	}
	return &Store{
		db:                          db,
		incidentAccess:              incidents.NewAccess(db),
		importedAttributionResolver: resolver,
		linkRollbackTargets:         linkTargets,
		tagRollbackTargets:          tagTargets,
		projectionRebuilder:         projectionRebuilder,
	}
}

func (s *Store) InsertChangeSetTx(ctx context.Context, tx pgx.Tx, params ChangeSetParams) (uuid.UUID, error) {
	if params.ChangeSetID != nil {
		if _, err := tx.Exec(ctx, `
INSERT INTO change_sets (
    change_set_id,
    incident_id,
    actor_user_id,
    source,
    reason,
    client_txn_id,
    request_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, *params.ChangeSetID, params.IncidentID, params.ActorUserID, params.Source, params.Reason, params.ClientTxnID, params.RequestID, params.CreatedAt.UTC()); err != nil {
			return uuid.UUID{}, fmt.Errorf("insert change set: %w", err)
		}
		return *params.ChangeSetID, nil
	}
	var changeSetID uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO change_sets (
    incident_id,
    actor_user_id,
    source,
    reason,
    client_txn_id,
    request_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING change_set_id
`, params.IncidentID, params.ActorUserID, params.Source, params.Reason, params.ClientTxnID, params.RequestID, params.CreatedAt.UTC()).Scan(&changeSetID); err != nil {
		return uuid.UUID{}, fmt.Errorf("insert change set: %w", err)
	}
	return changeSetID, nil
}

func (s *Store) InsertMutationTx(ctx context.Context, tx pgx.Tx, params MutationParams) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO change_set_mutations (
    change_set_id,
    sequence_no,
    target_kind,
    target_id,
    operation_kind,
    before_version_id,
    after_version_id,
    before_value,
    after_value
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, params.ChangeSetID, params.SequenceNo, params.TargetKind, params.TargetID, params.OperationKind, params.BeforeVersionID, params.AfterVersionID, jsonOrNil(params.BeforeValue), jsonOrNil(params.AfterValue)); err != nil {
		return fmt.Errorf("insert change-set mutation: %w", err)
	}
	return nil
}

func (s *Store) InsertRecordRevisionTx(ctx context.Context, tx pgx.Tx, params RecordRevisionParams) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO record_revisions (
    change_set_id,
    record_id,
    row_version,
    before_json,
    after_json
)
VALUES ($1, $2, $3, $4, $5)
`, params.ChangeSetID, params.RecordID, params.RowVersion, jsonOrNil(params.BeforeValue), jsonOrNil(params.AfterValue)); err != nil {
		return fmt.Errorf("insert record revision: %w", err)
	}
	return nil
}

func (s *Store) GetHistoryRecord(ctx context.Context, recordID uuid.UUID) (RecordHistoryRecord, error) {
	if s.db == nil {
		return RecordHistoryRecord{}, errors.New("revisions history store: postgres dependency is nil")
	}
	var (
		record       RecordHistoryRecord
		deletedAt    sql.NullTime
		deletedByRaw sql.NullString
	)
	if err := s.db.QueryRow(ctx, `
SELECT incident_id, record_id, record_type, row_version, deleted_at, deleted_by_user_id::text
  FROM records
 WHERE record_id = $1
`, recordID).Scan(&record.IncidentID, &record.RecordID, &record.RecordType, &record.RowVersion, &deletedAt, &deletedByRaw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RecordHistoryRecord{}, ErrRecordNotFound
		}
		return RecordHistoryRecord{}, fmt.Errorf("load record history envelope: %w", err)
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
		record.Deleted = true
	}
	if deletedByRaw.Valid {
		parsed, err := uuid.Parse(deletedByRaw.String)
		if err != nil {
			return RecordHistoryRecord{}, fmt.Errorf("parse deleted_by_user_id: %w", err)
		}
		record.DeletedByID = &parsed
	}
	return record, nil
}

func (s *Store) ListRecordHistory(ctx context.Context, record RecordHistoryRecord) ([]map[string]any, error) {
	if s.db == nil {
		return nil, errors.New("revisions history store: postgres dependency is nil")
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin record history transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	items, err := s.loadMutationHistoryItemsTx(ctx, tx, record)
	if err != nil {
		return nil, err
	}
	revisionItems, err := s.loadRevisionOnlyHistoryItemsTx(ctx, tx, record, items)
	if err != nil {
		return nil, err
	}
	items = append(items, revisionItems...)
	sortHistoryItems(items)

	rollbackRecord := rollbackRecordEnvelope{
		IncidentID:      record.IncidentID,
		RecordID:        record.RecordID,
		RecordType:      record.RecordType,
		RowVersion:      record.RowVersion,
		DeletedAt:       record.DeletedAt,
		DeletedByUserID: record.DeletedByID,
	}
	changeSetExecutable := make(map[uuid.UUID]bool)
	changeSetChecked := make(map[uuid.UUID]bool)
	for i := range items {
		items[i].AvailableRollbackActions = nil
		if rollbackRecord.DeletedAt != nil {
			items[i].RevisionNo = nil
			items[i].Reversible = false
			continue
		}
		if items[i].HistoryEntryRef != nil {
			executable, err := s.historyEntryRollbackExecutableTx(ctx, tx, rollbackRecord, *items[i].HistoryEntryRef)
			if err != nil {
				return nil, err
			}
			if executable {
				items[i].AvailableRollbackActions = append(items[i].AvailableRollbackActions, "history_entry")
			}
		}
		if !changeSetChecked[items[i].ChangeSetID] {
			executable, err := s.changeSetRollbackExecutableTx(ctx, tx, rollbackRecord, items[i].ChangeSetID)
			if err != nil {
				return nil, err
			}
			changeSetChecked[items[i].ChangeSetID] = true
			changeSetExecutable[items[i].ChangeSetID] = executable
		}
		if changeSetExecutable[items[i].ChangeSetID] {
			items[i].AvailableRollbackActions = append(items[i].AvailableRollbackActions, "change_set")
		}
		if items[i].RevisionNo != nil {
			executable, err := s.rowRestoreExecutableTx(ctx, tx, rollbackRecord, *items[i].RevisionNo)
			if err != nil {
				return nil, err
			}
			if executable {
				items[i].AvailableRollbackActions = append(items[i].AvailableRollbackActions, "row_restore")
			} else {
				items[i].RevisionNo = nil
			}
		}
		items[i].Reversible = len(items[i].AvailableRollbackActions) > 0
	}

	resources := make([]map[string]any, 0, len(items))
	for _, item := range items {
		resources = append(resources, item.Resource())
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit record history transaction: %w", err)
	}
	return resources, nil
}

func (s *Store) historyEntryRollbackExecutableTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, historyEntryRef string) (bool, error) {
	plan, err := loadHistoryEntryRollbackPlanTx(ctx, tx, record, historyEntryRef)
	if err != nil {
		if errors.Is(err, ErrRollbackTargetNotFound) || errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	return s.rollbackPlanExecutableTx(ctx, tx, plan)
}

func (s *Store) changeSetRollbackExecutableTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, changeSetID uuid.UUID) (bool, error) {
	plan, err := loadChangeSetRollbackPlanTx(ctx, tx, record, changeSetID.String())
	if err != nil {
		if errors.Is(err, ErrRollbackTargetNotFound) || errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	return s.rollbackPlanExecutableTx(ctx, tx, plan)
}

func (s *Store) rollbackPlanExecutableTx(ctx context.Context, tx pgx.Tx, plan rollbackPlan) (bool, error) {
	if err := s.validateRollbackPlan(plan); err != nil {
		if errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	if err := ensureNoLaterRollbackPlanMutationTx(ctx, tx, plan); err != nil {
		if errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Store) rowRestoreExecutableTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, revisionNo int64) (bool, error) {
	_, err := loadRowRestorePlanTx(ctx, tx, record, revisionNo)
	if err != nil {
		if errors.Is(err, ErrRollbackTargetNotFound) || errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Store) loadMutationHistoryItemsTx(ctx context.Context, tx pgx.Tx, record RecordHistoryRecord) ([]RecordHistoryItem, error) {
	rows, err := tx.Query(ctx, `
SELECT cs.change_set_id,
       cs.actor_user_id,
       cs.created_at,
       cs.source,
       csm.sequence_no,
       csm.target_kind,
       csm.target_id,
       csm.operation_kind,
       csm.before_value,
       csm.after_value,
       rr.row_version,
       href.history_entry_ref
  FROM change_sets cs
  JOIN change_set_mutations csm
    ON csm.change_set_id = cs.change_set_id
  LEFT JOIN record_revisions rr
    ON rr.change_set_id = cs.change_set_id
   AND rr.record_id = $1
  LEFT JOIN record_history_entry_refs href
    ON href.record_id = $1
	   AND href.change_set_id = csm.change_set_id
	   AND href.mutation_sequence_no = csm.sequence_no
	 WHERE cs.incident_id = $2
	   AND (
	       csm.target_id = $3
	       OR (
	           csm.target_kind = 'record_link'
	           AND (
	               csm.before_value ->> 'src_record_id' = $3
	               OR csm.before_value ->> 'dst_record_id' = $3
	               OR csm.after_value ->> 'src_record_id' = $3
	               OR csm.after_value ->> 'dst_record_id' = $3
	           )
	       )
	       OR (
	           csm.target_kind = 'entity_mention'
	           AND (
	               csm.before_value ->> 'source_record_id' = $3
	               OR csm.after_value ->> 'source_record_id' = $3
	           )
	       )
	       OR (
	           csm.target_kind = 'record_tag'
	           AND (
	               csm.before_value ->> 'record_id' = $3
	               OR csm.after_value ->> 'record_id' = $3
	           )
	       )
	   )
	 ORDER BY cs.created_at DESC, cs.change_set_id DESC, csm.sequence_no ASC
	`, record.RecordID, record.IncidentID, record.RecordID.String())
	if err != nil {
		return nil, fmt.Errorf("query record history mutations: %w", err)
	}
	defer rows.Close()

	items := make([]RecordHistoryItem, 0)
	for rows.Next() {
		var (
			item          RecordHistoryItem
			source        string
			targetKind    string
			targetID      string
			operationKind string
			beforeValue   []byte
			afterValue    []byte
			revisionNo    sql.NullInt64
			ref           sql.NullString
		)
		if err := rows.Scan(
			&item.ChangeSetID,
			&item.ActorUserID,
			&item.CommittedAt,
			&source,
			&item.sequenceNo,
			&targetKind,
			&targetID,
			&operationKind,
			&beforeValue,
			&afterValue,
			&revisionNo,
			&ref,
		); err != nil {
			return nil, fmt.Errorf("scan record history mutation: %w", err)
		}
		if ref.Valid {
			item.HistoryEntryRef = &ref.String
		}
		if revisionNo.Valid {
			value := revisionNo.Int64
			item.RevisionNo = &value
		}
		item.HistoryItemRef = historyItemRefForMutation(record.RecordID, item.ChangeSetID, item.sequenceNo)
		item.Operation = historyOperation(source, operationKind)
		item.DiffSummary = mutationDiffSummary(targetKind, targetID, operationKind, item.sequenceNo, beforeValue, afterValue)
		item.AvailableRollbackActions = nil
		item.Reversible = false
		item.createdAt = item.CommittedAt
		item.changeSetID = item.ChangeSetID
		item.syntheticRank = 0
		item.targetKey = targetKind + ":" + targetID
		item.hasTargetEntry = singleEntryAddressable(targetKind, targetID, record.RecordID, beforeValue, afterValue)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record history mutations: %w", err)
	}
	if err := s.attachImportedSourceActorsTx(ctx, tx, record.IncidentID, items); err != nil {
		return nil, err
	}
	rows.Close()
	for i := range items {
		if !items[i].hasTargetEntry || items[i].HistoryEntryRef != nil {
			continue
		}
		generated, err := ensureHistoryEntryRefTx(ctx, tx, record.RecordID, items[i].ChangeSetID, items[i].sequenceNo)
		if err != nil {
			return nil, err
		}
		items[i].HistoryEntryRef = &generated
		items[i].AvailableRollbackActions = nil
		items[i].Reversible = false
	}
	return items, nil
}

func (s *Store) loadRevisionOnlyHistoryItemsTx(ctx context.Context, tx pgx.Tx, record RecordHistoryRecord, mutationItems []RecordHistoryItem) ([]RecordHistoryItem, error) {
	changeSetsWithMutation := make(map[uuid.UUID]bool, len(mutationItems))
	for _, item := range mutationItems {
		changeSetsWithMutation[item.ChangeSetID] = true
	}

	rows, err := tx.Query(ctx, `
SELECT cs.change_set_id,
       cs.actor_user_id,
       cs.created_at,
       cs.source,
       rr.row_version,
       rr.before_json,
       rr.after_json
  FROM record_revisions rr
  JOIN change_sets cs
    ON cs.change_set_id = rr.change_set_id
 WHERE rr.record_id = $1
   AND cs.incident_id = $2
 ORDER BY cs.created_at DESC, cs.change_set_id DESC, rr.row_version DESC
`, record.RecordID, record.IncidentID)
	if err != nil {
		return nil, fmt.Errorf("query record history revisions: %w", err)
	}
	defer rows.Close()

	items := make([]RecordHistoryItem, 0)
	for rows.Next() {
		var (
			item        RecordHistoryItem
			source      string
			revisionNo  int64
			beforeValue []byte
			afterValue  []byte
		)
		if err := rows.Scan(&item.ChangeSetID, &item.ActorUserID, &item.CommittedAt, &source, &revisionNo, &beforeValue, &afterValue); err != nil {
			return nil, fmt.Errorf("scan record history revision: %w", err)
		}
		if changeSetsWithMutation[item.ChangeSetID] {
			continue
		}
		item.HistoryItemRef = historyItemRefForRevision(record.RecordID, item.ChangeSetID, revisionNo)
		item.Operation = historyOperation(source, "row_revision")
		item.DiffSummary = revisionDiffSummary(record.RecordID, revisionNo, beforeValue, afterValue)
		item.RevisionNo = &revisionNo
		item.AvailableRollbackActions = nil
		item.Reversible = false
		item.createdAt = item.CommittedAt
		item.changeSetID = item.ChangeSetID
		item.sequenceNo = int(^uint(0) >> 1)
		item.syntheticRank = 1
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record history revisions: %w", err)
	}
	if err := s.attachImportedSourceActorsTx(ctx, tx, record.IncidentID, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) attachImportedSourceActorsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, items []RecordHistoryItem) error {
	if len(items) == 0 || s.importedAttributionResolver == nil {
		return nil
	}
	rowIDs := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		rowID := item.ChangeSetID.String()
		if _, ok := seen[rowID]; ok {
			continue
		}
		seen[rowID] = struct{}{}
		rowIDs = append(rowIDs, rowID)
	}
	sourceActors, err := s.importedAttributionResolver.ResolveImportedSourceActorsTx(ctx, tx, incidentID, "change_sets", "actor_user_id", rowIDs)
	if err != nil {
		return err
	}
	for idx := range items {
		sourceActorID := sourceActors[items[idx].ChangeSetID.String()]
		if sourceActorID == "" {
			continue
		}
		items[idx].SourceActorID = &sourceActorID
	}
	return nil
}

func (item RecordHistoryItem) Resource() map[string]any {
	actions := make([]string, 0, len(item.AvailableRollbackActions))
	actions = append(actions, item.AvailableRollbackActions...)
	resource := map[string]any{
		"actor_user_id":              item.ActorUserID.String(),
		"committed_at":               item.CommittedAt.UTC().Format(time.RFC3339Nano),
		"history_item_ref":           item.HistoryItemRef,
		"operation":                  item.Operation,
		"diff_summary":               item.DiffSummary,
		"change_set_id":              item.ChangeSetID.String(),
		"reversible":                 item.Reversible,
		"available_rollback_actions": actions,
	}
	if item.HistoryEntryRef != nil {
		resource["history_entry_ref"] = *item.HistoryEntryRef
	}
	if item.SourceActorID != nil {
		resource["source_actor_id"] = *item.SourceActorID
	}
	if item.RevisionNo != nil {
		resource["revision_no"] = *item.RevisionNo
	}
	return resource
}

func ensureHistoryEntryRefTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changeSetID uuid.UUID, sequenceNo int) (string, error) {
	var existing string
	err := tx.QueryRow(ctx, `
SELECT history_entry_ref
  FROM record_history_entry_refs
 WHERE record_id = $1
   AND change_set_id = $2
   AND mutation_sequence_no = $3
`, recordID, changeSetID, sequenceNo).Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("lookup history entry ref: %w", err)
	}

	for attempts := 0; attempts < 3; attempts++ {
		candidate, err := generateHistoryEntryRef()
		if err != nil {
			return "", err
		}
		err = tx.QueryRow(ctx, `
INSERT INTO record_history_entry_refs (history_entry_ref, record_id, change_set_id, mutation_sequence_no)
VALUES ($1, $2, $3, $4)
ON CONFLICT (record_id, change_set_id, mutation_sequence_no) DO UPDATE
SET created_at = record_history_entry_refs.created_at
RETURNING history_entry_ref
`, candidate, recordID, changeSetID, sequenceNo).Scan(&existing)
		if err == nil {
			return existing, nil
		}
	}
	return "", fmt.Errorf("insert history entry ref after retries: %w", err)
}

func generateHistoryEntryRef() (string, error) {
	var payload [16]byte
	if _, err := rand.Read(payload[:]); err != nil {
		return "", fmt.Errorf("generate history entry ref: %w", err)
	}
	return "href_" + base64.RawURLEncoding.EncodeToString(payload[:]), nil
}

func historyItemRefForMutation(recordID uuid.UUID, changeSetID uuid.UUID, sequenceNo int) string {
	return historyItemRef("mutation", recordID.String(), changeSetID.String(), fmt.Sprintf("%d", sequenceNo))
}

func historyItemRefForRevision(recordID uuid.UUID, changeSetID uuid.UUID, revisionNo int64) string {
	return historyItemRef("revision", recordID.String(), changeSetID.String(), fmt.Sprintf("%d", revisionNo))
}

func historyItemRef(parts ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(parts, ":")))
	return "hitem_" + base64.RawURLEncoding.EncodeToString(digest[:])
}

func singleEntryAddressable(targetKind string, targetID string, recordID uuid.UUID, beforeValue []byte, afterValue []byte) bool {
	if targetID != recordID.String() {
		switch targetKind {
		case "record_link":
			return mutationJSONReferencesRecord(beforeValue, recordID, "src_record_id", "dst_record_id") ||
				mutationJSONReferencesRecord(afterValue, recordID, "src_record_id", "dst_record_id")
		case "entity_mention":
			return mutationJSONReferencesRecord(beforeValue, recordID, "source_record_id") ||
				mutationJSONReferencesRecord(afterValue, recordID, "source_record_id")
		case "record_tag":
			return mutationJSONReferencesRecord(beforeValue, recordID, "record_id") ||
				mutationJSONReferencesRecord(afterValue, recordID, "record_id")
		default:
			return false
		}
	}
	switch targetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
		return true
	case "record_tag":
		return mutationJSONReferencesRecord(beforeValue, recordID, "record_id") ||
			mutationJSONReferencesRecord(afterValue, recordID, "record_id")
	default:
		return false
	}
}

func mutationJSONReferencesRecord(raw []byte, recordID uuid.UUID, keys ...string) bool {
	if len(raw) == 0 {
		return false
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	recordIDText := recordID.String()
	for _, key := range keys {
		if text, ok := value[key].(string); ok && text == recordIDText {
			return true
		}
	}
	return false
}

func historyOperation(source string, operationKind string) string {
	if operationKind != "" {
		return operationKind
	}
	if source != "" {
		return source
	}
	return "unknown"
}

func mutationDiffSummary(targetKind string, targetID string, operationKind string, sequenceNo int, beforeValue []byte, afterValue []byte) map[string]any {
	return map[string]any{
		"summary": fmt.Sprintf("%s %s", historyOperation("", operationKind), targetKind),
		"units": []map[string]any{
			{
				"target_kind":       targetKind,
				"target_id":         targetID,
				"operation":         historyOperation("", operationKind),
				"sequence_no":       sequenceNo,
				"has_before_value":  len(beforeValue) > 0,
				"has_after_value":   len(afterValue) > 0,
				"history_unit_kind": "mutation",
			},
		},
	}
}

func revisionDiffSummary(recordID uuid.UUID, revisionNo int64, beforeValue []byte, afterValue []byte) map[string]any {
	return map[string]any{
		"summary": fmt.Sprintf("row revision %d", revisionNo),
		"units": []map[string]any{
			{
				"record_id":         recordID.String(),
				"revision_no":       revisionNo,
				"has_before_value":  len(beforeValue) > 0,
				"has_after_value":   len(afterValue) > 0,
				"history_unit_kind": "row_revision",
			},
		},
	}
}

func sortHistoryItems(items []RecordHistoryItem) {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if !left.createdAt.Equal(right.createdAt) {
			return left.createdAt.After(right.createdAt)
		}
		if left.changeSetID != right.changeSetID {
			return left.changeSetID.String() > right.changeSetID.String()
		}
		if left.syntheticRank != right.syntheticRank {
			return left.syntheticRank < right.syntheticRank
		}
		return left.sequenceNo < right.sequenceNo
	})
}

func jsonOrNil(value any) any {
	if value == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return payload
}
