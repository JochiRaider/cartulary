package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// countingBundle counts admitted bundle-member opens after preparation has
// returned, which is before the final transaction begins.
type countingBundle struct {
	incidentportability.MemberReader
	frozen atomic.Bool
	late   atomic.Int64
}

func (c *countingBundle) OpenMember(ctx context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	if c.frozen.Load() {
		c.late.Add(1)
	}
	return c.MemberReader.OpenMember(ctx, path)
}

// portableWriteFakeTx records COPY statements. A per-row statement is the
// write path this slice removes, so it is recorded and refused.
type portableWriteFakeTx struct {
	pgx.Tx
	copied   map[string]int64
	perRow   []string
	rejected error
}

func (tx *portableWriteFakeTx) CopyFrom(_ context.Context, table pgx.Identifier, _ []string, source pgx.CopyFromSource) (int64, error) {
	var count int64
	for source.Next() {
		if _, err := source.Values(); err != nil {
			return count, err
		}
		count++
	}
	if err := source.Err(); err != nil {
		return count, err
	}
	tx.copied[strings.Join(table, ".")] += count
	return count, nil
}

func (tx *portableWriteFakeTx) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	tx.perRow = append(tx.perRow, sql)
	return pgconn.CommandTag{}, tx.rejected
}

func (tx *portableWriteFakeTx) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	tx.perRow = append(tx.perRow, sql)
	return rejectedRow{err: tx.rejected}
}

type rejectedRow struct{ err error }

func (r rejectedRow) Scan(...any) error { return r.err }

// recordingPortableAttributions is the attribution owner's adapter double.
type recordingPortableAttributions struct {
	flushed int64
}

func (r *recordingPortableAttributions) FlushImportedAttributions(ctx context.Context, source incidentportability.MemberReader, streams ...incidentportability.AttributionStream) error {
	reader := incidentportability.OpenAttributions(ctx, source, streams...)
	for reader.Next() {
		r.flushed++
	}
	return errors.Join(reader.Err(), reader.Close())
}

func (r *recordingPortableAttributions) CountImportedAttributions(context.Context, []string) (map[string]int64, error) {
	return nil, errors.New("final write must not count attributions")
}

