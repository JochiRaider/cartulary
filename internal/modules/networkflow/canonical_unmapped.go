package networkflow

import (
	"encoding/json"
	"slices"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// NF-106 assigns numeric order to this typed ordinal object. It is not a
// generic dynamic dictionary; each value has Table 12-A1's closed shape.
type retainedRawValue struct {
	SourceColumnOrdinal int    `json:"source_column_ordinal"`
	RawHeaderText       string `json:"raw_header_text"`
	RawHeaderSHA256     string `json:"raw_header_sha256"`
	DecodedValue        string `json:"decoded_value"`
	DecodedValueSHA256  string `json:"decoded_value_sha256"`
}

type retainedRawValues map[int]retainedRawValue

func (values retainedRawValues) canonicalObject() (canonicalObject, error) {
	if values == nil {
		return nil, errCanonicalValue
	}
	ordinals := make([]int, 0, len(values))
	for ordinal, value := range values {
		if ordinal < 1 || value.SourceColumnOrdinal != ordinal ||
			value.RawHeaderSHA256 != sha256Hex([]byte(value.RawHeaderText)) ||
			value.DecodedValueSHA256 != sha256Hex([]byte(value.DecodedValue)) {
			return nil, errCanonicalValue
		}
		ordinals = append(ordinals, ordinal)
	}
	slices.Sort(ordinals)
	object := make(canonicalObject, 0, len(ordinals))
	for _, ordinal := range ordinals {
		object = append(object, canonicalMember{strconv.Itoa(ordinal), values[ordinal]})
	}
	return object, nil
}

func (values *retainedRawValues) UnmarshalJSON(data []byte) error {
	decoded, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return err
	}
	object, ok := decoded.(map[string]any)
	if !ok {
		return errCanonicalValue
	}
	result := retainedRawValues{}
	for key, raw := range object {
		ordinal, err := strconv.Atoi(key)
		if err != nil || ordinal < 1 || strconv.Itoa(ordinal) != key {
			return errCanonicalValue
		}
		members, ok := raw.(map[string]any)
		if !ok || len(members) != 5 {
			return errCanonicalValue
		}
		number, numberOK := members["source_column_ordinal"].(json.Number)
		header, headerOK := members["raw_header_text"].(string)
		headerSHA, headerSHAOK := members["raw_header_sha256"].(string)
		value, valueOK := members["decoded_value"].(string)
		valueSHA, valueSHAOK := members["decoded_value_sha256"].(string)
		if !numberOK || string(number) != key || !headerOK || !headerSHAOK || !valueOK || !valueSHAOK {
			return errCanonicalValue
		}
		result[ordinal] = retainedRawValue{ordinal, header, headerSHA, value, valueSHA}
	}
	if _, err := result.canonicalObject(); err != nil {
		return err
	}
	*values = result
	return nil
}
