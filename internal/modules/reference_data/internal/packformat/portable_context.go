package packformat

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type PortableVerificationContext struct {
	SchemaID            string `json:"schema_id"`
	ConfigurationSHA256 string `json:"configuration_sha256"`
	ClockTrusted        bool   `json:"clock_trusted"`
	TimeoutSeconds      int64  `json:"timeout_seconds"`
}

var portableContextShape = compileProjection("portable_verification_context.v1.schema.json")

func DecodePortableVerificationContext(data []byte) (PortableVerificationContext, error) {
	invalid := errors.New("reference pack: invalid frozen portable verification context")
	if !admittedPortableContextSize(len(data)) {
		return PortableVerificationContext{}, invalid
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil || !portableContextShape.matches(value) {
		return PortableVerificationContext{}, invalid
	}
	canonical, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(canonical, data) {
		return PortableVerificationContext{}, invalid
	}
	var result PortableVerificationContext
	if err := json.Unmarshal(data, &result); err != nil {
		return PortableVerificationContext{}, invalid
	}
	return result, nil
}

func admittedPortableContextSize(size int) bool { return size >= 1 && size <= 4096 }
