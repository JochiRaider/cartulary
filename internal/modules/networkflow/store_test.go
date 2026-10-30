package networkflow_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

const (
	testSHA1 = "1111111111111111111111111111111111111111111111111111111111111111"
	testSHA2 = "2222222222222222222222222222222222222222222222222222222222222222"
	testSHA3 = "3333333333333333333333333333333333333333333333333333333333333333"
	testSHA4 = "4444444444444444444444444444444444444444444444444444444444444444"
)

func newTestNetworkFlowStore(
	t testing.TB,
	db postgres.DB,
	appender *revisions.Appender,
	options ...StoreOption,
) *Store {
	t.Helper()
	projection, err := projectionassembly.Build(db)
	if err != nil {
		t.Fatalf("compose Projections: %v", err)
	}
	indicatorOwner, err := indicators.NewApplication(indicators.ApplicationDependencies{
		ReferencePacks:  referencefixture.IndicatorRegistry{},
		Postgres:        db,
		Idempotency:     indicatorassembly.NewIdempotencyPort(authn.NewStore(db)),
		IncidentState:   admission.NewChecker(db),
		Revisions:       appender,
		RecordEnvelopes: records.NewStore(db),
		Projections:     projection.IndicatorPorts().Rows,
		SourceText:      indicatorassembly.NewSourceTextPort(projection.SourceTextRows()),
		Collaboration:   collaborationsupport.NewRecordChangedAppender(),
		Clock:           func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("compose Indicator test owner: %v", err)
	}
	defaults := []StoreOption{WithOwnerParticipants(
		incidents.NewTransactionParticipant(),
		authn.NewAdministrativeAuditAppender(),
		indicatorOwner,
	), WithSafeDigester(testNetworkFlowSafeDigester{}), WithResourceIntentAppender(networkflowsupport.NewResourceIntentAppender())}
	return NewStore(db, DefaultEffectiveLimits(), append(defaults, options...)...)
}

type testNetworkFlowSafeDigester struct{}

type invalidKeySafeDigester struct{}

func (invalidKeySafeDigester) Digest(valueClass, value string) (string, string, error) {
	return SafeDigest("missing", nil, valueClass, value)
}

type failingResourceIntentAppender struct {
	err error
}

func (a failingResourceIntentAppender) AppendResourceIntentTx(context.Context, pgx.Tx, ResourceIntent) error {
	return a.err
}

func (testNetworkFlowSafeDigester) Digest(valueClass string, canonicalValue string) (string, string, error) {
	return SafeDigest("test-safe-v1", []byte("0123456789abcdef0123456789abcdef"), valueClass, canonicalValue)
}

func assertNetworkFlowStoreCreateTableDerivesNamePersistsRowsAndCounts(t *testing.T) {
	t.Run("functional table entropy", assertNetworkFlowTableEntropyTransactions)
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-create")
	sessionID, unitID := seedImportSessionUnit(t, harness.DB, incidentID, actor.ID, "C:\\tmp\\flows.csv")
	store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
	now := time.Date(2026, 7, 10, 11, 30, 0, 0, time.UTC)

	table, err := store.CreateTable(context.Background(), CreateTableParams{
		IncidentID:                incidentID,
		ActorUserID:               actor.ID,
		ImportSessionID:           sessionID,
		ImportUnitID:              unitID,
		SourceContentSHA256:       testSHA1,
		OriginalFilename:          "C:\\tmp\\flows.csv",
		SourceFilenameDigest:      testSHA2,
		SourceFilenameDigestKeyID: "nf-test-key",
		MappingFingerprint:        testSHA3,
		SourceProfileID:           SourceProfileCiscoSNANetFlowCSV,
		ParserProfileID:           ParserProfileRFC4180HeaderedCSV,
		Rows: []FlowRow{
			testFlowRow(2, "2"),
			testFlowRow(1, "1"),
		},
		RowCountRejected: 1, Diagnostics: []RejectedRowDiagnostic{testDiagnostic(3)},
		Now: now,
	})
	if err != nil {
		t.Fatalf("create network flow table: %v", err)
	}
	if table.DisplayName != "flows" || table.SourceFilenameDisplay != "flows.csv" {
		t.Fatalf("unexpected derived names: %#v", table)
	}
	if table.TableStatus != TableStatusActive || table.TableVersion != 1 || table.RowCountAccepted != 2 || table.RowCountRejected != 1 {
		t.Fatalf("unexpected table lifecycle/counts: %#v", table)
	}

	listed, err := store.ListActiveTables(context.Background(), incidentID)
	if err != nil {
		t.Fatalf("list active tables: %v", err)
	}
	if len(listed) != 1 || listed[0].TableID != table.TableID {
		t.Fatalf("active table list got %#v want table %s", listed, table.TableID)
	}
	loadedRows := make([]FlowRow, 0, 2)
	for _, suffix := range []string{"1", "2"} {
		row, err := store.GetAcceptedRow(context.Background(), incidentID, table.TableID, testFlowRow(0, suffix).RowID)
		if err != nil {
			t.Fatalf("get accepted row: %v", err)
		}
		loadedRows = append(loadedRows, row)
	}
	if loadedRows[0].SourceRowNumber != 1 || loadedRows[1].SourceRowNumber != 2 {
		t.Fatalf("source row identity changed: %#v", loadedRows)
	}
	if _, err := store.GetAcceptedRow(context.Background(), incidentID, table.TableID, testFlowRow(0, "f").RowID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing row: %v", err)
	}
	if _, err := store.GetAcceptedRow(context.Background(), uuid.New(), table.TableID, loadedRows[0].RowID); err == nil {
		t.Fatal("cross-incident row admitted")
	}
	if _, err := store.GetAcceptedRow(context.Background(), incidentID, "nft_"+strings.Repeat("f", 32), loadedRows[0].RowID); err == nil {
		t.Fatal("cross-table row admitted")
	}
	var observation map[string]any
	if err := json.Unmarshal(loadedRows[0].ObservationSourceRef, &observation); err != nil {
		t.Fatalf("unmarshal observation source ref: %v", err)
	}
	if observation["import_session_id"] != sessionID.String() || observation["source_row_number"].(float64) != 1 {
		t.Fatalf("unexpected observation source ref: %#v", observation)
	}
	counts, err := store.RetainedCounts(context.Background(), incidentID)
	if err != nil {
		t.Fatalf("retained counts: %v", err)
	}
	if counts.Active != 1 || counts.Retained != 1 {
		t.Fatalf("counts got %#v want active=1 retained=1", counts)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM network_flow_rejected_row_diagnostics WHERE network_flow_table_id = $1`, table.TableID); got != 1 {
		t.Fatalf("diagnostic rows got %d want 1", got)
	}
}

func assertNetworkFlowStoreRenameSoftDeleteAndRetainedNameReuse(t *testing.T) {
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-lifecycle")
	store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
	first := createTestTable(t, harness.DB, store, actor.ID, incidentID, "flows.csv", nil, 1)
	second := createTestTable(t, harness.DB, store, actor.ID, incidentID, "flows.csv", nil, 2)
	if first.DisplayName != "flows" || second.DisplayName != "flows (2)" {
		t.Fatalf("unexpected suffix allocation: first=%q second=%q", first.DisplayName, second.DisplayName)
	}

	_, err := store.RenameTable(context.Background(), RenameTableParams{
		IncidentID:       incidentID,
		TableID:          first.TableID,
		BaseTableVersion: first.TableVersion,
		DisplayName:      "flows (2)",
		Now:              time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC),
	})
	var invalidName *InvalidDisplayNameError
	if !errors.As(err, &invalidName) || invalidName.ReasonCode != "duplicate_display_name" {
		t.Fatalf("duplicate rename got %T %[1]v want duplicate display-name error", err)
	}

	unchanged, err := store.RenameTable(context.Background(), RenameTableParams{
		IncidentID:       incidentID,
		TableID:          first.TableID,
		BaseTableVersion: first.TableVersion,
		DisplayName:      " flows ",
		Now:              time.Date(2026, 7, 10, 12, 1, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("same-name rename: %v", err)
	}
	if unchanged.TableVersion != first.TableVersion || !unchanged.UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("same-name rename must not change version/timestamp: before=%#v after=%#v", first, unchanged)
	}

	renamed, err := store.RenameTable(context.Background(), RenameTableParams{
		IncidentID:       incidentID,
		TableID:          first.TableID,
		BaseTableVersion: first.TableVersion,
		DisplayName:      "Renamed Flows",
		Now:              time.Date(2026, 7, 10, 12, 2, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("rename table: %v", err)
	}
	if renamed.DisplayName != "Renamed Flows" || renamed.TableVersion != first.TableVersion+1 {
		t.Fatalf("unexpected rename result: %#v", renamed)
	}
	_, err = store.RenameTable(context.Background(), RenameTableParams{
		IncidentID:       incidentID,
		TableID:          first.TableID,
		BaseTableVersion: first.TableVersion,
		DisplayName:      "Another Name",
		Now:              time.Date(2026, 7, 10, 12, 3, 0, 0, time.UTC),
	})
	if !errors.Is(err, ErrTableVersionConflict) {
		t.Fatalf("stale rename got %T %[1]v want version conflict", err)
	}

	deleted, err := store.SoftDeleteTable(context.Background(), SoftDeleteTableParams{
		IncidentID:       incidentID,
		TableID:          renamed.TableID,
		BaseTableVersion: renamed.TableVersion,
		Now:              time.Date(2026, 7, 10, 12, 4, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("soft delete table: %v", err)
	}
	if deleted.TableStatus != TableStatusSoftDeleted || deleted.DeletedAt == nil || deleted.TableVersion != renamed.TableVersion+1 {
		t.Fatalf("unexpected soft-delete result: %#v", deleted)
	}
	if _, err := store.GetActiveTable(context.Background(), incidentID, deleted.TableID); !errors.Is(err, ErrTableNotActive) {
		t.Fatalf("soft-deleted active lookup got %T %[1]v want not-active", err)
	}
	if _, err := store.GetAcceptedRow(context.Background(), incidentID, deleted.TableID, testFlowRow(0, "1").RowID); !errors.Is(err, ErrTableNotActive) {
		t.Fatalf("soft-deleted row lookup got %T %[1]v want not-active", err)
	}
	third := createTestTable(t, harness.DB, store, actor.ID, incidentID, "Renamed Flows.csv", nil, 3)
	if third.DisplayName != "Renamed Flows" {
		t.Fatalf("soft-deleted names must not reserve active display names, got %q", third.DisplayName)
	}
}

func TestNetworkFlowResourceIntentFailureRollsBackSourceMutation_Integration(t *testing.T) {
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-intent-rollback")
	appender := revisionsupport.MustAppender(t)
	store := newTestNetworkFlowStore(t, harness.DB, appender)
	table := createTestTable(t, harness.DB, store, actor.ID, incidentID, "flows.csv", nil, 1)
	t.Run("invalid digest key rolls back mutation and effects", func(t *testing.T) {
		count := func(table string) int {
			t.Helper()
			var n int
			if err := harness.DB.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		audits, receipts := count("deployment_admin_audit_events"), count("route_idempotency")
		broken := newTestNetworkFlowStore(t, harness.DB, appender, WithSafeDigester(invalidKeySafeDigester{}))
		_, err := broken.RenameTable(context.Background(), RenameTableParams{IncidentID: incidentID, ActorUserID: actor.ID, TableID: table.TableID, BaseTableVersion: table.TableVersion, DisplayName: "must not commit", ClientTxnID: "digest-failure", RequestID: "digest-failure", Now: time.Now().UTC()})
		if err == nil {
			t.Fatal("missing digest key accepted")
		}
		current, err := store.GetActiveTable(context.Background(), incidentID, table.TableID)
		if err != nil || current.DisplayName != table.DisplayName || current.TableVersion != table.TableVersion {
			t.Fatalf("digest failure changed table: %#v %v", current, err)
		}
		if count("deployment_admin_audit_events") != audits || count("route_idempotency") != receipts {
			t.Fatal("digest failure committed effects")
		}
	})
	intentFailure := errors.New("reject invalid extension resource payload")
	failingStore := newTestNetworkFlowStore(
		t,
		harness.DB,
		appender,
		WithResourceIntentAppender(failingResourceIntentAppender{err: intentFailure}),
	)

	_, err := failingStore.RenameTable(context.Background(), RenameTableParams{
		IncidentID:       incidentID,
		TableID:          table.TableID,
		BaseTableVersion: table.TableVersion,
		DisplayName:      "must roll back",
		Now:              time.Date(2026, 7, 10, 12, 30, 0, 0, time.UTC),
	})
	if !errors.Is(err, intentFailure) {
		t.Fatalf("rename intent failure = %v want %v", err, intentFailure)
	}
	unchanged, err := store.GetActiveTable(context.Background(), incidentID, table.TableID)
	if err != nil {
		t.Fatalf("load table after rejected intent: %v", err)
	}
	if unchanged.DisplayName != table.DisplayName || unchanged.TableVersion != table.TableVersion || !unchanged.UpdatedAt.Equal(table.UpdatedAt) {
		t.Fatalf("source mutation escaped rejected intent rollback: before=%#v after=%#v", table, unchanged)
	}
}

func assertNetworkFlowStoreLimitsUseActiveAndRetainedCounts(t *testing.T) {
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-limits")
	store := newTestNetworkFlowStore(
		t,
		harness.DB,
		revisionsupport.MustAppender(t),
		WithEffectiveLimits(func() EffectiveLimits {
			limits := DefaultEffectiveLimits()
			limits.MaxActiveTablesPerIncident = 1
			limits.MaxRetainedTablesPerIncident = 2
			return limits
		}()),
	)
	first := createTestTable(t, harness.DB, store, actor.ID, incidentID, "one.csv", nil, 1)

	_, err := createTestTableResult(t, harness.DB, store, actor.ID, incidentID, "two.csv", nil, 2)
	var limitErr *TableLimitError
	if !errors.As(err, &limitErr) || limitErr.LimitName != "network_flow.max_active_tables_per_incident" {
		t.Fatalf("active limit got %T %[1]v want active table limit", err)
	}
	deletedFirst, err := store.SoftDeleteTable(context.Background(), SoftDeleteTableParams{
		IncidentID:       incidentID,
		TableID:          first.TableID,
		BaseTableVersion: first.TableVersion,
		Now:              time.Date(2026, 7, 10, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("soft delete first table: %v", err)
	}
	second := createTestTable(t, harness.DB, store, actor.ID, incidentID, "two.csv", nil, 2)
	if _, err := store.SoftDeleteTable(context.Background(), SoftDeleteTableParams{
		IncidentID:       incidentID,
		TableID:          second.TableID,
		BaseTableVersion: second.TableVersion,
		Now:              time.Date(2026, 7, 10, 13, 1, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("soft delete second table: %v", err)
	}
	_, err = createTestTableResult(t, harness.DB, store, actor.ID, incidentID, "three.csv", nil, 3)
	if !errors.As(err, &limitErr) || limitErr.LimitName != "network_flow.max_retained_tables_per_incident" {
		t.Fatalf("retained limit got %T %[1]v want retained table limit", err)
	}
	counts, err := store.RetainedCounts(context.Background(), incidentID)
	if err != nil {
		t.Fatalf("retained counts: %v", err)
	}
	if counts.Active != 0 || counts.Retained != 2 || deletedFirst.TableStatus != TableStatusSoftDeleted {
		t.Fatalf("counts after retained-limit setup got %#v", counts)
	}
}

func assertNetworkFlowStoreRejectsInvalidExplicitDisplayNames(t *testing.T) {
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-invalid-names")
	store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))

	for _, testCase := range []struct {
		name       string
		value      string
		reasonCode string
	}{
		{name: "whitespace-control", value: " \t ", reasonCode: "forbidden_control"},
		{name: "empty", value: " \u3000 ", reasonCode: "empty_display_name"},
		{name: "too-long", value: strings.Repeat("a", 65), reasonCode: "display_name_too_long"},
		{name: "control", value: "bad\u007fvalue", reasonCode: "forbidden_control"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := createTestTableResult(t, harness.DB, store, actor.ID, incidentID, testCase.name+".csv", &testCase.value, len(testCase.name))
			var invalidName *InvalidDisplayNameError
			if !errors.As(err, &invalidName) || invalidName.ReasonCode != testCase.reasonCode {
				t.Fatalf("got %T %[1]v want invalid display name reason %q", err, testCase.reasonCode)
			}
		})
	}
}

func assertNetworkFlowCommittedRowsAreImmutable(t *testing.T) {
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-immutable")
	store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
	table := createTestTable(t, harness.DB, store, actor.ID, incidentID, "immutable.csv", nil, 1)
	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin immutable check transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	_, err = tx.Exec(context.Background(), `UPDATE network_flow_rows SET bytes_count = '2' WHERE network_flow_table_id = $1`, table.TableID)
	if err == nil {
		t.Fatalf("expected immutable row update to fail")
	}
}

func assertNetworkFlowDisplayNameAlgorithmsAndLifecycleVocabulary(t *testing.T) {
	cases := []struct {
		input      string
		sourceName string
		tableName  string
	}{
		{input: `C:\tmp\flows.csv`, sourceName: "flows.csv", tableName: "flows"},
		{input: `/tmp/flows.csv`, sourceName: "flows.csv", tableName: "flows"},
		{input: `.csv`, sourceName: ".csv", tableName: ".csv"},
		{input: `file.`, sourceName: "file.", tableName: "file"},
		{input: "///\u0000\t", sourceName: "uploaded.csv", tableName: "uploaded"},
	}
	for _, testCase := range cases {
		if got := SanitizeSourceFilenameDisplay(testCase.input); got != testCase.sourceName {
			t.Fatalf("sanitize %q got %q want %q", testCase.input, got, testCase.sourceName)
		}
		displayName, err := DeriveTableDisplayName(testCase.input, map[string]struct{}{})
		if err != nil {
			t.Fatalf("derive display name for %q: %v", testCase.input, err)
		}
		if displayName != testCase.tableName {
			t.Fatalf("derive %q got %q want %q", testCase.input, displayName, testCase.tableName)
		}
	}
	states := LifecycleStates()
	if len(states) != 2 || states[0] != TableStatusActive || states[1] != TableStatusSoftDeleted {
		t.Fatalf("unexpected lifecycle states: %#v", states)
	}
	for _, state := range states {
		if state == "renamed" {
			t.Fatalf("renamed must not be a lifecycle state")
		}
	}
}

func TestNetworkFlow_TableLifecycleVocabularyHasNoRenamedState_Unit(t *testing.T) {
	states := LifecycleStates()
	if len(states) != 2 || states[0] != TableStatusActive || states[1] != TableStatusSoftDeleted {
		t.Fatalf("unexpected lifecycle states: %#v", states)
	}
	for _, state := range states {
		if state == "renamed" {
			t.Fatalf("renamed must not be a lifecycle state")
		}
	}
}

func TestNetworkFlow_StoreLifecycleAndLimitAccounting_Integration(t *testing.T) {
	t.Run("exact row lookup does not enumerate and rechecks deletion", func(t *testing.T) {
		harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-exact-row")
		store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
		table := createTestTable(t, harness.DB, store, actor.ID, incidentID, "lookup.csv", nil, 1)
		rowID := testFlowRow(0, rowIDSuffix(1)).RowID
		lookup := newTestNetworkFlowStore(t, rowLookupOnlyDB{DB: harness.DB}, revisionsupport.MustAppender(t))
		row, err := lookup.GetAcceptedRow(context.Background(), incidentID, table.TableID, rowID)
		if err != nil || row.RowID != rowID {
			t.Fatalf("exact lookup used enumeration or lost identity: %v", err)
		}
		racing := newTestNetworkFlowStore(t, rowLookupOnlyDB{DB: harness.DB, deleteBeforeRead: true, t: t}, revisionsupport.MustAppender(t))
		if _, err := racing.GetAcceptedRow(context.Background(), incidentID, table.TableID, rowID); !errors.Is(err, ErrTableNotActive) {
			t.Fatalf("deleted source admitted during lookup: %v", err)
		}
	})
	t.Run("diagnostic identity scoped to table", func(t *testing.T) {
		harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-diagnostic-identity")
		store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
		other := storetest.CreateIncidentInStore(t, harness.Incidents, actor, "txn-diagnostic-other", "IR-NF-DIAGNOSTIC-OTHER", "Other diagnostics")
		for i, incidentID := range []uuid.UUID{incidentID, incidentID, other.Incident.ID} {
			session, unit := seedImportSessionUnit(t, harness.DB, incidentID, actor.ID, "same-source.csv")
			table, err := store.CreateTable(context.Background(), CreateTableParams{
				IncidentID: incidentID, ActorUserID: actor.ID, ImportSessionID: session, ImportUnitID: unit,
				SourceContentSHA256: testSHA1, OriginalFilename: "same-source.csv", SourceFilenameDigest: testSHA2,
				SourceFilenameDigestKeyID: "nf-test-key", MappingFingerprint: testSHA3,
				SourceProfileID: SourceProfileCiscoSNANetFlowCSV, ParserProfileID: ParserProfileRFC4180HeaderedCSV,
				Rows: []FlowRow{testFlowRow(2, []string{"1", "2", "3"}[i])}, RowCountRejected: 1, Diagnostics: []RejectedRowDiagnostic{testDiagnostic(3)},
				Now: time.Date(2026, 7, 10, 11, 30, i, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("same diagnostic in table %d: %v", i, err)
			}
			diagnostics, more, err := store.QueryRejectedDiagnosticsPage(context.Background(), incidentID, table.TableID, RejectedRowsQueryRequest{}, nil, 10)
			if err != nil || more || len(diagnostics) != 1 || ValidateRetainedExtensionState(context.Background(), harness.DB) != nil {
				t.Fatalf("stored diagnostic validity: %v", err)
			}
			wrongIncident := other.Incident.ID
			if wrongIncident == incidentID {
				wrongIncident = uuid.New()
			}
			hidden, _, err := store.QueryRejectedDiagnosticsPage(context.Background(), wrongIncident, table.TableID, RejectedRowsQueryRequest{}, nil, 10)
			if err != nil || len(hidden) != 0 {
				t.Fatalf("cross-incident diagnostics escaped: %v", err)
			}
			tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = tx.Exec(context.Background(), `INSERT INTO network_flow_rejected_row_diagnostics SELECT * FROM network_flow_rejected_row_diagnostics WHERE network_flow_table_id = $1`, table.TableID)
			_ = tx.Rollback(context.Background())
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
				t.Fatalf("same-table duplicate outcome: %v", err)
			}
		}
	})
	t.Run("create_rows_counts", assertNetworkFlowStoreCreateTableDerivesNamePersistsRowsAndCounts)
	t.Run("rename_soft_delete_reuse", assertNetworkFlowStoreRenameSoftDeleteAndRetainedNameReuse)
	t.Run("limits", assertNetworkFlowStoreLimitsUseActiveAndRetainedCounts)
	t.Run("invalid_names", assertNetworkFlowStoreRejectsInvalidExplicitDisplayNames)
	t.Run("immutable_rows", assertNetworkFlowCommittedRowsAreImmutable)
	t.Run("vocabulary", assertNetworkFlowDisplayNameAlgorithmsAndLifecycleVocabulary)
}

type rowLookupOnlyDB struct {
	postgres.DB
	deleteBeforeRead bool
	t                testing.TB
}

func (db rowLookupOnlyDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("single-row lookup attempted enumeration")
}

func (db rowLookupOnlyDB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if db.deleteBeforeRead && len(args) == 3 {
		if _, err := db.DB.Exec(ctx, `UPDATE network_flow_tables SET table_status = 'soft_deleted', table_version = table_version + 1, deleted_at = now() WHERE incident_id = $1 AND network_flow_table_id = $2`, args[0], args[1]); err != nil {
			db.t.Fatal(err)
		}
	}
	return db.DB.QueryRow(ctx, query, args...)
}

func startNetworkFlowStoreTest(t testing.TB, prefix string) (*storetest.StoreHarness, authn.UserRecord, uuid.UUID) {
	t.Helper()
	harness := storetest.StartStore(t, prefix)
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: prefix + "@example.test", DisplayName: "Network Flow Tester"})
	result := storetest.CreateIncidentInStore(
		t, harness.Incidents, actor,
		"txn-"+prefix,
		"IR-"+strings.ToUpper(strings.ReplaceAll(prefix, "-", "")),
		"Network Flow "+prefix,
	)

	return harness, actor, result.Incident.ID
}

func seedImportSessionUnit(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorID uuid.UUID, originalFilename string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	return seedImportSessionUnitForSequence(t, db, incidentID, actorID, originalFilename, 1)
}

func createTestTable(t testing.TB, db postgres.DB, store *Store, actorID uuid.UUID, incidentID uuid.UUID, filename string, displayNameOverride *string, sequence int) TableRecord {
	t.Helper()
	table, err := createTestTableResult(t, db, store, actorID, incidentID, filename, displayNameOverride, sequence)
	if err != nil {
		t.Fatalf("create test table %s: %v", filename, err)
	}
	return table
}

func createTestTableResult(t testing.TB, db postgres.DB, store *Store, actorID uuid.UUID, incidentID uuid.UUID, filename string, displayNameOverride *string, sequence int) (TableRecord, error) {
	t.Helper()
	sessionID, unitID := seedImportSessionUnitForSequence(t, db, incidentID, actorID, filename, sequence)
	return store.CreateTable(context.Background(), CreateTableParams{
		IncidentID:                incidentID,
		ActorUserID:               actorID,
		ImportSessionID:           sessionID,
		ImportUnitID:              unitID,
		SourceContentSHA256:       testSHA1,
		OriginalFilename:          filename,
		SourceFilenameDigest:      testSHA2,
		SourceFilenameDigestKeyID: "nf-test-key",
		MappingFingerprint:        testSHA3,
		SourceProfileID:           SourceProfileCiscoSNANetFlowCSV,
		ParserProfileID:           ParserProfileRFC4180HeaderedCSV,
		DisplayNameOverride:       displayNameOverride,
		Rows:                      []FlowRow{testFlowRow(int64(sequence), rowIDSuffix(sequence))},
		Now:                       time.Date(2026, 7, 10, 11, sequence%60, 0, 0, time.UTC),
	})
}

func seedImportSessionUnitForSequence(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorID uuid.UUID, filename string, sequence int) (uuid.UUID, uuid.UUID) {
	t.Helper()
	sessionID := uuid.New()
	unitID := uuid.New()
	now := time.Date(2026, 7, 10, 10, sequence%60, 0, 0, time.UTC)
	if _, err := db.Exec(context.Background(), `
INSERT INTO import_sessions (
    import_session_id, incident_id, created_by_user_id, client_txn_id, assistant_profile,
    source_file_kind, original_filename, source_content_sha256, source_media_type, source_byte_size,
    parser_profile_id, parser_version, session_status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'network_flow_test', 'csv', $5, $6, 'text/csv', 12,
    $7, 'test', 'ready_to_apply', $8, $8
)
`, sessionID, incidentID, actorID, "txn-import-"+unitID.String(), filename, testSHA1, ParserProfileRFC4180HeaderedCSV, now); err != nil {
		t.Fatalf("seed import session: %v", err)
	}
	if _, err := db.Exec(context.Background(), `
INSERT INTO import_units (
    import_unit_id, import_session_id, unit_status, locator_kind, locator, source_rect_a1,
    header_row_ref, data_start_row_ref, inferred_row_count, inferred_column_count,
    warning_codes, mapping_fingerprint, approved_mapping_json, columns_json, source_rows_json,
    preview_rows_json, approved_target_kind, approved_extension_profile_id, discovery_sequence,
    created_at, updated_at
) VALUES (
    $1, $2, 'ready', 'csv', 'unit-1', 'A1:Z2', 1, 2, 1, 9,
    '{}', $3, '{}'::jsonb, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb,
    'network_flow_table', $4, 1, $5, $5
)
`, unitID, sessionID, testSHA3, ProfileID, now); err != nil {
		t.Fatalf("seed import unit: %v", err)
	}
	return sessionID, unitID
}

func testFlowRow(sourceRowNumber int64, suffix string) FlowRow {
	srcPort := int32(443)
	dstPort := int32(51515)
	tcpFlags := int32(24)
	return FlowRow{
		RowID:                     "nfr_" + strings.Repeat(suffix, 64),
		SourceRowNumber:           sourceRowNumber,
		SourceRowDigestSHA256:     testSHA1,
		NormalizedRowDigestSHA256: testSHA4,
		FlowStartUTC:              time.Date(2026, 7, 10, 9, int(sourceRowNumber), 0, 0, time.UTC),
		FlowEndUTC:                time.Date(2026, 7, 10, 9, int(sourceRowNumber), 30, 0, time.UTC),
		SrcIP:                     "192.0.2.10",
		DstIP:                     "2001:db8::1",
		SrcPort:                   &srcPort,
		DstPort:                   &dstPort,
		IPProtocol:                6,
		BytesCount:                "42",
		PacketsCount:              "2",
		ExporterID:                stringPtr("collector-a"),
		InputInterface:            stringPtr("inside"),
		OutputInterface:           stringPtr("outside"),
		TCPFlags:                  &tcpFlags,
		ApplicationLabel:          stringPtr("https"),
		UnmappedRaw:               json.RawMessage(`{}`),
	}
}

func testDiagnostic(sourceRowNumber int64) RejectedRowDiagnostic {
	sourceColumn := int64(2)
	fieldKey := FieldSrcIP
	const code = "network_flow_invalid_ip"
	const reason = "missing_or_empty"
	const message = "network_flow.diagnostic." + code + "." + reason
	// Independent fixture bytes follow the owner stream, including its final NUL.
	parts := []string{"cartulary.network_flow.diagnostic_id.v1", testSHA1, testSHA3, strconv.FormatInt(sourceRowNumber, 10), "p:2", "p:" + fieldKey, code, reason}
	identity := sha256.Sum256([]byte(strings.Join(parts, "\x00") + "\x00"))
	rawValue := sha256.Sum256(nil)
	return RejectedRowDiagnostic{
		DiagnosticID: "nfd_" + hex.EncodeToString(identity[:]), SourceRowNumber: sourceRowNumber,
		SourceColumnOrdinal: &sourceColumn, RawHeaderSHA256: stringPtr(testSHA2), FieldKey: &fieldKey,
		ErrorCode: code, ReasonCode: reason, RawValueSHA256: stringPtr(hex.EncodeToString(rawValue[:])),
		MessageKey: message, MessageArgs: json.RawMessage(`{}`), Message: message,
	}
}

func rowIDSuffix(sequence int) string {
	const digits = "0123456789abcdef"
	index := sequence % len(digits)
	return digits[index : index+1]
}

func stringPtr(value string) *string {
	return &value
}
