package savedviews

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type savedViewPortableAttributionResolver interface {
	ResolvePortableSourceActors(
		context.Context,
		incidentportability.Queryer,
		uuid.UUID,
		string,
		string,
		[]string,
	) (map[string]string, error)
}

type savedViewExportContext struct {
	Actors               incidentportability.ActorReferences
	Query                incidentportability.Queryer
	IncidentID           uuid.UUID
	PortableAttributions savedViewPortableAttributionResolver
}

type savedViewImportContext struct {
	BundleVersion int
	IncidentID    uuid.UUID
	ActorUserID   uuid.UUID
	Attributions  incidentportability.AttributionRecorder
	ActorAdmitted func(string) bool
	// Preparation is the port's preparation scope (Core 01 REQ-01-641):
	// preparation stages the admitted rows and owner attribution there, and
	// the final transaction reads them back without reopening the bundle.
	Preparation sourceport.PreparationStore
}

type savedViewInvariantError struct {
	InvariantID string
}

func (e *savedViewInvariantError) Error() string {
	return "saved views failed invariant " + e.InvariantID
}

func exportIncidentBundleFiles(ctx context.Context, exportContext savedViewExportContext, members incidentportability.MemberWriter) error {
	return incidentportability.WriteMemberFrom(ctx, members, "data/saved_views.ndjson", func(ctx context.Context, output io.Writer) error {
		var after *uuid.UUID
		for {
			next, err := exportIncidentBundlePage(ctx, exportContext, after, output)
			if err != nil {
				return err
			}
			if next == nil {
				return nil
			}
			after = next
		}
	})
}

