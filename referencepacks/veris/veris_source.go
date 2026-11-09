package producer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	verisinputs "github.com/JochiRaider/cartulary/referencepacks/veris/inputs"
)

const VERISKey = "framework.veris"
const VERISSourceProfile = "cartulary.veris.community.dataset.v1"
const VERISRepository = "cartulary.veris.community"
const VERISFixtureRepository = DatasetFixtureRepository

type VERISSource struct {
	Path      string `json:"path"`
	URL       string `json:"url"`
	SHA256    string `json:"sha256"`
	Size      int64  `json:"size_bytes"`
	MediaType string `json:"media_type"`
	Role      string `json:"role"`
}
type VERISInventory struct {
	SchemaID string        `json:"schema_id"`
	Commit   string        `json:"commit"`
	Version  string        `json:"source_version"`
	AsOf     string        `json:"source_as_of"`
	Sources  []VERISSource `json:"sources"`
}
type VERISAcquired struct {
	Source      VERISSource  `json:"source"`
	Evidence    EvidenceFile `json:"evidence"`
	RetrievedAt string       `json:"retrieved_at"`
	ContentType string       `json:"content_type"`
}
type VERISLock struct {
	SchemaID        string          `json:"schema_id"`
	InventorySHA256 string          `json:"inventory_sha256"`
	ObservedAt      string          `json:"source_set_observed_at"`
	Sources         []VERISAcquired `json:"sources"`
}
type VERISDownload struct {
	Bytes                    []byte
	RetrievedAt, ContentType string
}
type VERISFetcher interface {
	FetchVERIS(context.Context, VERISSource) (VERISDownload, error)
}
type VERISHTTPFetcher struct{}

func verisInventory(b []byte) (VERISInventory, error) {
	var result VERISInventory
	if err := DecodeClosed(b, &result); err != nil {
		return result, err
	}
	approved, err := verisinputs.Files.ReadFile("inventory.v1.json")
	if err != nil {
		return result, err
	}
	if !bytes.Equal(b, approved) {
		return result, errors.New("VERIS inventory differs from frozen source profile")
	}
	return result, nil
}

func (VERISHTTPFetcher) FetchVERIS(ctx context.Context, source VERISSource) (VERISDownload, error) {
	b, err := verisinputs.Files.ReadFile("inventory.v1.json")
	if err != nil {
		return VERISDownload{}, err
	}
	inventory, err := verisInventory(b)
	if err != nil {
		return VERISDownload{}, err
	}
	allowed := false
	for _, pin := range inventory.Sources {
		allowed = allowed || pin == source
	}
	if !allowed {
		return VERISDownload{}, errors.New("unlisted VERIS source")
	}
	client := &http.Client{Timeout: 2 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("VERIS redirects forbidden") }}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return VERISDownload{}, err
	}
	response, err := client.Do(req)
	if err != nil {
		return VERISDownload{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return VERISDownload{}, fmt.Errorf("VERIS source HTTP status %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, source.Size+1))
	if err != nil {
		return VERISDownload{}, err
	}
	d := VERISDownload{data, time.Now().UTC().Truncate(time.Second).Format(time.RFC3339), response.Header.Get("Content-Type")}
	if err := verisDownloadValid(source, d); err != nil {
		return VERISDownload{}, err
	}
	return d, nil
}

func verisDownloadValid(s VERISSource, d VERISDownload) error {
	media, _, err := mime.ParseMediaType(d.ContentType)
	if err != nil || media != s.MediaType || int64(len(d.Bytes)) != s.Size || Digest(d.Bytes) != s.SHA256 || !validBuildTime(d.RetrievedAt) {
		return fmt.Errorf("VERIS source identity or receipt mismatch: %s", s.Path)
	}
	return nil
}

func VERISAcquire(ctx context.Context, root *os.Root, raw []byte, fetch VERISFetcher) (VERISLock, error) {
	inventory, err := verisInventory(raw)
	if err != nil {
		return VERISLock{}, err
	}
	if _, err := root.Lstat("source-lock.json"); err == nil {
		return VERISCheck(root, raw)
	} else if !errors.Is(err, os.ErrNotExist) {
		return VERISLock{}, err
	}
	lock := VERISLock{SchemaID: "cartulary.veris_source_lock.v1", InventorySHA256: Digest(raw), Sources: []VERISAcquired{}}
	for _, source := range inventory.Sources {
		if err := ctx.Err(); err != nil {
			return lock, err
		}
		download, err := fetch.FetchVERIS(ctx, source)
		if err != nil {
			return lock, err
		}
		if err := verisDownloadValid(source, download); err != nil {
			return lock, err
		}
		name := "sources/" + source.Path
		if err := Retain(root, name, download.Bytes, 0600); err != nil {
			return lock, err
		}
		lock.Sources = append(lock.Sources, VERISAcquired{source, EvidenceFile{name, source.SHA256, source.Size}, download.RetrievedAt, download.ContentType})
	}
	lock.ObservedAt = time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
	for _, s := range lock.Sources {
		if s.RetrievedAt > lock.ObservedAt {
			return lock, errors.New("future VERIS receipt")
		}
	}
	b, err := canonicaljson.Marshal(lock)
	if err != nil {
		return lock, err
	}
	if err := Retain(root, "inventory.json", raw, 0600); err != nil {
		return lock, err
	}
	return lock, Retain(root, "source-lock.json", b, 0600)
}

func VERISCheck(root *os.Root, raw []byte) (VERISLock, error) {
	inventory, err := verisInventory(raw)
	if err != nil {
		return VERISLock{}, err
	}
	if _, err := ReadEvidence(root, EvidenceFile{"inventory.json", Digest(raw), int64(len(raw))}); err != nil {
		return VERISLock{}, err
	}
	if err := noSymlinks(root, "source-lock.json", false); err != nil {
		return VERISLock{}, err
	}
	f, err := root.Open("source-lock.json")
	if err != nil {
		return VERISLock{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return VERISLock{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > 65536 {
		return VERISLock{}, errors.New("bounded regular VERIS source lock required")
	}
	b, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil {
		return VERISLock{}, err
	}
	var lock VERISLock
	if err := DecodeClosed(b, &lock); err != nil {
		return lock, err
	}
	if len(b) > 65536 || lock.SchemaID != "cartulary.veris_source_lock.v1" || lock.InventorySHA256 != Digest(raw) || !validBuildTime(lock.ObservedAt) || len(lock.Sources) != len(inventory.Sources) {
		return lock, errors.New("invalid VERIS source lock")
	}
	for i, s := range lock.Sources {
		if s.Source != inventory.Sources[i] || s.Evidence != (EvidenceFile{"sources/" + s.Source.Path, s.Source.SHA256, s.Source.Size}) || s.RetrievedAt > lock.ObservedAt {
			return lock, errors.New("VERIS source pin mismatch")
		}
		b, err := ReadEvidence(root, s.Evidence)
		if err != nil {
			return lock, err
		}
		if err := verisDownloadValid(s.Source, VERISDownload{b, s.RetrievedAt, s.ContentType}); err != nil {
			return lock, err
		}
	}
	expected := map[string]bool{}
	for _, source := range lock.Sources {
		expected[source.Evidence.Path] = true
	}
	if err := fs.WalkDir(root.FS(), "sources", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() || !expected[name] {
			return errors.New("undeclared VERIS source member")
		}
		delete(expected, name)
		return nil
	}); err != nil {
		return lock, err
	}
	if len(expected) != 0 {
		return lock, errors.New("missing VERIS source member")
	}
	return lock, nil
}
