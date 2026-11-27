package imports_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtureartifacts"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// The scaling sweep is retained capacity evidence for the Network Flow default
// row limit and the default cross-owner participant budget. It times the real
// public import path at several sizes. Each size gets a fresh runtime so btree
// depth and WAL history never carry over. Only the default-budget case records
// its outcome instead of asserting it; every raised-budget case must commit.
const networkFlowScalingArtifact = "network-flow-import-scaling.json"

type networkFlowScalingCase struct {
	Name          string `json:"name"`
	Rows          int    `json:"rows"`
	BudgetSeconds int    `json:"participant_budget_seconds"`
	DefaultBudget bool   `json:"default_budget"`
}

type networkFlowScalingPass struct {
	Bytes     int64   `json:"bytes"`
	ReadMS    float64 `json:"read_ms"`
	ElapsedMS float64 `json:"elapsed_ms"`
}

type networkFlowScalingCopy struct {
	Calls        int64   `json:"calls"`
	Rows         int64   `json:"rows"`
	WallMS       float64 `json:"wall_ms"`
	MaxCallMS    float64 `json:"max_call_ms"`
	SourceMS     float64 `json:"source_callback_ms"`
	CommitMS     float64 `json:"commit_ms"`
	Committed    bool    `json:"committed"`
	RolledBackMS float64 `json:"rollback_ms,omitempty"`
}

type networkFlowScalingIO struct {
	WALRecords       int64 `json:"wal_records"`
	WALFPI           int64 `json:"wal_fpi"`
	WALBytes         int64 `json:"wal_bytes"`
	WALBuffersFull   int64 `json:"wal_buffers_full"`
	RelationWrites   int64 `json:"relation_writes"`
	RelationExtends  int64 `json:"relation_extends"`
	RelationEvicts   int64 `json:"relation_evictions"`
	RelationFsyncs   int64 `json:"relation_fsyncs"`
	WALWrites        int64 `json:"wal_writes"`
	WALFsyncs        int64 `json:"wal_fsyncs"`
	CheckpointsTimed int64 `json:"checkpoints_timed"`
	CheckpointsReq   int64 `json:"checkpoints_requested"`
}

type networkFlowScalingCheck struct {
	Constraint string  `json:"constraint"`
	NSPerRow   float64 `json:"ns_per_row"`
}

type networkFlowScalingObservation struct {
	Case               networkFlowScalingCase    `json:"case"`
	SourceBytes        int                       `json:"source_bytes"`
	JobStatus          string                    `json:"job_status"`
	ErrorCode          string                    `json:"error_code,omitempty"`
	ReasonCode         string                    `json:"reason_code,omitempty"`
	AcceptedRows       int                       `json:"accepted_rows"`
	ApplyWallMS        float64                   `json:"apply_wall_ms"`
	ParticipantWriteMS float64                   `json:"participant_write_ms"`
	UnitElapsedMS      float64                   `json:"unit_elapsed_ms"`
	Passes             []networkFlowScalingPass  `json:"passes"`
	Copy               networkFlowScalingCopy    `json:"copy"`
	IO                 networkFlowScalingIO      `json:"pg_stat_delta"`
	Checks             []networkFlowScalingCheck `json:"check_costs,omitempty"`
	CheckBaselineMS    float64                   `json:"check_scan_baseline_ms,omitempty"`
	Plans              []networkFlowScalingPlan  `json:"query_plans,omitempty"`
	Environment        *networkFlowEnvironment   `json:"environment,omitempty"`
}

// networkFlowScalingPlan records which scans the planner chooses for the owner
// query shapes that filter rows by incident and table, so an index change can
// be compared before and after.
type networkFlowScalingPlan struct {
	Query   string   `json:"query"`
	Scans   []string `json:"scans"`
	Indexes []string `json:"indexes"`
}

type networkFlowScalingEvidence struct {
	ArtifactKind string                          `json:"artifact_kind"`
	ClaimPosture string                          `json:"claim_posture"`
	ObservedAt   string                          `json:"observed_at"`
	GoMaxProcs   int                             `json:"gomaxprocs"`
	NumCPU       int                             `json:"num_cpu"`
	Postgres     map[string]string               `json:"postgres_settings"`
	Cases        []networkFlowScalingObservation `json:"cases"`
}