// checkPortableWriteOpensNoBundleMember is the WU red test "Apply opens no
// bundle member after Begin" (Core 01 REQ-01-641): after preparation returns,
// the final write opens no admitted bundle member and writes only staged rows.
func checkPortableWriteOpensNoBundleMember(t *testing.T) {
	invocation, members := portableAdmissionFixture(t)
	bundle := &countingBundle{MemberReader: members}
	invocation.Members = bundle
	participant := &portabilityParticipant{limits: defaultLimits(), digester: &queryCompletionProbe{}}
	prepared, err := participant.PrepareImport(t.Context(), invocation, nil)
	if err != nil {
		t.Fatal(err)
	}
	bundle.frozen.Store(true)
	tx := &portableWriteFakeTx{copied: map[string]int64{}, rejected: errors.New("per-row statement in the final transaction")}
	attributions := &recordingPortableAttributions{}
	module := &Module{safeDigester: &queryCompletionProbe{}}
	_, write, err := module.PortabilityTransactionCapabilities(tx, func(context.Context, uuid.UUID, uuid.UUID) (incidentportability.AttributionRecorder, error) {
		return attributions, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := prepared.TransactionParticipant.Write(t.Context(), crossownertransaction.Invocation{Phase: "write", WriteAccess: write})
	if opened := bundle.late.Load(); opened != 0 {
		t.Fatalf("Apply opened %d bundle members after Begin", opened)
	}
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	if tx.copied["network_flow_tables"] != 1 || tx.copied["network_flow_rows"] != 2 || len(tx.perRow) != 0 || attributions.flushed != 3 {
		t.Fatalf("final write copied %v, ran %d per-row statements and flushed %d", tx.copied, len(tx.perRow), attributions.flushed)
	}
}

// portableFixtureSourceSet serves several committed sources with one CSV body.
type portableFixtureSourceSet map[uuid.UUID]portableFixtureSources

func (s portableFixtureSourceSet) ReadSource(ctx context.Context, id uuid.UUID) (imports.CommittedSource, error) {
	source, ok := s[id]
	if !ok {
		return imports.CommittedSource{}, errors.New("source not found")
	}
	return source.ReadSource(ctx, id)
}
func (s portableFixtureSourceSet) OpenSource(ctx context.Context, id uuid.UUID) (imports.CommittedSource, io.ReadCloser, error) {
	source, ok := s[id]
	if !ok {
		return imports.CommittedSource{}, nil, errors.New("source not found")
	}
	return source.OpenSource(ctx, id)
}

// portableSharedNameFixture admits an active table and a soft-deleted table
// with the same display name, the active one first in identity order: valid
// retained state, since only active names are unique.
func portableSharedNameFixture(t *testing.T, incidentID, localActor uuid.UUID) incidentbundles.ImportInvocation {
	t.Helper()
	limits := defaultLimits()
	f, request, sourceFixture := streamingApplyFixture(t, validPreviewRow+validPreviewRow, limits)
	request.IncidentID = incidentID
	request.ActorUserID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	request.ImportSessionID = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	request.ImportUnitID = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	applied, err := f.prepareImportApply(t.Context(), nil, request)
	if err != nil {
		t.Fatal(err)
	}
	defer applied.close()
	params := applied.params
	rowContext, err := validatedRowImportContext(params)
	if err != nil {
		t.Fatal(err)
	}
	var normalized []normalizedRow
	summary, err := params.ImportRows(nil, func(row normalizedRow) error {
		normalized = append(normalized, row)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	envelope := mappingTestJSON(t, map[string]any{"owner_mapping": json.RawMessage(request.OwnerMapping), "owner_mapping_schema_id": importMappingSchemaID})
	members := portableFixtureMembers{MapBundle: portabilitysupport.MapBundle{}}
	for _, path := range portabilityFamilyPaths {
		members.MapBundle[portabilityMemberPrefix+path] = []byte{}
	}
	encode := func(path string, value any) {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		members.MapBundle[portabilityMemberPrefix+path] = append(members.MapBundle[portabilityMemberPrefix+path], append(raw, '\n')...)
	}
	sources := portableFixtureSourceSet{}
	deleted := params.Now.Add(time.Hour)
	for index, status := range []string{tableStatusActive, "soft_deleted"} {
		committed := params.CommittedSourceID
		if index > 0 {
			committed = uuid.MustParse("55555555-5555-4555-8555-555555555555")
		}
		table := tableRecord{TableID: "nft_" + strings.Repeat(string(rune('a'+index)), 32), IncidentID: incidentID, DisplayName: "Shared flow source", TableVersion: 1, TableStatus: status, CommittedSourceID: committed, SourceInterpretationSHA256: summary.SourceInterpretationSHA256, SourceRawCellScalarLimit: params.SourceRawCellScalarLimit, SourceContentSHA256: params.SourceContentSHA256, SourceFilenameDisplay: "source.csv", MappingFingerprint: params.MappingFingerprint, SourceProfileID: params.SourceProfileID, ParserProfileID: params.ParserProfileID, RowCountAccepted: 2, CreatedByUserID: request.ActorUserID, CreatedAt: params.Now, UpdatedAt: params.Now}
		if status != tableStatusActive {
			table.TableVersion, table.UpdatedAt, table.DeletedAt = 2, deleted, &deleted
		}
		encode("tables.ndjson", portableTableRecord(table))
		rowSource := rowContext
		rowSource.committedSourceID, rowSource.committedSourceText = committed, committed.String()
		var rows []flowRow
		for _, row := range normalized {
			retained, err := rowSource.persistedRow(table.TableID, row)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, retained)
		}
		sort.Slice(rows, func(i, j int) bool { return comparePortableRows(rows[i], rows[j]) < 0 })
		for _, row := range rows {
			encode("observations.ndjson", portableFlowRow(row))
		}
		sources[committed] = portableFixtureSources{source: imports.CommittedSource{CommittedSourceID: committed, IncidentID: incidentID, ContentSHA256: request.ExpectedSourceContentSHA256, MappingFingerprint: request.MappingFingerprint, TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID, SourceFileKind: "csv", OriginalFilename: "source.csv", HeaderRowRef: 1, DataStartRowRef: 2, LocatorKind: "csv_file", Locator: "file", CreatedByUserID: request.ActorUserID, ApprovedMapping: envelope, CommittedAt: params.Now}, bytes: []byte(sourceFixture.source)}
	}
	catalog, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: request.ActorUserID.String(), DisplayName: "Historical author"}})
	if err != nil {
		t.Fatal(err)
	}
	dependencies, err := NewPortabilityDependencies(incidentID, sources, portableFixtureIndicators{})
	if err != nil {
		t.Fatal(err)
	}
	return incidentbundles.ImportInvocation{PortabilityImportContext: incidentbundles.PortabilityImportContext{OperationID: uuid.NewString(), IncidentID: incidentID, ActorUserID: localActor, Actors: catalog}, ProfileID: ProfileID, ContractMajor: 10, ClaimState: incidentbundles.ClaimStateClaimed, StateVersion: 7, ContentSchemaID: portabilityContentSchema, CatalogSHA256: strings.Repeat("b", 64), Members: members, Dependencies: dependencies, PreparationMembers: portableFixtureMembers{MapBundle: portabilitysupport.MapBundle{}}}
}

