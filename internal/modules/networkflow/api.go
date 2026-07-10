package networkflow

import (
	"errors"

	"github.com/google/uuid"
)

const (
	ProfileID = "network_flow_activity"

	SourceProfileCiscoSNANetFlowCSV = "cisco_sna_netflow_csv_v1"
	ParserProfileRFC4180HeaderedCSV = "rfc4180_headered_csv_v1"

	TableStatusActive      = "active"
	TableStatusSoftDeleted = "soft_deleted"

	DefaultMaxActiveTablesPerIncident   = 128
	DefaultMaxRetainedTablesPerIncident = 512
	DefaultMaxColumnsPerCSV             = 256
	DefaultMaxRowsPerCSV                = 250000
	DefaultMaxAcceptedRowsPerTable      = 250000
	DefaultMaxRejectedRowDiagnostics    = 10000
	DefaultMaxHeaderScalarLength        = 1024
	DefaultMaxRawCellScalarLength       = 4096
	DefaultMaxSelectedTablesPerQuery    = 16
	DefaultMaxFiltersPerQuery           = 16
	DefaultMaxSortsPerQuery             = 8
	DefaultMaxQueryLimit                = 500
	DefaultMaxGraphVertices             = 5000
	DefaultMaxGraphEdges                = 10000
	DefaultMaxExampleRowRefsPerEdge     = 10
	DefaultMaxBindingSourceRowRefs      = 16
	DefaultMaxAggregateCounterDigits    = 39
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

type Limits struct {
	MaxActiveTablesPerIncident   int64
	MaxRetainedTablesPerIncident int64
	MaxColumnsPerCSV             int64
	MaxRowsPerCSV                int64
	MaxAcceptedRowsPerTable      int64
	MaxRejectedRowDiagnostics    int64
	MaxHeaderScalarLength        int64
	MaxRawCellScalarLength       int64
	MaxSelectedTablesPerQuery    int64
	MaxFiltersPerQuery           int64
	MaxSortsPerQuery             int64
	MaxQueryLimit                int64
	MaxGraphVertices             int64
	MaxGraphEdges                int64
	MaxExampleRowRefsPerEdge     int64
	MaxBindingSourceRowRefs      int64
	MaxAggregateCounterDigits    int64
}

func DefaultLimits() Limits {
	return Limits{
		MaxActiveTablesPerIncident:   DefaultMaxActiveTablesPerIncident,
		MaxRetainedTablesPerIncident: DefaultMaxRetainedTablesPerIncident,
		MaxColumnsPerCSV:             DefaultMaxColumnsPerCSV,
		MaxRowsPerCSV:                DefaultMaxRowsPerCSV,
		MaxAcceptedRowsPerTable:      DefaultMaxAcceptedRowsPerTable,
		MaxRejectedRowDiagnostics:    DefaultMaxRejectedRowDiagnostics,
		MaxHeaderScalarLength:        DefaultMaxHeaderScalarLength,
		MaxRawCellScalarLength:       DefaultMaxRawCellScalarLength,
		MaxSelectedTablesPerQuery:    DefaultMaxSelectedTablesPerQuery,
		MaxFiltersPerQuery:           DefaultMaxFiltersPerQuery,
		MaxSortsPerQuery:             DefaultMaxSortsPerQuery,
		MaxQueryLimit:                DefaultMaxQueryLimit,
		MaxGraphVertices:             DefaultMaxGraphVertices,
		MaxGraphEdges:                DefaultMaxGraphEdges,
		MaxExampleRowRefsPerEdge:     DefaultMaxExampleRowRefsPerEdge,
		MaxBindingSourceRowRefs:      DefaultMaxBindingSourceRowRefs,
		MaxAggregateCounterDigits:    DefaultMaxAggregateCounterDigits,
	}
}

func (l Limits) normalized() Limits {
	defaults := DefaultLimits()
	if l.MaxActiveTablesPerIncident <= 0 {
		l.MaxActiveTablesPerIncident = defaults.MaxActiveTablesPerIncident
	}
	if l.MaxRetainedTablesPerIncident <= 0 {
		l.MaxRetainedTablesPerIncident = defaults.MaxRetainedTablesPerIncident
	}
	if l.MaxActiveTablesPerIncident > l.MaxRetainedTablesPerIncident {
		l.MaxActiveTablesPerIncident = l.MaxRetainedTablesPerIncident
	}
	if l.MaxColumnsPerCSV <= 0 {
		l.MaxColumnsPerCSV = defaults.MaxColumnsPerCSV
	}
	if l.MaxRowsPerCSV <= 0 {
		l.MaxRowsPerCSV = defaults.MaxRowsPerCSV
	}
	if l.MaxAcceptedRowsPerTable <= 0 {
		l.MaxAcceptedRowsPerTable = defaults.MaxAcceptedRowsPerTable
	}
	if l.MaxRejectedRowDiagnostics < 0 {
		l.MaxRejectedRowDiagnostics = defaults.MaxRejectedRowDiagnostics
	}
	if l.MaxHeaderScalarLength <= 0 {
		l.MaxHeaderScalarLength = defaults.MaxHeaderScalarLength
	}
	if l.MaxRawCellScalarLength <= 0 {
		l.MaxRawCellScalarLength = defaults.MaxRawCellScalarLength
	}
	if l.MaxSelectedTablesPerQuery <= 0 {
		l.MaxSelectedTablesPerQuery = defaults.MaxSelectedTablesPerQuery
	}
	if l.MaxFiltersPerQuery <= 0 {
		l.MaxFiltersPerQuery = defaults.MaxFiltersPerQuery
	}
	if l.MaxSortsPerQuery <= 0 {
		l.MaxSortsPerQuery = defaults.MaxSortsPerQuery
	}
	if l.MaxQueryLimit <= 0 {
		l.MaxQueryLimit = defaults.MaxQueryLimit
	}
	if l.MaxGraphVertices <= 0 {
		l.MaxGraphVertices = defaults.MaxGraphVertices
	}
	if l.MaxGraphEdges <= 0 {
		l.MaxGraphEdges = defaults.MaxGraphEdges
	}
	if l.MaxExampleRowRefsPerEdge < 0 {
		l.MaxExampleRowRefsPerEdge = defaults.MaxExampleRowRefsPerEdge
	}
	if l.MaxBindingSourceRowRefs <= 0 {
		l.MaxBindingSourceRowRefs = defaults.MaxBindingSourceRowRefs
	}
	if l.MaxAggregateCounterDigits <= 0 {
		l.MaxAggregateCounterDigits = defaults.MaxAggregateCounterDigits
	}
	return l
}

func LifecycleStates() []string {
	return []string{TableStatusActive, TableStatusSoftDeleted}
}