// networkFlowScalingProbe observes one runtime. Each runtime applies exactly one
// Network Flow unit, so every Network Flow row copy belongs to that unit.
type networkFlowScalingProbe struct {
	mu            sync.Mutex
	passes        []networkFlowScalingPass
	copy          networkFlowScalingCopy
	firstPassOpen time.Time
	writeStart    time.Time
	writeEnd      time.Time
	unitEnd       time.Time
}

func (p *networkFlowScalingProbe) WrapDatabase(db dbport.DB) dbport.DB {
	return &networkFlowScalingDB{DB: db, probe: p}
}

func (p *networkFlowScalingProbe) ConfigureNetworkFlow(deps networkflow.ModuleDependencies) networkflow.ModuleDependencies {
	deps.ImportSources = networkFlowScalingSources{ImportSourcePort: deps.ImportSources, probe: p}
	return deps
}

func (p *networkFlowScalingProbe) WrapImportFacade(f imports.ExtensionImportFacade) imports.ExtensionImportFacade {
	return networkFlowScalingFacade{ExtensionImportFacade: f, probe: p}
}

type networkFlowScalingSources struct {
	networkflow.ImportSourcePort
	probe *networkFlowScalingProbe
}

func (s networkFlowScalingSources) OpenSourceStreamTx(ctx context.Context, tx pgx.Tx, ref string) (imports.ImportSourceStream, error) {
	stream, err := s.ImportSourcePort.OpenSourceStreamTx(ctx, tx, ref)
	if err != nil {
		return stream, err
	}
	now := time.Now()
	s.probe.mu.Lock()
	if s.probe.firstPassOpen.IsZero() {
		s.probe.firstPassOpen = now
	}
	s.probe.mu.Unlock()
	stream.Reader = &networkFlowScalingReader{ReadCloser: stream.Reader, probe: s.probe, started: now}
	return stream, nil
}

type networkFlowScalingReader struct {
	io.ReadCloser
	probe    *networkFlowScalingProbe
	started  time.Time
	readTime time.Duration
	bytes    int64
}

func (r *networkFlowScalingReader) Read(b []byte) (int, error) {
	start := time.Now()
	n, err := r.ReadCloser.Read(b)
	r.readTime += time.Since(start)
	r.bytes += int64(n)
	return n, err
}

func (r *networkFlowScalingReader) Close() error {
	err := r.ReadCloser.Close()
	r.probe.mu.Lock()
	r.probe.passes = append(r.probe.passes, networkFlowScalingPass{Bytes: r.bytes, ReadMS: milliseconds(r.readTime), ElapsedMS: milliseconds(time.Since(r.started))})
	r.probe.mu.Unlock()
	return err
}

type networkFlowScalingFacade struct {
	imports.ExtensionImportFacade
	probe *networkFlowScalingProbe
}

func (f networkFlowScalingFacade) ImportTransaction(request imports.ExtensionImportApplyRequest) (crossownertransaction.Participant, error) {
	participant, err := f.ExtensionImportFacade.ImportTransaction(request)
	if err != nil {
		return nil, err
	}
	return &networkFlowScalingParticipant{Participant: participant, probe: f.probe}, nil
}

type networkFlowScalingParticipant struct {
	crossownertransaction.Participant
	probe *networkFlowScalingProbe
}

func (p *networkFlowScalingParticipant) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	start := time.Now()
	result, err := p.Participant.Write(ctx, invocation)
	end := time.Now()
	p.probe.mu.Lock()
	p.probe.writeStart, p.probe.writeEnd = start, end
	p.probe.mu.Unlock()
	return result, err
}

type networkFlowScalingDB struct {
	dbport.DB
	probe *networkFlowScalingProbe
}

func (db *networkFlowScalingDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &networkFlowScalingTx{Tx: tx, probe: db.probe}, nil
}

type networkFlowScalingTx struct {
	pgx.Tx
	probe  *networkFlowScalingProbe
	parent *networkFlowScalingTx
	copied bool
}

func (tx *networkFlowScalingTx) root() *networkFlowScalingTx {
	for tx.parent != nil {
		tx = tx.parent
	}
	return tx
}

// Begin keeps savepoint-scoped work observable; only the outer transaction's
// end is the unit's commit or rollback.
func (tx *networkFlowScalingTx) Begin(ctx context.Context) (pgx.Tx, error) {
	nested, err := tx.Tx.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &networkFlowScalingTx{Tx: nested, probe: tx.probe, parent: tx}, nil
}

