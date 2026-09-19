package networkflow

import (
	"errors"

	"github.com/google/uuid"
)

const (
	ProfileID                   = "network_flow_activity"
	workspaceKeyNetworkAnalysis = "network_analysis"
	RouteContributionID         = "network_flow_activity.route_family"
	WorkspaceContributionID     = "network_flow_activity.network_analysis_workspace"

	sourceProfileCiscoSNANetFlowCSV = "cisco_sna_netflow_csv_v1"
	parserProfileRFC4180HeaderedCSV = "rfc4180_headered_csv_v1"

	tableStatusActive      = "active"
	tableStatusSoftDeleted = "soft_deleted"

	defaultMaxActiveTablesPerIncident         = 128
	defaultMaxRetainedTablesPerIncident       = 512
	defaultMaxSelectedTablesPerQuery          = 16
	defaultMaxColumnsPerCSV                   = 256
	defaultMaxHeaderScalarLength              = 256
	defaultMaxRawCellScalarLength             = 4096
	defaultMaxRowsPerCSV                      = 250000
	defaultMaxAcceptedRowsPerTable            = 250000
	defaultMaxRejectedRowDiagnostics          = 10000
	defaultMaxFiltersPerQuery                 = 16
	defaultMaxSortsPerQuery                   = 8
	defaultMaxQueryLimit                      = 500
	defaultMaxGraphVertices                   = 5000
	defaultMaxGraphEdges                      = 10000
	defaultMaxActiveGraphViewsPerIncident     = 32
	defaultMaxRetainedGraphViewsPerIncident   = 128
	defaultMaxNonterminalGraphJobsPerIncident = 4
	defaultMaxExampleRowRefsPerEdge           = 10
	defaultMaxBindingSourceRowRefs            = 16
	defaultMaxAggregateCounterDigits          = 39
	defaultMaxContributingRowsPerGraph        = 250000
	defaultMaxTimeBucketsPerGraph             = 256
	defaultGraphMaterializationTimeoutSeconds = 300
)

var (
	errIncidentNotFound       = errors.New("networkflow: incident not found")
	errNoAcceptedRows         = errors.New("networkflow: no accepted rows")
	errTableNotFound          = errors.New("networkflow: table not found")
	errTableNotActive         = errors.New("networkflow: table not active")
	errTableVersionConflict   = errors.New("networkflow: table version conflict")
	errInvalidDisplayName     = errors.New("networkflow: invalid display name")
	errTableNameExhausted     = errors.New("networkflow: table name exhausted")
	errTableLimitExceeded     = errors.New("networkflow: table limit exceeded")
	errIDGenerationFailed     = errors.New("networkflow: id generation failed")
	errInvalidStorageArgument = errors.New("networkflow: invalid storage argument")
	errInvalidSource          = errors.New("networkflow: invalid source")
	errInvalidMapping         = errors.New("networkflow: invalid mapping")
	errSourceChanged          = errors.New("networkflow: source changed")
	errInvalidCursor          = errors.New("networkflow: invalid cursor")
)

type sourceValidationError struct {
	Code       string
	ReasonCode string
}

func (e *sourceValidationError) Error() string {
	if e.Code == "" {
		return errInvalidSource.Error()
	}
	return e.Code
}

func (e *sourceValidationError) Unwrap() error {
	return errInvalidSource
}

type mappingValidationError struct {
	Code       string
	ReasonCode string
	FieldKey   string
}

func (e *mappingValidationError) Error() string {
	if e.Code == "" {
		return errInvalidMapping.Error()
	}
	return e.Code
}

func (e *mappingValidationError) Unwrap() error {
	return errInvalidMapping
}

type invalidDisplayNameError struct {
	NormalizedLength int
	ReasonCode       string
}

func (e *invalidDisplayNameError) Error() string {
	return errInvalidDisplayName.Error()
}

func (e *invalidDisplayNameError) Unwrap() error {
	return errInvalidDisplayName
}

type tableVersionConflictError struct {
	TableID             string
	BaseTableVersion    int64
	CurrentTableVersion int64
}

