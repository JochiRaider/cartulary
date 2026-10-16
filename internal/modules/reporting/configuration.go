package reporting

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
)

type Configuration struct {
	Claimed            bool   `toml:"claimed"`
	RuntimeReceiptPath string `toml:"runtime_receipt_path"`
}

func ValidateConfiguration(configuration Configuration) error {
	path := configuration.RuntimeReceiptPath
	// Omission is admitted structurally for an inactive profile. A claimed
	// application must open this exact installation before state admission.
	if path == "" {
		return nil
	}
	if len(path) > 4096 || !filepath.IsAbs(path) || filepath.Clean(path) != path || filepath.Base(path) != "installation.json" || filepath.Dir(path) == "/" || strings.ContainsRune(path, 0) {
		return errors.New("reporting runtime receipt requires a canonical absolute installation.json path")
	}
	return nil
}

func ApplyConfigurationOverlay(configuration Configuration, segments []string, raw string) (Configuration, error) {
	if strings.Join(segments, ".") == "snapshot_reporting.claimed" {
		claimed, err := strconv.ParseBool(raw)
		if err != nil {
			return configuration, errors.New("reporting claim must be boolean")
		}
		configuration.Claimed = claimed
		return configuration, nil
	}
	if strings.Join(segments, ".") != "snapshot_reporting.runtime_receipt_path" {
		return configuration, errors.New("unknown Reporting configuration key")
	}
	configuration.RuntimeReceiptPath = raw
	if raw == "" {
		return configuration, errors.New("reporting runtime receipt path is empty")
	}
	return configuration, ValidateConfiguration(configuration)
}
