package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"github.com/google/uuid"
)

const snapshotCaptureSchemaID = "cartulary.reporting_snapshot_capture.v1"

// A capture retains source-owner facts exactly once. Sections, Core record
// summaries and Timeline presentation are projections of these facts, not
// additional durable copies of every source field.
type snapshotCapture struct {
	SchemaID                     string                            `json:"schema_id"`
	SnapshotID                   string                            `json:"snapshot_id"`
	IncidentID                   string                            `json:"incident_id"`
	SnapshotAt                   time.Time                         `json:"snapshot_at"`
	SourceChangeSetHighWatermark string                            `json:"source_change_set_high_watermark"`
	DerivationVersion            string                            `json:"derivation_version"`
	ReferencePacks               reference_data.SetBinding         `json:"reference_packs"`
	Incident                     capturedIncident                  `json:"incident"`
	SourceFacts                  exportprovider.SourceFacts        `json:"source_facts"`
	Fields                       []capturedSourceField             `json:"fields"`
	Timeline                     []exportprovider.TimelineSortFact `json:"timeline_sort"`
}

type capturedIncident struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	Status       string  `json:"status"`
	Severity     *string `json:"severity"`
	TLP          *string `json:"tlp"`
	CurrentPhase *string `json:"current_phase"`
	Version      int64   `json:"version"`
}

type capturedSourceField struct {
	SourceRecordID          string   `json:"source_record_id"`
	SourceRecordType        string   `json:"source_record_type"`
	Path                    string   `json:"path"`
	ContentClass            string   `json:"content_class"`
	SourceFamily            string   `json:"source_family"`
	Value                   any      `json:"value"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs"`
	SupportRefs             []string `json:"support_refs"`
}

func freezeSnapshotCapture(incident IncidentMetadataSnapshot, snapshotID string, at time.Time, boundary string, binding reference_data.SetBinding, source SnapshotSource) ([]byte, string, error) {
	capture := snapshotCapture{SchemaID: snapshotCaptureSchemaID, SnapshotID: snapshotID, IncidentID: incident.ID, SnapshotAt: at, SourceChangeSetHighWatermark: boundary, DerivationVersion: DerivationVersion, ReferencePacks: binding,
		Incident: capturedIncident(incident), SourceFacts: source.SourceFacts.Clone(), Fields: []capturedSourceField{}, Timeline: []exportprovider.TimelineSortFact{}}
	for _, field := range source.Fields {
		if field.RawBlobSource || field.OpaqueBinary || field.GeneratedPresentation {
			return nil, "", errors.New("reporting: unsupported snapshot source")
		}
		capture.Fields = append(capture.Fields, capturedSourceField{field.SourceRecordID, field.SourceRecordType, field.Path, field.ContentClass, field.SourceFamily, field.Value, append([]string{}, field.DisclosurePartitionRefs...), append([]string{}, field.SupportRefs...)})
	}
	for _, fact := range source.Timeline {
		capture.Timeline = append(capture.Timeline, fact.Clone())
	}
	sort.Slice(capture.Fields, func(i, j int) bool { return capture.Fields[i].Path < capture.Fields[j].Path })
	sort.Slice(capture.Timeline, func(i, j int) bool { return capture.Timeline[i].RecordID < capture.Timeline[j].RecordID })
	if err := validateSnapshotCapture(capture); err != nil {
		return nil, "", err
	}
	raw, err := canonicalJSON(capture)
	if err != nil {
		return nil, "", err
	}
	return raw, reportingObjectDigest(snapshotCaptureSchemaID, raw), nil
}

