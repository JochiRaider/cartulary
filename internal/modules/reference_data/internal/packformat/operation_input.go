package packformat

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// OperationInput is the current closed durable operation format shared by
// startup admission and execution. Content profiles are validated on their
// manifests and sets; operation inputs contain references, not nested entries.
type OperationInput struct {
	SchemaID            string    `json:"schema_id"`
	Contract            string    `json:"pack_contract_version"`
	Kind                string    `json:"kind"`
	AdmittedAt          time.Time `json:"admitted_at"`
	PackSetID           *string   `json:"pack_set_id"`
	SetRevision         int64     `json:"set_revision"`
	ConfigurationSHA256 string    `json:"configuration_sha256"`
	ClockTrusted        bool      `json:"clock_trusted"`
	TimeoutSeconds      int64     `json:"timeout_seconds"`
	ContainerReference  *string   `json:"container_reference"`
	ContainerSHA256     *string   `json:"container_sha256"`
	ContainerBytes      *int64    `json:"container_bytes"`
}

func DecodeOperationInput(data []byte) (OperationInput, error) {
	var value OperationInput
	decoded, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return value, errors.New("reference pack: invalid retained operation input")
	}
	object, ok := decoded.(map[string]any)
	if !ok || len(object) != 12 {
		return value, errors.New("reference pack: invalid retained operation input")
	}
	for _, name := range []string{"schema_id", "pack_contract_version", "kind", "admitted_at", "set_revision", "configuration_sha256", "clock_trusted", "timeout_seconds"} {
		if object[name] == nil {
			return value, errors.New("reference pack: incomplete retained operation input")
		}
	}
	for _, name := range []string{"pack_set_id", "container_reference", "container_sha256", "container_bytes"} {
		if _, ok := object[name]; !ok {
			return value, errors.New("reference pack: omitted nullable operation input")
		}
	}
	canonical, err := canonicaljson.Marshal(decoded)
	if err != nil || !bytes.Equal(canonical, data) {
		return value, errors.New("reference pack: noncanonical operation input")
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	if value.Contract != ContractID || value.SchemaID != "cartulary.reference_pack_operation_input.v3" || value.AdmittedAt.IsZero() || value.TimeoutSeconds < 60 || value.TimeoutSeconds > 86400 || !slices.Contains([]string{"import", "reverify", "refresh", "activate", "disable", "remove", "integrity"}, value.Kind) || (value.PackSetID == nil && value.Kind != "integrity" || value.PackSetID != nil && !validOperationSetID(*value.PackSetID)) || value.SetRevision < 1 || !validOperationDigest(value.ConfigurationSHA256) {
		return value, errors.New("reference pack: invalid retained operation input")
	}
	if value.Kind == "import" {
		if value.ContainerReference == nil || value.ContainerSHA256 == nil || value.ContainerBytes == nil || *value.ContainerBytes < 0 || !validOperationDigest(*value.ContainerSHA256) {
			return value, errors.New("reference pack: incomplete retained import input")
		}
		if *value.ContainerReference == "" {
			return value, errors.New("reference pack: empty retained input reference")
		}
	} else if value.ContainerReference != nil || value.ContainerSHA256 != nil || value.ContainerBytes != nil {
		return value, errors.New("reference pack: unexpected retained upload input")
	}
	return value, nil
}

func validOperationDigest(value string) bool {
	_, ok := hexBytes(value, 32)
	return ok
}

func validOperationSetID(value string) bool {
	return strings.HasPrefix(value, "rpset_") && validOperationDigest(strings.TrimPrefix(value, "rpset_"))
}

// ValidatePendingInput admits only the two current closed input contracts.
// It does not search arbitrary JSON for a growing list of obsolete schema IDs.
func ValidatePendingInput(data []byte) error {
	var header struct {
		SchemaID string `json:"schema_id"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.SchemaID {
	case "cartulary.reference_pack_operation_input.v3":
		_, err := DecodeOperationInput(data)
		return err
	case "cartulary.reference_pack_portability_input.v2":
		_, err := decodePortabilityObject(data, portabilityInputShape)
		return err
	default:
		return errors.New("reference pack: unsupported pending input contract")
	}
}
