package envelope

import (
	"archive/tar"
	"context"
	"io"
	"slices"
	"time"
)

// WriteUSTAR serializes a closed logical inventory deterministically. Envelope
// members are ordinary inputs; neither the clock nor filesystem metadata enters
// the archive. Admission still runs through the production archive reader.
func WriteUSTAR(ctx context.Context, output io.Writer, members map[string][]byte) error {
	names := make([]string, 0, len(members))
	for name := range members {
		if err := ValidatePath(name, false); err != nil {
			return err
		}
		names = append(names, name)
	}
	slices.Sort(names)
	w := tar.NewWriter(output)
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return err
		}
		data := members[name]
		if err := w.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(data)), Typeflag: tar.TypeReg, ModTime: time.Unix(0, 0).UTC(), Format: tar.FormatUSTAR}); err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	return w.Close()
}
