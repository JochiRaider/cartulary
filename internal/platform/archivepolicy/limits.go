// Package archivepolicy owns the small value object shared by archive-reading
// owners. Wire defaults and deployment admission remain in platform/config.
package archivepolicy

type Limits struct {
	DefaultMaxExtractedBytes int64
	MaxCompressionRatio      int64
	MaxMembers               int64
}
