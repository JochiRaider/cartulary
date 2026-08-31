package imports

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var sourceRectA1Pattern = regexp.MustCompile(`^([A-Z]+)([1-9][0-9]*):([A-Z]+)([1-9][0-9]*)$`)

var errInvalidPersistedSourceRegion = errors.New("imports: invalid persisted source region")

type sourceRectangle struct {
	left   int
	top    int
	right  int
	bottom int
}

type selectedSourceRegion struct {
	unitID uuid.UUID
	sheet  string
	rect   sourceRectangle
}

type unitSelectionState struct {
	status             string
	mappingFingerprint *string
	approvedMapping    []byte
	locatorKind        string
	locator            string
	sourceRectA1       string
	blockingColumns    []int32
}

func unitSelectionStateTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, unitID uuid.UUID) (unitSelectionState, error) {
	var state unitSelectionState
	if err := tx.QueryRow(ctx, `
SELECT unit_status, mapping_fingerprint, approved_mapping_json, locator_kind, locator,
       source_rect_a1, blocking_source_column_ordinals
  FROM import_units
 WHERE import_session_id = $1
   AND import_unit_id = $2
 FOR UPDATE
`, sessionID, unitID).Scan(
		&state.status,
		&state.mappingFingerprint,
		&state.approvedMapping,
		&state.locatorKind,
		&state.locator,
		&state.sourceRectA1,
		&state.blockingColumns,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return unitSelectionState{}, errNotFound
		}
		return unitSelectionState{}, err
	}
	return state, nil
}

func (state unitSelectionState) statusAfterSelection() string {
	if state.mappingFingerprint != nil && *state.mappingFingerprint != "" && len(state.approvedMapping) != 0 {
		if mappingUsesBlockingColumn(state.approvedMapping, state.blockingColumns) {
			return "mapped"
		}
		return "ready"
	}
	return "selected"
}

func mappingUsesBlockingColumn(approvedMappingJSON []byte, blockingColumns []int32) bool {
	if len(approvedMappingJSON) == 0 || len(blockingColumns) == 0 {
		return false
	}
	var mapping approvedMapping
	if err := json.Unmarshal(approvedMappingJSON, &mapping); err != nil {
		return true
	}
	blocked := make(map[int]struct{}, len(blockingColumns))
	for _, ordinal := range blockingColumns {
		blocked[int(ordinal)] = struct{}{}
	}
	for _, column := range mapping.SourceColumns {
		if column.FieldKey == nil {
			continue
		}
		if _, blockedColumn := blocked[column.SourceColumnOrdinal]; blockedColumn {
			return true
		}
	}
	return false
}

func validateProposedSelectionDoesNotOverlapTx(
	ctx context.Context,
	tx pgx.Tx,
	sessionID uuid.UUID,
	candidateID uuid.UUID,
	candidateState unitSelectionState,
) error {
	candidate, err := sourceRegion(candidateID, candidateState.locatorKind, candidateState.locator, candidateState.sourceRectA1)
	if err != nil {
		return err
	}
	if candidate == nil {
		return nil
	}

	rows, err := tx.Query(ctx, `
SELECT u.import_unit_id, u.locator_kind, u.locator, u.source_rect_a1
  FROM import_sessions s
  JOIN import_units u
    ON u.import_session_id = s.import_session_id
   AND u.import_unit_id = ANY(s.selected_unit_ids)
 WHERE s.import_session_id = $1
   AND u.import_unit_id <> $2
 ORDER BY u.discovery_sequence, u.import_unit_id
 FOR UPDATE OF u
`, sessionID, candidateID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var unitID uuid.UUID
		var locatorKind string
		var locator string
		var sourceRectA1 string
		if err := rows.Scan(&unitID, &locatorKind, &locator, &sourceRectA1); err != nil {
			return err
		}
		other, err := sourceRegion(unitID, locatorKind, locator, sourceRectA1)
		if err != nil {
			return err
		}
		if sourceRegionsOverlap(candidate, other) {
			return importApplyBlockedError("overlapping_units")
		}
	}
	return rows.Err()
}

func validateSelectedUnitsDoNotOverlap(regions []*selectedSourceRegion) error {
	for leftIndex := 0; leftIndex < len(regions); leftIndex++ {
		for rightIndex := leftIndex + 1; rightIndex < len(regions); rightIndex++ {
			if sourceRegionsOverlap(regions[leftIndex], regions[rightIndex]) {
				return importApplyBlockedError("overlapping_units")
			}
		}
	}
	return nil
}

func sourceRegionsOverlap(left *selectedSourceRegion, right *selectedSourceRegion) bool {
	if left == nil || right == nil || left.sheet != right.sheet {
		return false
	}
	return left.rect.left <= right.rect.right &&
		right.rect.left <= left.rect.right &&
		left.rect.top <= right.rect.bottom &&
		right.rect.top <= left.rect.bottom
}

func sourceRegion(unitID uuid.UUID, locatorKind string, locator string, sourceRectA1 string) (*selectedSourceRegion, error) {
	if locatorKind == "csv_file" {
		return nil, nil
	}
	if !strings.HasPrefix(locatorKind, "xlsx_") && locatorKind != "operator_region" {
		return nil, fmt.Errorf("%w: unsupported locator kind", errInvalidPersistedSourceRegion)
	}
	var locatorObject struct {
		SheetName string `json:"sheet_name"`
	}
	if err := json.Unmarshal([]byte(locator), &locatorObject); err != nil || locatorObject.SheetName == "" {
		return nil, fmt.Errorf("%w: invalid locator", errInvalidPersistedSourceRegion)
	}
	rect, err := parseSourceRectangle(sourceRectA1)
	if err != nil {
		return nil, err
	}
	return &selectedSourceRegion{
		unitID: unitID,
		sheet:  locatorObject.SheetName,
		rect:   rect,
	}, nil
}

func parseSourceRectangle(value string) (sourceRectangle, error) {
	match := sourceRectA1Pattern.FindStringSubmatch(value)
	if match == nil {
		return sourceRectangle{}, fmt.Errorf("%w: invalid rectangle", errInvalidPersistedSourceRegion)
	}
	left, err := sourceColumnNumber(match[1])
	if err != nil {
		return sourceRectangle{}, err
	}
	top, err := strconv.Atoi(match[2])
	if err != nil {
		return sourceRectangle{}, fmt.Errorf("%w: invalid row", errInvalidPersistedSourceRegion)
	}
	right, err := sourceColumnNumber(match[3])
	if err != nil {
		return sourceRectangle{}, err
	}
	bottom, err := strconv.Atoi(match[4])
	if err != nil {
		return sourceRectangle{}, fmt.Errorf("%w: invalid row", errInvalidPersistedSourceRegion)
	}
	if left > right || top > bottom {
		return sourceRectangle{}, fmt.Errorf("%w: reversed rectangle", errInvalidPersistedSourceRegion)
	}
	return sourceRectangle{left: left, top: top, right: right, bottom: bottom}, nil
}

func sourceColumnNumber(value string) (int, error) {
	result := 0
	for _, character := range value {
		column := int(character-'A') + 1
		if result > (math.MaxInt-column)/26 {
			return 0, fmt.Errorf("%w: column overflow", errInvalidPersistedSourceRegion)
		}
		result = result*26 + column
	}
	return result, nil
}
