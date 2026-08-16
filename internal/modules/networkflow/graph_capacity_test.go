package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const graphCapacityEvidenceArtifact = "network-flow-capacity-evidence.json"

type graphCapacityWorkload struct {
	Name              string
	Profile           string
	Mode              string
	Rows              int
	Vertices          int
	Edges             int
	Buckets           int
	BucketWidth       int64
	Limits            EffectiveLimits
	SelectionEligible bool
}

type graphCapacityObservation struct {
	Workload                  string           `json:"workload"`
	Profile                   string           `json:"profile"`
	GraphMode                 string           `json:"graph_mode"`
	EffectiveLimits           map[string]int64 `json:"effective_limits"`
	SourceRowsVisited         int              `json:"source_rows_visited"`
	ResultVertices            int              `json:"result_vertices"`
	ResultEdges               int              `json:"result_edges"`
	ResultBuckets             int              `json:"result_buckets"`
	ProjectionInputBytes      int              `json:"projection_input_bytes"`
	ProjectionResultBytes     int              `json:"projection_result_bytes"`
	AllocationBytes           uint64           `json:"allocation_bytes"`
	ProcessPeakResidentBytes  uint64           `json:"process_peak_resident_bytes"`
	WallTimeMilliseconds      int64            `json:"wall_time_milliseconds"`
	DatabaseStatements        int              `json:"database_statements"`
	DatabaseRowsRead          int              `json:"database_rows_read"`
	Outcome                   string           `json:"outcome"`
	DeploymentSelectionStatus string           `json:"deployment_selection_status"`
}

