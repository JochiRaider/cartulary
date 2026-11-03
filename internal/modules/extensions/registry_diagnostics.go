package extensions

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var safeArtifactIdentity = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~-]{0,381}$`)

// A broken package cannot supply its own trusted diagnostic schema or registry.
// This bootstrap mapping uses only the declared package slot, never its bytes
// or a parser's error. The generated condition/schema project the same owner
// contract and are checked against this mapping using valid packaged inputs.
func registryArtifactFinding(slot string) StartupFinding {
	kind := "registry"
	for _, candidate := range []string{"dependency", "owner", "configuration", "descriptor", "integrity", "binding", "client", "route", "state", "backup", "job", "participant", "validation", "schema", "generator"} {
		if strings.Contains(slot, candidate) {
			kind = candidate
			break
		}
	}
	identity := slot
	if !safeArtifactIdentity.MatchString(identity) {
		sum := sha256.Sum256([]byte(slot))
		identity = "sha256-" + hex.EncodeToString(sum[:])
	}
	return StartupFinding{Path: "$", ReasonCode: "extension_registry_invalid", Message: "The extension contract artifact set is invalid.", Details: map[string]any{"artifact_kind": kind, "safe_ref": "cartulary:" + kind + ":" + identity, "expected": nil, "actual": "redacted"}}
}

func implementationUnavailable(profile string, expected int, actual *int, cause error) error {
	var actualValue any
	if actual != nil {
		actualValue = *actual
	}
	return startupFailure(cause, StartupFinding{Path: "$", ReasonCode: "extension_implementation_unavailable", Message: "The requested extension implementation is unavailable or incompatible.", Details: map[string]any{"profile_id": profile, "expected_contract_major": expected, "actual_contract_major": actualValue}})
}
