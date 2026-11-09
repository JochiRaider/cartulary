// Package dataset implements the finite generic Reference Pack dataset profile.
// It has no source-specific interpretation, publication or application policy.
package dataset

import (
	"context"
	"encoding/json"
	"errors"
	"io"
)

const (
	ProfileID          = "cartulary.reference_pack.dataset.v1"
	DescriptorSchema   = "cartulary.reference_pack.dataset.descriptor.v1"
	ObjectSchema       = "cartulary.reference_pack.dataset.object.v1"
	RelationshipSchema = "cartulary.reference_pack.dataset.relationship.v1"
	DescriptorPath     = "payload/descriptor.json"
	ObjectsPath        = "payload/objects.ndjson"
	RelationshipsPath  = "payload/relationships.ndjson"
	MaxDescriptorBytes = 65536
	MaxLookups         = 32
	MaxComponents      = 4
	MaxObjectPostings  = 128
	MaxKeyBytes        = 65536
	MaxPostings        = 8000000
	MaxTotalKeyBytes   = 268435456
	MaxDepth           = 64
	MaxRowBytes        = 4194304
	MaxObjects         = 2000000
	MaxRelationships   = 5000000
	MaxSafeInteger     = 9007199254740991
)

// Failure contains only engine-owned tokens and bounded numeric diagnostics.
// Untrusted names, payload values and host paths never appear in its message.
type Failure struct {
	Code, Limit       string
	Maximum, Observed uint64
}

func (e *Failure) Error() string { return "reference dataset: " + e.Code }
func invalid() error             { return &Failure{Code: "content_schema_invalid"} }
func semantic() error            { return &Failure{Code: "content_semantic_invalid"} }
func excess(id string, maximum, observed uint64, structural bool) error {
	code := "content_semantic_invalid"
	if structural {
		code = "content_schema_invalid"
	}
	return &Failure{Code: code, Limit: id, Maximum: maximum, Observed: observed}
}

var (
	ErrStorage           = errors.New("reference dataset: member storage unavailable")
	ErrUnsupportedLookup = errors.New("reference dataset: unsupported lookup")
	ErrInvalidInput      = errors.New("reference dataset: malformed lookup input")
)

type Descriptor struct {
	SchemaID string   `json:"schema_id"`
	Lookups  []Lookup `json:"lookups"`
}
type Lookup struct {
	Name       string      `json:"name"`
	InputType  string      `json:"input_type"`
	Components []Component `json:"components"`
	Unique     bool        `json:"unique"`
	Coverage   Coverage    `json:"coverage"`
}
type Component struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Normalization string `json:"normalization"`
}
type Coverage struct {
	Objects      int64 `json:"objects"`
	Postings     int64 `json:"postings"`
	DistinctKeys int64 `json:"distinct_keys"`
}
type SourceRef struct {
	ArtifactIndex int    `json:"artifact_index"`
	Locator       string `json:"locator"`
}
type Object struct {
	inputPostings int
	SchemaID      string                       `json:"schema_id"`
	ID            string                       `json:"object_id"`
	Type          string                       `json:"object_type"`
	DisplayLabel  string                       `json:"display_label"`
	Description   *string                      `json:"description"`
	Deprecated    bool                         `json:"deprecated"`
	Replacement   *string                      `json:"replacement_object_id"`
	SourceRefs    []SourceRef                  `json:"source_refs"`
	Data          json.RawMessage              `json:"data"`
	LookupValues  map[string][]json.RawMessage `json:"lookup_values"`
}
type Relationship struct {
	SchemaID   string          `json:"schema_id"`
	ID         string          `json:"relationship_id"`
	Type       string          `json:"relationship_type"`
	Source     string          `json:"source_object_id"`
	Target     string          `json:"target_object_id"`
	SourceRefs []SourceRef     `json:"source_refs"`
	Data       json.RawMessage `json:"data"`
}

// Source opens an already inventoried logical member, never a host path. Bytes
// must remain immutable throughout validation and indexing.
type Source interface {
	Open(context.Context, string) (io.ReadCloser, error)
}
type Posting struct{ Lookup, Key string }
type Row struct {
	ID, Shape string
	Canonical []byte
	Postings  []Posting
}

// Sink writes a private index. The caller owns its discard/atomic publication.
type Sink interface {
	Append(context.Context, Row) error
}
type Expected struct {
	Objects, Relationships int64
	SourceArtifacts        int
}
type Result struct {
	Descriptor                                        Descriptor
	DescriptorBytes                                   []byte
	Objects, Relationships, Postings, EncodedKeyBytes int64
}
