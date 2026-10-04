package incidentbundles

import (
	"context"
	"errors"
	"io"
)

// Enforce the enclosing bundle budget before allocating a container, including
// files already emitted by source owners. The callback receives no filesystem
// path or Reference Data persistence identity.
func boundedContainerSink(files map[string][]byte, limits Limits) func(context.Context, string, int64, io.Reader) error {
	limit := limits.IncidentBundles.MaxExtractedBytes
	if limit <= 0 {
		limit = defaultIncidentBundleMaxExtractedBytes
	}
	used := int64(0)
	valid := true
	for _, data := range files {
		if int64(len(data)) > limit-used {
			valid = false
			break
		}
		used += int64(len(data))
	}
	return func(ctx context.Context, path string, size int64, reader io.Reader) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !valid || size < 0 || size > limit-used {
			return &verificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
		}
		if _, exists := files[path]; exists {
			return errors.New("incident bundle: duplicate owner container")
		}
		// Reserve outer manifest and checksum members before adding one container.
		if err := checkMemberCount(len(files)+3, limits); err != nil {
			return err
		}
		data, err := io.ReadAll(io.LimitReader(reader, size+1))
		if err != nil {
			return err
		}
		if int64(len(data)) != size {
			return errors.New("incident bundle: incomplete owner container")
		}
		files[path] = data
		used += size
		return nil
	}
}
