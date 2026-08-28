package sourcestate

import (
	"bytes"
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func NewSourcePort() (sourceport.Port, error) {
	stateManifest, err := loadManifest()
	if err != nil {
		return nil, err
	}
	descriptor := stateManifest.descriptor()
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export: func(ctx context.Context, exportContext sourceport.ExportContext) ([]incidentportability.File, error) {
			return exportFiles(ctx, exportContext.Query, exportContext.IncidentID, stateManifest)
		},
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareImport(descriptor, stateManifest, bundle, importContext)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return applyPreparedTx(ctx, tx, value, importContext, stateManifest, descriptor)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			return validatePreparedTx(ctx, tx, value, importContext, stateManifest, descriptor)
		},
		ValidateContract: func() error {
			_, err := validateManifest(authoredManifestInput())
			return err
		},
	}), nil
}

func RecoveryStateContribution() (recoverystate.Contribution, error) {
	stateManifest, err := loadManifest()
	if err != nil {
		return recoverystate.Contribution{}, err
	}
	return recoverystate.NewContribution(
		"module.links",
		recoverystate.AuthoritativeTables(stateManifest.tableNames()...),
	), nil
}

func exportFiles(
	ctx context.Context,
	query incidentportability.Queryer,
	incidentID uuid.UUID,
	stateManifest manifest,
) ([]incidentportability.File, error) {
	files := make([]incidentportability.File, 0, len(stateManifest.paths))
	for _, path := range stateManifest.pathSpecs() {
		queryText, err := exportQuery(path.kind)
		if err != nil {
			return nil, err
		}
		file, err := incidentportability.ExportNDJSON(ctx, query, incidentID, path.logicalPath, queryText)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func exportQuery(kind pathKind) (string, error) {
	switch kind {
	case pathRecordLinks:
		return `SELECT to_jsonb(t) FROM record_links t WHERE incident_id = $1 ORDER BY record_link_id`, nil
	case pathTagCatalog:
		return `SELECT jsonb_build_object('tag_name', tag_name, 'normalized_tag_name', normalized_tag_name) FROM (SELECT DISTINCT tag_name, normalized_tag_name FROM record_tags WHERE incident_id = $1 ORDER BY normalized_tag_name, tag_name) tags`, nil
	case pathRecordTags:
		return `SELECT to_jsonb(t) FROM record_tags t WHERE incident_id = $1 ORDER BY record_id, record_tag_id`, nil
	default:
		return "", manifestError("unknown export path kind")
	}
}

func applyPreparedTx(
	ctx context.Context,
	tx pgx.Tx,
	value any,
	importContext sourceport.ImportContext,
	stateManifest manifest,
	descriptor sourceport.Descriptor,
) error {
	prepared, ok := value.(preparedImport)
	if !ok {
		return &sourceStateError{reason: "wrong prepared value"}
	}
	if tx == nil || !prepared.matches(importContext) || importContext.ActorUserID == uuid.Nil || importContext.Attributions == nil {
		return &sourceStateError{reason: "prepared value binding mismatch"}
	}
	if invariantID, err := validateDatabasePreconditions(ctx, tx, prepared, importContext); err != nil {
		return err
	} else if invariantID != "" {
		return descriptor.DeclaredFailure(invariantID)
	}
	for _, path := range stateManifest.pathSpecs() {
		rows, ok := prepared.rows(path.kind)
		if !ok {
			continue
		}
		spec, err := fixedImportSpec(path)
		if err != nil {
			return err
		}
		if err := incidentportability.ImportFixedRows(
			ctx, tx, spec, rows, importContext.ActorUserID, importContext.Attributions,
		); err != nil {
			if _, fixed := incidentportability.FixedImportFailurePath(err); fixed {
				return err
			}
			return descriptor.DeclaredFailure("links_tags.link_tuple_legal")
		}
	}
	return nil
}

func validatePreparedTx(
	ctx context.Context,
	tx pgx.Tx,
	value any,
	importContext sourceport.ImportContext,
	stateManifest manifest,
	descriptor sourceport.Descriptor,
) error {
	prepared, ok := value.(preparedImport)
	if !ok {
		return &sourceStateError{reason: "wrong prepared value"}
	}
	if tx == nil || !prepared.matches(importContext) || importContext.Attributions == nil {
		return &sourceStateError{reason: "prepared value binding mismatch"}
	}
	if invariantID, err := validatePersistedState(ctx, tx, prepared, importContext, stateManifest); err != nil {
		return err
	} else if invariantID != "" {
		return descriptor.DeclaredFailure(invariantID)
	}
	if invariantID, err := validateDatabaseScope(ctx, tx, importContext.IncidentID); err != nil {
		return err
	} else if invariantID != "" {
		return descriptor.DeclaredFailure(invariantID)
	}
	return nil
}

func (value preparedImport) matches(importContext sourceport.ImportContext) bool {
	return value.incidentID != uuid.Nil && value.incidentID == importContext.IncidentID &&
		value.bundleVersion == importContext.BundleVersion && value.bundleVersion == 3 &&
		value.operationID != "" && value.operationID == importContext.OperationID
}

func (value preparedImport) rows(kind pathKind) ([]map[string]any, bool) {
	switch kind {
	case pathRecordLinks:
		rows := make([]map[string]any, 0, len(value.links))
		for _, row := range value.links {
			rows = append(rows, cloneRow(row.raw))
		}
		return rows, true
	case pathTagCatalog:
		return nil, false
	case pathRecordTags:
		rows := make([]map[string]any, 0, len(value.recordTags))
		for _, row := range value.recordTags {
			rows = append(rows, cloneRow(row.raw))
		}
		return rows, true
	default:
		return nil, false
	}
}

func fixedImportSpec(path pathSpec) (incidentportability.FixedImportSpec, error) {
	var table string
	var insertSQL string
	switch path.kind {
	case pathRecordLinks:
		table = "record_links"
		insertSQL = `INSERT INTO record_links SELECT * FROM jsonb_populate_record(NULL::record_links, $1::jsonb)`
	case pathRecordTags:
		table = "record_tags"
		insertSQL = `INSERT INTO record_tags SELECT * FROM jsonb_populate_record(NULL::record_tags, $1::jsonb)`
	default:
		return incidentportability.FixedImportSpec{}, manifestError("path has no import relation")
	}
	return incidentportability.FixedImportSpec{
		LogicalBundlePath: path.logicalPath,
		AttributionTable:  table,
		StableIdentity:    slices.Clone(path.stableIdentity),
		RequiredColumns:   slices.Clone(path.requiredColumns),
		AllowedColumns:    slices.Clone(path.allowedColumns),
		InsertSQL:         insertSQL,
	}, nil
}

func validateDatabasePreconditions(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedImport,
	importContext sourceport.ImportContext,
) (string, error) {
	linkIDs := make([]uuid.UUID, 0, len(prepared.links))
	tagIDs := make([]uuid.UUID, 0, len(prepared.recordTags))
	for _, row := range prepared.links {
		linkIDs = append(linkIDs, uuid.MustParse(row.targetKey))
	}
	for _, row := range prepared.recordTags {
		tagIDs = append(tagIDs, uuid.MustParse(row.targetKey))
	}
	var duplicateIdentity bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (SELECT 1 FROM record_links WHERE record_link_id = ANY($1::uuid[]))
    OR EXISTS (SELECT 1 FROM record_tags WHERE record_tag_id = ANY($2::uuid[]))
`, linkIDs, tagIDs).Scan(&duplicateIdentity); err != nil {
		return "", err
	}
	if duplicateIdentity {
		return "links_tags.source_identity_admitted", nil
	}
	if duplicate, err := activeDatabaseCollision(ctx, tx, prepared); err != nil {
		return "", err
	} else if duplicate {
		return "links_tags.link_unique", nil
	}
	return validatePreparedDatabaseScope(ctx, tx, prepared, importContext.IncidentID)
}

func validatePreparedDatabaseScope(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedImport,
	incidentID uuid.UUID,
) (string, error) {
	for _, row := range prepared.links {
		var endpointsPresent bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (SELECT 1 FROM records WHERE incident_id = $1 AND record_id = $2)
   AND EXISTS (SELECT 1 FROM records WHERE incident_id = $1 AND record_id = $3)
`, incidentID, row.source, row.target).Scan(&endpointsPresent); err != nil {
			return "", err
		}
		if !endpointsPresent {
			return "links_tags.endpoints_same_incident", nil
		}
	}
	for _, row := range prepared.recordTags {
		var endpointPresent bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (SELECT 1 FROM records WHERE incident_id = $1 AND record_id = $2)
`, incidentID, row.recordID).Scan(&endpointPresent); err != nil {
			return "", err
		}
		if !endpointPresent {
			return "links_tags.endpoints_same_incident", nil
		}
	}
	return "", nil
}

func activeDatabaseCollision(ctx context.Context, tx pgx.Tx, prepared preparedImport) (bool, error) {
	for _, row := range prepared.links {
		var duplicate bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM record_links
     WHERE incident_id = $1
       AND src_record_id = $2
       AND dst_record_id = $3
       AND link_type = $4
       AND field_key IS NOT DISTINCT FROM $5::text
       AND deleted_at IS NULL
)
OR ($4::text = 'supersedes' AND EXISTS (
    SELECT 1
      FROM record_links
     WHERE incident_id = $1
       AND dst_record_id = $3
       AND link_type = 'supersedes'
       AND deleted_at IS NULL
))
`, row.incident, row.source, row.target, row.linkType, nullableFieldKey(row.fieldKey)).Scan(&duplicate); err != nil {
			return false, err
		}
		if duplicate {
			return true, nil
		}
	}
	for _, row := range prepared.recordTags {
		var duplicate bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM record_tags
     WHERE incident_id = $1
       AND record_id = $2
       AND normalized_tag_name = $3
       AND deleted_at IS NULL
)
`, row.incident, row.recordID, row.normalized).Scan(&duplicate); err != nil {
			return false, err
		}
		if duplicate {
			return true, nil
		}
	}
	return false, nil
}

func nullableFieldKey(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func validateDatabaseScope(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) (string, error) {
	var invalid bool
	if err := tx.QueryRow(ctx, `
