package incidentbundles

import "github.com/JochiRaider/cartulary/internal/platform/archivepolicy"

const (
	defaultArchiveMaxCompressionRatio      int64 = 100
	defaultArchiveMaxMembers               int64 = 10000
	defaultIncidentBundleMaxExtractedBytes int64 = 68719476736
)

// Limits contains the admitted archive ceilings used by incident portability.
type Limits struct {
	Archives        archivepolicy.Limits
	IncidentBundles IncidentBundleLimits
}

type ArchiveLimits = archivepolicy.Limits

type IncidentBundleLimits struct {
	MaxExtractedBytes int64
}
