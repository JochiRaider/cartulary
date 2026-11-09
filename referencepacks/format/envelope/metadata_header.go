package envelope

import "time"

// MetadataHeader is structural metadata only. It grants no trust, signature,
// freshness, rollback or repository assurance; use VerifyTrust for admission.
type MetadataHeader struct {
	Version int64
	Expires time.Time
}

// ReadMetadataHeader admits the closed metadata shape before retained-state
// comparisons. It never repairs or canonicalizes the caller's bytes.
func ReadMetadataHeader(data []byte, role string) (MetadataHeader, error) {
	maximum := 64
	if role == "root" {
		maximum = 128
	}
	m, err := decodeMetadataShape(data, role, maximum)
	if err == nil && role == "root" {
		root, rootErr := decodeRootShape(data)
		m, err = root.metadata, rootErr
	}
	if err != nil {
		return MetadataHeader{}, err
	}
	return MetadataHeader{Version: m.version, Expires: m.expires}, nil
}

// MetadataFindings enumerates a single bounded metadata document. The role
// selects a fixed built-in shape and signature bound, never a supplied schema.
func MetadataFindings(data []byte, role, path string, emit FindingSink) error {
	maximum := 64
	if role == "root" {
		maximum = 128
	}
	_, err := metadataSchemaFindings(data, role, maximum, path, emit)
	return err
}
