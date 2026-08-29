package exportprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("reporting export provider: not found")

const (
	ProviderOutputSchemaID          = "cartulary.reporting_owner_provider_output.v1"
	FieldFactSchemaID               = "cartulary.reporting_owner_field_fact.v1"
	RecordFactSchemaID              = "cartulary.reporting_owner_record_fact.v1"
	TimelineEventFactSchemaID       = "cartulary.reporting_owner_timeline_event_fact.v1"
	SubjectFactSchemaID             = "cartulary.reporting_owner_subject_fact.v1"
	SupportRefFactSchemaID          = "cartulary.reporting_owner_support_ref_fact.v1"
	DisclosurePartitionFactSchemaID = "cartulary.reporting_owner_disclosure_partition_fact.v1"
)

type IncidentSnapshot struct {
	ID           string
	Title        string
	Description  *string
	Status       string
	Severity     *string
	TLP          *string
	CurrentPhase *string
	Version      int64
}

type ProviderOutput struct {
	SchemaID                 string                    `json:"schema_id"`
	ProviderKey              string                    `json:"provider_key"`
	FieldFacts               []FieldFact               `json:"field_facts"`
	RecordFacts              []RecordFact              `json:"record_facts"`
	TimelineEventFacts       []TimelineEventFact       `json:"timeline_event_facts"`
	SubjectFacts             []SubjectFact             `json:"subject_facts"`
	SupportRefFacts          []SupportRefFact          `json:"support_ref_facts"`
	DisclosurePartitionFacts []DisclosurePartitionFact `json:"disclosure_partition_facts"`
}

// FieldProvider is the source-owner reporting boundary. Implementations load
// authoritative facts with the caller's transaction and never coordinate or
// commit that transaction.
type FieldProvider interface {
	ProviderKey() string
	CollectFactsTx(context.Context, pgx.Tx, uuid.UUID, map[string][]string) (ProviderOutput, error)
}

// SupportReferenceProvider supplies source-record-to-target-path relationships
// under the caller's transaction before immutable snapshot materialization.
type SupportReferenceProvider interface {
	ProviderKey() string
	CollectSupportRefsTx(context.Context, pgx.Tx, uuid.UUID) (map[string][]string, error)
}

// LogicalSupportTargetProvider contributes source-owned logical paths that
// Reporting uses when materializing support references. Implementations run
// inside the caller's transaction and never coordinate or commit it.
type LogicalSupportTargetProvider interface {
	ProviderKey() string
	CollectLogicalSupportTargetsTx(context.Context, pgx.Tx, uuid.UUID) (map[string]string, error)
}

type FieldFact struct {
	SchemaID                string `json:"schema_id"`
	Path                    string
	ContentClass            string
	SourceFamily            string
	Value                   any
	DisclosurePartitionRefs []string
	SupportRefs             []string
	RawBlobSource           bool
	OpaqueBinary            bool
	GeneratedPresentation   bool
}

type Field = FieldFact

type RecordFact struct {
	SchemaID                string   `json:"schema_id"`
	SourceFamily            string   `json:"source_family"`
	RecordID                string   `json:"record_id"`
	RecordType              string   `json:"record_type"`
	FieldPaths              []string `json:"field_paths"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs,omitempty"`
	SupportRefs             []string `json:"support_refs,omitempty"`
}

type TimelineEventFact struct {
	SchemaID                string   `json:"schema_id"`
	TimelineEventID         string   `json:"timeline_event_id"`
	SourceFamily            string   `json:"source_family"`
	FieldPaths              []string `json:"field_paths"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs,omitempty"`
	SupportRefs             []string `json:"support_refs,omitempty"`
}

type SubjectFact struct {
	SchemaID                string   `json:"schema_id"`
	StableSubjectRef        string   `json:"stable_subject_ref"`
	SubjectKind             string   `json:"subject_kind"`
	SourceFamily            string   `json:"source_family"`
	SourceRecordID          string   `json:"source_record_id"`
	FieldPaths              []string `json:"field_paths"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs,omitempty"`
}

