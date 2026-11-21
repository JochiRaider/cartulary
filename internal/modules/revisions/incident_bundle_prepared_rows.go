package revisions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// Staged write-ready Revisions streams (Core 01 REQ-01-641). Preparation proves
// every bundle-only fact and stages these rows; the final transaction copies
// them and checks only database-side facts. Paths are relative to the port's
// preparation scope.
const (
	preparedChangeSetsPath  = "change_sets.rows"
	preparedMutationsPath   = "change_set_mutations.rows"
	preparedRevisionsPath   = "record_revisions.rows"
	preparedAttributionPath = "attribution.rows"
	preparedReferencesPath  = "references/"
	preparedMembershipPath  = "membership/"
	preparedVersionsPath    = "versions/"
)

// Bundle-only verdicts that the Revisions precedence ranks after a database
// check are recorded while preparing and reported at these ranks.
const (
	rankMembership = iota + 1
	rankBundleReferences
	rankRecordVersions
	rankHistoryShapes
	rankActorContexts
)

// stagedChangeSet is one write-ready change set. Contexts already carry this
// import's appended actor context; the importing actor is bound by the codec.
type stagedChangeSet struct {
	ChangeSetID     uuid.UUID
	IncidentID      uuid.UUID
	PortableActorID uuid.UUID
	Source          string
	Reason          *string
	ClientTxnID     *string
	RequestID       *string
	CreatedAt       time.Time
	Contexts        []revisionActorImportContext
	HistoryOrder    int64
}

var changeSetColumns = []string{
	"change_set_id", "incident_id", "actor_user_id", "source", "reason",
	"client_txn_id", "request_id", "created_at", "actor_import_contexts", "history_order",
}

type changeSetRowCodec struct{ actor uuid.UUID }

func (changeSetRowCodec) Encode(row stagedChangeSet) ([]byte, error) {
	return json.Marshal([]any{row.ChangeSetID, row.IncidentID, row.PortableActorID, row.Source, row.Reason,
		row.ClientTxnID, row.RequestID, row.CreatedAt, row.Contexts, row.HistoryOrder})
}
func (changeSetRowCodec) Decode(line []byte) (stagedChangeSet, error) {
	var row stagedChangeSet
	err := decodeStagedPositional(line, &row.ChangeSetID, &row.IncidentID, &row.PortableActorID, &row.Source, &row.Reason,
		&row.ClientTxnID, &row.RequestID, &row.CreatedAt, &row.Contexts, &row.HistoryOrder)
	row.CreatedAt = row.CreatedAt.UTC()
	return row, err
}
func (changeSetRowCodec) Columns() []string { return changeSetColumns }
func (c changeSetRowCodec) Values(row stagedChangeSet) ([]any, error) {
	contexts, err := json.Marshal(row.Contexts)
	if err != nil {
		return nil, err
	}
	return []any{row.ChangeSetID, row.IncidentID, c.actor, row.Source, row.Reason, row.ClientTxnID,
		row.RequestID, row.CreatedAt, contexts, row.HistoryOrder}, nil
}

// stagedMutation is one write-ready mutation with the history identities
// derived while preparing. The trailing fields carry the record references
// and declared snapshot types whose database side the transaction checks.
type stagedMutation struct {
	ChangeSetID           uuid.UUID
	SequenceNo            int
	TargetKind            string
	TargetID              string
	OperationKind         string
	BeforeVersionID       *string
	AfterVersionID        *string
	BeforeValue           json.RawMessage
	AfterValue            json.RawMessage
	HistoryRecordIDs      []uuid.UUID
	HistoryEntryRecordIDs []uuid.UUID
	DescribeFailed        bool
	References            []uuid.UUID
	RowRecord             *uuid.UUID
	Declared              []string
}

var mutationColumns = []string{
	"change_set_id", "sequence_no", "target_kind", "target_id", "operation_kind",
	"before_version_id", "after_version_id", "before_value", "after_value",
	"history_record_ids", "history_entry_record_ids",
}

type mutationRowCodec struct{}