func exportIncidentBundlePage(ctx context.Context, exportContext savedViewExportContext, after *uuid.UUID, output io.Writer) (*uuid.UUID, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT saved_view_id,
       incident_id,
       view_schema_id,
       scope,
       display_name,
       query_json,
       layout_json,
       owner_user_id,
       created_at,
       updated_at,
       saved_view_version
  FROM saved_views
 WHERE incident_id = $1 AND ($2::uuid IS NULL OR saved_view_id > $2)
 ORDER BY saved_view_id
 LIMIT 128
`, exportContext.IncidentID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exportRows := make([]savedViewPortableRow, 0)
	sourceRowIDs := make([]string, 0)
	for rows.Next() {
		var (
			row          savedViewPortableRow
			savedViewID  pgtype.UUID
			incidentID   pgtype.UUID
			ownerUserID  pgtype.UUID
			createdAt    pgtype.Timestamptz
			updatedAt    pgtype.Timestamptz
			queryJSON    []byte
			layoutJSON   []byte
			scopeValue   string
			viewSchemaID string
			displayName  string
			version      int64
		)
		if err := rows.Scan(
			&savedViewID,
			&incidentID,
			&viewSchemaID,
			&scopeValue,
			&displayName,
			&queryJSON,
			&layoutJSON,
			&ownerUserID,
			&createdAt,
			&updatedAt,
			&version,
		); err != nil {
			return nil, err
		}
		parsedSavedViewID, err := uuidFromPG(savedViewID)
		if err != nil {
			return nil, fmt.Errorf("export saved view id: %w", err)
		}
		parsedIncidentID, err := uuidFromPG(incidentID)
		if err != nil {
			return nil, fmt.Errorf("export saved view incident id: %w", err)
		}
		parsedCreatedAt, err := timeFromPG(createdAt)
		if err != nil {
			return nil, fmt.Errorf("export saved view created at: %w", err)
		}
		parsedUpdatedAt, err := timeFromPG(updatedAt)
		if err != nil {
			return nil, fmt.Errorf("export saved view updated at: %w", err)
		}
		parsedScope, ok := parseScope(scopeValue)
		if !ok {
			return nil, fmt.Errorf("export saved view scope is invalid")
		}
		normalizedDisplayName, ok := normalizeDisplayName(displayName)
		if !ok {
			return nil, fmt.Errorf("export saved view display name is invalid")
		}
		normalizedQuery, queryErr := viewquery.NormalizePersisted(queryJSON, viewSchemaID)
		if queryErr != nil {
			return nil, fmt.Errorf("export saved view query is invalid")
		}
		normalizedLayout, layoutErr := viewschema.NormalizeLayout(layoutJSON, viewSchemaID)
		if layoutErr != nil {
			return nil, fmt.Errorf("export saved view layout is invalid")
		}
		if version < 1 || parsedUpdatedAt.Before(parsedCreatedAt) {
			return nil, fmt.Errorf("export saved view version or timestamps are invalid")
		}
		if (parsedScope == scopeSystem && ownerUserID.Valid) ||
			(parsedScope != scopeSystem && !ownerUserID.Valid) {
			return nil, fmt.Errorf("export saved view owner tuple is invalid")
		}
		queryValue, err := decodeSavedViewPortableJSON(normalizedQuery)
		if err != nil {
			return nil, fmt.Errorf("export saved view query: %w", err)
		}
		layoutValue, err := decodeSavedViewPortableJSON(normalizedLayout)
		if err != nil {
			return nil, fmt.Errorf("export saved view layout: %w", err)
		}
		row = savedViewPortableRow{
			SavedViewID:      parsedSavedViewID,
			IncidentID:       parsedIncidentID,
			ViewSchemaID:     viewSchemaID,
			Scope:            parsedScope,
			DisplayName:      normalizedDisplayName,
			QueryJSON:        queryValue,
			LayoutJSON:       layoutValue,
			OwnerUserID:      optionalUUIDFromPG(ownerUserID),
			CreatedAt:        parsedCreatedAt,
			UpdatedAt:        parsedUpdatedAt,
			SavedViewVersion: version,
		}
		if parsedScope != scopeSystem {
			sourceRowIDs = append(sourceRowIDs, parsedSavedViewID.String())
		}
		exportRows = append(exportRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	portableOwners := map[string]string{}
	if exportContext.PortableAttributions != nil && len(sourceRowIDs) > 0 {
		portableOwners, err = exportContext.PortableAttributions.ResolvePortableSourceActors(
			ctx,
			exportContext.Query,
			exportContext.IncidentID,
			"saved_views",
			"owner_user_id",
			sourceRowIDs,
		)
		if err != nil {
			return nil, err
		}
	}

	rows.Close()
	for _, row := range exportRows {
		owner := any(nil)
		if row.Scope != scopeSystem {
			if row.OwnerUserID == nil {
				return nil, fmt.Errorf("export saved view runtime owner is missing")
			}
			owner, err = incidentportability.ExportActorID(exportContext.Actors, portableOwners[row.SavedViewID.String()], *row.OwnerUserID)
			if err != nil {
				return nil, err
			}
		}
		err = incidentportability.WriteNDJSONRecord(output, map[string]any{
			"saved_view_id":      row.SavedViewID.String(),
			"incident_id":        row.IncidentID.String(),
			"view_schema_id":     row.ViewSchemaID,
			"scope":              string(row.Scope),
			"display_name":       row.DisplayName,
			"query_json":         row.QueryJSON,
			"layout_json":        row.LayoutJSON,
			"owner_user_id":      owner,
			"created_at":         row.CreatedAt.UTC().Format(time.RFC3339Nano),
			"updated_at":         row.UpdatedAt.UTC().Format(time.RFC3339Nano),
			"saved_view_version": row.SavedViewVersion,
		})
		if err != nil {
			return nil, err
		}
	}
	if len(exportRows) == 0 {
		return nil, nil
	}
	last := exportRows[len(exportRows)-1].SavedViewID
	return &last, nil
}

type savedViewPortableRow struct {
	SavedViewID      uuid.UUID
	IncidentID       uuid.UUID
	ViewSchemaID     string
	Scope            scope
	DisplayName      string
	QueryJSON        any
	LayoutJSON       any
	OwnerUserID      *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SavedViewVersion int64
}

func decodeSavedViewPortableJSON(raw []byte) (any, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

// preparedSavedViewImport holds the staged rows and owner attribution; the
// final transaction reads no bundle member.
type preparedSavedViewImport struct {
	rows                incidentportability.PreparedRowStream
	attribution         incidentportability.AttributionStream
	incidentID, actorID uuid.UUID
	count, owners       uint64
	staged              bool
}

type preparedSavedViewRow struct {
	SavedViewID      uuid.UUID       `json:"id"`
	IncidentID       uuid.UUID       `json:"incident"`
	ViewSchemaID     string          `json:"schema"`
	Scope            scope           `json:"scope"`
	DisplayName      string          `json:"name"`
	QueryJSON        json.RawMessage `json:"query"`
	LayoutJSON       json.RawMessage `json:"layout"`
	PortableOwnerID  *uuid.UUID      `json:"owner"`
	RuntimeOwnerID   *uuid.UUID      `json:"runtime_owner"`
	CreatedAt        time.Time       `json:"created"`
	UpdatedAt        time.Time       `json:"updated"`
	SavedViewVersion int64           `json:"version"`
}

const (
	savedViewMemberPath        = "data/saved_views.ndjson"
	preparedSavedViewsPath     = "saved_views.rows"
	preparedSavedViewAttrPath  = "attribution.rows"
	savedViewOwnerColumnSource = "owner_user_id"
)

// savedViewRowCodec stages one admitted row.
type savedViewRowCodec struct{}

func (savedViewRowCodec) Encode(row preparedSavedViewRow) ([]byte, error) {
	return incidentportability.EncodeJSONRow(row)
}

func (savedViewRowCodec) Decode(line []byte) (preparedSavedViewRow, error) {
	var row preparedSavedViewRow
	return row, incidentportability.DecodeJSONRow(line, &row)
}

func (savedViewRowCodec) Columns() []string { return nil }

func (savedViewRowCodec) Values(preparedSavedViewRow) ([]any, error) {
	return nil, incidentportability.ErrPreparedStream
}

func prepareSavedViewImport(ctx context.Context, bundle incidentportability.MemberReader, importContext savedViewImportContext) (preparedSavedViewImport, error) {
	if importContext.BundleVersion != 8 {
		return preparedSavedViewImport{}, savedViewInvariantFailure("saved_views.query_layout_legal")
	}
	members, err := incidentportability.BindMembers(bundle, []string{savedViewMemberPath})
	if err != nil {
		return preparedSavedViewImport{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	store := importContext.Preparation
	if store == nil {
		return preparedSavedViewImport{}, errors.New("saved views: incident bundle preparation staging is required")
	}
	prepared := preparedSavedViewImport{incidentID: importContext.IncidentID, actorID: importContext.ActorUserID}
	previous := ""
	prepared.rows, err = incidentportability.StagePreparedRows(ctx, store, preparedSavedViewsPath, savedViewRowCodec{}, func(emit func(preparedSavedViewRow) error) error {
		return walkSavedViewMembers(ctx, members, importContext, func(row preparedSavedViewRow) error {
			id := row.SavedViewID.String()
			if id <= previous {
				return savedViewInvariantFailure("saved_views.identity_scope_legal")
			}
			previous = id
			prepared.count++
			if row.PortableOwnerID != nil {
				prepared.owners++
			}
			return emit(row)
		})
	})
	if err != nil {
		return preparedSavedViewImport{}, err
	}
	admitted := func(actorID string) bool {
		return importContext.ActorAdmitted != nil && importContext.ActorAdmitted(actorID)
	}
	prepared.attribution, err = incidentportability.StageAttributions(ctx, store, preparedSavedViewAttrPath, admitted, func(emit func(incidentportability.AttributionRow) error) error {
		return incidentportability.ReadPreparedRows(ctx, store, prepared.rows, savedViewRowCodec{}, func(row preparedSavedViewRow) error {
			if row.PortableOwnerID == nil {
				return nil
			}
			return emit(incidentportability.AttributionRow{SourceTable: "saved_views", SourceRowID: row.SavedViewID.String(), SourceColumn: savedViewOwnerColumnSource, SourceActorID: row.PortableOwnerID.String()})
		})
	})
	if err != nil {
		return preparedSavedViewImport{}, err
	}
	prepared.staged = true
	return prepared, nil
}

// walk visits the staged rows in order; the final transaction reads no bundle
// member.
func (p preparedSavedViewImport) walk(ctx context.Context, importContext savedViewImportContext, visit func(preparedSavedViewRow) error) error {
	if !p.staged || importContext.Preparation == nil || p.incidentID != importContext.IncidentID || p.actorID != importContext.ActorUserID {
		return savedViewInvariantFailure("saved_views.identity_scope_legal")
	}
	return incidentportability.ReadPreparedRows(ctx, importContext.Preparation, p.rows, savedViewRowCodec{}, visit)
}

// walkSavedViewMembers admits the bundle rows while preparing.
func walkSavedViewMembers(ctx context.Context, members *incidentportability.BoundMembers, importContext savedViewImportContext, visit func(preparedSavedViewRow) error) error {
	err := incidentportability.WalkPath(ctx, members, savedViewMemberPath, func(raw map[string]any) error {
		row, err := prepareSavedViewRow(raw, importContext)
		if err != nil {
			return err
		}
		return visit(row)
	})
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	return err
}

func prepareSavedViewRow(row map[string]any, importContext savedViewImportContext) (preparedSavedViewRow, error) {
	required := []string{
		"saved_view_id", "incident_id", "view_schema_id", "scope",
		"display_name", "query_json", "layout_json", "owner_user_id",
		"created_at", "updated_at", "saved_view_version",
	}
	if len(row) != len(required) {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	for _, member := range required {
		if _, ok := row[member]; !ok {
			return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
		}
	}
	savedViewIDText, ok := row["saved_view_id"].(string)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	savedViewID, err := uuid.Parse(savedViewIDText)
	if err != nil || savedViewID.String() != savedViewIDText {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.identity_scope_legal")
	}
	incidentIDText, ok := row["incident_id"].(string)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	incidentID, err := uuid.Parse(incidentIDText)
	if err != nil || incidentID.String() != incidentIDText || incidentID != importContext.IncidentID {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.identity_scope_legal")
	}
	viewSchemaID, ok := row["view_schema_id"].(string)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	if _, ok := viewschema.Lookup(viewSchemaID); !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.identity_scope_legal")
	}
	scopeText, ok := row["scope"].(string)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	parsedScope, ok := parseScope(scopeText)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.identity_scope_legal")
	}
	displayName, ok := row["display_name"].(string)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	normalizedDisplayName, ok := normalizeDisplayName(displayName)
	if !ok || normalizedDisplayName != displayName {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.display_name_normalized")
	}
	queryValue, ok := row["query_json"].(map[string]any)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	if !savedViewQueryShapeExact(queryValue) {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	queryRaw, err := json.Marshal(queryValue)
	if err != nil {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.query_layout_legal")
	}
	queryJSON, validationErr := viewquery.NormalizePersisted(queryRaw, viewSchemaID)
	if validationErr != nil {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.query_layout_legal")
	}
	queryEqual, err := jsonStructurallyEqual(queryRaw, queryJSON)
	if err != nil || !queryEqual {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.query_layout_legal")
	}
	layoutValue, ok := row["layout_json"].(map[string]any)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	if !savedViewLayoutShapeExact(layoutValue) {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	layoutRaw, err := json.Marshal(layoutValue)
	if err != nil {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.query_layout_legal")
	}
	layoutJSON, layoutErr := viewschema.NormalizeLayout(layoutRaw, viewSchemaID)
	if layoutErr != nil {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.query_layout_legal")
	}
	layoutEqual, err := jsonStructurallyEqual(layoutRaw, layoutJSON)
	if err != nil || !layoutEqual {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.query_layout_legal")
	}

	var portableOwnerID *uuid.UUID
	var runtimeOwnerID *uuid.UUID
	switch parsedScope {
	case scopePrivate, scopeShared:
		ownerText, ok := row["owner_user_id"].(string)
		if !ok {
			return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.owner_tuple_legal")
		}
		ownerID, err := uuid.Parse(ownerText)
		if err != nil || ownerID.String() != ownerText {
			return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.owner_tuple_legal")
		}
		if importContext.ActorAdmitted == nil || !importContext.ActorAdmitted(ownerText) {
			return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.owner_tuple_legal")
		}
		portableOwnerID = &ownerID
		targetOwnerID := importContext.ActorUserID
		runtimeOwnerID = &targetOwnerID
	case scopeSystem:
		if row["owner_user_id"] != nil {
			return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.owner_tuple_legal")
		}
	default:
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.identity_scope_legal")
	}

	createdAtText, ok := row["created_at"].(string)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	createdAt, err := parseCanonicalPortableTimestamp(createdAtText)
	if err != nil {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.version_timestamps_legal")
	}
	updatedAtText, ok := row["updated_at"].(string)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	updatedAt, err := parseCanonicalPortableTimestamp(updatedAtText)
	if err != nil || updatedAt.Before(createdAt) {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.version_timestamps_legal")
	}
	versionNumber, ok := row["saved_view_version"].(json.Number)
	if !ok {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	version, err := strconv.ParseInt(versionNumber.String(), 10, 64)
	if err != nil || version < 1 {
		return preparedSavedViewRow{}, savedViewInvariantFailure("saved_views.version_timestamps_legal")
	}
	return preparedSavedViewRow{
		SavedViewID:      savedViewID,
		IncidentID:       incidentID,
		ViewSchemaID:     viewSchemaID,
		Scope:            parsedScope,
		DisplayName:      displayName,
		QueryJSON:        append([]byte(nil), queryJSON...),
		LayoutJSON:       append([]byte(nil), layoutJSON...),
		PortableOwnerID:  portableOwnerID,
		RuntimeOwnerID:   runtimeOwnerID,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		SavedViewVersion: version,
	}, nil
}

func savedViewQueryShapeExact(query map[string]any) bool {
	// An empty query is structurally an object, but is noncanonical and therefore
	// belongs to the query/layout invariant rather than the row-shape invariant.
	if len(query) == 0 {
		return true
	}
	if !hasExactMembers(query, []string{"sort", "filters"}, []string{"group_by"}) {
		return false
	}
	sortEntries, ok := query["sort"].([]any)
	if !ok {
		return false
	}
	for _, rawEntry := range sortEntries {
		entry, ok := rawEntry.(map[string]any)
		if !ok || !hasExactMembers(entry, []string{"field_key", "direction"}, nil) {
			return false
		}
		if _, ok := entry["field_key"].(string); !ok {
			return false
		}
		if _, ok := entry["direction"].(string); !ok {
			return false
		}
	}
	filters, ok := query["filters"].([]any)
	if !ok {
		return false
	}
	for _, rawFilter := range filters {
		filter, ok := rawFilter.(map[string]any)
		if !ok || !hasExactMembers(filter, []string{"field_key", "op", "arg"}, nil) {
			return false
		}
		if _, ok := filter["field_key"].(string); !ok {
			return false
		}
		op, ok := filter["op"].(string)
		if !ok {
			return false
		}
		arg, ok := filter["arg"].(map[string]any)
		if !ok || !savedViewFilterArgShapeExact(op, arg) {
			return false
		}
	}
	if groupBy, present := query["group_by"]; present && groupBy != nil {
		if _, ok := groupBy.(string); !ok {
			return false
		}
	}
	return true
}

func savedViewFilterArgShapeExact(op string, arg map[string]any) bool {
	switch op {
	case "eq":
		value, hasValue := arg["value"]
		values, hasValues := arg["values"]
		if hasValue == hasValues || len(arg) != 1 {
			return false
		}
		if hasValue {
			return isSavedViewScalar(value, true)
		}
		return isSavedViewScalarArray(values)
	case "range":
		if len(arg) == 0 {
			return true
		}
		for key, value := range arg {
			switch key {
			case "gt", "gte", "lt", "lte":
				if !isSavedViewScalar(value, false) {
					return false
				}
			default:
				return false
			}
		}
		return true
	case "contains_any", "contains_all":
		values, ok := arg["values"]
		return ok && len(arg) == 1 && isSavedViewScalarArray(values)
	case "prefix":
		value, ok := arg["value"]
		if !ok || len(arg) != 1 {
			return false
		}
		_, ok = value.(string)
		return ok
	case "full_text":
		query, ok := arg["query"]
		if !ok || len(arg) != 1 {
			return false
		}
		_, ok = query.(string)
		return ok
	default:
		// The operator registry owns whether an operator is admitted. The row
		// shape still requires a non-empty object, and semantic validation will
		// reject an unknown operator without disclosing its contents.
		return len(arg) > 0
	}
}

func isSavedViewScalar(value any, allowNull bool) bool {
	if value == nil {
		return allowNull
	}
	switch value.(type) {
	case string, bool, json.Number:
		return true
	default:
		return false
	}
}

func isSavedViewScalarArray(value any) bool {
	values, ok := value.([]any)
	if !ok {
		return false
	}
	for _, item := range values {
		if !isSavedViewScalar(item, false) {
			return false
		}
	}
	return true
}

func savedViewLayoutShapeExact(layout map[string]any) bool {
	// As with queries, {} is a closed-object canonicality failure.
	if len(layout) == 0 {
		return true
	}
	members := []string{"layout_schema_id", "column_order", "hidden_field_keys", "column_widths", "frozen_through_field_key"}
	if !hasExactMembers(layout, members, nil) {
		return false
	}
	if _, ok := layout["layout_schema_id"].(string); !ok {
		return false
	}
	if !isStringArray(layout["column_order"]) || !isStringArray(layout["hidden_field_keys"]) {
		return false
	}
	widths, ok := layout["column_widths"].([]any)
	if !ok {
		return false
	}
	for _, rawWidth := range widths {
		width, ok := rawWidth.(map[string]any)
		if !ok || !hasExactMembers(width, []string{"field_key", "width_px"}, nil) {
			return false
		}
		if _, ok := width["field_key"].(string); !ok {
			return false
		}
		if _, ok := width["width_px"].(json.Number); !ok {
			return false
		}
	}
	return true
}

func isStringArray(value any) bool {
	values, ok := value.([]any)
	if !ok {
		return false
	}
	for _, item := range values {
		if _, ok := item.(string); !ok {
			return false
		}
	}
	return true
}

func hasExactMembers(value map[string]any, required []string, optional []string) bool {
	if len(value) < len(required) || len(value) > len(required)+len(optional) {
		return false
	}
	admitted := make(map[string]struct{}, len(required)+len(optional))
	for _, member := range required {
		if _, ok := value[member]; !ok {
			return false
		}
		admitted[member] = struct{}{}
	}
	for _, member := range optional {
		admitted[member] = struct{}{}
	}
	for member := range value {
		if _, ok := admitted[member]; !ok {
			return false
		}
	}
	return true
}

func parseCanonicalPortableTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Nanosecond()%1_000 != 0 || parsed.UTC().Format(time.RFC3339Nano) != value {
		return time.Time{}, fmt.Errorf("noncanonical timestamp")
	}
	return parsed.UTC(), nil
}

func savedViewInvariantFailure(invariantID string) error {
	return &savedViewInvariantError{InvariantID: invariantID}
}

func applyPreparedSavedViewImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedSavedViewImport,
	importContext savedViewImportContext,
) error {
	queries := sqlc.New(tx)
	return prepared.walk(ctx, importContext, func(row preparedSavedViewRow) error {
		affected, err := queries.ImportSavedView(ctx, sqlc.ImportSavedViewParams{
			SavedViewID:      pgUUID(row.SavedViewID),
			IncidentID:       pgUUID(row.IncidentID),
			ViewSchemaID:     row.ViewSchemaID,
			Scope:            string(row.Scope),
			DisplayName:      row.DisplayName,
			QueryJson:        append([]byte(nil), row.QueryJSON...),
			LayoutJson:       append([]byte(nil), row.LayoutJSON...),
			OwnerUserID:      optionalPGUUID(row.RuntimeOwnerID),
			CreatedAt:        pgtype.Timestamptz{Time: row.CreatedAt, Valid: true},
			UpdatedAt:        pgtype.Timestamptz{Time: row.UpdatedAt, Valid: true},
			SavedViewVersion: row.SavedViewVersion,
		})
		if err != nil {
			return classifySavedViewApplyError(err)
		}
		if affected != 1 {
			return savedViewInvariantFailure("saved_views.identity_scope_legal")
		}
		// The owner attribution was staged while preparing; the core flushes it
		// in step 9. The capability check keeps its rank.
		if row.PortableOwnerID != nil && importContext.Attributions == nil {
			return savedViewInvariantFailure("saved_views.owner_tuple_legal")
		}
		return nil
	})
}

func optionalPGUUID(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return pgUUID(*value)
}

func classifySavedViewApplyError(err error) error {
	var postgresFailure *pgconn.PgError
	if !errors.As(err, &postgresFailure) {
		return err
	}
	switch postgresFailure.ConstraintName {
	case "saved_views_pkey", "saved_views_incident_id_fkey", "saved_views_scope_check":
		return savedViewInvariantFailure("saved_views.identity_scope_legal")
	case "saved_views_owner_user_id_fkey", "saved_views_owner_scope_ck":
		return savedViewInvariantFailure("saved_views.owner_tuple_legal")
	case "saved_views_version_positive_ck", "saved_views_timestamp_order_ck":
		return savedViewInvariantFailure("saved_views.version_timestamps_legal")
	default:
		return err
	}
}

func validatePreparedSavedViewImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedSavedViewImport,
	importContext savedViewImportContext,
) error {

	rows, err := tx.Query(ctx, `
SELECT saved_view_id,
       incident_id,
       view_schema_id,
       scope,
       display_name,
       query_json,
       layout_json,
       owner_user_id,
       created_at,
       updated_at,
       saved_view_version
  FROM saved_views
 WHERE incident_id = $1
 ORDER BY saved_view_id
`, importContext.IncidentID)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = prepared.walk(ctx, importContext, func(expectedRow preparedSavedViewRow) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return savedViewInvariantFailure("saved_views.row_shape_exact")
		}

		var (
			savedViewID  pgtype.UUID
			incidentID   pgtype.UUID
			viewSchemaID string
			scopeValue   string
			displayName  string
			queryJSON    []byte
			layoutJSON   []byte
			ownerUserID  pgtype.UUID
			createdAt    pgtype.Timestamptz
			updatedAt    pgtype.Timestamptz
			version      int64
		)
		if err := rows.Scan(
			&savedViewID,
			&incidentID,
			&viewSchemaID,
			&scopeValue,
			&displayName,
			&queryJSON,
			&layoutJSON,
			&ownerUserID,
			&createdAt,
			&updatedAt,
			&version,
		); err != nil {
			return err
		}
		parsedSavedViewID, err := uuidFromPG(savedViewID)
		if err != nil {
			return savedViewInvariantFailure("saved_views.identity_scope_legal")
		}
		if parsedSavedViewID != expectedRow.SavedViewID {
			return savedViewInvariantFailure("saved_views.identity_scope_legal")
		}

		parsedIncidentID, err := uuidFromPG(incidentID)
		if err != nil ||
			parsedIncidentID != expectedRow.IncidentID ||
			parsedIncidentID != importContext.IncidentID ||
			viewSchemaID != expectedRow.ViewSchemaID ||
			scopeValue != string(expectedRow.Scope) {
			return savedViewInvariantFailure("saved_views.identity_scope_legal")
		}
		if !pgUUIDMatches(ownerUserID, expectedRow.RuntimeOwnerID) {
			return savedViewInvariantFailure("saved_views.owner_tuple_legal")
		}
		if displayName != expectedRow.DisplayName {
			return savedViewInvariantFailure("saved_views.display_name_normalized")
		}
		queryEqual, err := jsonStructurallyEqual(queryJSON, expectedRow.QueryJSON)
		if err != nil || !queryEqual {
			return savedViewInvariantFailure("saved_views.query_layout_legal")
		}
		layoutEqual, err := jsonStructurallyEqual(layoutJSON, expectedRow.LayoutJSON)
		if err != nil || !layoutEqual {
			return savedViewInvariantFailure("saved_views.query_layout_legal")
		}
		if !createdAt.Valid || !updatedAt.Valid ||
			!createdAt.Time.Equal(expectedRow.CreatedAt) ||
			!updatedAt.Time.Equal(expectedRow.UpdatedAt) ||
			version != expectedRow.SavedViewVersion {
			return savedViewInvariantFailure("saved_views.version_timestamps_legal")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return savedViewInvariantFailure("saved_views.row_shape_exact")
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	if err := validateSavedViewAttributions(prepared, importContext); err != nil {
		return err
	}
	// Saved Views neither consumes nor mutates Reference Packs. Exact equality
	// with the prepared query/layout values above proves that optional-pack
	// degradation did not substitute a schema, delete a filter, or repair a row.
	return nil
}

func pgUUIDMatches(actual pgtype.UUID, expected *uuid.UUID) bool {
	if expected == nil {
		return !actual.Valid
	}
	parsed, err := uuidFromPG(actual)
	return err == nil && parsed == *expected
}

// validateSavedViewAttributions keeps the capability's rank; the owner
// attribution itself is flushed by the core in step 9 and checked there by
// counts.
func validateSavedViewAttributions(prepared preparedSavedViewImport, importContext savedViewImportContext) error {
	if importContext.Attributions == nil && prepared.owners != 0 {
		return savedViewInvariantFailure("saved_views.owner_tuple_legal")
	}
	return nil
}