type SupportRefFact struct {
	SchemaID         string `json:"schema_id"`
	SupportRefID     string `json:"support_ref_id"`
	SourcePath       string `json:"source_path"`
	TargetPath       string `json:"target_path"`
	SupportTargetRef string `json:"support_target_ref"`
}

type DisclosurePartitionFact struct {
	SchemaID       string `json:"schema_id"`
	PartitionRef   string `json:"partition_ref"`
	SourceFamily   string `json:"source_family"`
	SourceRecordID string `json:"source_record_id"`
	FieldPath      string `json:"field_path"`
}

type FieldQuery struct {
	Prefix                       string
	SQL                          string
	DisclosurePartitionRefPrefix string
}

func CollectQueryFieldsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string, queries []FieldQuery) ([]Field, error) {
	output, err := CollectQueryProviderOutputTx(ctx, tx, incidentID, "legacy", supportRefs, queries)
	if err != nil {
		return nil, err
	}
	return output.Fields(), nil
}

func CollectQueryProviderOutputTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, providerKey string, supportRefs map[string][]string, queries []FieldQuery) (ProviderOutput, error) {
	fields := []FieldFact{}
	for _, query := range queries {
		rows, err := tx.Query(ctx, query.SQL, incidentID)
		if err != nil {
			return ProviderOutput{}, err
		}
		for rows.Next() {
			var id string
			var sourceFamily string
			var contentClass string
			var raw []byte
			if err := rows.Scan(&id, &sourceFamily, &contentClass, &raw); err != nil {
				rows.Close()
				return ProviderOutput{}, err
			}
			var value any
			if err := json.Unmarshal(raw, &value); err != nil {
				rows.Close()
				return ProviderOutput{}, err
			}
			field := FieldFact{
				SchemaID:     FieldFactSchemaID,
				Path:         fmt.Sprintf("/%s/%s", query.Prefix, id),
				ContentClass: contentClass,
				SourceFamily: sourceFamily,
				Value:        value,
				SupportRefs:  CloneStrings(supportRefs[id]),
			}
			if query.DisclosurePartitionRefPrefix != "" {
				field.DisclosurePartitionRefs = []string{query.DisclosurePartitionRefPrefix + id}
			}
			fields = append(fields, field)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return ProviderOutput{}, err
		}
		rows.Close()
	}
	return NewProviderOutput(providerKey, fields)
}

// NewProviderOutput constructs and validates the complete owner output from
// typed field facts. It supports owners whose derived facts arrive through a
// typed reader instead of an owner-authored SQL query contract.
func NewProviderOutput(providerKey string, fields []FieldFact) (ProviderOutput, error) {
	fields = append([]FieldFact(nil), fields...)
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})
	output := ProviderOutput{
		SchemaID:                 ProviderOutputSchemaID,
		ProviderKey:              providerKey,
		FieldFacts:               fields,
		RecordFacts:              recordFactsFromFields(fields),
		TimelineEventFacts:       timelineEventFactsFromFields(fields),
		SubjectFacts:             subjectFactsFromFields(fields),
		SupportRefFacts:          supportRefFactsFromFields(fields),
		DisclosurePartitionFacts: disclosurePartitionFactsFromFields(fields),
	}
	return output, output.Validate()
}

