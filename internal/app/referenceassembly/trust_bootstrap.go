package referenceassembly

import (
	"errors"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

func ReadTrustBootstrap(configuration reference_data.Configuration) (reference_data.TrustBootstrap, error) {
	if configuration.TrustBootstrapPath == "" {
		return reference_data.TrustBootstrap{}, errors.New("reference pack: trust bootstrap configuration is required")
	}
	if err := reference_data.ValidateReferencePackConfiguration(configuration); err != nil {
		return reference_data.TrustBootstrap{}, err
	}
	root, err := rootedfs.Open(filepath.Dir(configuration.TrustBootstrapPath))
	if err != nil {
		return reference_data.TrustBootstrap{}, errors.New("reference pack: trust bootstrap file unavailable")
	}
	defer root.Close()
	ref, err := rootedfs.ParseReference(filepath.Base(configuration.TrustBootstrapPath))
	if err != nil {
		return reference_data.TrustBootstrap{}, errors.New("reference pack: invalid trust bootstrap file reference")
	}
	data, _, err := root.ReadRegular(ref, 8388608)
	if err != nil {
		return reference_data.TrustBootstrap{}, errors.New("reference pack: trust bootstrap file unavailable or exceeds byte limit")
	}
	return reference_data.AdmitTrustBootstrap(data)
}
