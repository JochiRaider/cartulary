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
)

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
}

func DefaultLimits() Limits {
	return Limits{
		MaxActiveTablesPerIncident:   DefaultMaxActiveTablesPerIncident,
		MaxRetainedTablesPerIncident: DefaultMaxRetainedTablesPerIncident,
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
	return l
}

func LifecycleStates() []string {
	return []string{TableStatusActive, TableStatusSoftDeleted}
}
