package reference_data

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/archivepolicy"
)

type ArchiveLimits = archivepolicy.Limits

// MaxAdministrativeRequestBytes bounds JSON bodies and upload metadata before
// parsing. It is fixed independently of the container's streaming byte limit.
const MaxAdministrativeRequestBytes = 65536

// ReferenceLimits contains Reference Pack-specific extraction ceilings.
type ReferenceLimits struct {
	MaxExtractedBytes      int64
	MaxContainerBytes      int64
	MaxVerificationSeconds int64
}

type Limits struct {
	Archives       archivepolicy.Limits
	ReferencePacks ReferenceLimits
}

func DefaultLimits() Limits {
	limits := packformat.DefaultArchiveLimits()
	return Limits{Archives: ArchiveLimits{DefaultMaxExtractedBytes: limits.ExtractedBytes, MaxCompressionRatio: limits.CompressionRatio, MaxMembers: limits.Members}, ReferencePacks: ReferenceLimits{MaxExtractedBytes: limits.ExtractedBytes, MaxContainerBytes: limits.ContainerBytes, MaxVerificationSeconds: 1800}}
}

func (l Limits) verificationArchiveLimits() packformat.ArchiveLimits {
	extracted := l.ReferencePacks.MaxExtractedBytes
	if extracted == 0 {
		extracted = l.Archives.DefaultMaxExtractedBytes
	}
	return packformat.ArchiveLimits{ContainerBytes: l.ReferencePacks.MaxContainerBytes, ExtractedBytes: extracted, CompressionRatio: l.Archives.MaxCompressionRatio, Members: l.Archives.MaxMembers}
}