// AssertPortableImportWritesFinalTableStatus is the D8 regression: an active
// and a soft-deleted table share a display name and the active one sorts
// first. Each table must be written with its final status; writing every table
// as active first collides on the active display-name index.
func AssertPortableImportWritesFinalTableStatus(t *testing.T, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}, incidentID, localActor uuid.UUID) {
	t.Helper()
	invocation := portableSharedNameFixture(t, incidentID, localActor)
	prepared, err := (&portabilityParticipant{limits: defaultLimits(), digester: &queryCompletionProbe{}}).PrepareImport(t.Context(), invocation, nil)
	if err != nil {
		t.Fatal("shared-name retained state was not admitted", err)
	}
	tx, err := db.BeginTx(t.Context(), pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(t.Context())) }()
	module := &Module{safeDigester: &queryCompletionProbe{}}
	_, write, err := module.PortabilityTransactionCapabilities(tx, func(ctx context.Context, incident, actor uuid.UUID) (incidentportability.AttributionRecorder, error) {
		return incidentbundles.NewImportedAttributionRecorder(ctx, tx, incident, actor)
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := prepared.TransactionParticipant.Write(t.Context(), crossownertransaction.Invocation{Phase: "write", WriteAccess: write}); err != nil {
		t.Fatal("portable import of an active and a soft-deleted table sharing a name failed", err)
	}
	rows, err := tx.Query(t.Context(), `SELECT network_flow_table_id, table_status, created_by_user_id FROM network_flow_tables WHERE incident_id=$1 ORDER BY network_flow_table_id`, incidentID)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for rows.Next() {
		var id, status string
		var creator uuid.UUID
		if err := rows.Scan(&id, &status, &creator); err != nil {
			t.Fatal(err)
		}
		if creator != localActor {
			t.Fatal("imported table is not created by the importing actor")
		}
		got = append(got, id[:5]+":"+status)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != "nft_a:active,nft_b:soft_deleted" {
		t.Fatalf("written table states = %v", got)
	}
	var attributions int
	if err := tx.QueryRow(t.Context(), `SELECT count(*) FROM incident_bundle_imported_attributions WHERE incident_id=$1`, incidentID).Scan(&attributions); err != nil || attributions != 6 {
		t.Fatalf("historical attributions = %d (%v), want 2 tables and 4 observations", attributions, err)
	}
}
