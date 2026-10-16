package exportprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("reporting export provider: not found")

const (
	ProviderOutputSchemaID          = "cartulary.reporting_owner_provider_output.v4"
	FieldFactSchemaID               = "cartulary.reporting_owner_field_fact.v1"
	RecordFactSchemaID              = "cartulary.reporting_owner_record_fact.v1"
	TimelineEventFactSchemaID       = "cartulary.reporting_owner_timeline_event_fact.v2"
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
	SourceFacts              SourceFacts               `json:"source_facts"`
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
	Sort                    TimelineSortFact `json:"sort"`
	SchemaID                string           `json:"schema_id"`
	TimelineEventID         string           `json:"timeline_event_id"`
	SourceFamily            string           `json:"source_family"`
	FieldPaths              []string         `json:"field_paths"`
	DisclosurePartitionRefs []string         `json:"disclosure_partition_refs,omitempty"`
	SupportRefs             []string         `json:"support_refs,omitempty"`
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

func (output ProviderOutput) Validate() error {
	_, err := output.validatedRecordBindings()
	return err
}

func (output ProviderOutput) validatedRecordBindings() (map[string]RecordFact, error) {
	if output.SchemaID != ProviderOutputSchemaID {
		return nil, fmt.Errorf("invalid provider output schema %q", output.SchemaID)
	}
	if err := output.SourceFacts.Validate(); err != nil {
		return nil, err
	}
	if len(output.SourceFacts.Parties) > 0 && output.ProviderKey != "parties" || len(output.SourceFacts.Assignments) > 0 && output.ProviderKey != "links" || len(output.SourceFacts.DeletedRecords) > 0 && output.ProviderKey != "records" {
		return nil, errors.New("source facts supplied by an unowned provider")
	}
	if output.ProviderKey == "" {
		return nil, errors.New("provider output omits provider key")
	}
	seenPaths := map[string]struct{}{}
	for _, fact := range output.FieldFacts {
		if fact.SchemaID != FieldFactSchemaID {
			return nil, fmt.Errorf("invalid field fact schema %q", fact.SchemaID)
		}
		if fact.Path == "" || fact.ContentClass == "" || fact.SourceFamily == "" {
			return nil, fmt.Errorf("field fact is incomplete: path=%q content_class=%q source_family=%q", fact.Path, fact.ContentClass, fact.SourceFamily)
		}
		if _, exists := seenPaths[fact.Path]; exists {
			return nil, fmt.Errorf("duplicate field fact path %s", fact.Path)
		}
		seenPaths[fact.Path] = struct{}{}
	}
	for _, fact := range output.RecordFacts {
		if fact.SchemaID != RecordFactSchemaID || fact.SourceFamily == "" || fact.RecordID == "" || fact.RecordType == "" {
			return nil, fmt.Errorf("record fact is incomplete: %+v", fact)
		}
	}
	for _, fact := range output.TimelineEventFacts {
		if fact.Sort.Validate() != nil || fact.Sort.RecordID != fact.TimelineEventID || fact.SchemaID != TimelineEventFactSchemaID || fact.TimelineEventID == "" || fact.SourceFamily == "" {
			return nil, fmt.Errorf("timeline event fact is incomplete: %+v", fact)
		}
	}
	for _, fact := range output.SubjectFacts {
		if fact.SchemaID != SubjectFactSchemaID || fact.StableSubjectRef == "" || fact.SubjectKind == "" || fact.SourceRecordID == "" {
			return nil, fmt.Errorf("subject fact is incomplete: %+v", fact)
		}
	}
	for _, fact := range output.SupportRefFacts {
		if fact.SchemaID != SupportRefFactSchemaID || fact.SupportRefID == "" || fact.SourcePath == "" || fact.TargetPath == "" {
			return nil, fmt.Errorf("support ref fact is incomplete: %+v", fact)
		}
	}
	for _, fact := range output.DisclosurePartitionFacts {
		if fact.SchemaID != DisclosurePartitionFactSchemaID || fact.PartitionRef == "" || fact.FieldPath == "" {
			return nil, fmt.Errorf("disclosure partition fact is incomplete: %+v", fact)
		}
	}
	return output.recordBindings()
}

func CloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}
