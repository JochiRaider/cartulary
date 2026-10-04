package packformat

import (
	"bytes"
	"encoding/base64"
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var successfulEnvelopeShape = compileProjection("successful_envelope.v1.schema.json")

func admittedSuccessfulEnvelopeSize(size int) bool { return size > 0 && size <= 67108864 }

// ValidateSuccessfulEnvelope admits persisted success evidence without
// authenticating it. Restore separately verifies signatures and retained-root
// linkage. Signed metadata remains opaque original bytes at this boundary.
func ValidateSuccessfulEnvelope(data []byte) error {
	invalid := errors.New("reference pack: invalid successful envelope")
	if !admittedSuccessfulEnvelopeSize(len(data)) {
		return invalid
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil || !successfulEnvelopeShape.matches(value) {
		return invalid
	}
	canonical, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(canonical, data) {
		return invalid
	}
	object := value.(map[string]any)
	if _, known := profiles[object["pack_key"].(string)]; !known {
		return invalid
	}
	verified, err := time.Parse(time.RFC3339Nano, object["verified_at"].(string))
	if err != nil || verified.IsZero() {
		return invalid
	}
	operatorFields := []string{"container_sha256", "container_ref", "trust_valid_until", "trust_snapshot", "trust_proposal"}
	for _, field := range operatorFields {
		if (object[field] != nil) != (object["distribution_kind"] == "operator_imported") {
			return invalid
		}
	}
	if object["distribution_kind"] == "packaged_builtin" {
		return nil
	}
	expires, err := time.Parse(time.RFC3339Nano, object["trust_valid_until"].(string))
	if err != nil || !verified.Before(expires) {
		return invalid
	}
	snapshot := object["trust_snapshot"].(map[string]any)
	proposal := object["trust_proposal"].(map[string]any)
	if proposal["valid_until"] != object["trust_valid_until"] {
		return invalid
	}
	binding := proposal["binding"].(map[string]any)
	for _, field := range []string{"pack_key", "pack_version", "manifest_sha256", "payload_sha256"} {
		if binding[field] != object[field] {
			return invalid
		}
	}
	for _, state := range []map[string]any{snapshot, proposal} {
		root, err := decodeEnvelopeBytes(state["root"])
		if err != nil {
			return invalid
		}
		rootVersion, err := RootVersion(root)
		if err != nil {
			return invalid
		}
		var historyBytes int64
		history := state["root_history"].(map[string]any)
		for key, value := range history {
			version, err := strconv.ParseInt(key, 10, 64)
			if err != nil || version < 1 || version > rootVersion || strconv.FormatInt(version, 10) != key {
				return invalid
			}
			data, err := decodeEnvelopeBytes(value)
			if err != nil {
				return invalid
			}
			actual, err := RootVersion(data)
			if err != nil || version != actual {
				return invalid
			}
			historyBytes += int64(len(data))
		}
		// A frozen snapshot includes the admission root in addition to the
		// container's bounded metadata; a proposal contains only new roots.
		if historyBytes > 10485760 {
			return invalid
		}
	}
	root, _ := decodeEnvelopeBytes(snapshot["root"])
	rootVersion, _ := RootVersion(root)
	if snapshot["root_history"].(map[string]any)[strconv.FormatInt(rootVersion, 10)] != snapshot["root"] {
		return invalid
	}
	for _, state := range []map[string]any{snapshot["highest"].(map[string]any), proposal["metadata"].(map[string]any)} {
		for role, value := range state {
			metadata := value.(map[string]any)
			data, err := decodeEnvelopeBytes(metadata["bytes"])
			if err != nil {
				return invalid
			}
			admitted, err := decodeMetadataShape(data, role, 64)
			version, ok := integer(metadata["version"], 1, 9007199254740991)
			if err != nil || !ok || admitted.version != version {
				return invalid
			}
			canonical, err := canonicaljson.Canonicalize(data)
			if err != nil || !bytes.Equal(canonical, data) {
				return invalid
			}
		}
	}
	sortedSigners := func(value any) bool {
		previous := ""
		for _, value := range value.([]any) {
			id := value.(string)
			if id <= previous {
				return false
			}
			previous = id
		}
		return true
	}
	for _, value := range proposal["signers"].(map[string]any) {
		if !sortedSigners(value) {
			return invalid
		}
	}
	transitions := proposal["root_transitions"].([]any)
	history := proposal["root_history"].(map[string]any)
	if len(history) != len(transitions) {
		return invalid
	}
	for _, value := range transitions {
		transition := value.(map[string]any)
		version, ok := integer(transition["version"], 1, 9007199254740991)
		if !ok || version != rootVersion+1 || history[strconv.FormatInt(version, 10)] == nil || !sortedSigners(transition["previous_signers"]) || !sortedSigners(transition["next_signers"]) {
			return invalid
		}
		rootVersion = version
	}
	proposalRoot, _ := decodeEnvelopeBytes(proposal["root"])
	proposalVersion, _ := RootVersion(proposalRoot)
	if proposalVersion != rootVersion || len(transitions) == 0 && !bytes.Equal(root, proposalRoot) || len(transitions) > 0 && history[strconv.FormatInt(rootVersion, 10)] != proposal["root"] {
		return invalid
	}
	return nil
}

func decodeEnvelopeBytes(value any) ([]byte, error) {
	data, err := base64.StdEncoding.Strict().DecodeString(value.(string))
	if err != nil || len(data) == 0 || len(data) > 2097152 || !slices.Equal([]byte(base64.StdEncoding.EncodeToString(data)), []byte(value.(string))) {
		return nil, errors.New("reference pack: invalid retained metadata bytes")
	}
	return data, nil
}
