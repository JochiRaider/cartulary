package packformat

import "github.com/JochiRaider/cartulary/referencepacks/format/envelope"

func validLicense(license License) bool {
	return envelope.ValidateManifestLicense(Manifest{License: license}) == nil
}
func ManifestLicenseFindings(m Manifest, emit FindingSink) error {
	return formatError(envelope.ManifestLicenseFindings(m, emit))
}
