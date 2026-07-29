package workbookprobe

import (
	"context"

	"github.com/google/uuid"
)

const (
	RegistrationSchemaID = "cartulary.restore_workbook_probe_registration.v1"
	BaseProfile          = "base"
)

type Filter struct {
	FieldKey string         `json:"field_key"`
	Op       string         `json:"op"`
	Arg      map[string]any `json:"arg"`
}

type Sort struct {
	FieldKey  string `json:"field_key"`
	Direction string `json:"direction"`
}

type Registration struct {
	SchemaID       string   `json:"schema_id"`
	RegistrationID string   `json:"registration_id"`
	OwnerID        string   `json:"owner_id"`
	Profile        string   `json:"profile"`
	IsDefault      bool     `json:"is_default"`
	ViewSchemaID   string   `json:"view_schema_id"`
	Filters        []Filter `json:"filters"`
	Sort           []Sort   `json:"sort"`
	GroupBy        *string  `json:"group_by"`
	RowRequirement string   `json:"row_requirement"`
}

type Result struct {
	RegistrationID string
	ViewSchemaID   string
	RowCount       int64
}

type Executor interface {
	ExecuteDefault(context.Context, string, uuid.UUID) (Result, error)
}