func (mutationRowCodec) Encode(row stagedMutation) ([]byte, error) {
	return json.Marshal([]any{row.ChangeSetID, row.SequenceNo, row.TargetKind, row.TargetID, row.OperationKind,
		row.BeforeVersionID, row.AfterVersionID, row.BeforeValue, row.AfterValue, row.HistoryRecordIDs,
		row.HistoryEntryRecordIDs, row.DescribeFailed, row.References, row.RowRecord, row.Declared})
}
func (mutationRowCodec) Decode(line []byte) (stagedMutation, error) {
	var row stagedMutation
	err := decodeStagedPositional(line, &row.ChangeSetID, &row.SequenceNo, &row.TargetKind, &row.TargetID, &row.OperationKind,
		&row.BeforeVersionID, &row.AfterVersionID, &row.BeforeValue, &row.AfterValue, &row.HistoryRecordIDs,
		&row.HistoryEntryRecordIDs, &row.DescribeFailed, &row.References, &row.RowRecord, &row.Declared)
	return row, err
}
func (mutationRowCodec) Columns() []string { return mutationColumns }
func (mutationRowCodec) Values(row stagedMutation) ([]any, error) {
	return []any{row.ChangeSetID, row.SequenceNo, row.TargetKind, row.TargetID, row.OperationKind,
		row.BeforeVersionID, row.AfterVersionID, stagedMutationValue(row.BeforeValue), stagedMutationValue(row.AfterValue),
		row.HistoryRecordIDs, row.HistoryEntryRecordIDs}, nil
}

// A portable mutation value of JSON null is stored as SQL NULL.
func stagedMutationValue(raw json.RawMessage) any {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil
	}
	return []byte(raw)
}

// stagedRevision is one write-ready record revision. Snapshots keep the exact
// bytes the live import stored, including a JSON null before the first version.
type stagedRevision struct {
	RevisionID  int64
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeJSON  json.RawMessage
	AfterJSON   json.RawMessage
	CreatedAt   time.Time
	Declared    []string
}

var revisionColumns = []string{"revision_id", "change_set_id", "record_id", "row_version", "before_json", "after_json", "created_at"}

type revisionRowCodec struct{}

func (revisionRowCodec) Encode(row stagedRevision) ([]byte, error) {
	return json.Marshal([]any{row.RevisionID, row.ChangeSetID, row.RecordID, row.RowVersion, row.BeforeJSON,
		row.AfterJSON, row.CreatedAt, row.Declared})
}
func (revisionRowCodec) Decode(line []byte) (stagedRevision, error) {
	var row stagedRevision
	err := decodeStagedPositional(line, &row.RevisionID, &row.ChangeSetID, &row.RecordID, &row.RowVersion,
		&row.BeforeJSON, &row.AfterJSON, &row.CreatedAt, &row.Declared)
	row.CreatedAt = row.CreatedAt.UTC()
	return row, err
}
func (revisionRowCodec) Columns() []string { return revisionColumns }
func (revisionRowCodec) Values(row stagedRevision) ([]any, error) {
	return []any{row.RevisionID, row.ChangeSetID, row.RecordID, row.RowVersion, []byte(row.BeforeJSON),
		[]byte(row.AfterJSON), row.CreatedAt}, nil
}

func decodeStagedPositional(line []byte, targets ...any) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil || len(raw) != len(targets) {
		return incidentportability.ErrPreparedStream
	}
	for index, target := range targets {
		if err := json.Unmarshal(raw[index], target); err != nil {
			return incidentportability.ErrPreparedStream
		}
	}
	return nil
}

// stagedJSON keeps the canonical bytes the live import passed to the database.
func stagedJSON(value any) (json.RawMessage, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, revisionsFailure(revisionsReferencesInvariant)
	}
	return encoded, nil
}

// Reference keys group one record's database-side checks: "r" requires its
// envelope and current source row, "k" a row provider for a target kind and
// "t" an envelope type equal to a snapshot's declared record type.
func referenceKey(recordID uuid.UUID, tag string, value string) string {
	if value == "" {
		return recordID.String() + "|" + tag
	}
	return recordID.String() + "|" + tag + "|" + value
}

func membershipKey(changeSetID uuid.UUID, reference bool) string {
	if reference {
		return changeSetID.String() + "|1"
	}
	return changeSetID.String() + "|0"
}

func versionKey(recordID uuid.UUID, revisionID, rowVersion int64) string {
	return fmt.Sprintf("%s|%020d|%020d", recordID, revisionID, rowVersion)
}

// declaredSnapshotType proves a snapshot's shape under the record type it
// declares. A nil snapshot declares nothing; the transaction then only
// requires the envelope's type to equal each declared type.
func (c *incidentBundleValidationCatalog) declaredSnapshotType(incidentID, recordID uuid.UUID, value any) (string, bool) {
	if value == nil {
		return "", true
	}
	snapshot, ok := value.(map[string]any)
	if !ok {
		return "", false
	}
	record, ok := snapshot["record"].(map[string]any)
	if !ok {
		return "", false
	}
	declared, ok := record["record_type"].(string)
	if !ok || c.validateSnapshot(incidentID, recordID, declared, value) != nil {
		return "", false
	}
	return declared, true
}

