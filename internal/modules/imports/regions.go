package imports

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

var ErrInvalidSourceRect = errors.New("imports: invalid source rectangle")

type RegionBase struct {
	Unit                map[string]any
	IncidentID          uuid.UUID
	SourceBytes         []byte
	SourceContentSHA256 string
}

type CreateOperatorRegionParams struct {
	ActorUserID                 uuid.UUID
	SessionID                   uuid.UUID
	BaseUnitID                  uuid.UUID
	Request                     RegionRequest
	Unit                        DiscoveredUnit
	ExpectedSourceContentSHA256 string
	Now                         time.Time
}

type CreateOperatorRegionResult struct {
	Unit     map[string]any
	Replayed bool
}

func (s *Service) handleRegion(
	w http.ResponseWriter,
	r *http.Request,
	principal httpauth.Principal,
	route importSessionRoute,
) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, incidentID, err := s.store.GetSession(r.Context(), route.SessionID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{
			Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{},
		})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(
		r.Context(),
		incidentID,
		principal.User.ID,
		"editor",
		"reviewer",
		"admin",
	); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeRegionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	base, err := s.store.GetRegionBase(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{
			Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{},
		})
		return
	}
	if errors.Is(err, ErrInvalidSourceRect) {
		writeAPIError(w, r, invalidImportRequest("source_rect", "invalid_source_rect"))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	requested := request.SourceRect.sourceRectangle()
	baseRect, rectErr := parseSourceRectangle(base.Unit["source_rect_a1"].(string))
	if rectErr != nil || !sourceRectangleContains(baseRect, requested) ||
		!operatorRegionWithinLimits(requested, s.limits) {
		writeAPIError(w, r, invalidImportRequest("source_rect", "invalid_source_rect"))
		return
	}
	locator, ok := base.Unit["locator"].(map[string]any)
	if !ok {
		writeAPIError(w, r, invalidImportRequest("source_rect", "invalid_source_rect"))
		return
	}
	sheetName, ok := locator["sheet_name"].(string)
	if !ok || strings.TrimSpace(sheetName) == "" {
		writeAPIError(w, r, invalidImportRequest("source_rect", "invalid_source_rect"))
		return
	}
	workbook, apiErr := indexXLSXWorkbook(base.SourceBytes, s.limits, s.archiveLimits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	sheet := workbook.sheetsByName[sheetName]
	decoded, apiErr := workbook.decodeRectangle(sheet, requested, s.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	locatorJSON, err := json.Marshal(map[string]any{
		"sheet_name": sheetName,
		"rect_a1":    sourceRectangleA1(requested),
	})
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	unit, apiErr := s.discoveredImportUnitAt(
		decoded.rows,
		"operator_region",
		string(locatorJSON),
		requested,
		decoded.warningCodes,
		decoded.blockingColumnOrdinals,
	)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.store.CreateOperatorRegion(r.Context(), CreateOperatorRegionParams{
		ActorUserID:                 principal.User.ID,
		SessionID:                   route.SessionID,
		BaseUnitID:                  route.UnitID,
		Request:                     request,
		Unit:                        unit,
		ExpectedSourceContentSHA256: base.SourceContentSHA256,
		Now:                         s.now(),
	})
	if errors.Is(err, ErrInvalidSourceRect) {
		writeAPIError(w, r, invalidImportRequest("source_rect", "invalid_source_rect"))
		return
	}
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, result.Unit)
}

func sourceRectangleContains(outer sourceRectangle, inner sourceRectangle) bool {
	return inner.left >= outer.left && inner.right <= outer.right &&
		inner.top >= outer.top && inner.bottom <= outer.bottom
}

func operatorRegionWithinLimits(rect sourceRectangle, limits Limits) bool {
	width := int64(rect.right - rect.left + 1)
	height := int64(rect.bottom - rect.top + 1)
	dataRows := height - 1
	if dataRows < 0 {
		dataRows = 0
	}
	return width > 0 &&
		width <= limits.MaxColumns &&
		dataRows <= limits.MaxRows &&
		(dataRows == 0 || width <= limits.MaxCells/dataRows)
}

func (s *Store) GetRegionBase(ctx context.Context, sessionID uuid.UUID, baseUnitID uuid.UUID) (RegionBase, error) {
	unit, incidentID, err := s.GetUnit(ctx, sessionID, baseUnitID)
	if err != nil {
		return RegionBase{}, err
	}
	if unit["locator_kind"] != "xlsx_used_range" {
		return RegionBase{}, ErrInvalidSourceRect
	}
	capability, err := s.SourceCapabilityForUnit(ctx, sessionID, baseUnitID)
	if err != nil {
		return RegionBase{}, err
	}
	_, sourceBytes, err := s.loadSourceStream(ctx, s.pool, capability.SourceStreamRef)
	if err != nil {
		return RegionBase{}, err
	}
	return RegionBase{
		Unit:                unit,
		IncidentID:          incidentID,
		SourceBytes:         sourceBytes,
		SourceContentSHA256: capability.SourceContentSHA256,
	}, nil
}

func (s *Store) CreateOperatorRegion(
	ctx context.Context,
	params CreateOperatorRegionParams,
) (CreateOperatorRegionResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey:    "imports.units.regions.create",
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.SessionID.String() + ":" + params.BaseUnitID.String(),
		ClientTxnID: params.Request.ClientTxnID,
	}
	sum := sha256.Sum256(params.Request.Normalized)
	requestHash := sum[:]
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CreateOperatorRegionResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return CreateOperatorRegionResult{}, authn.ErrClientTxnConflict
		}
		var unit map[string]any
		if err := json.Unmarshal(existing.ResponseJSON, &unit); err != nil {
			return CreateOperatorRegionResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return CreateOperatorRegionResult{}, err
		}
		return CreateOperatorRegionResult{Unit: unit, Replayed: true}, nil
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return CreateOperatorRegionResult{}, err
	}
	sessionStatus, incidentID, err := sessionStatusTx(ctx, tx, params.SessionID)
	if err != nil {
		return CreateOperatorRegionResult{}, err
	}
	switch sessionStatus {
	case "applying":
		return CreateOperatorRegionResult{}, importConflictError("session_applying")
	case "applied", "partially_applied", "failed", "canceled":
		return CreateOperatorRegionResult{}, importConflictError("session_terminal")
	}
	if _, err := s.incidentAccess.AuthorizeMutationTx(
		ctx,
		tx,
		incidentID,
		params.ActorUserID,
		"editor",
		"reviewer",
		"admin",
	); err != nil {
		return CreateOperatorRegionResult{}, err
	}
	var locatorKind string
	var sourceStreamRef *string
	var currentSourceSHA string
	if err := tx.QueryRow(ctx, `
SELECT u.locator_kind, u.source_stream_ref, streams.source_content_sha256
  FROM import_units u
  LEFT JOIN import_source_streams streams
    ON streams.source_stream_ref = u.source_stream_ref
 WHERE u.import_session_id = $1
   AND u.import_unit_id = $2
 FOR UPDATE OF u
`, params.SessionID, params.BaseUnitID).Scan(&locatorKind, &sourceStreamRef, &currentSourceSHA); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CreateOperatorRegionResult{}, ErrNotFound
		}
		return CreateOperatorRegionResult{}, err
	}
	if locatorKind != "xlsx_used_range" || sourceStreamRef == nil ||
		currentSourceSHA != params.ExpectedSourceContentSHA256 {
		return CreateOperatorRegionResult{}, ErrInvalidSourceRect
	}
	if existingUnit, findErr := scanUnitResource(tx.QueryRow(
		ctx,
		unitResourceSQL()+`
 WHERE import_session_id = $1
   AND base_import_unit_id = $2
   AND locator_kind = 'operator_region'
   AND source_rect_a1 = $3
 FOR UPDATE`,
		params.SessionID,
		params.BaseUnitID,
		params.Unit.SourceRectA1,
	)); findErr == nil {
		if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusCreated, existingUnit); err != nil {
			return CreateOperatorRegionResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return CreateOperatorRegionResult{}, err
		}
		return CreateOperatorRegionResult{Unit: existingUnit, Replayed: true}, nil
	} else if !errors.Is(findErr, pgx.ErrNoRows) {
		return CreateOperatorRegionResult{}, findErr
	}

	previewRows, err := json.Marshal(params.Unit.PreviewRows)
	if err != nil {
		return CreateOperatorRegionResult{}, err
	}
	sourceRows, err := json.Marshal(params.Unit.SourceRows)
	if err != nil {
		return CreateOperatorRegionResult{}, err
	}
	columns, err := json.Marshal(params.Unit.Columns)
	if err != nil {
		return CreateOperatorRegionResult{}, err
	}
	var regionSequence int
	var discoverySequence int
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(MAX(operator_region_sequence), 0) + 1,
       COALESCE(MAX(discovery_sequence), 0) + 1
  FROM import_units
 WHERE import_session_id = $1
