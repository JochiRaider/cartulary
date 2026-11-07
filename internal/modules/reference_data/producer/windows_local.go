package producer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func windowsBinary(kind string) bool { return kind == "original_docx" || kind == "original_xlsx" }
func windowsSourceURL(s WindowsSource) bool {
	if windowsURL(s.URL) {
		return true
	}
	return s.ExpectedSHA256 != nil && s.URL == "urn:cartulary:supplied:"+*s.ExpectedSHA256 && (s.Kind == "combined_csv" || windowsBinary(s.Kind) || s.Kind == "acquisition_receipt" || s.Kind == "license")
}
func validWindowsAcquired(a WindowsAcquired) bool {
	if !windowsSourceURL(a.Source) || !validBuildTime(a.ReceivedAt) {
		return false
	}
	if a.RetrievedAt != nil && (!validBuildTime(*a.RetrievedAt) || *a.RetrievedAt > a.ReceivedAt) {
		return false
	}
	switch a.Method {
	case "https":
		return windowsURL(a.ResolvedURL) && a.RetrievedAt != nil && a.OriginReceipt == nil
	case "retained":
		return windowsURL(a.ResolvedURL) && a.RetrievedAt != nil && a.OriginReceipt != nil && digestPattern.MatchString(*a.OriginReceipt)
	case "supplied":
		return a.ResolvedURL == a.Source.URL && strings.HasPrefix(a.Source.URL, "urn:cartulary:supplied:") && a.RetrievedAt == nil && a.OriginReceipt == nil
	}
	return false
}

type WindowsLocalInput struct {
	SchemaID string `json:"schema_id"`
	Files    []struct {
		SourceID    string  `json:"source_id"`
		Path        string  `json:"path"`
		ReceiptPath *string `json:"receipt_path"`
	} `json:"files"`
}

// Local paths are acquisition parameters, never embedded in the portable lock.
// A retained HTTP artifact must bind its metadata to a separately retained raw
// acquisition receipt. This reads source evidence, not a legacy pack format.
func WindowsLocalFetcher(inventoryBytes, inputBytes []byte, next WindowsFetcher) (WindowsFetcher, error) {
	inv, err := windowsInventory(inventoryBytes)
	if err != nil {
		return nil, err
	}
	var input WindowsLocalInput
	if err = windowsDecode(inputBytes, "acquisition_inputs", &input); err != nil {
		return nil, err
	}
	sources := map[string]WindowsSource{}
	for _, s := range inv.Sources {
		sources[s.ID] = s
	}
	f := &windowsLocalFetcher{next: next, values: map[string]WindowsDownload{}}
	seen := map[string]bool{}
	for _, item := range input.Files {
		s, ok := sources[item.SourceID]
		if !ok || seen[s.ID] || s.ExpectedSHA256 == nil {
			return nil, errors.New("local admission requires a unique inventory member with exact digest")
		}
		seen[s.ID] = true
		b, e := readWindowsLocal(item.Path)
		if e != nil {
			return nil, e
		}
		if Digest(b) != *s.ExpectedSHA256 {
			return nil, windowsFailure("source_lock_mismatch", s.ID, "supplied bytes differ")
		}
		d := WindowsDownload{Bytes: b, URL: s.URL, ContentType: windowsMedia(s.Kind), ReceivedAt: time.Now().UTC().Format(time.RFC3339), Method: "supplied"}
		if item.ReceiptPath != nil {
			raw, e := readWindowsLocal(*item.ReceiptPath)
			if e != nil {
				return nil, e
			}
			digest := Digest(raw)
			bound := false
			for _, other := range inv.Sources {
				if other.Kind == "acquisition_receipt" && other.ExpectedSHA256 != nil && *other.ExpectedSHA256 == digest {
					bound = true
				}
			}
			if !bound {
				return nil, errors.New("origin receipt must be included in frozen inventory")
			}
			value, e := canonicaljson.DecodeStrict(raw)
			if e != nil {
				return nil, e
			}
			doc, ok := value.(map[string]any)
			if !ok {
				return nil, errors.New("invalid source receipt")
			}
			members, ok := doc["sources"].([]any)
			if !ok {
				return nil, errors.New("source receipt lacks members")
			}
			matches := 0
			for _, member := range members {
				m, ok := member.(map[string]any)
				if !ok {
					continue
				}
				src, _ := m["source"].(map[string]any)
				file, _ := m["evidence"].(map[string]any)
				if src["url"] != s.URL || file["sha256"] != Digest(b) {
					continue
				}
				matches++
				d.URL, _ = m["resolved_url"].(string)
				d.RetrievedAt, _ = m["retrieved_at"].(string)
				d.ContentType, _ = m["content_type"].(string)
			}
			if matches != 1 {
				return nil, errors.New("retained origin must resolve uniquely")
			}
			d.Method = "retained"
			d.OriginReceipt = &digest
		}
		f.values[s.URL] = d
	}
	return f, nil
}

type windowsLocalFetcher struct {
	next   WindowsFetcher
	values map[string]WindowsDownload
}

func (f *windowsLocalFetcher) FetchWindows(ctx context.Context, url string) (WindowsDownload, error) {
	if e := ctx.Err(); e != nil {
		return WindowsDownload{}, e
	}
	if d, ok := f.values[url]; ok {
		return d, nil
	}
	if strings.HasPrefix(url, "urn:") {
		return WindowsDownload{}, errors.New("supplied source requires explicit local admission input")
	}
	return f.next.FetchWindows(ctx, url)
}
func readWindowsLocal(name string) ([]byte, error) {
	lower := strings.ToLower(filepath.ToSlash(name))
	if strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown") || strings.HasPrefix(strings.ToLower(filepath.Base(name)), "readme") || slices.Contains(strings.Split(lower, "/"), "docs") {
		return nil, errors.New("prose paths are not executable source inputs")
	}
	root, e := os.OpenRoot(filepath.Dir(name))
	if e != nil {
		return nil, e
	}
	defer root.Close()
	base := filepath.Base(name)
	if e = noSymlinks(root, base, false); e != nil {
		return nil, e
	}
	f, e := root.Open(base)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if !st.Mode().IsRegular() || st.Size() > windowsArtifactLimit {
		return nil, errors.New("bounded regular source required")
	}
	b, e := io.ReadAll(io.LimitReader(f, windowsArtifactLimit+1))
	if e != nil {
		return nil, e
	}
	if len(b) == 0 || int64(len(b)) > windowsArtifactLimit {
		return nil, errors.New("source byte budget")
	}
	return b, nil
}
func windowsMedia(kind string) string {
	switch kind {
	case "combined_csv":
		return "text/csv"
	case "original_docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "original_xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "acquisition_receipt":
		return "application/json"
	case "license":
		return "text/plain"
	}
	return fmt.Sprintf("application/x-cartulary-%s", kind)
}