func (output ProviderOutput) Validate() error {
	if output.SchemaID != ProviderOutputSchemaID {
		return fmt.Errorf("invalid provider output schema %q", output.SchemaID)
	}
	if output.ProviderKey == "" {
		return errors.New("provider output omits provider key")
	}
	seenPaths := map[string]struct{}{}
	for _, fact := range output.FieldFacts {
		if fact.SchemaID != FieldFactSchemaID {
			return fmt.Errorf("invalid field fact schema %q", fact.SchemaID)
		}
		if fact.Path == "" || fact.ContentClass == "" || fact.SourceFamily == "" {
			return fmt.Errorf("field fact is incomplete: path=%q content_class=%q source_family=%q", fact.Path, fact.ContentClass, fact.SourceFamily)
		}
		if _, exists := seenPaths[fact.Path]; exists {
			return fmt.Errorf("duplicate field fact path %s", fact.Path)
		}
		seenPaths[fact.Path] = struct{}{}
	}
	for _, fact := range output.RecordFacts {
		if fact.SchemaID != RecordFactSchemaID || fact.SourceFamily == "" || fact.RecordID == "" || fact.RecordType == "" {
			return fmt.Errorf("record fact is incomplete: %+v", fact)
		}
	}
	for _, fact := range output.TimelineEventFacts {
		if fact.SchemaID != TimelineEventFactSchemaID || fact.TimelineEventID == "" || fact.SourceFamily == "" {
			return fmt.Errorf("timeline event fact is incomplete: %+v", fact)
		}
	}
	for _, fact := range output.SubjectFacts {
		if fact.SchemaID != SubjectFactSchemaID || fact.StableSubjectRef == "" || fact.SubjectKind == "" || fact.SourceRecordID == "" {
			return fmt.Errorf("subject fact is incomplete: %+v", fact)
		}
	}
	for _, fact := range output.SupportRefFacts {
		if fact.SchemaID != SupportRefFactSchemaID || fact.SupportRefID == "" || fact.SourcePath == "" || fact.TargetPath == "" {
			return fmt.Errorf("support ref fact is incomplete: %+v", fact)
		}
	}
	for _, fact := range output.DisclosurePartitionFacts {
		if fact.SchemaID != DisclosurePartitionFactSchemaID || fact.PartitionRef == "" || fact.FieldPath == "" {
			return fmt.Errorf("disclosure partition fact is incomplete: %+v", fact)
		}
	}
	return nil
}

func (output ProviderOutput) Fields() []Field {
	if len(output.FieldFacts) == 0 {
		return nil
	}
	fields := make([]Field, len(output.FieldFacts))
	copy(fields, output.FieldFacts)
	return fields
}

func CloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func recordFactsFromFields(fields []FieldFact) []RecordFact {
	byKey := map[string]RecordFact{}
	for _, field := range fields {
		recordID := recordIDFromPath(field.Path)
		if recordID == "" {
			continue
		}
		key := field.SourceFamily + ":" + recordID
		fact := byKey[key]
		if fact.SchemaID == "" {
			fact = RecordFact{
				SchemaID:     RecordFactSchemaID,
				SourceFamily: field.SourceFamily,
				RecordID:     recordID,
				RecordType:   field.SourceFamily,
			}
		}
		fact.FieldPaths = appendUnique(fact.FieldPaths, field.Path)
		fact.DisclosurePartitionRefs = appendUnique(fact.DisclosurePartitionRefs, field.DisclosurePartitionRefs...)
		fact.SupportRefs = appendUnique(fact.SupportRefs, field.SupportRefs...)
		byKey[key] = fact
	}
	out := make([]RecordFact, 0, len(byKey))
	for _, fact := range byKey {
		sort.Strings(fact.FieldPaths)
		sort.Strings(fact.DisclosurePartitionRefs)
		sort.Strings(fact.SupportRefs)
		out = append(out, fact)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SourceFamily == out[j].SourceFamily {
			return out[i].RecordID < out[j].RecordID
		}
		return out[i].SourceFamily < out[j].SourceFamily
	})
	return out
}

func timelineEventFactsFromFields(fields []FieldFact) []TimelineEventFact {
	byID := map[string]TimelineEventFact{}
	for _, field := range fields {
		if field.SourceFamily != "timeline_event" {
			continue
		}
		recordID := recordIDFromPath(field.Path)
		if recordID == "" {
			continue
		}
		fact := byID[recordID]
		if fact.SchemaID == "" {
			fact = TimelineEventFact{
				SchemaID:        TimelineEventFactSchemaID,
				TimelineEventID: recordID,
				SourceFamily:    field.SourceFamily,
			}
		}
		fact.FieldPaths = appendUnique(fact.FieldPaths, field.Path)
		fact.DisclosurePartitionRefs = appendUnique(fact.DisclosurePartitionRefs, field.DisclosurePartitionRefs...)
		fact.SupportRefs = appendUnique(fact.SupportRefs, field.SupportRefs...)
		byID[recordID] = fact
	}
	out := make([]TimelineEventFact, 0, len(byID))
	for _, fact := range byID {
		sort.Strings(fact.FieldPaths)
		sort.Strings(fact.DisclosurePartitionRefs)
		sort.Strings(fact.SupportRefs)
		out = append(out, fact)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].TimelineEventID < out[j].TimelineEventID
	})
	return out
}

