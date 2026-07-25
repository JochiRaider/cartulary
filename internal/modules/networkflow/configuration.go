package networkflow

import (
	"errors"
	"fmt"
	pathpkg "path"
	"strings"
)

type Configuration struct {
	Claimed             bool   `toml:"claimed,omitempty"`
	KeyRingManifestPath string `toml:"key_ring_manifest_path,omitempty"`
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

// NormalizeAndValidateConfiguration is pure owner policy. It performs no file,
// secret, network, database, process, or other external access.
func NormalizeAndValidateConfiguration(configuration Configuration) (Configuration, []ConfigurationFinding) {
	if !configuration.Claimed {
		configuration.KeyRingManifestPath = ""
		return configuration, nil
	}
	if configuration.KeyRingManifestPath == "" {
		return configuration, []ConfigurationFinding{
			{
				Path:       keyRingConfigPath,
				ReasonCode: "network_flow_cursor_key_missing",
				Message:    "Network Flow cursor key-ring configuration is required when Network Flow Activity is claimed",
			},
			{
				Path:       keyRingConfigPath,
				ReasonCode: "network_flow_safe_digest_key_missing",
				Message:    "Network Flow safe-digest key-ring configuration is required when Network Flow Activity is claimed",
			},
		}
	}

	normalized, finding := normalizeManifestPath(configuration.KeyRingManifestPath)
	if finding != nil {
		return configuration, []ConfigurationFinding{*finding}
	}
	configuration.KeyRingManifestPath = normalized
	return configuration, nil
}

func normalizeManifestPath(raw string) (string, *ConfigurationFinding) {
	if !strings.HasPrefix(raw, "/") {
		return "", &ConfigurationFinding{
			Path:       keyRingConfigPath,
			ReasonCode: "path_not_absolute",
			Message:    "bootstrap manifest path must be an absolute POSIX path",
		}
	}
	if strings.ContainsRune(raw, '\x00') {
		return "", &ConfigurationFinding{
			Path:       keyRingConfigPath,
			ReasonCode: "path_forbidden_segment",
			Message:    "bootstrap manifest path must not contain NUL",
		}
	}
	if strings.HasPrefix(raw, "~") || strings.Contains(raw, "$") {
		return "", &ConfigurationFinding{
			Path:       keyRingConfigPath,
			ReasonCode: "path_forbidden_segment",
			Message:    "bootstrap manifest path must not use shell expansion segments",
		}
	}
	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return "", &ConfigurationFinding{
				Path:       keyRingConfigPath,
				ReasonCode: "path_forbidden_segment",
				Message:    "bootstrap manifest path must not contain . or .. segments",
			}
		}
	}
	return pathpkg.Clean(raw), nil
}

type ManifestReadFailure string

const (
	ManifestUnreadable ManifestReadFailure = "unreadable"
	ManifestUnsafe     ManifestReadFailure = "unsafe_object"
	ManifestTooLarge   ManifestReadFailure = "too_large"
)

func KeyRingManifestReadError(failure ManifestReadFailure) error {
	switch failure {
	case ManifestTooLarge:
		return keyRingConfigError(
			keyRingConfigPath,
			"network_flow_cursor_key_invalid",
			"Network Flow key-ring manifest exceeds 65536 bytes",
		)
	case ManifestUnsafe:
		return keyRingConfigError(
			keyRingConfigPath,
			"network_flow_cursor_key_invalid",
			"Network Flow key-ring manifest path must reference one regular file",
		)
	default:
		return keyRingConfigError(
			keyRingConfigPath,
			"network_flow_cursor_key_missing",
			"read Network Flow key-ring manifest: file is unavailable",
		)
	}
}
