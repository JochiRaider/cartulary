package networkflow

import (
	"errors"

	"github.com/google/uuid"
)

const (
	ProfileID                   = "network_flow_activity"
	WorkspaceKeyNetworkAnalysis = "network_analysis"
	RouteContributionID         = "network_flow_activity.route_family"
	WorkspaceContributionID     = "network_flow_activity.network_analysis_workspace"

	SourceProfileCiscoSNANetFlowCSV = "cisco_sna_netflow_csv_v1"
	ParserProfileRFC4180HeaderedCSV = "rfc4180_headered_csv_v1"

	TableStatusActive      = "active"
	TableStatusSoftDeleted = "soft_deleted"

	DefaultMaxActiveTablesPerIncident         = 128
	DefaultMaxRetainedTablesPerIncident       = 512
	DefaultMaxSelectedTablesPerQuery          = 16
	DefaultMaxColumnsPerCSV                   = 256
	DefaultMaxHeaderScalarLength              = 256
	DefaultMaxRawCellScalarLength             = 4096
	DefaultMaxRowsPerCSV                      = 250000
	DefaultMaxAcceptedRowsPerTable            = 250000
	DefaultMaxRejectedRowDiagnostics          = 10000
	DefaultMaxFiltersPerQuery                 = 16
	DefaultMaxSortsPerQuery                   = 8
	DefaultMaxQueryLimit                      = 500
	DefaultMaxGraphVertices                   = 5000
	DefaultMaxGraphEdges                      = 10000
	DefaultMaxActiveGraphViewsPerIncident     = 32
	DefaultMaxRetainedGraphViewsPerIncident   = 128
	DefaultMaxNonterminalGraphJobsPerIncident = 4
	DefaultMaxExampleRowRefsPerEdge           = 10
	DefaultMaxBindingSourceRowRefs            = 16
	DefaultMaxAggregateCounterDigits          = 39
	DefaultMaxContributingRowsPerGraph        = 250000
	DefaultMaxTimeBucketsPerGraph             = 256
	DefaultGraphMaterializationTimeoutSeconds = 300
)

var (
	ErrIncidentNotFound       = errors.New("networkflow: incident not found")
	ErrNoAcceptedRows         = errors.New("networkflow: no accepted rows")
	ErrTableNotFound          = errors.New("networkflow: table not found")
	ErrTableNotActive         = errors.New("networkflow: table not active")
	ErrTableVersionConflict   = errors.New("networkflow: table version conflict")
	ErrInvalidDisplayName     = errors.New("networkflow: invalid display name")
	ErrTableNameExhausted     = errors.New("networkflow: table name exhausted")
	ErrTableLimitExceeded     = errors.New("networkflow: table limit exceeded")
	ErrIDGenerationFailed     = errors.New("networkflow: id generation failed")
	ErrInvalidStorageArgument = errors.New("networkflow: invalid storage argument")
	ErrInvalidSource          = errors.New("networkflow: invalid source")
	ErrInvalidMapping         = errors.New("networkflow: invalid mapping")
	ErrSourceChanged          = errors.New("networkflow: source changed")
	ErrInvalidQuery           = errors.New("networkflow: invalid query")
	ErrInvalidCursor          = errors.New("networkflow: invalid cursor")
)

type SourceValidationError struct {
	Code       string
	ReasonCode string
}

func (e *SourceValidationError) Error() string {
	if e.Code == "" {
		return ErrInvalidSource.Error()
	}
	return e.Code
}

func (e *SourceValidationError) Unwrap() error {
	return ErrInvalidSource
}

type MappingValidationError struct {
	Code       string
	ReasonCode string
	FieldKey   string
}

func (e *MappingValidationError) Error() string {
	if e.Code == "" {
		return ErrInvalidMapping.Error()
	}
	return e.Code
}

func (e *MappingValidationError) Unwrap() error {
	return ErrInvalidMapping
}

type InvalidDisplayNameError struct {
	ReasonCode string
}

func (e *InvalidDisplayNameError) Error() string {
	return ErrInvalidDisplayName.Error()
}

func (e *InvalidDisplayNameError) Unwrap() error {
	return ErrInvalidDisplayName
}

type TableVersionConflictError struct {
	TableID             string
	BaseTableVersion    int64
	CurrentTableVersion int64
}

func (e *TableVersionConflictError) Error() string {
	return ErrTableVersionConflict.Error()
}

func (e *TableVersionConflictError) Unwrap() error {
	return ErrTableVersionConflict
}

type TableLimitError struct {
	IncidentID uuid.UUID
	LimitName  string
	Limit      int64
	Current    int64
}

func (e *TableLimitError) Error() string {
	return ErrTableLimitExceeded.Error()
}

func (e *TableLimitError) Unwrap() error {
	return ErrTableLimitExceeded
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

func DefaultEffectiveLimits() EffectiveLimits {
	return EffectiveLimits{
		MaxActiveTablesPerIncident:         DefaultMaxActiveTablesPerIncident,
		MaxRetainedTablesPerIncident:       DefaultMaxRetainedTablesPerIncident,
		MaxSelectedTablesPerQuery:          DefaultMaxSelectedTablesPerQuery,
		MaxColumnsPerCSV:                   DefaultMaxColumnsPerCSV,
		MaxHeaderScalarLength:              DefaultMaxHeaderScalarLength,
		MaxRawCellScalarLength:             DefaultMaxRawCellScalarLength,
		MaxRowsPerCSV:                      DefaultMaxRowsPerCSV,
		MaxAcceptedRowsPerTable:            DefaultMaxAcceptedRowsPerTable,
		MaxRejectedRowDiagnostics:          DefaultMaxRejectedRowDiagnostics,
		MaxFiltersPerQuery:                 DefaultMaxFiltersPerQuery,
		MaxSortsPerQuery:                   DefaultMaxSortsPerQuery,
		MaxQueryLimit:                      DefaultMaxQueryLimit,
		MaxGraphVertices:                   DefaultMaxGraphVertices,
		MaxGraphEdges:                      DefaultMaxGraphEdges,
		MaxActiveGraphViewsPerIncident:     DefaultMaxActiveGraphViewsPerIncident,
		MaxRetainedGraphViewsPerIncident:   DefaultMaxRetainedGraphViewsPerIncident,
		MaxNonterminalGraphJobsPerIncident: DefaultMaxNonterminalGraphJobsPerIncident,
		MaxExampleRowRefsPerEdge:           DefaultMaxExampleRowRefsPerEdge,
		MaxBindingSourceRowRefs:            DefaultMaxBindingSourceRowRefs,
		MaxAggregateCounterDigits:          DefaultMaxAggregateCounterDigits,
		MaxContributingRowsPerGraph:        DefaultMaxContributingRowsPerGraph,
		MaxTimeBucketsPerGraph:             DefaultMaxTimeBucketsPerGraph,
		GraphMaterializationTimeoutSeconds: DefaultGraphMaterializationTimeoutSeconds,
	}
}

// DefaultLimits is retained as the explicit constructor for callers that need
// the adopted default policy. It performs no normalization of caller input.
func DefaultLimits() EffectiveLimits {
	return DefaultEffectiveLimits()
}

func LifecycleStates() []string {
	return []string{TableStatusActive, TableStatusSoftDeleted}
}