`, params.SessionID).Scan(&regionSequence, &discoverySequence); err != nil {
		return CreateOperatorRegionResult{}, err
	}
	unitID := uuid.New()
	regionStreamRef := newImportSourceStreamRef()
	if _, err := tx.Exec(ctx, `
INSERT INTO import_units (
    import_unit_id, import_session_id, unit_status, locator_kind, locator, source_rect_a1,
    header_row_ref, data_start_row_ref, inferred_row_count, inferred_column_count,
    warning_codes, blocking_source_column_ordinals, columns_json, source_rows_json,
    preview_rows_json, source_stream_ref, discovery_sequence, base_import_unit_id,
    operator_region_sequence, created_at, updated_at
)
VALUES (
    $1, $2, 'discovered', 'operator_region', $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $18
)
`, unitID, params.SessionID, params.Unit.Locator, params.Unit.SourceRectA1,
		params.Unit.HeaderRowRef, params.Unit.DataStartRowRef, params.Unit.InferredRowCount,
		params.Unit.InferredColumnCount, params.Unit.WarningCodes, params.Unit.BlockingColumns,
		columns, sourceRows, previewRows, regionStreamRef, discoverySequence, params.BaseUnitID,
		regionSequence, params.Now.UTC()); err != nil {
		return CreateOperatorRegionResult{}, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO import_source_streams (
    source_stream_ref, import_session_id, import_unit_id, source_content_sha256,
    source_media_type, source_byte_size, source_bytes, created_at
)
SELECT $1, $2, $3, source_content_sha256, source_media_type, source_byte_size, source_bytes, $4
  FROM import_source_streams
 WHERE source_stream_ref = $5
`, regionStreamRef, params.SessionID, unitID, params.Now.UTC(), *sourceStreamRef); err != nil {
		return CreateOperatorRegionResult{}, err
	}
	unit, err := scanUnitResource(tx.QueryRow(
		ctx,
		unitResourceSQL()+` WHERE import_session_id = $1 AND import_unit_id = $2`,
		params.SessionID,
		unitID,
	))
	if err != nil {
		return CreateOperatorRegionResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusCreated, unit); err != nil {
		return CreateOperatorRegionResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateOperatorRegionResult{}, err
	}
	return CreateOperatorRegionResult{Unit: unit}, nil
}