type graphCapacityRunManifest struct {
	RunID           string `json:"run_id"`
	SourceCommit    string `json:"source_commit"`
	SourceState     string `json:"source_state"`
	SourceDigest    string `json:"source_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	SystemDigest    string `json:"system_digest"`
	GraphDigest     string `json:"graph_digest"`
}

type graphCapacityEvidence struct {
	ArtifactKind              string                     `json:"artifact_kind"`
	ClaimPosture              string                     `json:"claim_posture"`
	RunManifest               graphCapacityRunManifest   `json:"run_manifest"`
	ObservedAt                string                     `json:"observed_at"`
	Runtime                   map[string]any             `json:"runtime"`
	SelectedDeploymentProfile string                     `json:"selected_deployment_profile"`
	SelectedEffectiveLimits   map[string]int64           `json:"selected_effective_limits"`
	SelectionRationale        string                     `json:"selection_rationale"`
	Workloads                 []graphCapacityObservation `json:"workloads"`
	PairedEvidence            []string                   `json:"paired_evidence"`
	ResidualRisks             []string                   `json:"residual_risks"`
}

func TestNetworkFlowGraphCapacityCertification_Integration(t *testing.T) {
	defaultLimits := DefaultEffectiveLimits()
	raisedLimits := defaultLimits
	raisedLimits.MaxGraphVertices = 10_000
	raisedLimits.MaxGraphEdges = 20_000
	raisedLimits.MaxContributingRowsPerGraph = 500_000
	raisedLimits.MaxTimeBucketsPerGraph = 512
	raisedLimits.MaxExampleRowRefsPerEdge = 20
	maximumLimits := maximumEffectiveLimitsForCapacityTest()
	for name, limits := range map[string]EffectiveLimits{
		"default": defaultLimits, "raised": raisedLimits, "semantic_maximum": maximumLimits,
	} {
		if err := ValidateEffectiveLimits(limits); err != nil {
			t.Fatalf("%s capacity profile is not an admitted EffectiveLimits value: %v", name, err)
		}
	}

	workloads := []graphCapacityWorkload{
		{
			Name: "default_skew", Profile: "default", Mode: "default_flow_edge_v1",
			Rows: 250_000, Vertices: 2, Edges: 1, Limits: defaultLimits, SelectionEligible: true,
		},
		{
			Name: "default_dense", Profile: "default", Mode: "default_flow_edge_v1",
			Rows: 250_000, Vertices: 5_000, Edges: 10_000, Limits: defaultLimits, SelectionEligible: true,
		},
		{
			Name: "raised_dense", Profile: "raised", Mode: "default_flow_edge_v1",
			Rows: 500_000, Vertices: 10_000, Edges: 20_000, Limits: raisedLimits,
		},
		{
			Name: "temporal_default", Profile: "default", Mode: "time_bucket_v1",
			Rows: 250_000, Vertices: 2, Edges: 256, Buckets: 256, BucketWidth: 60,
			Limits: defaultLimits, SelectionEligible: true,
		},
		{
			Name: "semantic_maximum_rows", Profile: "semantic_maximum_diagnostic", Mode: "default_flow_edge_v1",
			Rows: 5_000_000, Vertices: 2, Edges: 1, Limits: maximumLimits,
		},
		{
			Name: "semantic_maximum_buckets", Profile: "semantic_maximum_diagnostic", Mode: "time_bucket_v1",
			Rows: 1_024, Vertices: 2, Edges: 1_024, Buckets: 1_024, BucketWidth: 60,
			Limits: maximumLimits,
		},
	}

	observations := make([]graphCapacityObservation, 0, len(workloads))
	for _, workload := range workloads {
		workload := workload
		t.Run(workload.Name, func(t *testing.T) {
			observations = append(observations, runGraphCapacityWorkload(t, workload))
		})
	}

	evidence := graphCapacityEvidence{
		ArtifactKind:              "cartulary.network_flow_capacity_evidence.v1",
		ClaimPosture:              "informative_environment_specific",
		ObservedAt:                time.Now().UTC().Format(time.RFC3339Nano),
		Runtime:                   graphCapacityRuntimeEvidence(),
		SelectedDeploymentProfile: "conservative_defaults",
		SelectedEffectiveLimits:   effectiveLimitsEvidence(defaultLimits),
		SelectionRationale:        "default end-to-end semantics are proven; raised and semantic-maximum executions are diagnostic and do not establish a portable deployment capacity claim",
		Workloads:                 observations,
		PairedEvidence: []string{
			"module.networkflow.integration.graph_result_cleanup",
			"module.networkflow.integration.graph_result_cleanup_races",
			"module.networkflow.integration.saved_graph_lifecycle_v2",
			"module.networkflow.integration.time_bucket_saved_graph_lifecycle",
			"module.networkflow.unit.cleanup_dispatcher_lifecycle",
			"platform.jobs.integration.worker_capacity_and_mixed_kind_fairness",
			"app.server.unit.network_flow_cleanup_lifecycle",
			"module.recovery.unit.graph_restore_v3_contract_projection",
		},
		ResidualRisks: []string{
			"capacity measurements are host-specific and are not a hardware-independent SLO",
			"raised and semantic-maximum profiles require deployment-local database, browser, restore, and shutdown qualification before selection",
		},
	}
	artifactDir := strings.TrimSpace(os.Getenv("CARTULARY_STEP_ARTIFACT_DIR"))
	if artifactDir == "" {
		t.Log("capacity evidence retained only by the Make-owned harness; CARTULARY_STEP_ARTIFACT_DIR is unset")
		return
	}
	evidence.RunManifest = readGraphCapacityRunManifest(t, artifactDir)
	body, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatalf("encode Network Flow capacity evidence: %v", err)
	}
	if err := os.MkdirAll(artifactDir, 0o700); err != nil {
		t.Fatalf("create Network Flow capacity evidence directory: %v", err)
	}
	path := filepath.Join(artifactDir, graphCapacityEvidenceArtifact)
	if err := os.WriteFile(path, append(body, '\n'), 0o600); err != nil {
		t.Fatalf("write Network Flow capacity evidence: %v", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatalf("secure Network Flow capacity evidence: %v", err)
	}
}

func runGraphCapacityWorkload(t testing.TB, workload graphCapacityWorkload) graphCapacityObservation {
	t.Helper()
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	startedAt := time.Now()

	incidentID := uuid.MustParse("00000000-0000-4000-8000-00000000c311")
	table := TableRecord{
		IncidentID:         incidentID,
		TableID:            "nft_" + strings.Repeat("c", 64),
		MappingFingerprint: strings.Repeat("d", 64),
	}
	composition := graphComposition{
		SemanticSchemaID: schemaGraphSemanticQueryV2,
		Aggregation:      graphAggregation{Mode: workload.Mode, BucketWidthSeconds: workload.BucketWidth},
		Digest:           "capacity_" + workload.Name,
		ResultLimits:     effectiveGraphResultLimits(workload.Limits),
		SourceTables:     []TableRecord{table},
		SourceTableRefs:  graphSourceTableRefs([]TableRecord{table}),
		TableRanks:       map[string]int{table.TableID: 0},
		Vertices:         map[string]*graphVertex{},
		Edges:            map[string]*graphEdge{},
		SelectedTableIDs: []string{table.TableID},
	}
	if workload.Buckets > 0 {
		start := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
		end := start.Add(time.Duration(int64(workload.Buckets)*workload.BucketWidth) * time.Second)
		buckets, apiErr := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &end}, workload.BucketWidth, int(workload.Limits.MaxTimeBucketsPerGraph))
		if apiErr != nil {
			t.Fatalf("construct %s buckets: %#v", workload.Name, apiErr)
		}
		composition.TimeBuckets = buckets
	}

	patterns := graphCapacityPatterns(workload, table)
	for index := 0; index < workload.Rows; index++ {
		row := patterns[index%len(patterns)]
		row.SourceRowNumber = int64(index + 1)
		if workload.Buckets > 0 {
			bucket := composition.TimeBuckets[index%len(composition.TimeBuckets)]
			row.FlowStartUTC = bucket.StartUTC.Add(time.Second)
			row.FlowEndUTC = row.FlowStartUTC.Add(time.Second)
		}
		if apiErr := composeGraphRow(incidentID, row, map[string]TableRecord{table.TableID: table}, &composition); apiErr != nil {
			t.Fatalf("execute %s row %d: %#v", workload.Name, index+1, apiErr)
		}
	}
	if composition.ContributingRows != workload.Rows || len(composition.Vertices) != workload.Vertices || len(composition.Edges) != workload.Edges || len(composition.TimeBuckets) != workload.Buckets {
		t.Fatalf(
			"%s cardinality rows/vertices/edges/buckets = %d/%d/%d/%d; want %d/%d/%d/%d",
			workload.Name, composition.ContributingRows, len(composition.Vertices), len(composition.Edges), len(composition.TimeBuckets),
			workload.Rows, workload.Vertices, workload.Edges, workload.Buckets,
		)
	}
	if apiErr := validateGraphLimits(composition); apiErr != nil {
		t.Fatalf("validate %s result limits: %#v", workload.Name, apiErr)
	}

	sourceSnapshotID := "nfsnap_" + strings.Repeat("e", 64)
	projectionInput := canonicalJSON(networkFlowProjectionInput(sourceSnapshotID, composition))
	projector := newGraphProjectionAdapter()
	graphViewID, err := projector.GraphViewID("capacity:" + workload.Name)
	if err != nil {
		t.Fatalf("derive %s graph view ID: %v", workload.Name, err)
	}
	projection, err := projector.ProjectEphemeral(context.Background(), graphViewID, projectionInput)
	if err != nil {
		t.Fatalf("project %s workload: %v", workload.Name, err)
	}
	projectionResult := canonicalJSON(projection)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	selectionStatus := "diagnostic_not_selected"
	if workload.SelectionEligible {
		selectionStatus = "supports_selected_default_profile"
	}
	return graphCapacityObservation{
		Workload: workload.Name, Profile: workload.Profile, GraphMode: workload.Mode,
		EffectiveLimits:   effectiveLimitsEvidence(workload.Limits),
		SourceRowsVisited: composition.ContributingRows,
		ResultVertices:    len(composition.Vertices), ResultEdges: len(composition.Edges), ResultBuckets: len(composition.TimeBuckets),
		ProjectionInputBytes: len(projectionInput), ProjectionResultBytes: len(projectionResult),
		AllocationBytes:          after.TotalAlloc - before.TotalAlloc,
		ProcessPeakResidentBytes: graphCapacityPeakRSSBytes(),
		WallTimeMilliseconds:     time.Since(startedAt).Milliseconds(),
		DatabaseStatements:       0, DatabaseRowsRead: 0,
		Outcome: "passed", DeploymentSelectionStatus: selectionStatus,
	}
}

func graphCapacityPatterns(workload graphCapacityWorkload, table TableRecord) []FlowRow {
	patternCount := workload.Edges
	if patternCount < 1 {
		patternCount = 1
	}
	patterns := make([]FlowRow, 0, patternCount)
	port := int32(443)
	startedAt := time.Date(2026, 8, 16, 0, 0, 1, 0, time.UTC)
	for index := 0; index < patternCount; index++ {
		source := index % workload.Vertices
		destination := (source + 1 + index/workload.Vertices) % workload.Vertices
		patterns = append(patterns, FlowRow{
			NetworkFlowTableID: table.TableID,
			RowID:              "nfr_" + strings.Repeat("f", 64),
			FlowStartUTC:       startedAt, FlowEndUTC: startedAt.Add(time.Second),
			SrcIP: graphCapacityIPv4(source), DstIP: graphCapacityIPv4(destination),
			DstPort: &port, IPProtocol: 6, BytesCount: "1", PacketsCount: "1",
			MappingFingerprint: table.MappingFingerprint,
		})
	}
	return patterns
}

func graphCapacityIPv4(index int) string {
	return fmt.Sprintf("10.%d.%d.%d", (index>>16)&255, (index>>8)&255, index&255)
}

func maximumEffectiveLimitsForCapacityTest() EffectiveLimits {
	return EffectiveLimits{
		MaxActiveTablesPerIncident: 128, MaxRetainedTablesPerIncident: 512,
		MaxSelectedTablesPerQuery: 64, MaxColumnsPerCSV: 512,
		MaxHeaderScalarLength: 256, MaxRawCellScalarLength: 16_384,
		MaxRowsPerCSV: 5_000_000, MaxAcceptedRowsPerTable: 5_000_000,
		MaxRejectedRowDiagnostics: 100_000, MaxFiltersPerQuery: 16,
		MaxSortsPerQuery: 8, MaxQueryLimit: 1_000,
		MaxGraphVertices: 100_000, MaxGraphEdges: 250_000,
		MaxActiveGraphViewsPerIncident: 32, MaxRetainedGraphViewsPerIncident: 128,
		MaxNonterminalGraphJobsPerIncident: 4, MaxExampleRowRefsPerEdge: 100,
		MaxBindingSourceRowRefs: 1_000, MaxAggregateCounterDigits: 128,
		MaxContributingRowsPerGraph: 5_000_000, MaxTimeBucketsPerGraph: 1_024,
		GraphMaterializationTimeoutSeconds: 3_600,
	}
}

func effectiveLimitsEvidence(limits EffectiveLimits) map[string]int64 {
	return map[string]int64{
		"max_active_tables_per_incident":          limits.MaxActiveTablesPerIncident,
		"max_retained_tables_per_incident":        limits.MaxRetainedTablesPerIncident,
		"max_selected_tables_per_query":           limits.MaxSelectedTablesPerQuery,
		"max_columns_per_csv":                     limits.MaxColumnsPerCSV,
		"max_header_scalar_length":                limits.MaxHeaderScalarLength,
		"max_raw_cell_scalar_length":              limits.MaxRawCellScalarLength,
		"max_rows_per_csv":                        limits.MaxRowsPerCSV,
		"max_accepted_rows_per_table":             limits.MaxAcceptedRowsPerTable,
		"max_rejected_row_diagnostics":            limits.MaxRejectedRowDiagnostics,
		"max_filters_per_query":                   limits.MaxFiltersPerQuery,
		"max_sorts_per_query":                     limits.MaxSortsPerQuery,
		"max_query_limit":                         limits.MaxQueryLimit,
		"max_graph_vertices":                      limits.MaxGraphVertices,
		"max_graph_edges":                         limits.MaxGraphEdges,
		"max_active_graph_views_per_incident":     limits.MaxActiveGraphViewsPerIncident,
		"max_retained_graph_views_per_incident":   limits.MaxRetainedGraphViewsPerIncident,
		"max_nonterminal_graph_jobs_per_incident": limits.MaxNonterminalGraphJobsPerIncident,
		"max_example_row_refs_per_edge":           limits.MaxExampleRowRefsPerEdge,
		"max_binding_source_row_refs":             limits.MaxBindingSourceRowRefs,
		"max_aggregate_counter_digits":            limits.MaxAggregateCounterDigits,
		"max_contributing_rows_per_graph":         limits.MaxContributingRowsPerGraph,
		"max_time_buckets_per_graph":              limits.MaxTimeBucketsPerGraph,
		"graph_materialization_timeout_seconds":   limits.GraphMaterializationTimeoutSeconds,
	}
}

func graphCapacityRuntimeEvidence() map[string]any {
	return map[string]any{
		"go_version": runtime.Version(), "goos": runtime.GOOS, "goarch": runtime.GOARCH,
		"logical_cpu_count": runtime.NumCPU(), "host_memory_bytes": graphCapacityHostMemoryBytes(),
		"peak_rss_source": "/proc/self/status VmHWM", "host_memory_source": "/proc/meminfo MemTotal",
	}
}

func graphCapacityHostMemoryBytes() uint64 {
	body, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	return parseProcKilobytes(string(body), "MemTotal:")
}

func graphCapacityPeakRSSBytes() uint64 {
	body, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0
	}
	return parseProcKilobytes(string(body), "VmHWM:")
}

func parseProcKilobytes(body string, field string) uint64 {
	for _, line := range strings.Split(body, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 || parts[0] != field {
			continue
		}
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err == nil {
			return value * 1024
		}
	}
	return 0
}

func readGraphCapacityRunManifest(t testing.TB, artifactDir string) graphCapacityRunManifest {
	t.Helper()
	directory := filepath.Clean(artifactDir)
	for {
		candidate := filepath.Join(directory, "run-manifest.json")
		body, err := os.ReadFile(candidate)
		if err == nil {
			var manifest graphCapacityRunManifest
			if err := json.Unmarshal(body, &manifest); err != nil {
				t.Fatalf("decode capacity run manifest %s: %v", candidate, err)
			}
			if manifest.RunID == "" || manifest.SourceCommit == "" || manifest.SourceDigest == "" || manifest.ToolchainDigest == "" || manifest.SystemDigest == "" || manifest.GraphDigest == "" {
				t.Fatalf("capacity run manifest is incomplete: %#v", manifest)
			}
			return manifest
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	buildCommit := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				buildCommit = setting.Value
			}
		}
	}
	t.Fatalf("locate Make-owned run manifest for capacity artifact (build commit %q)", buildCommit)
	return graphCapacityRunManifest{}
}
