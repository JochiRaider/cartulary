package packformat

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var attemptResultShape = compileProjection("attempt_result.v1.schema.json")

// A terminal operation uses the same result shape even if no verification
// attempt ever started. The attempt table, not this result, proves execution.
func ValidateAttemptResult(data []byte) error {
	invalid := errors.New("reference pack: invalid retained attempt result")
	if !admittedAttemptResultSize(len(data)) {
		return invalid
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil || !attemptResultShape.matches(value) {
		return invalid
	}
	canonical, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(canonical, data) {
		return invalid
	}
	var result struct {
		Outcome string `json:"outcome"`
		Members int64  `json:"member_count"`
		Failed  int64  `json:"failed_count"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return invalid
	}
	if result.Failed > result.Members || (result.Outcome == "content_rejected") != (result.Failed > 0) {
		return invalid
	}
	return nil
}

func EncodeAttemptResult(outcome string, members, failed int64) ([]byte, error) {
	value := map[string]any{"schema_id": "cartulary.reference_pack_attempt_result.v1", "outcome": outcome}
	if outcome != "interrupted" {
		value["member_count"] = members
		value["failed_count"] = failed
	} else if members != 0 || failed != 0 {
		return nil, errors.New("reference pack: interrupted result cannot publish member counts")
	}
	data, err := canonicaljson.Marshal(value)
	if err != nil {
		return nil, err
	}
	if err := ValidateAttemptResult(data); err != nil {
		return nil, err
	}
	return data, nil
}

func admittedAttemptResultSize(size int) bool { return size > 0 && size <= 4096 }
