package sourcestate

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/google/uuid"
)

type preparedLinkRow struct {
	raw                      map[string]any
	incident, source, target uuid.UUID
	linkType, fieldKey       string
	deleted                  bool
	targetKey                string
}
type preparedTagCatalogRow struct {
	raw                 map[string]any
	tagName, normalized string
}
type preparedRecordTagRow struct {
	raw                 map[string]any
	incident, recordID  uuid.UUID
	tagName, normalized string
	deleted             bool
	targetKey           string
}

// preparedImport holds the staged source streams and their historical
// attribution (Core 01 REQ-01-641); the final transaction reads no bundle
// member.
type preparedImport struct {
	incidentID    uuid.UUID
	bundleVersion int
	operationID   string
	links         incidentportability.PreparedRowStream
	catalog       incidentportability.PreparedRowStream
	tags          incidentportability.PreparedRowStream
	attribution   incidentportability.AttributionStream
	// unadmittedActor records an attributed source actor the operation's
	// catalog does not admit; Apply reports it at the write rank.
	unadmittedActor bool
	staged          bool
}
type sourceStateError struct{ reason string }

func (err *sourceStateError) Error() string { return "links: invalid prepared source state" }

func prepareImport(ctx context.Context, descriptor sourceport.Descriptor, stateManifest manifest, bundle incidentportability.MemberReader, importContext sourceport.ImportContext) (preparedImport, error) {
	if importContext.IncidentID == uuid.Nil {
		return preparedImport{}, descriptor.DeclaredFailure("links_tags.source_identity_admitted")
	}
	members, err := sourceport.PrepareMembers(ctx, descriptor, bundle, importContext.BundleVersion)
	if err != nil {
		var malformed *incidentportability.StrictNDJSONDecodeError
		if errors.As(err, &malformed) {
			return preparedImport{}, &incidentportability.MalformedPayloadError{Err: err}
		}
		return preparedImport{}, err
	}
	result := preparedImport{incidentID: importContext.IncidentID, bundleVersion: importContext.BundleVersion, operationID: importContext.OperationID}
	// Separate bounded passes preserve invariant precedence across every source row.
	for _, invariant := range []string{"links_tags.source_identity_admitted", "links_tags.link_tuple_legal", "links_tags.deletion_tuple_legal", "links_tags.tag_normalized"} {
		for _, path := range stateManifest.pathSpecs() {
			err := walkMembers(ctx, members, path.kind, func(raw map[string]any) error {
				if !rowAdmitted(raw, path, invariant) {
					return descriptor.DeclaredFailure(invariant)
				}
				return nil
			})
			if err != nil {
				return preparedImport{}, err
			}
		}
	}
	if err := result.stage(ctx, members, importContext, stateManifest); err != nil {
		return preparedImport{}, err
	}
	return result, nil
}

