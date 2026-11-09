// Package producer owns offline ATT&CK source preparation. It is not imported
// by application runtime assembly and does not admit or publish reference packs.
package producer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

const LockSchema = "cartulary.attack_source_lock.v1"

type SourceLock struct {
	SchemaID         string           `json:"schema_id"`
	Repository       string           `json:"repository"`
	Tag              string           `json:"tag"`
	TagObject        string           `json:"tag_object"`
	Commit           string           `json:"commit"`
	CatalogVersion   string           `json:"catalog_version"`
	PublishedAt      string           `json:"published_at"`
	MajorReleaseDate string           `json:"major_release_date"`
	ResearchCutoff   string           `json:"research_cutoff"`
	ReleaseURL       string           `json:"release_url"`
	TagURL           string           `json:"tag_url"`
	Artifacts        []SourceArtifact `json:"artifacts"`
}

type SourceArtifact struct {
	Name         string  `json:"name"`
	URL          string  `json:"url"`
	SHA256       string  `json:"sha256"`
	Size         int64   `json:"size_bytes"`
	Domain       string  `json:"domain"`
	CollectionID *string `json:"collection_id"`
}

type EvidenceFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size_bytes"`
}

type Acquisition struct {
	SchemaID     string         `json:"schema_id"`
	LockSHA256   string         `json:"source_lock_sha256"`
	Release      EvidenceFile   `json:"release_metadata"`
	Tag          EvidenceFile   `json:"tag_metadata"`
	TagReference EvidenceFile   `json:"tag_reference"`
	Sources      []EvidenceFile `json:"sources"`
}

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func DecodeClosed(b []byte, target any) error {
	value, err := canonicaljson.DecodeStrict(b)
	if err != nil {
		return errors.New("invalid JSON structure")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return errors.New("invalid closed input shape")
	}
	// Every producer contract member is required, including explicit nulls.
	// Re-encoding detects omitted struct fields without materializing defaults.
	original, err := canonicaljson.Marshal(value)
	if err != nil {
		return err
	}
	decoded, err := canonicaljson.Marshal(target)
	if err != nil {
		return err
	}
	if !bytes.Equal(original, decoded) {
		return errors.New("missing required producer member")
	}
	return nil
}

func ReadLock(b []byte) (SourceLock, error) {
	var l SourceLock
	if err := DecodeClosed(b, &l); err != nil {
		return l, err
	}
	if l.SchemaID != LockSchema || l.Repository != "https://github.com/mitre-attack/attack-stix-data" || l.Tag != "v"+l.CatalogVersion || !commitPattern.MatchString(l.Commit) || !commitPattern.MatchString(l.TagObject) || len(l.Artifacts) != 4 {
		return l, errors.New("source lock identity invalid")
	}
	pub, err := time.Parse(time.RFC3339, l.PublishedAt)
	if err != nil || pub.UTC().Format(time.RFC3339) != l.PublishedAt {
		return l, errors.New("source publication time invalid")
	}
	for _, date := range []string{l.MajorReleaseDate, l.ResearchCutoff} {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return l, errors.New("source date invalid")
		}
	}
	if l.MajorReleaseDate > l.PublishedAt[:10] || l.PublishedAt[:10] > l.ResearchCutoff {
		return l, errors.New("source chronology invalid")
	}
	api := "https://api.github.com/repos/mitre-attack/attack-stix-data/"
	if l.ReleaseURL != api+"releases/tags/"+l.Tag || l.TagURL != api+"git/tags/"+l.TagObject {
		return l, errors.New("source metadata locator invalid")
	}
	domains := map[string]bool{}
	previous := ""
	for _, a := range l.Artifacts {
		if a.Name <= previous || path.Base(a.Name) != a.Name || !digestPattern.MatchString(a.SHA256) || a.Size < 1 || a.Size > 128<<20 || domains[a.Domain] {
			return l, errors.New("source artifact identity invalid")
		}
		previous = a.Name
		domains[a.Domain] = true
		if a.Domain == "license" {
			if a.Name != "LICENSE.txt" || a.CollectionID != nil || a.URL != "https://raw.githubusercontent.com/mitre-attack/attack-stix-data/"+l.Commit+"/LICENSE.txt" {
				return l, errors.New("license locator invalid")
			}
		} else if !slices.Contains([]string{"enterprise-attack", "mobile-attack", "ics-attack"}, a.Domain) || a.Name != a.Domain+".json" || a.CollectionID == nil || !strings.HasPrefix(*a.CollectionID, "x-mitre-collection--") || a.URL != l.Repository+"/releases/download/"+l.Tag+"/"+a.Name {
			return l, errors.New("domain locator invalid")
		}
	}
	if !domains["license"] || !domains["enterprise-attack"] || !domains["mobile-attack"] || !domains["ics-attack"] {
		return l, errors.New("all-domain source scope required")
	}
	return l, nil
}

