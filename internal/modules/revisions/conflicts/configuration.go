package conflicts

import (
	"errors"
	"fmt"
	pathpkg "path"
	"strings"
)

const (
	ConflictTokenKeyRingManifestMaximumSize int64 = 65536
	conflictTokenKeyRingConfigPath                = "revisions.conflict_token_key_ring_manifest_path"
)

type Configuration struct {
	ConflictTokenKeyRingManifestPath string `toml:"conflict_token_key_ring_manifest_path,omitempty"`
}

type ConfigurationFinding struct {
	Path       string
	ReasonCode string
	Message    string
}

type ConfigurationError struct {
	Finding ConfigurationFinding
}

func (err *ConfigurationError) Error() string {
	return fmt.Sprintf("%s: %s: %s", err.Finding.Path, err.Finding.ReasonCode, err.Finding.Message)
}

func ConfigurationFindingFromError(err error) (ConfigurationFinding, bool) {
	var configurationError *ConfigurationError
	if !errors.As(err, &configurationError) {
		return ConfigurationFinding{}, false
	}
	return configurationError.Finding, true
}

func NormalizeAndValidateConfiguration(configuration Configuration) (Configuration, []ConfigurationFinding) {
	if configuration.ConflictTokenKeyRingManifestPath == "" {
		return configuration, []ConfigurationFinding{{
			Path:       conflictTokenKeyRingConfigPath,
			ReasonCode: "revisions_conflict_token_manifest_missing",
			Message:    "Revisions conflict-token key-ring manifest path is required",
		}}
	}
	normalized, finding := normalizeConflictTokenManifestPath(configuration.ConflictTokenKeyRingManifestPath)
	if finding != nil {
		return configuration, []ConfigurationFinding{*finding}
	}
	configuration.ConflictTokenKeyRingManifestPath = normalized
	return configuration, nil
}

func normalizeConflictTokenManifestPath(raw string) (string, *ConfigurationFinding) {
	if !strings.HasPrefix(raw, "/") {
		return "", &ConfigurationFinding{Path: conflictTokenKeyRingConfigPath, ReasonCode: "path_not_absolute", Message: "Revisions conflict-token key-ring manifest path must be absolute"}
	}
	if strings.ContainsRune(raw, '\x00') || strings.HasPrefix(raw, "~") || strings.Contains(raw, "$") || strings.Contains(raw, `\`) {
		return "", &ConfigurationFinding{Path: conflictTokenKeyRingConfigPath, ReasonCode: "path_forbidden_segment", Message: "Revisions conflict-token key-ring manifest path contains a forbidden segment"}
	}
	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return "", &ConfigurationFinding{Path: conflictTokenKeyRingConfigPath, ReasonCode: "path_forbidden_segment", Message: "Revisions conflict-token key-ring manifest path must not contain . or .. segments"}
		}
	}
	cleaned := pathpkg.Clean(raw)
	if cleaned == "/" || cleaned != raw {
		return "", &ConfigurationFinding{Path: conflictTokenKeyRingConfigPath, ReasonCode: "path_forbidden_segment", Message: "Revisions conflict-token key-ring manifest path must be normalized"}
	}
	return cleaned, nil
}

type ManifestReadFailure string

const (
	ManifestUnreadable ManifestReadFailure = "unreadable"
	ManifestUnsafe     ManifestReadFailure = "unsafe_object"
	ManifestTooLarge   ManifestReadFailure = "too_large"
)

func ConflictTokenKeyRingManifestReadError(failure ManifestReadFailure) error {
	switch failure {
	case ManifestTooLarge:
		return conflictTokenConfigError(conflictTokenKeyRingConfigPath, "revisions_conflict_token_manifest_invalid", "Revisions conflict-token key-ring manifest exceeds 65536 bytes")
	case ManifestUnsafe:
		return conflictTokenConfigError(conflictTokenKeyRingConfigPath, "revisions_conflict_token_manifest_invalid", "Revisions conflict-token key-ring manifest path must reference one secure regular file")
	default:
		return conflictTokenConfigError(conflictTokenKeyRingConfigPath, "revisions_conflict_token_manifest_missing", "Revisions conflict-token key-ring manifest is unavailable")
	}
}

func conflictTokenConfigError(path string, reasonCode string, message string) error {
	return &ConfigurationError{Finding: ConfigurationFinding{Path: path, ReasonCode: reasonCode, Message: message}}
}