// stage writes the admitted source streams and the historical attribution of
// every top-level actor column; the core flushes the attribution in step 9.
func (value *preparedImport) stage(ctx context.Context, members *incidentportability.BoundMembers, importContext sourceport.ImportContext, stateManifest manifest) error {
	store := importContext.Preparation
	if store == nil {
		return &sourceStateError{reason: "preparation staging is required"}
	}
	paths := stateManifest.pathSpecs()
	for _, path := range paths {
		staged, err := incidentportability.StagePreparedRows(ctx, store, strings.TrimSuffix(strings.TrimPrefix(path.logicalPath, "data/"), ".ndjson")+".rows", incidentportability.RawRowCodec{}, func(emit func(map[string]any) error) error {
			return walkMembers(ctx, members, path.kind, emit)
		})
		if err != nil {
			return err
		}
		stream, err := value.stream(path.kind)
		if err != nil {
			return err
		}
		*stream = staged
	}
	admitted := func(actorID string) bool { _, ok := importContext.Actors.Lookup(actorID); return ok }
	attribution, err := incidentportability.StageAttributions(ctx, store, "attribution.rows", admitted, func(emit func(incidentportability.AttributionRow) error) error {
		for _, path := range paths {
			table, attributed := attributedTable(path.kind)
			if !attributed {
				continue
			}
			stream, err := value.stream(path.kind)
			if err != nil {
				return err
			}
			if err := incidentportability.ReadPreparedRows(ctx, store, *stream, incidentportability.RawRowCodec{}, func(row map[string]any) error {
				rowID := incidentportability.SourceRowID(row, path.stableIdentity)
				columns := make([]string, 0, len(row))
				for column := range row {
					columns = append(columns, column)
				}
				slices.Sort(columns)
				for _, column := range columns {
					if !strings.HasSuffix(column, "_user_id") || row[column] == nil {
						continue
					}
					actorID := incidentportability.StringFromAny(row[column])
					if strings.TrimSpace(actorID) == "" || rowID == "" {
						continue
					}
					if !admitted(actorID) {
						value.unadmittedActor = true
						continue
					}
					if err := emit(incidentportability.AttributionRow{SourceTable: table, SourceRowID: rowID, SourceColumn: column, SourceActorID: actorID}); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	value.attribution = attribution
	value.staged = true
	return nil
}

func (value *preparedImport) stream(kind pathKind) (*incidentportability.PreparedRowStream, error) {
	switch kind {
	case pathRecordLinks:
		return &value.links, nil
	case pathTagCatalog:
		return &value.catalog, nil
	case pathRecordTags:
		return &value.tags, nil
	default:
		return nil, manifestError("unknown path kind")
	}
}

// attributedTable names the authoritative relation whose actor columns carry
// historical attribution.
func attributedTable(kind pathKind) (string, bool) {
	switch kind {
	case pathRecordLinks:
		return "record_links", true
	case pathRecordTags:
		return "record_tags", true
	default:
		return "", false
	}
}

// walk visits one staged stream in order.
func (value preparedImport) walk(ctx context.Context, importContext sourceport.ImportContext, kind pathKind, visit func(map[string]any) error) error {
	if !value.staged || importContext.Preparation == nil {
		return &sourceStateError{reason: "prepared value is not staged"}
	}
	stream, err := value.stream(kind)
	if err != nil {
		return err
	}
	return incidentportability.ReadPreparedRows(ctx, importContext.Preparation, *stream, incidentportability.RawRowCodec{}, visit)
}

func walkMembers(ctx context.Context, members *incidentportability.BoundMembers, kind pathKind, visit func(map[string]any) error) error {
	var path string
	switch kind {
	case pathRecordLinks:
		path = "data/record_links.ndjson"
	case pathTagCatalog:
		path = "data/tags.ndjson"
	case pathRecordTags:
		path = "data/record_tags.ndjson"
	default:
		return manifestError("unknown path kind")
	}
	return incidentportability.WalkPath(ctx, members, path, visit)
}

func rowAdmitted(raw map[string]any, path pathSpec, invariant string) bool {
	switch invariant {
	case "links_tags.source_identity_admitted":
		if len(raw) != len(path.allowedColumns) {
			return false
		}
		for _, column := range path.allowedColumns {
			if _, ok := raw[column]; !ok {
				return false
			}
		}
		switch path.kind {
		case pathRecordLinks:
			_, ok := parseLink(raw)
			return ok
		case pathTagCatalog:
			_, ok := parseTag(raw)
			return ok
		case pathRecordTags:
			_, ok := parseRecordTag(raw)
			return ok
		}
	case "links_tags.link_tuple_legal":
		if path.kind != pathRecordLinks {
			return true
		}
		canonical, ok := canonicalCodecRow(raw, "decided_at", "created_at", "deleted_at")
		if !ok {
			return false
		}
		canonical["deleted_at"], canonical["deleted_by_user_id"] = nil, nil
		_, err := valuecodec.DecodeRecordLinkMutationValue(canonical)
		return err == nil
	case "links_tags.deletion_tuple_legal":
		switch path.kind {
		case pathRecordLinks:
			canonical, ok := canonicalCodecRow(raw, "decided_at", "created_at", "deleted_at")
			if !ok {
				return false
			}
			_, err := valuecodec.DecodeRecordLinkMutationValue(canonical)
			return err == nil
		case pathRecordTags:
			canonical, ok := canonicalCodecRow(raw, "created_at", "updated_at", "deleted_at")
			if !ok {
				return false
			}
			_, err := valuecodec.DecodeRecordTagMutationValue(canonical)
			return err == nil
		default:
			return true
		}
	case "links_tags.tag_normalized":
		if path.kind == pathRecordLinks {
			return true
		}
		name, nameOK := raw["tag_name"].(string)
		normalized, normOK := raw["normalized_tag_name"].(string)
		return nameOK && normOK && fieldnorm.IsCanonicalTagLabel(name, normalized)
	}
	return false
}
func parseLink(raw map[string]any) (preparedLinkRow, bool) {
	linkID, ok := canonicalUUIDValue(raw["record_link_id"])
	incident, incidentOK := canonicalUUIDValue(raw["incident_id"])
	source, sourceOK := canonicalUUIDValue(raw["src_record_id"])
	target, targetOK := canonicalUUIDValue(raw["dst_record_id"])
	if !ok || !incidentOK || !sourceOK || !targetOK || !canonicalRequiredUUID(raw["owner_user_id"]) ||
		!canonicalRequiredUUID(raw["created_by_user_id"]) || !canonicalNullableUUID(raw["deleted_by_user_id"]) {
		return preparedLinkRow{}, false
	}
	fieldKey := ""
	if raw["field_key"] != nil {
		var fieldOK bool
		fieldKey, fieldOK = raw["field_key"].(string)
		if !fieldOK || strings.TrimSpace(fieldKey) == "" || strings.TrimSpace(fieldKey) != fieldKey {
			return preparedLinkRow{}, false
		}
	}
	linkType, typeOK := raw["link_type"].(string)
	result := preparedLinkRow{
		raw: cloneRow(raw), incident: incident, source: source, target: target, linkType: linkType,
		fieldKey: fieldKey, deleted: raw["deleted_at"] != nil, targetKey: linkID.String(),
	}
	if !typeOK {
		return preparedLinkRow{}, false
	}
	return result, true
}

func parseTag(raw map[string]any) (preparedTagCatalogRow, bool) {
	tagName, tagOK := raw["tag_name"].(string)
	normalized, normalizedOK := raw["normalized_tag_name"].(string)
	if !tagOK || !normalizedOK {
		return preparedTagCatalogRow{}, false
	}
	result := preparedTagCatalogRow{raw: cloneRow(raw), tagName: tagName, normalized: normalized}
	return result, true
}

func parseRecordTag(raw map[string]any) (preparedRecordTagRow, bool) {
	tagID, ok := canonicalUUIDValue(raw["record_tag_id"])
	incident, incidentOK := canonicalUUIDValue(raw["incident_id"])
	recordID, recordOK := canonicalUUIDValue(raw["record_id"])
	tagName, tagOK := raw["tag_name"].(string)
	normalized, normalizedOK := raw["normalized_tag_name"].(string)
	if !ok || !incidentOK || !recordOK || !tagOK || !normalizedOK ||
		!canonicalRequiredUUID(raw["created_by_user_id"]) || !canonicalNullableUUID(raw["deleted_by_user_id"]) {
		return preparedRecordTagRow{}, false
	}
	result := preparedRecordTagRow{
		raw: cloneRow(raw), incident: incident, recordID: recordID, tagName: tagName, normalized: normalized,
		deleted: raw["deleted_at"] != nil, targetKey: tagID.String(),
	}
	return result, true
}

var portableTimestampPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,6})?\+00:00$`)

func canonicalCodecRow(raw map[string]any, timestampFields ...string) (map[string]any, bool) {
	result := cloneRow(raw)
	for _, field := range timestampFields {
		if result[field] == nil {
			continue
		}
		text, ok := result[field].(string)
		if !ok || !portableTimestampPattern.MatchString(text) {
			return nil, false
		}
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil || parsed.IsZero() {
			return nil, false
		}
		result[field] = parsed.UTC().Format(time.RFC3339Nano)
	}
	return result, true
}


func canonicalUUIDValue(raw any) (uuid.UUID, bool) {
	text, ok := raw.(string)
	if !ok {
		return uuid.Nil, false
	}
	value, err := uuid.Parse(text)
	return value, err == nil && value != uuid.Nil && value.String() == text
}

func canonicalRequiredUUID(raw any) bool {
	_, ok := canonicalUUIDValue(raw)
	return ok
}

func canonicalNullableUUID(raw any) bool {
	if raw == nil {
		return true
	}
	return canonicalRequiredUUID(raw)
}

func cloneRow(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
