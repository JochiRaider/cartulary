package packformat

import "github.com/JochiRaider/cartulary/referencepacks/format/envelope"

type TrustSnapshot = envelope.TrustSnapshot
type RetainedMetadata = envelope.RetainedMetadata
type TrustProposal = envelope.TrustProposal
type RootTransition = envelope.RootTransition

func AdmitBootstrap(data []byte) (TrustSnapshot, error) {
	v, e := envelope.AdmitBootstrap(data)
	return v, formatError(e)
}
func VerifyHistoricalRoot(previous, data []byte, repository string) (RootTransition, error) {
	v, e := envelope.VerifyHistoricalRoot(previous, data, repository)
	return v, formatError(e)
}
func metadataVersion(data []byte, role string) (int64, error) {
	v, e := envelope.ReadMetadataHeader(data, role)
	return v.Version, formatError(e)
}