// stageKeys reads the staged streams once and stages the bounded sorted runs
// for change-set membership, record versions and record references.
func (p *preparedRevisionsImport) stageKeys(ctx context.Context, store incidentportability.MemberWriter, reader incidentportability.MemberReader) (membership, versions incidentportability.SortedKeyRuns, err error) {
	membershipKeys := incidentportability.NewSortedKeyStager(ctx, store, preparedMembershipPath, 0)
	versionKeys := incidentportability.NewSortedKeyStager(ctx, store, preparedVersionsPath, 0)
	referenceKeys := incidentportability.NewSortedKeyStager(ctx, store, preparedReferencesPath, 0)
	if err := incidentportability.ReadPreparedRows(ctx, reader, p.changeSets, changeSetRowCodec{}, func(row stagedChangeSet) error {
		return membershipKeys.Add(membershipKey(row.ChangeSetID, false))
	}); err != nil {
		return membership, versions, err
	}
	if err := incidentportability.ReadPreparedRows(ctx, reader, p.mutations, mutationRowCodec{}, func(row stagedMutation) error {
		if err := membershipKeys.Add(membershipKey(row.ChangeSetID, true)); err != nil {
			return err
		}
		for _, recordID := range row.References {
			if err := referenceKeys.Add(referenceKey(recordID, "r", "")); err != nil {
				return err
			}
		}
		if row.RowRecord == nil {
			return nil
		}
		keys := []string{referenceKey(*row.RowRecord, "r", ""), referenceKey(*row.RowRecord, "k", row.TargetKind)}
		for _, declared := range row.Declared {
			keys = append(keys, referenceKey(*row.RowRecord, "t", declared))
		}
		for _, key := range keys {
			if err := referenceKeys.Add(key); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return membership, versions, err
	}
	if err := incidentportability.ReadPreparedRows(ctx, reader, p.revisions, revisionRowCodec{}, func(row stagedRevision) error {
		keys := []string{referenceKey(row.RecordID, "r", "")}
		for _, declared := range row.Declared {
			keys = append(keys, referenceKey(row.RecordID, "t", declared))
		}
		for _, key := range keys {
			if err := referenceKeys.Add(key); err != nil {
				return err
			}
		}
		if err := membershipKeys.Add(membershipKey(row.ChangeSetID, true)); err != nil {
			return err
		}
		return versionKeys.Add(versionKey(row.RecordID, row.RevisionID, row.RowVersion))
	}); err != nil {
		return membership, versions, err
	}
	if membership, err = membershipKeys.Finish(); err != nil {
		return membership, versions, err
	}
	if versions, err = versionKeys.Finish(); err != nil {
		return membership, versions, err
	}
	p.references, err = referenceKeys.Finish()
	return membership, versions, err
}

// checkMembership records the membership verdict: every mutation and revision
// names a change set of this bundle.
func (p *preparedRevisionsImport) checkMembership(ctx context.Context, reader incidentportability.MemberReader, runs incidentportability.SortedKeyRuns) error {
	present := ""
	return incidentportability.MergeSortedKeys(ctx, reader, runs, func(key string) error {
		id, tag, ok := strings.Cut(key, "|")
		if !ok {
			return incidentportability.ErrPreparedStream
		}
		if tag == "0" {
			present = id
		} else if id != present {
			p.verdicts.Record(rankMembership, id, revisionsFailure(revisionsReferencesInvariant))
		}
		return nil
	})
}

// checkRecordVersions records the record-version verdict: in revision-identity
// order, each record's versions strictly increase.
func (p *preparedRevisionsImport) checkRecordVersions(ctx context.Context, reader incidentportability.MemberReader, runs incidentportability.SortedKeyRuns) error {
	record, previous := "", ""
	return incidentportability.MergeSortedKeys(ctx, reader, runs, func(key string) error {
		parts := strings.Split(key, "|")
		if len(parts) != 3 {
			return incidentportability.ErrPreparedStream
		}
		if parts[0] != record {
			record, previous = parts[0], parts[2]
			return nil
		}
		if parts[2] <= previous {
			p.verdicts.Record(rankRecordVersions, parts[0], revisionsFailure(revisionsRecordVersionInvariant))
		}
		previous = parts[2]
		return nil
	})
}

// revisionsReferenceGroup is one distinct referenced record and the database
// facts its references require.
type revisionsReferenceGroup struct {
	recordID uuid.UUID
	kinds    []string
	declared []string
}

func (p preparedRevisionsImport) visitReferenceGroups(ctx context.Context, reader incidentportability.MemberReader, visit func(revisionsReferenceGroup) error) error {
	var current *revisionsReferenceGroup
	err := incidentportability.MergeSortedKeys(ctx, reader, p.references, func(key string) error {
		idText, rest, ok := strings.Cut(key, "|")
		if !ok {
			return incidentportability.ErrPreparedStream
		}
		recordID, err := uuid.Parse(idText)
		if err != nil {
			return incidentportability.ErrPreparedStream
		}
		if current == nil || current.recordID != recordID {
			if current != nil {
				if err := visit(*current); err != nil {
					return err
				}
			}
			current = &revisionsReferenceGroup{recordID: recordID}
		}
		tag, value, _ := strings.Cut(rest, "|")
		switch tag {
		case "r":
		case "k":
			current.kinds = append(current.kinds, value)
		case "t":
			current.declared = append(current.declared, value)
		default:
			return incidentportability.ErrPreparedStream
		}
		return nil
	})
	if err != nil || current == nil {
		return err
	}
	return visit(*current)
}

// referenceOutcome carries database-side verdicts computed in the reference
// pass and reported at their own ranks.
type referenceOutcome struct {
	typeMismatch     bool
	terminalMismatch bool
}

// referenceReadPage bounds the reference groups resolved together, so the pass
// reads each owner once per page rather than once per record.
const referenceReadPage = 128

// checkReferencesTx resolves each distinct referenced record once: its envelope
// in this incident, its current source row and every required row provider.
// It returns the references failure at its rank and records the declared-type
// and terminal-row verdicts for later ranks. Groups are read a page at a time
// and judged in record order, so the first failing record decides the result.
func (p preparedRevisionsImport) checkReferencesTx(ctx context.Context, tx pgx.Tx, reader incidentportability.MemberReader, validation *incidentBundleValidationCatalog, terminal *terminalRevisionCursor) (referenceOutcome, error) {
	var outcome referenceOutcome
	page := make([]revisionsReferenceGroup, 0, referenceReadPage)
	flush := func() error {
		if len(page) == 0 {
			return nil
		}
		resolved, err := p.resolveReferencesTx(ctx, tx, validation, page)
		if err != nil {
			return err
		}
		for index, group := range page {
			if err := p.judgeReference(group, resolved[index], validation, terminal, &outcome); err != nil {
				return err
			}
		}
		page = page[:0]
		return nil
	}
	err := p.visitReferenceGroups(ctx, reader, func(group revisionsReferenceGroup) error {
		page = append(page, group)
		if len(page) < referenceReadPage {
			return nil
		}
		return flush()
	})
	if err == nil {
		err = flush()
	}
	return outcome, err
}

// resolvedReference is a group's record type and current source row; a nil
// row means the record does not resolve in this incident.
type resolvedReference struct {
	recordType string
	current    map[string]any
}

// resolveReferencesTx resolves a page of groups. A group whose declared types
// agree is read from that owner directly: the owner's current row carries its
// record envelope, so a row in this incident with the declared type is exactly
// what the envelope lookup and the owner read would return together. Every
// other group resolves its envelope type first and then its owner row. Each
// step is one read per record type for the page.
func (p preparedRevisionsImport) resolveReferencesTx(ctx context.Context, tx pgx.Tx, validation *incidentBundleValidationCatalog, page []revisionsReferenceGroup) ([]resolvedReference, error) {
	resolved := make([]resolvedReference, len(page))
	declared := make(map[string][]int)
	for index, group := range page {
		if len(group.declared) > 0 && !slices.ContainsFunc(group.declared, func(value string) bool { return value != group.declared[0] }) {
			declared[group.declared[0]] = append(declared[group.declared[0]], index)
		}
	}
	if err := readCurrentRowsTx(ctx, tx, validation, page, declared, resolved); err != nil {
		return nil, err
	}
	pending := make([]uuid.UUID, 0, len(page))
	for index, group := range page {
		if current := resolved[index].current; current != nil {
			envelope, ok := current["record"].(map[string]any)
			if ok && envelope["record_id"] == group.recordID.String() && envelope["incident_id"] == p.incidentID.String() && envelope["record_type"] == resolved[index].recordType {
				continue
			}
		}
		resolved[index] = resolvedReference{}
		pending = append(pending, group.recordID)
	}
	if len(pending) == 0 {
		return resolved, nil
	}
	types, err := validation.envelopes.RecordTypesTx(ctx, tx, p.incidentID, pending)
	if err != nil {
		return nil, err
	}
	byType := make(map[string][]int)
	for index, group := range page {
		if resolved[index].recordType != "" {
			continue
		}
		if recordType, found := types[group.recordID]; found {
			byType[recordType] = append(byType[recordType], index)
		}
	}
	if err := readCurrentRowsTx(ctx, tx, validation, page, byType, resolved); err != nil {
		return nil, err
	}
	return resolved, nil
}

// readCurrentRowsTx reads the current source rows of the indexed groups, one
// owner read per record type. A type with no snapshot owner resolves nothing.
func readCurrentRowsTx(ctx context.Context, tx pgx.Tx, validation *incidentBundleValidationCatalog, page []revisionsReferenceGroup, byType map[string][]int, resolved []resolvedReference) error {
	recordTypes := slices.Sorted(maps.Keys(byType))
	for _, recordType := range recordTypes {
		indexes := byType[recordType]
		recordIDs := make([]uuid.UUID, len(indexes))
		for position, index := range indexes {
			recordIDs[position] = page[index].recordID
		}
		rows, err := validation.currentSourceRowsTx(ctx, tx, recordIDs, recordType)
		if errors.Is(err, ErrUnexpectedHistoryTargetProvider) {
			continue
		}
		if err != nil {
			return err
		}
		for _, index := range indexes {
			if current, found := rows[page[index].recordID]; found {
				resolved[index] = resolvedReference{recordType: recordType, current: current}
			}
		}
	}
	return nil
}

// judgeReference applies one resolved group's verdicts in rank order.
func (p preparedRevisionsImport) judgeReference(group revisionsReferenceGroup, resolved resolvedReference, validation *incidentBundleValidationCatalog, terminal *terminalRevisionCursor, outcome *referenceOutcome) error {
	if resolved.current == nil {
		return revisionsFailure(revisionsReferencesInvariant)
	}
	for _, kind := range group.kinds {
		if _, err := validation.targets.rowProvider(kind, resolved.recordType); err != nil {
			return revisionsFailure(revisionsReferencesInvariant)
		}
	}
	if slices.ContainsFunc(group.declared, func(declared string) bool { return declared != resolved.recordType }) {
		outcome.typeMismatch = true
	}
	if terminal == nil {
		return nil
	}
	after, found, err := terminal.lookup(group.recordID)
	if err != nil || !found {
		return err
	}
	expected := normalizePortableSnapshotActors(after, p.actors, p.runtimeActor)
	if !canonicalHistoryRowMatchesSnapshot(group.recordID, expected, resolved.current) {
		outcome.terminalMismatch = true
	}
	return nil
}

// terminalRevisionCursor pages the incident's terminal revision of each record
// in record order, so the reference pass compares each record once.
type terminalRevisionCursor struct {
	ctx        context.Context
	tx         pgx.Tx
	incidentID uuid.UUID
	page       []terminalRevision
	index      int
	after      *uuid.UUID
	done       bool
}

type terminalRevision struct {
	recordID uuid.UUID
	after    []byte
}

const terminalRevisionPage = 128

func (c *terminalRevisionCursor) lookup(recordID uuid.UUID) (map[string]any, bool, error) {
	for {
		for c.index < len(c.page) {
			row := c.page[c.index]
			switch order := bytes.Compare(row.recordID[:], recordID[:]); {
			case order < 0:
				c.index++
			case order == 0:
				c.index++
				after, err := decodeDatabaseJSONObject(row.after)
				return after, err == nil && after != nil, err
			default:
				return nil, false, nil
			}
		}
		if c.done {
			return nil, false, nil
		}
		rows, err := c.tx.Query(c.ctx, `
SELECT DISTINCT ON (revision.record_id) revision.record_id, revision.after_json
FROM record_revisions revision JOIN change_sets changes USING (change_set_id)
WHERE changes.incident_id = $1 AND ($2::uuid IS NULL OR revision.record_id > $2)
ORDER BY revision.record_id, revision.row_version DESC
LIMIT $3
`, c.incidentID, c.after, terminalRevisionPage)
		if err != nil {
			return nil, false, err
		}
		page, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (terminalRevision, error) {
			var result terminalRevision
			return result, row.Scan(&result.recordID, &result.after)
		})
		if err != nil {
			return nil, false, err
		}
		c.page, c.index = page, 0
		c.done = len(page) < terminalRevisionPage
		if len(page) > 0 {
			last := page[len(page)-1].recordID
			c.after = &last
		}
	}
}
