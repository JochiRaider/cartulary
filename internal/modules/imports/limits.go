package imports

import "github.com/JochiRaider/cartulary/internal/platform/archivepolicy"

// Limits contains the import-owned deployment ceilings used while decoding
// source material.
type Limits struct {
	MaxCSVSourceBytes  int64
	MaxXLSXSourceBytes int64
	MaxRows            int64
	MaxColumns         int64
	MaxCells           int64
}

type ArchiveLimits = archivepolicy.Limits
