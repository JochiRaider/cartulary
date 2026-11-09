package producer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	entrainputs "github.com/JochiRaider/cartulary/referencepacks/entra/inputs"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

const EntraKey = "enrichment.entra_app_ids"
const EntraContentProfile = "cartulary.reference_pack.enrichment.entra_app_ids.v2"
const EntraSourceProfile = "cartulary.entra_app_ids.microsoft_docs.2026_11_06.dataset.v1"
const EntraRepository = "cartulary.entra_app_ids"
const EntraFixtureRepository = DatasetFixtureRepository

type EntraSource struct {
	ID              string   `json:"source_id"`
	Kind            string   `json:"kind"`
	URL             string   `json:"url"`
	Repository      string   `json:"repository"`
	Commit          string   `json:"commit"`
	CommitAt        string   `json:"commit_at"`
	FrontMatterDate *string  `json:"front_matter_date"`
	SHA256          string   `json:"sha256"`
	Size            int64    `json:"size_bytes"`
	Heading         *string  `json:"heading"`
	Header          []string `json:"header"`
	Rows            int      `json:"expected_rows"`
	Assertions      int      `json:"expected_assertions"`
	LicenseID       *string  `json:"license_input_id"`
}
type EntraInventory struct {
	SchemaID string        `json:"schema_id"`
	Sources  []EntraSource `json:"sources"`
}
type EntraAcquired struct {
	Source      EntraSource  `json:"source"`
	Evidence    EvidenceFile `json:"evidence"`
	RetrievedAt string       `json:"retrieved_at"`
	ContentType string       `json:"content_type"`
}
type EntraLock struct {
	SchemaID        string          `json:"schema_id"`
	InventorySHA256 string          `json:"inventory_sha256"`
	ObservedAt      string          `json:"source_set_observed_at"`
	Sources         []EntraAcquired `json:"sources"`
}
type EntraDownload struct {
	Bytes                    []byte
	RetrievedAt, ContentType string
}
type EntraFetcher interface {
	FetchEntra(context.Context, EntraSource) (EntraDownload, error)
}
type EntraHTTPFetcher struct{}

func entraInventory(b []byte) (EntraInventory, error) {
	var result EntraInventory
	if err := DecodeClosed(b, &result); err != nil {
		return result, err
	}
	approved, err := entrainputs.Files.ReadFile("inventory.v2.json")
	if err != nil {
		return result, err
	}
	if !bytes.Equal(b, approved) {
		return result, errors.New("entra inventory differs from the immutable source profile")
	}
	return result, nil
}

// FetchEntra never discovers URLs or follows redirects/includes. Every response
// is bounded before hashing. A server-requested delay is honored or rejected,
// never shortened; cancellation also bounds the retry wait.
func (EntraHTTPFetcher) FetchEntra(ctx context.Context, source EntraSource) (EntraDownload, error) {
	b, err := entrainputs.Files.ReadFile("inventory.v2.json")
	if err != nil {
		return EntraDownload{}, err
	}
	inventory, err := entraInventory(b)
	if err != nil {
		return EntraDownload{}, err
	}
	allowed := false
	for _, pin := range inventory.Sources {
		if pin.ID == source.ID && pin.URL == source.URL && pin.SHA256 == source.SHA256 && pin.Size == source.Size {
			allowed = true
		}
	}
	if !allowed {
		return EntraDownload{}, errors.New("unlisted Entra acquisition source")
	}
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return errors.New("entra source redirects are not admitted")
	}}
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
		if e != nil {
			return EntraDownload{}, e
		}
		response, e := client.Do(req)
		wait := time.Duration(attempt+1) * time.Second
		if e != nil {
			last = e
		} else {
			data, readErr := io.ReadAll(io.LimitReader(response.Body, source.Size+1))
			closeErr := response.Body.Close()
			if response.StatusCode == http.StatusOK {
				if readErr != nil {
					last = readErr
				} else if closeErr != nil {
					last = closeErr
				} else {
					if int64(len(data)) != source.Size || Digest(data) != source.SHA256 || !utf8.Valid(data) {
						return EntraDownload{}, fmt.Errorf("entra source %s byte identity mismatch", source.ID)
					}
					return EntraDownload{data, time.Now().UTC().Truncate(time.Second).Format(time.RFC3339), response.Header.Get("Content-Type")}, nil
				}
			} else if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 {
				last = fmt.Errorf("entra source %s HTTP status %d", source.ID, response.StatusCode)
				if value := response.Header.Get("Retry-After"); value != "" {
					wait, e = entraRetryDelay(value, time.Now())
					if e != nil {
						return EntraDownload{}, e
					}
				}
			} else {
				return EntraDownload{}, fmt.Errorf("entra source %s HTTP status %d", source.ID, response.StatusCode)
			}
		}
		if attempt == 2 {
			break
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return EntraDownload{}, ctx.Err()
		case <-timer.C:
		}
	}
	return EntraDownload{}, last
}

