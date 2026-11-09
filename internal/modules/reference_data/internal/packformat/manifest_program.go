package packformat

import "github.com/JochiRaider/cartulary/referencepacks/format/envelope"

func ManifestChecks(data []byte, size int64, operator bool) (map[string]CheckFunc, func() Manifest) {
	p, r := envelope.ManifestChecks(data, size, operator)
	return formatChecks(p), r
}
func HintChecks(data []byte, size int64) (map[string]CheckFunc, func() string) {
	p, r := envelope.HintChecks(data, size)
	return formatChecks(p), r
}
func DisallowedContentFindings(m Manifest, emit FindingSink) error {
	return formatError(envelope.DisallowedContentFindings(m, emit))
}
