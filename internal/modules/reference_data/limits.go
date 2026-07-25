package reference_data

import "github.com/JochiRaider/cartulary/internal/platform/archivepolicy"

type ArchiveLimits = archivepolicy.Limits

// ReferenceLimits contains Reference Pack-specific extraction ceilings.
type ReferenceLimits struct {
	MaxExtractedBytes int64
}

type Limits struct {
	Archives       archivepolicy.Limits
	ReferencePacks ReferenceLimits
}