func decodeSnapshotCapture(raw []byte, digest string) (snapshotCapture, error) {
	var capture snapshotCapture
	invalid := errors.New("reporting: invalid retained snapshot capture")
	canonical, err := canonicalReportingBytes(raw)
	if err != nil || reportingObjectDigest(snapshotCaptureSchemaID, canonical) != digest {
		return capture, invalid
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.UseNumber()
	if decoder.Decode(&capture) != nil {
		return capture, invalid
	}
	encoded, err := canonicalJSON(capture)
	if err != nil || !bytes.Equal(encoded, canonical) {
		return snapshotCapture{}, invalid
	}
	if err := validateSnapshotCapture(capture); err != nil {
		return snapshotCapture{}, err
	}
	return capture, nil
}

func validateSnapshotCapture(capture snapshotCapture) error {
	invalid := errors.New("reporting: invalid snapshot capture")
	snapshot, err := uuid.Parse(capture.SnapshotID)
	if err != nil || snapshot == uuid.Nil || snapshot.String() != capture.SnapshotID {
		return invalid
	}
	incident, err := uuid.Parse(capture.IncidentID)
	if err != nil || incident == uuid.Nil || incident.String() != capture.IncidentID || capture.Incident.ID != capture.IncidentID {
		return invalid
	}
	if _, err := normalizedReportingTimestamp(capture.SnapshotAt); err != nil {
		return invalid
	}
	if capture.SchemaID != snapshotCaptureSchemaID || capture.DerivationVersion != DerivationVersion || capture.SourceChangeSetHighWatermark == "" || capture.Incident.Version < 1 || capture.Fields == nil || capture.Timeline == nil {
		return invalid
	}
	if err := capture.SourceFacts.Validate(); err != nil {
		return invalid
	}
	metadata := SnapshotContent{IncidentID: capture.IncidentID, Records: []ReportingRecordSummary{}, SourceIndex: []ReportingSourceRef{}}
	owners := map[sourceOwner]string{}
	availableSupport := map[supportTarget]bool{{"source_record", "record:incident_metadata:" + capture.IncidentID}: true}
	timeline := map[string]bool{}
	previous := ""
	for _, field := range capture.Fields {
		if field.Path == "" || field.Path <= previous || field.SourceFamily == "" || field.SourceRecordID == "" || field.SourceRecordType == "" || field.DisclosurePartitionRefs == nil || field.SupportRefs == nil || validateSourceContentClass(field.ContentClass) != nil {
			return invalid
		}
		previous = field.Path
		if _, _, err := projectReportingSourceFields(ExportField{Path: field.Path, Value: field.Value, ContentClass: field.ContentClass, DisclosurePartitionRefs: field.DisclosurePartitionRefs, SupportRefs: field.SupportRefs}); err != nil {
			return invalid
		}
		availableSupport[supportTarget{"source_record", "record:" + field.SourceFamily + ":" + field.SourceRecordID}] = true
		switch field.SourceFamily {
		case "evidence":
			availableSupport[supportTarget{"evidence_item", "evidence:" + field.SourceRecordID}] = true
		case "timeline_event":
			availableSupport[supportTarget{"timeline_event", "timeline:" + field.SourceRecordID}] = true
		case "record_link":
			availableSupport[supportTarget{"relationship", "relationship:" + field.SourceRecordID}] = true
		}
		owner := sourceOwner{field.SourceFamily, field.SourceRecordID}
		if kind, exists := owners[owner]; exists {
			if kind != field.SourceRecordType {
				return invalid
			}
		} else {
			owners[owner] = field.SourceRecordType
			metadata.SourceIndex = append(metadata.SourceIndex, ReportingSourceRef{SourceFamily: field.SourceFamily, SourceRecordID: field.SourceRecordID})
			metadata.Records = append(metadata.Records, ReportingRecordSummary{RecordID: field.SourceRecordID, RecordType: field.SourceRecordType, SourceRecordRef: SourceRecordRef{SourceFamily: field.SourceFamily, SourceRecordID: field.SourceRecordID}})
		}
		if field.SourceFamily == "timeline_event" {
			if field.SourceRecordType != "timeline_event" {
				return invalid
			}
			timeline[field.SourceRecordID] = true
		}
	}
	previous = ""
	for _, fact := range capture.Timeline {
		if fact.Validate() != nil || fact.RecordID <= previous || !timeline[fact.RecordID] {
			return invalid
		}
		previous = fact.RecordID
		delete(timeline, fact.RecordID)
	}
	if len(timeline) != 0 {
		return invalid
	}
	for _, field := range capture.Fields {
		for _, ref := range field.SupportRefs {
			target, err := resolveSupportSelector(ref, capture.IncidentID)
			if err != nil || !availableSupport[target] {
				return invalid
			}
		}
	}
	return validateSnapshotSourceFacts(capture.SourceFacts, metadata)
}

func (capture snapshotCapture) source() (IncidentMetadataSnapshot, SnapshotSource) {
	incident := IncidentMetadataSnapshot(capture.Incident)
	source := SnapshotSource{SourceFacts: capture.SourceFacts.Clone(), Fields: []ExportField{}, Timeline: []exportprovider.TimelineSortFact{}}
	for _, field := range capture.Fields {
		source.Fields = append(source.Fields, ExportField{SourceRecordID: field.SourceRecordID, SourceRecordType: field.SourceRecordType, Path: field.Path, ContentClass: field.ContentClass, SourceFamily: field.SourceFamily, Value: field.Value, DisclosurePartitionRefs: append([]string{}, field.DisclosurePartitionRefs...), SupportRefs: append([]string{}, field.SupportRefs...)})
	}
	for _, fact := range capture.Timeline {
		source.Timeline = append(source.Timeline, fact.Clone())
	}
	return incident, source
}

func sameSnapshotCaptureBinding(binding reference_data.SetBinding, raw []byte, digest string) bool {
	capture, err := decodeSnapshotCapture(raw, digest)
	return err == nil && sameReportingValue(binding, capture.ReferencePacks)
}

// MaterializeSnapshotCapture derives a render projection from retained source
// evidence. It never consults current incident data or Reference Pack selection.
func MaterializeSnapshotCapture(ctx context.Context, raw []byte, digest string) (SnapshotModel, error) {
	capture, err := decodeSnapshotCapture(raw, digest)
	if err != nil {
		return SnapshotModel{}, err
	}
	return materializeSnapshotCapture(ctx, capture)
}

func materializeSnapshotCapture(ctx context.Context, capture snapshotCapture) (SnapshotModel, error) {
	incident, source := capture.source()
	model, _, err := buildSnapshotModel(ctx, incident, capture.SnapshotID, capture.SnapshotAt, capture.SourceChangeSetHighWatermark, source)
	model.ReferencePacks = capture.ReferencePacks
	return model, err
}