func (tx *networkFlowScalingTx) CopyFrom(ctx context.Context, table pgx.Identifier, columns []string, source pgx.CopyFromSource) (int64, error) {
	if len(table) == 0 || table[len(table)-1] != "network_flow_rows" {
		return tx.Tx.CopyFrom(ctx, table, columns, source)
	}
	tx.root().copied = true
	timed := &networkFlowScalingSource{CopyFromSource: source}
	start := time.Now()
	written, err := tx.Tx.CopyFrom(ctx, table, columns, timed)
	elapsed := time.Since(start)
	tx.probe.mu.Lock()
	tx.probe.copy.Calls++
	tx.probe.copy.Rows += written
	tx.probe.copy.WallMS += milliseconds(elapsed)
	tx.probe.copy.SourceMS += milliseconds(timed.elapsed)
	if ms := milliseconds(elapsed); ms > tx.probe.copy.MaxCallMS {
		tx.probe.copy.MaxCallMS = ms
	}
	tx.probe.mu.Unlock()
	return written, err
}

func (tx *networkFlowScalingTx) Commit(ctx context.Context) error {
	if tx.parent != nil || !tx.copied {
		return tx.Tx.Commit(ctx)
	}
	start := time.Now()
	err := tx.Tx.Commit(ctx)
	end := time.Now()
	tx.probe.mu.Lock()
	tx.probe.copy.CommitMS = milliseconds(end.Sub(start))
	tx.probe.copy.Committed = err == nil
	tx.probe.unitEnd = end
	tx.probe.mu.Unlock()
	return err
}

func (tx *networkFlowScalingTx) Rollback(ctx context.Context) error {
	if tx.parent != nil || !tx.copied {
		return tx.Tx.Rollback(ctx)
	}
	start := time.Now()
	err := tx.Tx.Rollback(ctx)
	end := time.Now()
	tx.probe.mu.Lock()
	if tx.probe.unitEnd.IsZero() {
		tx.probe.copy.RolledBackMS = milliseconds(end.Sub(start))
		tx.probe.unitEnd = end
	}
	tx.probe.mu.Unlock()
	return err
}

// networkFlowScalingSource times the owner's row production inside COPY. pgx
// calls it from its encoder goroutine, so it measures the Go producer only.
type networkFlowScalingSource struct {
	pgx.CopyFromSource
	elapsed time.Duration
}

func (s *networkFlowScalingSource) Next() bool {
	start := time.Now()
	next := s.CopyFromSource.Next()
	s.elapsed += time.Since(start)
	return next
}

func (s *networkFlowScalingSource) Values() ([]any, error) {
	start := time.Now()
	values, err := s.CopyFromSource.Values()
	s.elapsed += time.Since(start)
	return values, err
}

func milliseconds(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

func TestNetworkFlowImportScalingSweep(t *testing.T) {
	cases := []networkFlowScalingCase{
		// The default-budget case runs first so concurrent-load selections
		// overlap it with the rest of their work.
		{Name: "default_budget_250000", Rows: 250000, BudgetSeconds: 30, DefaultBudget: true},
		{Name: "rows_25000", Rows: 25000, BudgetSeconds: 300},
		{Name: "rows_50000", Rows: 50000, BudgetSeconds: 300},
		{Name: "rows_100001", Rows: 100001, BudgetSeconds: 300},
		{Name: "rows_250000", Rows: 250000, BudgetSeconds: 300},
	}
	evidence := networkFlowScalingEvidence{
		ArtifactKind: "cartulary.network_flow_import_scaling_evidence.v1",
		ClaimPosture: "environment_specific",
		ObservedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		GoMaxProcs:   runtime.GOMAXPROCS(0),
		NumCPU:       runtime.NumCPU(),
		Postgres:     map[string]string{},
	}
	var evidenceMu sync.Mutex
	t.Cleanup(func() {
		root, err := fixtureartifacts.SelectedUnitRoot()
		if err != nil {
			t.Logf("scaling evidence not retained: %v", err)
			return
		}
		evidenceMu.Lock()
		defer evidenceMu.Unlock()
		raw, err := json.MarshalIndent(evidence, "", "  ")
		if err != nil {
			t.Error(err)
			return
		}
		if err := os.WriteFile(filepath.Join(root, networkFlowScalingArtifact), append(raw, '\n'), 0o600); err != nil {
			t.Error(err)
		}
	})
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			observation := runNetworkFlowScalingCase(t, tc, &evidence, &evidenceMu)
			evidenceMu.Lock()
			evidence.Cases = append(evidence.Cases, observation)
			evidenceMu.Unlock()
			t.Logf("network flow scaling %s: status=%s code=%s rows=%d participant_write=%.0fms unit=%.0fms passes=%v copy=%+v",
				tc.Name, observation.JobStatus, observation.ErrorCode, observation.AcceptedRows,
				observation.ParticipantWriteMS, observation.UnitElapsedMS, observation.Passes, observation.Copy)
			if !tc.DefaultBudget && observation.JobStatus != "succeeded" {
				t.Fatalf("raised-budget import %s ended %s (%s/%s)", tc.Name, observation.JobStatus, observation.ErrorCode, observation.ReasonCode)
			}
		})
	}
}

