package recovery

import (
	"bytes"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const maxPostgresSnapshotMetadataBytes = 16777216

func postgresSnapshotDigest(snapshot VNextPostgresSnapshotArtifact) (string, error) {
	snapshot.SnapshotSHA256 = ""
	body, err := canonicaljson.MarshalExtension(snapshot)
	if err != nil {
		return "", err
	}
	return digestBytes(append([]byte(vNextPostgresSnapshotDigestDomain), body...)), nil
}

func decodePostgresSnapshot(body []byte, snapshot *VNextPostgresSnapshotArtifact) error {
	if len(body) == 0 || len(body) > maxPostgresSnapshotMetadataBytes {
		return fmt.Errorf("%w: snapshot metadata exceeds its bound", ErrVNextBackup)
	}
	if err := strictDecodeJSON(body, snapshot); err != nil {
		return err
	}
	canonical, err := canonicaljson.MarshalExtension(*snapshot)
	if err != nil || !bytes.Equal(body, canonical) {
		return fmt.Errorf("%w: snapshot metadata is not canonical", ErrVNextBackup)
	}
	return nil
}