func subjectFactsFromFields(fields []FieldFact) []SubjectFact {
	byRef := map[string]SubjectFact{}
	for _, field := range fields {
		kind := subjectKindForSourceFamily(field.SourceFamily)
		if kind == "" {
			continue
		}
		recordID := recordIDFromPath(field.Path)
		if recordID == "" {
			continue
		}
		stableRef := kind + ":" + recordID
		fact := byRef[stableRef]
		if fact.SchemaID == "" {
			fact = SubjectFact{
				SchemaID:         SubjectFactSchemaID,
				StableSubjectRef: stableRef,
				SubjectKind:      kind,
				SourceFamily:     field.SourceFamily,
				SourceRecordID:   recordID,
			}
		}
		fact.FieldPaths = appendUnique(fact.FieldPaths, field.Path)
		fact.DisclosurePartitionRefs = appendUnique(fact.DisclosurePartitionRefs, field.DisclosurePartitionRefs...)
		byRef[stableRef] = fact
	}
	out := make([]SubjectFact, 0, len(byRef))
	for _, fact := range byRef {
		sort.Strings(fact.FieldPaths)
		sort.Strings(fact.DisclosurePartitionRefs)
		out = append(out, fact)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StableSubjectRef < out[j].StableSubjectRef
	})
	return out
}

func supportRefFactsFromFields(fields []FieldFact) []SupportRefFact {
	seen := map[string]SupportRefFact{}
	for _, field := range fields {
		for _, ref := range field.SupportRefs {
			key := field.Path + "\x00" + ref
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = SupportRefFact{
				SchemaID:         SupportRefFactSchemaID,
				SupportRefID:     supportRefID(ref),
				SourcePath:       field.Path,
				TargetPath:       ref,
				SupportTargetRef: ref,
			}
		}
	}
	out := make([]SupportRefFact, 0, len(seen))
	for _, fact := range seen {
		out = append(out, fact)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SourcePath == out[j].SourcePath {
			return out[i].TargetPath < out[j].TargetPath
		}
		return out[i].SourcePath < out[j].SourcePath
	})
	return out
}

func disclosurePartitionFactsFromFields(fields []FieldFact) []DisclosurePartitionFact {
	seen := map[string]DisclosurePartitionFact{}
	for _, field := range fields {
		recordID := recordIDFromPath(field.Path)
		for _, ref := range field.DisclosurePartitionRefs {
			key := field.Path + "\x00" + ref
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = DisclosurePartitionFact{
				SchemaID:       DisclosurePartitionFactSchemaID,
				PartitionRef:   ref,
				SourceFamily:   field.SourceFamily,
				SourceRecordID: recordID,
				FieldPath:      field.Path,
			}
		}
	}
	out := make([]DisclosurePartitionFact, 0, len(seen))
	for _, fact := range seen {
		out = append(out, fact)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PartitionRef == out[j].PartitionRef {
			return out[i].FieldPath < out[j].FieldPath
		}
		return out[i].PartitionRef < out[j].PartitionRef
	})
	return out
}

func recordIDFromPath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 || idx == len(path)-1 {
		return ""
	}
	return path[idx+1:]
}

func subjectKindForSourceFamily(sourceFamily string) string {
	switch sourceFamily {
	case "host":
		return "host"
	case "identity":
		return "identity"
	case "party":
		return "party"
	case "entity_mention":
		return "mention"
	default:
		return ""
	}
}

func supportRefID(ref string) string {
	trimmed := strings.Trim(ref, "/")
	if trimmed == "" {
		return "support-root"
	}
	replacer := strings.NewReplacer("/", "-", ":", "-", " ", "-")
	return "support-" + replacer.Replace(trimmed)
}

func appendUnique(values []string, additions ...string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		seen[value] = struct{}{}
	}
	for _, value := range additions {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}