func runNetworkFlowScalingCase(t *testing.T, tc networkFlowScalingCase, evidence *networkFlowScalingEvidence, evidenceMu *sync.Mutex) networkFlowScalingObservation {
	t.Helper()
	env := map[string]string{
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}
	if !tc.DefaultBudget {
		env["CARTULARY__TIMEOUTS__EXTENSIONS__TRANSACTION_PARTICIPANT_SECONDS"] = fmt.Sprint(tc.BudgetSeconds)
	}
	probe := &networkFlowScalingProbe{}
	h := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{
		Prefix: "network-flow-scaling", Env: env, TestRouteMode: httptestx.TestRouteModeDisabled,
		ConfigureRuntime: func(options *server.Options) { options.NetworkFlowComposition = probe },
	})
	recordNetworkFlowScalingSettings(t, h.DB, evidence, evidenceMu)
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{
		"client_txn_id": "scaling-incident-" + tc.Name, "incident_key": "SCALING-" + strings.ToUpper(strings.ReplaceAll(tc.Name, "_", "-")), "title": "Network Flow scaling",
	})["incident_id"].(string)
	content := sourceAdmissionHeader + strings.Repeat(sourceAdmissionRow, tc.Rows)
	session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "scaling-"+tc.Name, content, tc.Name+".csv")
	base := "/api/v1/import-sessions/" + session + "/units/" + unit
	inspection := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"})
	columns := httptestx.RequireSuccessEnvelope(t, inspection, http.StatusOK)["data"].(map[string]any)["owner_result"].(map[string]any)["source_columns"].([]any)
	preview := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/mapping-preview", networkFlowMappingPreviewPayload())
	httptestx.RequireSuccessEnvelope(t, preview, http.StatusOK)
	mapping := networkFlowMappingPayload("map-" + unit)
	echoes := mapping["source_columns"].([]map[string]any)
	for i, raw := range columns {
		column := raw.(map[string]any)
		echoes[i]["source_header_text"] = column["raw_header_text"]
		echoes[i]["source_column_ordinal"] = column["source_column_ordinal"]
	}
	httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPut, base+"/mapping", mapping), http.StatusOK)
	httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/select", map[string]any{"client_txn_id": "select-" + unit}), http.StatusOK)

	before := readNetworkFlowScalingIO(t, h.DB)
	applyStart := time.Now()
	resp := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/apply", map[string]any{"client_txn_id": "apply-" + unit})
	jobID := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
	job := httptestx.PollJob(t, h.Server.HTTP.URL, jobID, httptestx.PollOptions{
		Timeout: 10 * time.Minute, Interval: 50 * time.Millisecond, Request: []func(*http.Request){httptestx.WithCookies(login.Session)}, LockWaits: h.DB,
	}, "succeeded", "failed", "canceled")
	applyWall := time.Since(applyStart)
	after := readNetworkFlowScalingIO(t, h.DB)

	observation := networkFlowScalingObservation{Case: tc, SourceBytes: len(content), ApplyWallMS: milliseconds(applyWall)}
	observation.JobStatus, _ = job["status"].(string)
	var code, reason sql.NullString
	if err := h.DB.QueryRowContext(t.Context(), `SELECT error_code,reason_code FROM import_unit_apply_outcomes WHERE import_unit_id::text=$1`, unit).Scan(&code, &reason); err == nil {
		observation.ErrorCode, observation.ReasonCode = code.String, reason.String
	}
	observation.AcceptedRows = dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM network_flow_rows r JOIN network_flow_tables t USING(network_flow_table_id) JOIN import_committed_sources s ON s.committed_source_id=t.committed_source_id WHERE s.origin_import_session_id::text=$1`, session)
	observation.IO = after.minus(before)
	probe.mu.Lock()
	observation.Passes = append([]networkFlowScalingPass(nil), probe.passes...)
	observation.Copy = probe.copy
	if !probe.writeStart.IsZero() {
		observation.ParticipantWriteMS = milliseconds(probe.writeEnd.Sub(probe.writeStart))
	}
	if !probe.firstPassOpen.IsZero() && !probe.unitEnd.IsZero() {
		observation.UnitElapsedMS = milliseconds(probe.unitEnd.Sub(probe.firstPassOpen))
	}
	probe.mu.Unlock()
	if observation.JobStatus == "succeeded" {
		if observation.AcceptedRows != tc.Rows {
			t.Fatalf("accepted rows=%d want=%d", observation.AcceptedRows, tc.Rows)
		}
		observation.CheckBaselineMS, observation.Checks = costNetworkFlowRowChecks(t, h.DB, session, tc.Rows)
		observation.Plans = planNetworkFlowRowQueries(t, h.DB, session)
		if tc.Rows == 100001 {
			observation.Environment = inspectNetworkFlowEnvironment(t, h.DB, session)
		}
	}
	return observation
}

func planNetworkFlowRowQueries(t *testing.T, db *sql.DB, session string) []networkFlowScalingPlan {
	t.Helper()
	ctx := t.Context()
	var incidentID, tableID, rowID string
	if err := db.QueryRowContext(ctx, `SELECT t.incident_id::text, t.network_flow_table_id, r.network_flow_row_id FROM network_flow_tables t JOIN import_committed_sources s ON s.committed_source_id=t.committed_source_id JOIN network_flow_rows r ON r.network_flow_table_id=t.network_flow_table_id WHERE s.origin_import_session_id::text=$1 ORDER BY r.source_row_number LIMIT 1`, session).Scan(&incidentID, &tableID, &rowID); err != nil {
		t.Fatalf("resolve planned table: %v", err)
	}
	if _, err := db.ExecContext(ctx, `ANALYZE public.network_flow_rows`); err != nil {
		t.Fatalf("analyze rows: %v", err)
	}
	shapes := []struct{ name, query string }{
		{"rows_query_default_order", `SELECT network_flow_row_id FROM network_flow_rows WHERE incident_id=$1::uuid AND network_flow_table_id = ANY(ARRAY[$2]::text[]) ORDER BY flow_start_utc ASC, flow_end_utc ASC, source_row_number ASC, network_flow_table_id COLLATE "C" ASC, network_flow_row_id COLLATE "C" ASC LIMIT 501`},
		{"graph_table_scan", `SELECT network_flow_row_id FROM unnest(ARRAY[$2]::text[]) WITH ORDINALITY AS selected(selected_table_id, table_rank) JOIN network_flow_rows ON incident_id = $1::uuid AND network_flow_table_id = selected.selected_table_id ORDER BY selected.table_rank ASC, flow_start_utc ASC, flow_end_utc ASC, source_row_number ASC, network_flow_row_id ASC`},
		{"graph_vertex_contributors", `SELECT network_flow_row_id FROM unnest(ARRAY[$2]::text[]) WITH ORDINALITY AS selected(selected_table_id,table_rank) JOIN network_flow_rows ON incident_id=$1::uuid AND network_flow_table_id=selected.selected_table_id WHERE (endpoint_a_ip = '192.0.2.10' OR endpoint_b_ip = '192.0.2.10')`},
		{"accepted_row", `SELECT network_flow_row_id FROM network_flow_rows WHERE incident_id = $1::uuid AND network_flow_table_id = $2 AND network_flow_row_id = $3`},
		{"table_scope_count", `SELECT count(*) FROM network_flow_rows WHERE incident_id = $1::uuid AND network_flow_table_id = $2`},
		{"table_cascade_delete", `DELETE FROM ONLY network_flow_rows WHERE incident_id = $1::uuid AND network_flow_table_id = $2`},
		{"incident_presence", `SELECT 1 FROM network_flow_rows WHERE incident_id = $1::uuid LIMIT 1`},
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve planning connection: %v", err)
	}
	defer conn.Close()
	plans := make([]networkFlowScalingPlan, 0, 2*len(shapes))
	// With one imported table every incident/table filter selects all rows, so
	// the planner scans sequentially. The forced variant shows which index would
	// serve the shape when the filter is selective.
	for _, variant := range []struct{ suffix, setting string }{{"", "on"}, {"/no_seqscan", "off"}} {
		if _, err := conn.ExecContext(ctx, `SELECT set_config('enable_seqscan', $1, false)`, variant.setting); err != nil {
			t.Fatalf("set planner variant: %v", err)
		}
		for _, shape := range shapes {
			args := []any{incidentID}
			if strings.Contains(shape.query, "$2") {
				args = append(args, tableID)
			}
			if strings.Contains(shape.query, "$3") {
				args = append(args, rowID)
			}
			var raw []byte
			if err := conn.QueryRowContext(ctx, `EXPLAIN (FORMAT JSON) `+shape.query, args...).Scan(&raw); err != nil {
				t.Fatalf("plan %s: %v", shape.name, err)
			}
			var decoded any
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("decode plan %s: %v", shape.name, err)
			}
			plan := networkFlowScalingPlan{Query: shape.name + variant.suffix}
			var walk func(any)
			walk = func(node any) {
				switch value := node.(type) {
				case []any:
					for _, item := range value {
						walk(item)
					}
				case map[string]any:
					if kind, ok := value["Node Type"].(string); ok && strings.Contains(kind, "Scan") && kind != "Function Scan" {
						plan.Scans = append(plan.Scans, kind)
					}
					if index, ok := value["Index Name"].(string); ok {
						plan.Indexes = append(plan.Indexes, index)
					}
					for _, child := range value {
						walk(child)
					}
				}
			}
			walk(decoded)
			sort.Strings(plan.Scans)
			sort.Strings(plan.Indexes)
			plans = append(plans, plan)
		}
	}
	return plans
}

func recordNetworkFlowScalingSettings(t *testing.T, db *sql.DB, evidence *networkFlowScalingEvidence, mu *sync.Mutex) {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()
	if len(evidence.Postgres) > 0 {
		return
	}
	for _, name := range []string{"server_version", "shared_buffers", "max_wal_size", "wal_buffers", "fsync", "synchronous_commit", "full_page_writes", "data_checksums", "wal_init_zero", "max_parallel_workers_per_gather"} {
		var value string
		if err := db.QueryRowContext(t.Context(), "SELECT current_setting($1)", name).Scan(&value); err != nil {
			t.Fatalf("read PostgreSQL setting %s: %v", name, err)
		}
		evidence.Postgres[name] = value
	}
}

func readNetworkFlowScalingIO(t *testing.T, db *sql.DB) networkFlowScalingIO {
	t.Helper()
	var stat networkFlowScalingIO
	ctx := t.Context()
	if err := db.QueryRowContext(ctx, `SELECT wal_records, wal_fpi, wal_bytes::bigint, wal_buffers_full FROM pg_stat_wal`).Scan(&stat.WALRecords, &stat.WALFPI, &stat.WALBytes, &stat.WALBuffersFull); err != nil {
		t.Fatalf("read pg_stat_wal: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT coalesce(sum(writes),0)::bigint, coalesce(sum(extends),0)::bigint, coalesce(sum(evictions),0)::bigint, coalesce(sum(fsyncs),0)::bigint FROM pg_stat_io WHERE object='relation'`).Scan(&stat.RelationWrites, &stat.RelationExtends, &stat.RelationEvicts, &stat.RelationFsyncs); err != nil {
		t.Fatalf("read pg_stat_io relation: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT coalesce(sum(writes),0)::bigint, coalesce(sum(fsyncs),0)::bigint FROM pg_stat_io WHERE object='wal'`).Scan(&stat.WALWrites, &stat.WALFsyncs); err != nil {
		t.Fatalf("read pg_stat_io wal: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT num_timed, num_requested FROM pg_stat_checkpointer`).Scan(&stat.CheckpointsTimed, &stat.CheckpointsReq); err != nil {
		t.Fatalf("read pg_stat_checkpointer: %v", err)
	}
	return stat
}

func (a networkFlowScalingIO) minus(b networkFlowScalingIO) networkFlowScalingIO {
	return networkFlowScalingIO{
		WALRecords: a.WALRecords - b.WALRecords, WALFPI: a.WALFPI - b.WALFPI, WALBytes: a.WALBytes - b.WALBytes,
		WALBuffersFull: a.WALBuffersFull - b.WALBuffersFull, RelationWrites: a.RelationWrites - b.RelationWrites,
		RelationExtends: a.RelationExtends - b.RelationExtends, RelationEvicts: a.RelationEvicts - b.RelationEvicts,
		RelationFsyncs: a.RelationFsyncs - b.RelationFsyncs, WALWrites: a.WALWrites - b.WALWrites, WALFsyncs: a.WALFsyncs - b.WALFsyncs,
		CheckpointsTimed: a.CheckpointsTimed - b.CheckpointsTimed, CheckpointsReq: a.CheckpointsReq - b.CheckpointsReq,
	}
}

// costNetworkFlowRowChecks re-evaluates every current CHECK expression of
// network_flow_rows over the imported rows with read-only queries. The excess
// over a plain scan approximates each constraint's per-row evaluation cost.
func costNetworkFlowRowChecks(t *testing.T, db *sql.DB, session string, rows int) (float64, []networkFlowScalingCheck) {
	t.Helper()
	ctx := t.Context()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve check-costing connection: %v", err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `SET max_parallel_workers_per_gather = 0`); err != nil {
		t.Fatalf("disable parallel check costing: %v", err)
	}
	var tableID string
	if err := conn.QueryRowContext(ctx, `SELECT t.network_flow_table_id FROM network_flow_tables t JOIN import_committed_sources s ON s.committed_source_id=t.committed_source_id WHERE s.origin_import_session_id::text=$1`, session).Scan(&tableID); err != nil {
		t.Fatalf("resolve imported table: %v", err)
	}
	type constraint struct{ name, expression string }
	var constraints []constraint
	result, err := conn.QueryContext(ctx, `SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid='public.network_flow_rows'::regclass AND contype='c' ORDER BY conname`)
	if err != nil {
		t.Fatalf("list row checks: %v", err)
	}
	for result.Next() {
		var c constraint
		if err := result.Scan(&c.name, &c.expression); err != nil {
			t.Fatal(err)
		}
		c.expression = strings.TrimSuffix(strings.TrimPrefix(c.expression, "CHECK "), " NOT VALID")
		constraints = append(constraints, c)
	}
	if err := result.Err(); err != nil {
		t.Fatal(err)
	}
	_ = result.Close()
	timeScan := func(predicate string) float64 {
		best := time.Duration(0)
		for attempt := 0; attempt < 2; attempt++ {
			var matched int
			start := time.Now()
			if err := conn.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE `+predicate+`) FROM public.network_flow_rows WHERE network_flow_table_id = $1`, tableID).Scan(&matched); err != nil {
				t.Fatalf("cost check %q: %v", predicate, err)
			}
			elapsed := time.Since(start)
			if matched != rows {
				t.Fatalf("check predicate %q accepted %d of %d committed rows", predicate, matched, rows)
			}
			if attempt == 0 || elapsed < best {
				best = elapsed
			}
		}
		return float64(best.Nanoseconds())
	}
	baseline := timeScan("true")
	checks := make([]networkFlowScalingCheck, 0, len(constraints))
	for _, c := range constraints {
		cost := (timeScan("("+c.expression+") IS NOT FALSE") - baseline) / float64(rows)
		checks = append(checks, networkFlowScalingCheck{Constraint: c.name, NSPerRow: cost})
	}
	return baseline / 1e6, checks
}

// networkFlowEnvironment records, read-only, what can change the cost of the
// row CHECK functions: server and session settings, database locale, role
// settings, and how the stored evidence is compressed.
type networkFlowEnvironment struct {
	Settings          map[string]string  `json:"settings"`
	Database          map[string]string  `json:"database"`
	RoleSettings      []string           `json:"role_settings"`
	EvidenceStorage   []map[string]any   `json:"evidence_storage"`
	EvidenceCheckCost map[string]float64 `json:"evidence_check_ns_per_row"`
}

func inspectNetworkFlowEnvironment(t *testing.T, db *sql.DB, session string) *networkFlowEnvironment {
	t.Helper()
	ctx := t.Context()
	env := &networkFlowEnvironment{Settings: map[string]string{}, Database: map[string]string{}, EvidenceCheckCost: map[string]float64{}}
	rows, err := db.QueryContext(ctx, `SELECT name, setting FROM pg_settings WHERE name = ANY($1) OR name LIKE 'plpgsql.%' OR name LIKE 'auto_explain.%' OR name LIKE 'pg_stat_statements.%' ORDER BY name`,
		"{shared_preload_libraries,session_preload_libraries,local_preload_libraries,track_functions,track_io_timing,track_activities,jit,jit_provider,jit_above_cost,default_toast_compression,search_path,server_version_num,debug_assertions,log_statement,log_min_duration_statement,max_stack_depth,work_mem,role,session_authorization}")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var name, setting string
		if err := rows.Scan(&name, &setting); err != nil {
			t.Fatal(err)
		}
		env.Settings[name] = setting
	}
	_ = rows.Close()
	var encoding, collate, ctype, provider, user string
	if err := db.QueryRowContext(ctx, `SELECT pg_encoding_to_char(encoding), datcollate, datctype, datlocprovider::text, current_user FROM pg_database WHERE datname = current_database()`).Scan(&encoding, &collate, &ctype, &provider, &user); err != nil {
		t.Fatal(err)
	}
	env.Database = map[string]string{"encoding": encoding, "collate": collate, "ctype": ctype, "locale_provider": provider, "current_user": user}
	rows, err = db.QueryContext(ctx, `SELECT coalesce(setrole::regrole::text, '*') || '@' || coalesce((SELECT datname FROM pg_database WHERE oid = setdatabase), '*') || ': ' || array_to_string(setconfig, ', ') FROM pg_db_role_setting ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var setting string
		if err := rows.Scan(&setting); err != nil {
			t.Fatal(err)
		}
		env.RoleSettings = append(env.RoleSettings, setting)
	}
	_ = rows.Close()
	var tableID string
	if err := db.QueryRowContext(ctx, `SELECT t.network_flow_table_id FROM network_flow_tables t JOIN import_committed_sources s ON s.committed_source_id=t.committed_source_id WHERE s.origin_import_session_id::text=$1`, session).Scan(&tableID); err != nil {
		t.Fatal(err)
	}
	rows, err = db.QueryContext(ctx, `SELECT coalesce(pg_column_compression(direction_evidence), 'none'), count(*), avg(pg_column_size(direction_evidence))::bigint, avg(octet_length(direction_evidence::text))::bigint, avg(pg_column_size(r.*))::bigint FROM network_flow_rows r WHERE network_flow_table_id = $1 GROUP BY 1 ORDER BY 1`, tableID)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var method string
		var count, stored, text, tuple int64
		if err := rows.Scan(&method, &count, &stored, &text, &tuple); err != nil {
			t.Fatal(err)
		}
		env.EvidenceStorage = append(env.EvidenceStorage, map[string]any{"compression": method, "rows": count, "stored_bytes": stored, "text_bytes": text, "row_bytes": tuple})
	}
	_ = rows.Close()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `SET max_parallel_workers_per_gather = 0`); err != nil {
		t.Fatal(err)
	}
	var count int
	timeScan := func(predicate string) float64 {
		best := time.Duration(0)
		for attempt := 0; attempt < 2; attempt++ {
			start := time.Now()
			if err := conn.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE `+predicate+`) FROM network_flow_rows WHERE network_flow_table_id = $1`, tableID).Scan(&count); err != nil {
				t.Fatalf("time %s: %v", predicate, err)
			}
			if elapsed := time.Since(start); attempt == 0 || elapsed < best {
				best = elapsed
			}
		}
		return float64(best.Nanoseconds()) / float64(count)
	}
	call := func(evidence string) string {
		return `public.network_flow_direction_evidence_is_valid_v2(` + evidence + `, direction_kind, from_endpoint, direction_reason, endpoint_a_ip = endpoint_b_ip AND endpoint_a_port IS NOT DISTINCT FROM endpoint_b_port) IS TRUE`
	}
	base := timeScan("true")
	detoast := timeScan("(direction_evidence #> '{}') IS NOT NULL")
	env.EvidenceCheckCost["stored_value"] = timeScan(call("direction_evidence")) - base
	env.EvidenceCheckCost["detoasted_once"] = timeScan(call("(direction_evidence #> '{}')")) - base
	env.EvidenceCheckCost["detoast_only"] = detoast - base
	return env
}
