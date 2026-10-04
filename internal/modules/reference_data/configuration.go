package reference_data

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

// Configuration contains deployment assertions, never inferred network-time
// health. Claim selection remains an Extensions/Core configuration concern.
type Configuration struct {
	ClockTrusted         bool   `toml:"clock_trusted"`
	TrustBootstrapPath   string `toml:"trust_bootstrap_path"`
	ClockTrustedExplicit bool   `toml:"-"`
}

func ValidateReferencePackConfiguration(c Configuration) error {
	if c.TrustBootstrapPath != "" && (len(c.TrustBootstrapPath) > 4096 || !filepath.IsAbs(c.TrustBootstrapPath) || filepath.Clean(c.TrustBootstrapPath) != c.TrustBootstrapPath || strings.ContainsRune(c.TrustBootstrapPath, 0)) {
		return errors.New("reference pack: trust bootstrap requires a canonical absolute file path")
	}
	return nil
}

func ApplyConfigurationOverlay(c Configuration, path []string, raw string) (Configuration, error) {
	switch strings.Join(path, ".") {
	case "reference_packs.clock_trusted":
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return c, errors.New("reference pack: clock_trusted must be a boolean")
		}
		c.ClockTrusted = value
	case "reference_packs.trust_bootstrap_path":
		c.TrustBootstrapPath = raw
	default:
		return c, errors.New("reference pack: unknown configuration key")
	}
	return c, nil
}

// TrustBootstrap is admitted through the owner boundary. No caller can supply
// an unchecked root or change a retained repository by mutating this value.
type TrustBootstrap struct {
	repositories map[string]packformat.TrustSnapshot
}

func AdmitTrustBootstrap(data []byte) (TrustBootstrap, error) {
	repositories, err := packformat.DecodeBootstrap(data)
	return TrustBootstrap{repositories: repositories}, err
}
