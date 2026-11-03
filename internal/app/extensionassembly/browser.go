package extensionassembly

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi/webassets"
)

// BrowserCoordinator joins physically verified serving bytes to the final
// Extensions integrity root before any configuration claim is evaluated.
func BrowserCoordinator(build *webassets.Build) (*extensions.Coordinator, error) {
	if build == nil {
		return nil, errors.New("admitted browser build is required")
	}
	c := build.Contracts()
	return extensions.NewPackagedCoordinator(extensions.BrowserPackage{IntegrityJSON: c.IntegrityJSON, IntegritySHA256: c.IntegritySHA256, ClientSupportJSON: c.ClientSupportJSON, AssetManifestJSON: c.AssetManifestJSON, BootstrapTemplateJSON: c.BootstrapTemplateJSON})
}
