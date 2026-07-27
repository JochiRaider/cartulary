package savedviews

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func ExportIncidentBundleFiles(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]incidentportability.File, error) {
	file, err := incidentportability.ExportNDJSON(ctx, q, incidentID, "data/saved_views.ndjson", `SELECT to_jsonb(t) FROM saved_views t WHERE incident_id = $1 ORDER BY saved_view_id`)
	if err != nil {
		return nil, err
	}
	return []incidentportability.File{file}, nil
}

func ImportIncidentBundleFilesTx(ctx context.Context, tx pgx.Tx, files map[string][]byte, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	spec := savedViewImportSpec()
	payload, ok := files[spec.LogicalBundlePath]
	if !ok {
		return &incidentportability.VerificationFailure{ReasonCode: "missing_required_file"}
	}
	rows, err := incidentportability.DecodeNDJSON(payload)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := validateAndNormalizeImportedSavedView(row, spec); err != nil {
			return err
		}
	}
	return incidentportability.ImportFixedRows(ctx, tx, spec, rows, actorUserID, attributions)
}

func savedViewImportSpec() incidentportability.FixedImportSpec {
	return incidentportability.FixedImportSpec{
		LogicalBundlePath: "data/saved_views.ndjson",
		AttributionTable:  "saved_views",
		StableIdentity:    []string{"saved_view_id"},
		RequiredColumns:   []string{"saved_view_id", "incident_id", "view_schema_id", "scope", "display_name", "query_json", "layout_json", "owner_user_id", "created_at", "updated_at", "saved_view_version"},
		InsertSQL:         `INSERT INTO saved_views SELECT * FROM jsonb_populate_record(NULL::saved_views, $1::jsonb)`,
	}
}

func validateAndNormalizeImportedSavedView(row map[string]any, spec incidentportability.FixedImportSpec) error {
	if err := incidentportability.ValidateRequiredColumns(row, spec.RequiredColumns, spec.StableIdentity); err != nil {
		return err
	}
	if _, err := uuid.Parse(requiredString(row, "saved_view_id")); err != nil {
		return malformedIncidentBundle()
	}
	if _, err := uuid.Parse(requiredString(row, "incident_id")); err != nil {
		return malformedIncidentBundle()
	}
	viewSchemaID := requiredString(row, "view_schema_id")
	if _, ok := viewschema.Lookup(viewSchemaID); !ok {
		return malformedIncidentBundle()
	}
	scope, ok := ParseScope(requiredString(row, "scope"))
	if !ok {
		return malformedIncidentBundle()
	}
	ownerUserID := strings.TrimSpace(incidentportability.StringFromAny(row["owner_user_id"]))
	switch scope {
	case ScopePrivate, ScopeShared:
		if _, err := uuid.Parse(ownerUserID); err != nil {
			return malformedIncidentBundle()
		}
	case ScopeSystem:
		if ownerUserID != "" {
			return malformedIncidentBundle()
		}
	default:
		return malformedIncidentBundle()
	}
	displayName, ok := NormalizeDisplayName(requiredString(row, "display_name"))
	if !ok {
		return malformedIncidentBundle()
	}
	row["display_name"] = displayName
	queryRaw, err := rawJSONFromImportedSavedView(row["query_json"])
	if err != nil {
		return err
	}
	queryJSON, validationErr := viewquery.NormalizePersisted(queryRaw, viewSchemaID)
	if validationErr != nil {
		return malformedIncidentBundle()
	}
	row["query_json"] = json.RawMessage(queryJSON)
	layoutRaw, err := rawJSONFromImportedSavedView(row["layout_json"])
	if err != nil {
		return err
	}
	layoutJSON, layoutErr := viewschema.NormalizeLayout(layoutRaw, viewSchemaID)
	if layoutErr != nil {
		return malformedIncidentBundle()
	}
	row["layout_json"] = json.RawMessage(layoutJSON)
	if _, err := strconv.ParseInt(requiredString(row, "saved_view_version"), 10, 64); err != nil {
		return malformedIncidentBundle()
	}
	if strings.TrimSpace(incidentportability.StringFromAny(row["created_at"])) == "" || strings.TrimSpace(incidentportability.StringFromAny(row["updated_at"])) == "" {
		return malformedIncidentBundle()
	}
	return nil
}

func rawJSONFromImportedSavedView(value any) (json.RawMessage, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(raw)) == "" || strings.TrimSpace(string(raw)) == "null" {
		return nil, malformedIncidentBundle()
	}
	return raw, nil
}

func requiredString(row map[string]any, key string) string {
	return strings.TrimSpace(incidentportability.StringFromAny(row[key]))
}

func malformedIncidentBundle() error {
	return &incidentportability.VerificationFailure{ReasonCode: "malformed_manifest"}
}
