package sourcestate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
		Export: func(ctx context.Context, exportContext sourceport.ExportContext) error {
			return exportFiles(ctx, exportContext, stateManifest)
		},
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareImport(ctx, descriptor, stateManifest, bundle, importContext)
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
	tables := recoverystate.AuthoritativeTables(stateManifest.tableNames()...)
	tables = append(tables, recoverystate.TransientStateTables("links.invalidate_bundle_validation.v1", "links_bundle_validation_catalog", "links_bundle_validation_links", "links_bundle_validation_tags")...)
	return recoverystate.NewContribution("module.links", tables), nil
}

func exportFiles(
	ctx context.Context,
	input sourceport.ExportContext,
	stateManifest manifest,
) error {
	for _, path := range stateManifest.pathSpecs() {
		queryText, err := sourceQuery(path.kind, true)
		if err != nil {
			return err
		}
		switch path.kind {
		case pathRecordLinks:
			err = sourceport.ExportAttributedRows(ctx, input, sourceport.AttributedRows{Path: path.logicalPath, Table: "record_links", Identity: "record_link_id", ActorColumns: []string{"owner_user_id", "created_by_user_id", "deleted_by_user_id"}, Query: queryText})
		case pathRecordTags:
			err = sourceport.ExportAttributedRows(ctx, input, sourceport.AttributedRows{Path: path.logicalPath, Table: "record_tags", Identity: "record_tag_id", ActorColumns: []string{"created_by_user_id", "deleted_by_user_id"}, Query: queryText})
		default:
			err = incidentportability.ExportNDJSONTo(ctx, input.Query, input.IncidentID, path.logicalPath, queryText, input.Members)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func sourceQuery(kind pathKind, paged bool) (string, error) {
	pagePredicate, pageLimit := "", ""
	if paged {
		pageLimit = " LIMIT 128"
	}
	switch kind {
	case pathRecordLinks:
		if paged {
			pagePredicate = ` AND ($2::uuid IS NULL OR record_link_id>$2)`
		}
		return `SELECT to_jsonb(t) FROM record_links t WHERE incident_id = $1` + pagePredicate + ` ORDER BY record_link_id` + pageLimit, nil
	case pathTagCatalog:
		return `SELECT jsonb_build_object('tag_name', tag_name, 'normalized_tag_name', normalized_tag_name) FROM (SELECT DISTINCT tag_name COLLATE "C" AS tag_name, normalized_tag_name COLLATE "C" AS normalized_tag_name FROM record_tags WHERE incident_id = $1 ORDER BY normalized_tag_name, tag_name) tags`, nil
	case pathRecordTags:
		if paged {
			pagePredicate = ` AND ($2::uuid IS NULL OR record_tag_id>$2)`
		}
		return `SELECT to_jsonb(t) FROM record_tags t WHERE incident_id = $1` + pagePredicate + ` ORDER BY record_tag_id` + pageLimit, nil
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
		if path.kind == pathTagCatalog {
			continue
		}
		spec, err := fixedImportSpec(path)
		if err != nil {
			return err
		}
		if err := incidentportability.ImportFixedMemberNDJSON(
			ctx, tx, spec, prepared.members, importContext.ActorUserID, importContext.Attributions,
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
		value.bundleVersion == importContext.BundleVersion && value.bundleVersion == 6 &&
		value.operationID != "" && value.operationID == importContext.OperationID
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

// Transaction-local relations let PostgreSQL perform collection-wide joins and
// uniqueness checks without retaining source collections in the process heap.
// Deferred owner guards forbid committing scratch rows. Runtime uses only DML.
func validateDatabasePreconditions(ctx context.Context, tx pgx.Tx, prepared preparedImport, importContext sourceport.ImportContext) (string, error) {
	for _, item := range []struct {
		kind      pathKind
		statement string
	}{
		{pathRecordLinks, `INSERT INTO public.links_bundle_validation_links SELECT * FROM jsonb_populate_record(NULL::public.record_links, $1::jsonb)`},
		{pathRecordTags, `INSERT INTO public.links_bundle_validation_tags SELECT * FROM jsonb_populate_record(NULL::public.record_tags, $1::jsonb)`},
		{pathTagCatalog, `INSERT INTO public.links_bundle_validation_catalog SELECT tag_name, normalized_tag_name FROM jsonb_to_record($1::jsonb) AS t(tag_name text, normalized_tag_name text)`},
	} {
		if err := prepared.walk(ctx, item.kind, func(row map[string]any) error {
			raw, err := json.Marshal(row)
			if err != nil {
				return err
			}
			_, err = tx.Exec(ctx, item.statement, raw)
			return err
		}); err != nil {
			return "", err
		}
	}
	checks := []struct{ invariant, statement string }{
		{"links_tags.source_identity_admitted", `SELECT
   EXISTS (SELECT 1 FROM public.links_bundle_validation_links s JOIN public.record_links t USING (record_link_id)) OR
   EXISTS (SELECT 1 FROM public.links_bundle_validation_tags s JOIN public.record_tags t USING (record_tag_id))`},
		{"links_tags.tag_catalog_exact", `SELECT EXISTS (
   SELECT 1 FROM public.links_bundle_validation_catalog c FULL JOIN
    (SELECT DISTINCT tag_name COLLATE "C" AS tag_name, normalized_tag_name COLLATE "C" AS normalized_tag_name FROM public.links_bundle_validation_tags) t
   USING (tag_name, normalized_tag_name) WHERE c.tag_name IS NULL OR t.tag_name IS NULL)`},
		{"links_tags.link_unique", `WITH links AS (
   SELECT incident_id, src_record_id, dst_record_id, link_type, field_key FROM public.links_bundle_validation_links WHERE deleted_at IS NULL
   UNION ALL SELECT incident_id, src_record_id, dst_record_id, link_type, field_key FROM public.record_links WHERE deleted_at IS NULL AND incident_id = $1
  ), tags AS (
   SELECT incident_id, record_id, normalized_tag_name FROM public.links_bundle_validation_tags WHERE deleted_at IS NULL
   UNION ALL SELECT incident_id, record_id, normalized_tag_name FROM public.record_tags WHERE deleted_at IS NULL AND incident_id = $1
  ) SELECT EXISTS (SELECT 1 FROM links GROUP BY incident_id, src_record_id, dst_record_id, link_type, field_key HAVING count(*) > 1)
   OR EXISTS (SELECT 1 FROM links WHERE link_type = 'supersedes' GROUP BY incident_id, dst_record_id HAVING count(*) > 1)
   OR EXISTS (SELECT 1 FROM tags GROUP BY incident_id, record_id, normalized_tag_name HAVING count(*) > 1)`},
		{"links_tags.endpoints_same_incident", `SELECT
   EXISTS (SELECT 1 FROM public.links_bundle_validation_links l
    LEFT JOIN public.records src ON src.record_id = l.src_record_id AND src.incident_id = $1
    LEFT JOIN public.records dst ON dst.record_id = l.dst_record_id AND dst.incident_id = $1
    WHERE l.incident_id <> $1 OR src.record_id IS NULL OR dst.record_id IS NULL)
   OR EXISTS (SELECT 1 FROM public.links_bundle_validation_tags t
    LEFT JOIN public.records r ON r.record_id = t.record_id AND r.incident_id = $1
    WHERE t.incident_id <> $1 OR r.record_id IS NULL)`},
	}
	for index, check := range checks {
		var invalid bool
		var args []any
		if index >= 2 {
			args = []any{importContext.IncidentID}
		}
		if err := tx.QueryRow(ctx, check.statement, args...).Scan(&invalid); err != nil {
			return "", err
		}
		if invalid {
			return check.invariant, nil
		}
	}
	for _, statement := range []string{`DELETE FROM public.links_bundle_validation_links WHERE transaction_id=pg_current_xact_id()`, `DELETE FROM public.links_bundle_validation_tags WHERE transaction_id=pg_current_xact_id()`, `DELETE FROM public.links_bundle_validation_catalog WHERE transaction_id=pg_current_xact_id()`} {
		if _, err := tx.Exec(ctx, statement); err != nil {
			return "", err
		}
	}
	return "", nil
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

func validatePersistedState(ctx context.Context, tx pgx.Tx, prepared preparedImport, importContext sourceport.ImportContext, stateManifest manifest) (string, error) {
	identities, values := map[pathKind]bool{}, map[pathKind]bool{}
	for _, path := range stateManifest.pathSpecs() {
		query, err := sourceQuery(path.kind, false)
		if err != nil {
			return "", err
		}
		sameIDs, sameValues, err := comparePersistedRows(ctx, tx, prepared, importContext, query, path)
		if err != nil {
			return "", err
		}
		identities[path.kind], values[path.kind] = sameIDs, sameValues
	}
	if !identities[pathRecordLinks] || !identities[pathRecordTags] {
		return "links_tags.source_identity_admitted", nil
	}
	if !values[pathRecordLinks] || !values[pathRecordTags] {
		return "links_tags.link_tuple_legal", nil
	}
	if !values[pathTagCatalog] {
		return "links_tags.tag_catalog_exact", nil
	}
	if err := validateAttributions(ctx, prepared, importContext); err != nil {
		return "", err
	}
	return "", nil
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

// Merge ordered source and database streams, retaining at most one row each.
func comparePersistedRows(ctx context.Context, tx pgx.Tx, prepared preparedImport, importContext sourceport.ImportContext, query string, path pathSpec) (bool, bool, error) {
	rows, err := tx.Query(ctx, query, importContext.IncidentID)
	if err != nil {
		return false, false, err
	}
	defer rows.Close()
	sameIDs, sameValues := true, true
	err = prepared.walk(ctx, path.kind, func(source map[string]any) error {
		if !rows.Next() {
			sameIDs, sameValues = false, false
			return rows.Err()
		}
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return err
		}
		decoded, err := incidentportability.DecodeNDJSON(raw)
		if err != nil || len(decoded) != 1 {
			return errors.New("invalid persisted link source row")
		}
		for _, field := range path.stableIdentity {
			if source[field] != decoded[0][field] {
				sameIDs = false
			}
		}
		expected, err := incidentportability.CanonicalJSONString(runtimeAttributedRow(source, importContext.ActorUserID))
		if err != nil {
			return err
		}
		actual, err := incidentportability.CanonicalJSONString(decoded[0])
		if err != nil {
			return err
		}
		sameValues = sameValues && bytes.Equal(expected, actual)
		return nil
	})
	if err != nil {
		return false, false, err
	}
	if rows.Next() {
		sameIDs, sameValues = false, false
	}
	if err := rows.Err(); err != nil {
		return false, false, err
	}
	return sameIDs, sameValues, nil
}

func validateAttributions(ctx context.Context, prepared preparedImport, input sourceport.ImportContext) error {
	for _, family := range []struct{ path, table, identity string }{
		{"data/record_links.ndjson", "record_links", "record_link_id"}, {"data/record_tags.ndjson", "record_tags", "record_tag_id"},
	} {
		if err := incidentportability.ValidateTopLevelAttributions(ctx, prepared.members, family.path, family.table, family.identity, input.Attributions); err != nil {
			stateManifest, manifestErr := loadManifest()
			if manifestErr != nil {
				return manifestErr
			}
			return stateManifest.descriptor().DeclaredFailure("links_tags.source_identity_admitted")
		}
	}
	return nil
}
