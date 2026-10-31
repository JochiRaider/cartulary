package networkflow

import (
	"bytes"
	"encoding/json"
)

// These closed projections give schema order to semantic values even after a
// transport or jsonb round trip has erased their original Go types.
type canonicalGraphTimeRange struct {
	StartUTC *string `json:"start_utc"`
	EndUTC   *string `json:"end_utc"`
}

type canonicalGraphAggregation struct {
	Mode                  string `json:"mode"`
	IncludeExampleRowRefs bool   `json:"include_example_row_refs"`
	BucketWidthSeconds    *int64 `json:"bucket_width_seconds,omitempty"`
}

type canonicalGraphSemantic struct {
	SchemaID         string                    `json:"schema_id"`
	SelectedTableIDs []string                  `json:"selected_table_ids"`
	Filters          []queryFilter             `json:"filters"`
	TimeRange        canonicalGraphTimeRange   `json:"time_range"`
	Aggregation      canonicalGraphAggregation `json:"aggregation"`
}

func graphCanonicalSemantic(value any) (canonicalGraphSemantic, error) {
	// Validate scalars before decoding: encoding/json alone replaces bad UTF-8.
	encoded, err := canonicalJSON(value)
	if err != nil {
		return canonicalGraphSemantic{}, err
	}
	var semantic canonicalGraphSemantic
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&semantic); err != nil {
		return canonicalGraphSemantic{}, err
	}
	if semantic.SchemaID != schemaGraphSemanticQueryV2 || semantic.SelectedTableIDs == nil || semantic.Filters == nil ||
		semantic.Aggregation.Mode != "default_flow_edge_v1" && semantic.Aggregation.Mode != "time_bucket_v1" ||
		(semantic.Aggregation.Mode == "time_bucket_v1") != (semantic.Aggregation.BucketWidthSeconds != nil) {
		return canonicalGraphSemantic{}, errCanonicalValue
	}
	return semantic, nil
}

func canonicalGraphSemanticJSON(value any) ([]byte, error) {
	semantic, err := graphCanonicalSemantic(value)
	if err != nil {
		return nil, err
	}
	return canonicalJSON(semantic)
}

func graphCanonicalTimeRange(value graphTimeRange) canonicalGraphTimeRange {
	resource := graphTimeRangeResource(value)
	var out canonicalGraphTimeRange
	if start, ok := resource["start_utc"].(string); ok {
		out.StartUTC = &start
	}
	if end, ok := resource["end_utc"].(string); ok {
		out.EndUTC = &end
	}
	return out
}

func graphCanonicalAggregation(value graphAggregation) canonicalGraphAggregation {
	out := canonicalGraphAggregation{Mode: value.Mode, IncludeExampleRowRefs: value.IncludeExampleRowRefs}
	if value.Mode == "time_bucket_v1" {
		out.BucketWidthSeconds = &value.BucketWidthSeconds
	}
	return out
}
