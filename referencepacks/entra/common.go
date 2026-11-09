package producer

import (
	"bytes"
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/publication"
	"io"
	"os"
	"path"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

type EvidenceFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size_bytes"`
}

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func Digest(b []byte) string                  { return envelope.Digest(b) }
func DecodeClosed(b []byte, target any) error { return publication.DecodeClosed(b, target) }
func Retain(root *os.Root, name string, b []byte, mode os.FileMode) error {
	return publication.Retain(root, name, b, mode)
}
func readRegular(name string) ([]byte, error) { return publication.ReadRegular(name, 536870912) }
func validBuildTime(value string) bool {
	at, err := time.Parse("2006-01-02T15:04:05Z", value)
	return err == nil && at.Year() >= 1 && at.Format("2006-01-02T15:04:05Z") == value
}

type packageSource map[string][]byte

func (s packageSource) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	b, ok := s[name]
	if !ok {
		return nil, errors.New("missing logical member")
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}
func publishTree(directory, relative string, members map[string][]byte) error {
	return publication.PublishTree(directory, relative, members)
}
func logicalMemberInventory(members map[string][]byte) []envelope.File {
	return publication.LogicalMemberInventory(members)
}
func validRelative(p string) bool {
	return p != "" && p == path.Clean(p) && p != "." && !strings.HasPrefix(p, "/") && !strings.Contains(p, "\\") && p != ".." && !strings.HasPrefix(p, "../")
}
func noSymlinks(root *os.Root, name string, allowMissing bool) error {
	parts := strings.Split(name, "/")
	for i := range parts {
		st, err := root.Lstat(strings.Join(parts[:i+1], "/"))
		if errors.Is(err, os.ErrNotExist) && allowMissing {
			return nil
		}
		if err != nil {
			return err
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return errors.New("symlinked evidence path")
		}
	}
	return nil
}
func ReadEvidence(root *os.Root, file EvidenceFile) ([]byte, error) {
	if !validRelative(file.Path) || !digestPattern.MatchString(file.SHA256) || file.Size < 1 || file.Size > 128<<20 {
		return nil, errors.New("invalid evidence descriptor")
	}
	if err := noSymlinks(root, file.Path, false); err != nil {
		return nil, err
	}
	st, err := root.Lstat(file.Path)
	if err != nil || !st.Mode().IsRegular() || st.Size() != file.Size {
		return nil, errors.New("missing or nonregular source evidence")
	}
	f, err := root.Open(file.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, file.Size+1))
	if err != nil || int64(len(b)) != file.Size || Digest(b) != file.SHA256 {
		return nil, errors.New("source hash or length mismatch")
	}
	return b, nil
}

type Builder struct {
	ID        string `json:"builder_id"`
	Version   string `json:"builder_version"`
	SHA256    string `json:"builder_source_sha256"`
	Size      int64  `json:"size_bytes"`
	GoVersion string `json:"go_version"`
}

func BuilderIdentity() (Builder, error) {
	// The Linux producer identifies the running image, even if its pathname
	// has been replaced or unlinked since launch.
	file, err := os.Open("/proc/self/exe")
	if err != nil {
		return Builder{}, err
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		return Builder{}, err
	}
	version := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		version = info.GoVersion
	}
	return Builder{ID: "cartulary.reference-pack.entra", Version: "dataset-v1", SHA256: Digest(b), Size: int64(len(b)), GoVersion: version}, nil
}