func (e *tableVersionConflictError) Error() string {
	return errTableVersionConflict.Error()
}

func (e *tableVersionConflictError) Unwrap() error {
	return errTableVersionConflict
}

type tableLimitError struct {
	IncidentID uuid.UUID
	LimitName  string
	Limit      int64
	Current    int64
}

func (e *tableLimitError) Error() string {
	return errTableLimitExceeded.Error()
}

func (e *tableLimitError) Unwrap() error {
	return errTableLimitExceeded
}

// EffectiveLimits is the fully resolved, immutable Network Flow resource
// policy injected by application composition. Zero values are meaningful for
// the owner-approved limits whose minima are zero, so runtime code must never
// infer defaults from individual fields.
type EffectiveLimits struct {
	MaxActiveTablesPerIncident         int64
	MaxRetainedTablesPerIncident       int64
	MaxSelectedTablesPerQuery          int64
	MaxColumnsPerCSV                   int64
	MaxHeaderScalarLength              int64
	MaxRawCellScalarLength             int64
	MaxRowsPerCSV                      int64
	MaxAcceptedRowsPerTable            int64
	MaxRejectedRowDiagnostics          int64
	MaxFiltersPerQuery                 int64
	MaxSortsPerQuery                   int64
	MaxQueryLimit                      int64
	MaxGraphVertices                   int64
	MaxGraphEdges                      int64
	MaxActiveGraphViewsPerIncident     int64
	MaxRetainedGraphViewsPerIncident   int64
	MaxNonterminalGraphJobsPerIncident int64
	MaxExampleRowRefsPerEdge           int64
	MaxBindingSourceRowRefs            int64
	MaxAggregateCounterDigits          int64
	MaxContributingRowsPerGraph        int64
	MaxTimeBucketsPerGraph             int64
	GraphMaterializationTimeoutSeconds int64
}

func defaultEffectiveLimits() EffectiveLimits {
	return EffectiveLimits{
		MaxActiveTablesPerIncident:         defaultMaxActiveTablesPerIncident,
		MaxRetainedTablesPerIncident:       defaultMaxRetainedTablesPerIncident,
		MaxSelectedTablesPerQuery:          defaultMaxSelectedTablesPerQuery,
		MaxColumnsPerCSV:                   defaultMaxColumnsPerCSV,
		MaxHeaderScalarLength:              defaultMaxHeaderScalarLength,
		MaxRawCellScalarLength:             defaultMaxRawCellScalarLength,
		MaxRowsPerCSV:                      defaultMaxRowsPerCSV,
		MaxAcceptedRowsPerTable:            defaultMaxAcceptedRowsPerTable,
		MaxRejectedRowDiagnostics:          defaultMaxRejectedRowDiagnostics,
		MaxFiltersPerQuery:                 defaultMaxFiltersPerQuery,
		MaxSortsPerQuery:                   defaultMaxSortsPerQuery,
		MaxQueryLimit:                      defaultMaxQueryLimit,
		MaxGraphVertices:                   defaultMaxGraphVertices,
		MaxGraphEdges:                      defaultMaxGraphEdges,
		MaxActiveGraphViewsPerIncident:     defaultMaxActiveGraphViewsPerIncident,
		MaxRetainedGraphViewsPerIncident:   defaultMaxRetainedGraphViewsPerIncident,
		MaxNonterminalGraphJobsPerIncident: defaultMaxNonterminalGraphJobsPerIncident,
		MaxExampleRowRefsPerEdge:           defaultMaxExampleRowRefsPerEdge,
		MaxBindingSourceRowRefs:            defaultMaxBindingSourceRowRefs,
		MaxAggregateCounterDigits:          defaultMaxAggregateCounterDigits,
		MaxContributingRowsPerGraph:        defaultMaxContributingRowsPerGraph,
		MaxTimeBucketsPerGraph:             defaultMaxTimeBucketsPerGraph,
		GraphMaterializationTimeoutSeconds: defaultGraphMaterializationTimeoutSeconds,
	}
}