// Only Acquire accepts a network dependency; checking and auditing never do.
type Fetcher interface {
	Fetch(context.Context, string, int64) ([]byte, error)
}

type HTTPFetcher struct{ Client *http.Client }

func (f HTTPFetcher) Fetch(ctx context.Context, location string, limit int64) ([]byte, error) {
	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 || !officialURL(req.URL) {
				return errors.New("unapproved source redirect")
			}
			return nil
		}}
	}
	u, err := url.Parse(location)
	if err != nil || !officialURL(u) {
		return nil, errors.New("unapproved source URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Cartulary-ATTACK-source-producer")
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.New("source download failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("source download HTTP %d", response.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(b)) > limit {
		return nil, errors.New("source download length invalid")
	}
	return b, nil
}
func officialURL(u *url.URL) bool {
	return u.Scheme == "https" && u.User == nil && u.Port() == "" && slices.Contains([]string{"github.com", "api.github.com", "raw.githubusercontent.com", "release-assets.githubusercontent.com", "objects.githubusercontent.com"}, u.Hostname())
}

// Retain uses no-replace publication. Existing differing bytes and symlinks are
// rejected. os.Root confines every relative operation to the selected store.
func Retain(root *os.Root, name string, b []byte, mode os.FileMode) error {
	if !validRelative(name) {
		return errors.New("invalid evidence path")
	}
	if err := noSymlinks(root, name, true); err != nil {
		return err
	}
	if st, err := root.Lstat(name); err == nil {
		if !st.Mode().IsRegular() {
			return errors.New("evidence is not a regular file")
		}
		old, err := root.ReadFile(name)
		if err != nil || !bytes.Equal(old, b) {
			return errors.New("immutable evidence collision")
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := root.MkdirAll(path.Dir(name), 0700); err != nil {
		return err
	}
	if err := noSymlinks(root, path.Dir(name), false); err != nil {
		return err
	}
	// The private temporary file is not an admitted artifact. Link publishes
	// atomically without replacing an existing name, including across producers.
	id, err := randomSuffix()
	if err != nil {
		return err
	}
	temp := path.Join(path.Dir(name), ".partial-"+id)
	f, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer root.Remove(temp)
	_, writeErr := f.Write(b)
	syncErr := f.Sync()
	closeErr := f.Close()
	if err = errors.Join(writeErr, syncErr, closeErr); err != nil {
		return err
	}
	if err = root.Link(temp, name); errors.Is(err, os.ErrExist) {
		return Retain(root, name, b, mode)
	}
	return err
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

func verifyMetadata(l SourceLock, releaseBytes, tagBytes, refBytes []byte) error {
	release, err := canonicaljson.DecodeStrict(releaseBytes)
	if err != nil {
		return errors.New("release metadata malformed")
	}
	r, ok := release.(map[string]any)
	if !ok {
		return errors.New("release metadata malformed")
	}
	if r["tag_name"] != l.Tag || r["draft"] != false || r["prerelease"] != false || r["published_at"] != l.PublishedAt {
		return errors.New("release identity mismatch")
	}
	tag, err := canonicaljson.DecodeStrict(tagBytes)
	if err != nil {
		return errors.New("tag metadata malformed")
	}
	t, ok := tag.(map[string]any)
	if !ok {
		return errors.New("tag metadata malformed")
	}
	obj, ok := t["object"].(map[string]any)
	if !ok || t["sha"] != l.TagObject || t["tag"] != l.Tag || obj["sha"] != l.Commit || obj["type"] != "commit" {
		return errors.New("tag commit mismatch")
	}
	reference, err := canonicaljson.DecodeStrict(refBytes)
	if err != nil {
		return errors.New("tag reference malformed")
	}
	ref, ok := reference.(map[string]any)
	if !ok {
		return errors.New("tag reference malformed")
	}
	target, ok := ref["object"].(map[string]any)
	if !ok || ref["ref"] != "refs/tags/"+l.Tag || target["sha"] != l.TagObject || target["type"] != "tag" {
		return errors.New("release tag moved")
	}
	assets, ok := r["assets"].([]any)
	if !ok {
		return errors.New("release assets missing")
	}
	for _, a := range l.Artifacts {
		if a.Domain == "license" {
			continue
		}
		matches := 0
		for _, item := range assets {
			asset, ok := item.(map[string]any)
			if !ok {
				return errors.New("release asset malformed")
			}
			if asset["name"] != a.Name {
				continue
			}
			matches++
			if asset["digest"] != "sha256:"+a.SHA256 || asset["browser_download_url"] != a.URL || asset["size"] != json.Number(fmt.Sprint(a.Size)) {
				return errors.New("release asset identity mismatch")
			}
		}
		if matches != 1 {
			return errors.New("release asset missing or duplicated")
		}
	}
	return nil
}

func Acquire(ctx context.Context, root *os.Root, lockBytes []byte, fetcher Fetcher) (Acquisition, error) {
	l, err := ReadLock(lockBytes)
	if err != nil {
		return Acquisition{}, err
	}
	if _, err = root.Lstat("acquisition.json"); err == nil {
		return Check(root, lockBytes)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Acquisition{}, err
	}
	if fetcher == nil {
		return Acquisition{}, errors.New("acquisition requires a fetcher")
	}
	release, err := fetcher.Fetch(ctx, l.ReleaseURL, 4<<20)
	if err != nil {
		return Acquisition{}, err
	}
	tag, err := fetcher.Fetch(ctx, l.TagURL, 1<<20)
	if err != nil {
		return Acquisition{}, err
	}
	ref, err := fetcher.Fetch(ctx, "https://api.github.com/repos/mitre-attack/attack-stix-data/git/ref/tags/"+l.Tag, 1<<20)
	if err != nil {
		return Acquisition{}, err
	}
	if err = verifyMetadata(l, release, tag, ref); err != nil {
		return Acquisition{}, err
	}
	a := Acquisition{SchemaID: "cartulary.attack_acquisition.v1", LockSHA256: Digest(lockBytes), Sources: []EvidenceFile{}}
	for _, item := range []struct {
		kind string
		b    []byte
		dst  *EvidenceFile
	}{{"release", release, &a.Release}, {"tag", tag, &a.Tag}, {"tag-ref", ref, &a.TagReference}} {
		*item.dst = EvidenceFile{Path: "metadata/" + item.kind + "-" + Digest(item.b) + ".json", SHA256: Digest(item.b), Size: int64(len(item.b))}
		if err = Retain(root, item.dst.Path, item.b, 0600); err != nil {
			return a, err
		}
	}
	for _, source := range l.Artifacts {
		file := EvidenceFile{Path: "sources/" + source.SHA256 + "/" + source.Name, SHA256: source.SHA256, Size: source.Size}
		_, err := ReadEvidence(root, file)
		if err != nil {
			if _, exists := root.Lstat(file.Path); !errors.Is(exists, os.ErrNotExist) {
				return a, err
			}
			b, err := fetcher.Fetch(ctx, source.URL, source.Size)
			if err != nil {
				return a, err
			}
			if int64(len(b)) != source.Size || Digest(b) != source.SHA256 {
				return a, errors.New("source hash or length mismatch")
			}
			if err = Retain(root, file.Path, b, 0600); err != nil {
				return a, err
			}
		}
		a.Sources = append(a.Sources, file)
	}
	encoded, err := canonicaljson.Marshal(a)
	if err != nil {
		return a, err
	}
	if err = Retain(root, "source-lock.json", lockBytes, 0600); err != nil {
		return a, err
	}
	if err = Retain(root, "acquisition.json", encoded, 0600); err != nil {
		return a, err
	}
	return Check(root, lockBytes)
}

func Check(root *os.Root, lockBytes []byte) (Acquisition, error) {
	l, err := ReadLock(lockBytes)
	if err != nil {
		return Acquisition{}, err
	}
	var a Acquisition
	if err := noSymlinks(root, "acquisition.json", false); err != nil {
		return a, err
	}
	st, err := root.Lstat("acquisition.json")
	if err != nil || !st.Mode().IsRegular() || st.Size() > 1<<20 {
		return a, errors.New("acquisition receipt missing or invalid")
	}
	b, err := root.ReadFile("acquisition.json")
	if err != nil {
		return a, err
	}
	if err = DecodeClosed(b, &a); err != nil {
		return a, err
	}
	if a.SchemaID != "cartulary.attack_acquisition.v1" || a.LockSHA256 != Digest(lockBytes) || len(a.Sources) != len(l.Artifacts) {
		return a, errors.New("acquisition lock mismatch")
	}
	retained, err := ReadEvidence(root, EvidenceFile{Path: "source-lock.json", SHA256: Digest(lockBytes), Size: int64(len(lockBytes))})
	if err != nil || !bytes.Equal(retained, lockBytes) {
		return a, errors.New("retained source lock mismatch")
	}
	release, err := ReadEvidence(root, a.Release)
	if err != nil {
		return a, err
	}
	tag, err := ReadEvidence(root, a.Tag)
	if err != nil {
		return a, err
	}
	ref, err := ReadEvidence(root, a.TagReference)
	if err != nil {
		return a, err
	}
	if err = verifyMetadata(l, release, tag, ref); err != nil {
		return a, err
	}
	for i, source := range l.Artifacts {
		file := a.Sources[i]
		if file.SHA256 != source.SHA256 || file.Size != source.Size || file.Path != "sources/"+source.SHA256+"/"+source.Name {
			return a, errors.New("acquisition inventory mismatch")
		}
		if _, err = ReadEvidence(root, file); err != nil {
			return a, err
		}
	}
	return a, nil
}