SELECT
    EXISTS (
        SELECT 1 FROM record_links link
        LEFT JOIN records source
          ON source.incident_id = link.incident_id AND source.record_id = link.src_record_id
        LEFT JOIN records target
          ON target.incident_id = link.incident_id AND target.record_id = link.dst_record_id
        WHERE link.incident_id = $1
          AND (source.record_id IS NULL OR target.record_id IS NULL)
    )
    OR EXISTS (
        SELECT 1 FROM record_tags tag
        LEFT JOIN records record
          ON record.incident_id = tag.incident_id AND record.record_id = tag.record_id
        WHERE tag.incident_id = $1
          AND record.record_id IS NULL
    )
`, incidentID).Scan(&invalid); err != nil {
		return "", err
	}
	if invalid {
		return "links_tags.endpoints_same_incident", nil
	}
	return "", nil
}

func validatePersistedState(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedImport,
	importContext sourceport.ImportContext,
	stateManifest manifest,
) (string, error) {
	files, err := exportFiles(ctx, tx, importContext.IncidentID, stateManifest)
	if err != nil {
		return "", err
	}
	actual := make(map[string][]map[string]any, len(files))
	for _, file := range files {
		rows, err := decodeStrictNDJSON(file.Payload)
		if err != nil {
			return "", err
		}
		actual[file.Path] = rows
	}
	expectedLinks := expectedRuntimeRows(prepared.links, importContext.ActorUserID)
	expectedTags := expectedRuntimeTagRows(prepared.recordTags, importContext.ActorUserID)
	if !rowIdentitySetsEqual(expectedLinks, actual["data/record_links.ndjson"], "record_link_id") ||
		!rowIdentitySetsEqual(expectedTags, actual["data/record_tags.ndjson"], "record_tag_id") {
		return "links_tags.source_identity_admitted", nil
	}
	if !canonicalRowSetsEqual(expectedLinks, actual["data/record_links.ndjson"], "record_link_id") ||
		!canonicalRowSetsEqual(expectedTags, actual["data/record_tags.ndjson"], "record_tag_id") {
		return "links_tags.link_tuple_legal", nil
	}
	expectedCatalog := make([]map[string]any, 0, len(prepared.tagCatalog))
	for _, row := range prepared.tagCatalog {
		expectedCatalog = append(expectedCatalog, cloneRow(row.raw))
	}
	if !canonicalRowSetsEqual(expectedCatalog, actual["data/tags.ndjson"], "normalized_tag_name", "tag_name") {
		return "links_tags.tag_catalog_exact", nil
	}
	if !attributionsEqual(prepared, importContext) {
		return "links_tags.source_identity_admitted", nil
	}
	return "", nil
}

func expectedRuntimeRows(rows []preparedLinkRow, actorUserID uuid.UUID) []map[string]any {
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		result = append(result, runtimeAttributedRow(row.raw, actorUserID))
	}
	return result
}

func expectedRuntimeTagRows(rows []preparedRecordTagRow, actorUserID uuid.UUID) []map[string]any {
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		result = append(result, runtimeAttributedRow(row.raw, actorUserID))
	}
	return result
}

func runtimeAttributedRow(row map[string]any, actorUserID uuid.UUID) map[string]any {
	result := cloneRow(row)
	for key, value := range result {
		if strings.HasSuffix(key, "_user_id") && value != nil {
			result[key] = actorUserID.String()
		}
	}
	return result
}

func rowIdentitySetsEqual(left []map[string]any, right []map[string]any, identity ...string) bool {
	return identitySet(left, identity...).equal(identitySet(right, identity...))
}

type stringSet map[string]struct{}

func identitySet(rows []map[string]any, identity ...string) stringSet {
	result := make(stringSet, len(rows))
	for _, row := range rows {
		parts := make([]string, 0, len(identity))
		for _, field := range identity {
			parts = append(parts, incidentportability.StringFromAny(row[field]))
		}
		result[strings.Join(parts, "\x00")] = struct{}{}
	}
	return result
}

func (value stringSet) equal(other stringSet) bool {
	if len(value) != len(other) {
		return false
	}
	for key := range value {
		if _, ok := other[key]; !ok {
			return false
		}
	}
	return true
}

func canonicalRowSetsEqual(left []map[string]any, right []map[string]any, identity ...string) bool {
	leftRows, err := canonicalRowsByIdentity(left, identity...)
	if err != nil {
		return false
	}
	rightRows, err := canonicalRowsByIdentity(right, identity...)
	if err != nil || len(leftRows) != len(rightRows) {
		return false
	}
	for key, expected := range leftRows {
		if !bytes.Equal(expected, rightRows[key]) {
			return false
		}
	}
	return true
}

func canonicalRowsByIdentity(rows []map[string]any, identity ...string) (map[string][]byte, error) {
	result := make(map[string][]byte, len(rows))
	for _, row := range rows {
		parts := make([]string, 0, len(identity))
		for _, field := range identity {
			parts = append(parts, incidentportability.StringFromAny(row[field]))
		}
		encoded, err := incidentportability.CanonicalJSONString(row)
		if err != nil {
			return nil, err
		}
		result[strings.Join(parts, "\x00")] = encoded
	}
	return result, nil
}

func attributionsEqual(prepared preparedImport, importContext sourceport.ImportContext) bool {
	expected := map[string]string{}
	add := func(table string, rowID string, row map[string]any) {
		for column, value := range row {
			if strings.HasSuffix(column, "_user_id") && value != nil {
				expected[table+"\x00"+rowID+"\x00"+column] = incidentportability.StringFromAny(value)
			}
		}
	}
	for _, row := range prepared.links {
		add("record_links", row.targetKey, row.raw)
	}
	for _, row := range prepared.recordTags {
		add("record_tags", row.targetKey, row.raw)
	}
	found := map[string]string{}
	relevantCount := 0
	for _, attribution := range importContext.Attributions.ImportedAttributions() {
		if attribution.SourceTable != "record_links" && attribution.SourceTable != "record_tags" {
			continue
		}
		relevantCount++
		found[attribution.SourceTable+"\x00"+attribution.SourceRowID+"\x00"+attribution.SourceColumn] = attribution.SourceActorID
	}
	if len(expected) != len(found) || relevantCount != len(expected) {
		return false
	}
	for key, value := range expected {
		if found[key] != value {
			return false
		}
	}
	return true
}