func entraRetryDelay(value string, now time.Time) (time.Duration, error) {
	var wait time.Duration
	if n, err := strconv.ParseInt(value, 10, 64); err == nil && n >= 0 && n <= 300 {
		wait = time.Duration(n) * time.Second
	} else {
		at, e := http.ParseTime(value)
		if e != nil {
			return 0, errors.New("invalid or excessive Entra Retry-After")
		}
		wait = at.Sub(now)
	}
	if wait < 0 {
		wait = 0
	}
	if wait > 300*time.Second {
		return 0, errors.New("entra Retry-After exceeds acquisition budget")
	}
	return wait, nil
}

func EntraAcquire(ctx context.Context, root *os.Root, inventoryBytes []byte, fetch EntraFetcher) (EntraLock, error) {
	inventory, err := entraInventory(inventoryBytes)
	if err != nil {
		return EntraLock{}, err
	}
	if _, err := root.Lstat("source-lock.json"); err == nil {
		return EntraCheck(root, inventoryBytes)
	} else if !errors.Is(err, os.ErrNotExist) {
		return EntraLock{}, err
	}
	lock := EntraLock{SchemaID: "cartulary.entra_source_lock.v2", InventorySHA256: Digest(inventoryBytes), Sources: []EntraAcquired{}}
	for _, source := range inventory.Sources {
		download, e := fetch.FetchEntra(ctx, source)
		if e != nil {
			return EntraLock{}, e
		}
		if int64(len(download.Bytes)) != source.Size || Digest(download.Bytes) != source.SHA256 || !utf8.Valid(download.Bytes) || !validBuildTime(download.RetrievedAt) || download.ContentType == "" {
			return EntraLock{}, fmt.Errorf("entra source %s receipt mismatch", source.ID)
		}
		extension := ".txt"
		if source.Kind == "dataset" {
			extension = ".md"
		}
		name := "sources/" + source.SHA256 + "/" + source.ID + extension
		if e = Retain(root, name, download.Bytes, 0600); e != nil {
			return EntraLock{}, e
		}
		lock.Sources = append(lock.Sources, EntraAcquired{source, EvidenceFile{name, source.SHA256, source.Size}, download.RetrievedAt, download.ContentType})
	}
	lock.ObservedAt = time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
	for _, source := range lock.Sources {
		if source.RetrievedAt > lock.ObservedAt {
			return EntraLock{}, errors.New("entra acquisition receipt is in the future")
		}
	}
	encoded, err := canonicaljson.Marshal(lock)
	if err != nil {
		return EntraLock{}, err
	}
	if err = Retain(root, "inventory.json", inventoryBytes, 0600); err != nil {
		return EntraLock{}, err
	}
	if err = Retain(root, "source-lock.json", encoded, 0600); err != nil {
		return EntraLock{}, err
	}
	return lock, nil
}

func EntraCheck(root *os.Root, inventoryBytes []byte) (EntraLock, error) {
	inventory, err := entraInventory(inventoryBytes)
	if err != nil {
		return EntraLock{}, err
	}
	if err = noSymlinks(root, "source-lock.json", false); err != nil {
		return EntraLock{}, err
	}
	info, err := root.Lstat("source-lock.json")
	if err != nil {
		return EntraLock{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > 65536 {
		return EntraLock{}, errors.New("bounded regular Entra source lock required")
	}
	file, err := root.Open("source-lock.json")
	if err != nil {
		return EntraLock{}, err
	}
	b, err := io.ReadAll(io.LimitReader(file, 65537))
	closeErr := file.Close()
	if err != nil {
		return EntraLock{}, err
	}
	if closeErr != nil {
		return EntraLock{}, closeErr
	}
	if len(b) > 65536 {
		return EntraLock{}, errors.New("entra source lock exceeds bound")
	}

	var lock EntraLock
	if err = DecodeClosed(b, &lock); err != nil {
		return lock, err
	}
	if lock.SchemaID != "cartulary.entra_source_lock.v2" || lock.InventorySHA256 != Digest(inventoryBytes) || !validBuildTime(lock.ObservedAt) || len(lock.Sources) != len(inventory.Sources) {
		return lock, errors.New("invalid Entra source lock")
	}
	for i, source := range lock.Sources {
		got, e := canonicaljson.Marshal(source.Source)
		if e != nil {
			return lock, e
		}
		want, e := canonicaljson.Marshal(inventory.Sources[i])
		if e != nil {
			return lock, e
		}
		extension := ".txt"
		if source.Source.Kind == "dataset" {
			extension = ".md"
		}
		if !bytes.Equal(got, want) || source.Evidence.Path != "sources/"+source.Source.SHA256+"/"+source.Source.ID+extension || source.Evidence.SHA256 != source.Source.SHA256 || source.Evidence.Size != source.Source.Size || !validBuildTime(source.RetrievedAt) || source.RetrievedAt > lock.ObservedAt || strings.TrimSpace(source.ContentType) == "" {
			return lock, errors.New("entra locked source differs from its pin")
		}
		data, e := ReadEvidence(root, source.Evidence)
		if e != nil {
			return lock, e
		}
		if !utf8.Valid(data) {
			return lock, errors.New("invalid Entra source UTF-8")
		}
	}
	return lock, nil
}
