package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	packformat "github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	inputs "github.com/JochiRaider/cartulary/referencepacks/windows/inputs"
	"github.com/JochiRaider/cartulary/referencepacks/windows/internal/native"
)

type sourceRecord struct {
	ObjectID          string              `json:"object_id"`
	EntryID           string              `json:"entry_id"`
	Type              string              `json:"object_type"`
	Label             string              `json:"display_label"`
	Description       *string             `json:"description"`
	Deprecated        bool                `json:"deprecated"`
	ReplacementObject *string             `json:"replacement_object_id"`
	ReplacementEntry  *string             `json:"replacement_entry_id"`
	Refs              []dataset.SourceRef `json:"source_refs"`
	Aliases           []string            `json:"aliases"`
	ExternalRefs      []struct {
		ID string `json:"external_id"`
	} `json:"external_refs"`
	Provider     string `json:"provider_name"`
	EventID      int64  `json:"event_id"`
	EventVersion *int64 `json:"event_version"`
}

func lookupValue(value any) json.RawMessage {
	b, err := canonicaljson.Marshal(value)
	if err != nil {
		panic("producer lookup projection is not JSON")
	}
	return b
}

func projectDataset(ctx context.Context, m *packformat.Manifest, members packageSource) (dataset.Result, error) {
	if err := native.ValidateContent(ctx, *m, members, nil); err != nil {
		return dataset.Result{}, err
	}
	profile, err := inputs.Files.ReadFile("dataset-profile.v1.json")
	if err != nil {
		return dataset.Result{}, err
	}
	var rules struct {
		ID      string           `json:"source_profile_id"`
		Lookups []dataset.Lookup `json:"lookup_definitions"`
	}
	if err = json.Unmarshal(profile, &rules); err != nil {
		return dataset.Result{}, err
	}
	if rules.ID != m.SourceProfileID || Digest(profile) != m.SourceProfileSHA256 {
		return dataset.Result{}, errors.New("generic source profile binding mismatch")
	}
	objects := []dataset.Object{}
	relationships := []dataset.Relationship{}
	objectPath := "payload/objects.ndjson"
	if m.Summary.Kind == "entries" {
		objectPath = "payload/entries.ndjson"
	}
	for raw := range bytes.SplitSeq(members[objectPath], []byte{'\n'}) {
		if len(raw) == 0 {
			continue
		}
		if err = ctx.Err(); err != nil {
			return dataset.Result{}, err
		}
		var record sourceRecord
		if err = json.Unmarshal(raw, &record); err != nil {
			return dataset.Result{}, err
		}
		object := dataset.Object{SchemaID: dataset.ObjectSchema, ID: record.ObjectID, Type: record.Type, DisplayLabel: record.Label, Description: record.Description, Deprecated: record.Deprecated, Replacement: record.ReplacementObject, SourceRefs: record.Refs, Data: bytes.Clone(raw), LookupValues: map[string][]json.RawMessage{}}
		projectLookups(&object, record)
		objects = append(objects, object)
	}
	for raw := range bytes.SplitSeq(members["payload/relationships.ndjson"], []byte{'\n'}) {
		if len(raw) == 0 {
			continue
		}
		var relationship dataset.Relationship
		// Source-model rows have additional native fields. They remain complete and
		// inert in Data; only explicitly named generic fields are projected here.
		if err = json.Unmarshal(raw, &relationship); err != nil {
			return dataset.Result{}, err
		}
		relationship.SchemaID = dataset.RelationshipSchema
		relationship.Data = bytes.Clone(raw)
		relationships = append(relationships, relationship)
	}
	encoded, result, err := dataset.Encode(ctx, rules.Lookups, objects, relationships, len(m.Artifacts))
	if err != nil {
		return dataset.Result{}, err
	}
	delete(members, "payload/entries.ndjson")
	for name, raw := range encoded {
		members[name] = raw
	}
	m.ProfileID = dataset.ProfileID
	m.ProfileVersion = "1"
	m.Summary.Kind = "objects_relationships"
	m.Summary.Entries = nil
	m.Summary.Objects = &result.Objects
	m.Summary.Relationships = &result.Relationships
	return result, nil
}
func projectLookups(object *dataset.Object, record sourceRecord) {
	object.ID = record.EntryID
	object.Type = "windows_event"
	object.Replacement = record.ReplacementEntry
	object.LookupValues["event_id"] = []json.RawMessage{lookupValue(record.EventID)}
	object.LookupValues["provider_event_id"] = []json.RawMessage{lookupValue(map[string]any{"event_id": record.EventID, "provider_name": record.Provider})}
	if record.EventVersion != nil {
		object.LookupValues["provider_event_version"] = []json.RawMessage{lookupValue(map[string]any{"event_id": record.EventID, "event_version": *record.EventVersion, "provider_name": record.Provider})}
	}
}
