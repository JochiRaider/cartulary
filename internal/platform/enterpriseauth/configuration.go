package enterpriseauth

import (
	"errors"
	"fmt"
	pathpkg "path"
	"strings"
)

const (
	providerManifestPathKey = "enterprise_authentication.provider_manifest_path"

	// ProviderManifestMaximumSize is the Core 04 bound applied before parsing.
	ProviderManifestMaximumSize int64 = 1048576
	// SigningCertificateMaximumSize is the Core 04 bound applied to each
	// referenced SAML signing certificate before parsing.
	SigningCertificateMaximumSize int64 = 262144
)

type Configuration struct {
	Claimed              bool   `toml:"claimed,omitempty"`
	ProviderManifestPath string `toml:"provider_manifest_path,omitempty"`
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
		configuration.ProviderManifestPath = ""
		return configuration, nil
	}
	if configuration.ProviderManifestPath == "" {
		return configuration, []ConfigurationFinding{{
			Path:       providerManifestPathKey,
			ReasonCode: "provider_manifest_path_missing",
			Message:    "enterprise provider manifest path is required when enterprise authentication is claimed",
		}}
	}

	normalized, finding := normalizeProviderManifestPath(configuration.ProviderManifestPath)
	if finding != nil {
		return configuration, []ConfigurationFinding{*finding}
	}
	configuration.ProviderManifestPath = normalized
	return configuration, nil
}

func normalizeProviderManifestPath(raw string) (string, *ConfigurationFinding) {
	if !strings.HasPrefix(raw, "/") {
		return "", &ConfigurationFinding{
			Path:       providerManifestPathKey,
			ReasonCode: "path_not_absolute",
			Message:    "enterprise provider manifest path must be an absolute POSIX path",
		}
	}
	if strings.ContainsRune(raw, '\x00') {
		return "", &ConfigurationFinding{
			Path:       providerManifestPathKey,
			ReasonCode: "path_forbidden_segment",
			Message:    "enterprise provider manifest path must not contain NUL",
		}
	}
	if strings.HasPrefix(raw, "~") || strings.Contains(raw, "$") {
		return "", &ConfigurationFinding{
			Path:       providerManifestPathKey,
			ReasonCode: "path_forbidden_segment",
			Message:    "enterprise provider manifest path must not use shell expansion segments",
		}
	}
	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return "", &ConfigurationFinding{
				Path:       providerManifestPathKey,
				ReasonCode: "path_forbidden_segment",
				Message:    "enterprise provider manifest path must not contain . or .. segments",
			}
		}
	}
	return pathpkg.Clean(raw), nil
}

func providerConfigError(path string, reasonCode string, message string) error {
	return &ConfigurationError{Finding: ConfigurationFinding{
		Path:       path,
		ReasonCode: reasonCode,
		Message:    message,
	}}
}
