package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"golang.org/x/net/html"
)

type WindowsDownload struct {
	Bytes         []byte
	URL           string
	ContentType   string
	RetrievedAt   string
	ReceivedAt    string
	Method        string
	OriginReceipt *string
}
type WindowsFetcher interface {
	FetchWindows(context.Context, string) (WindowsDownload, error)
}
type WindowsHTTPFetcher struct{ Client *http.Client }

func windowsURL(location string) bool {
	u, err := url.Parse(location)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Fragment != "" {
		return false
	}
	if strings.HasSuffix(strings.ToLower(u.Path), ".md") || strings.Contains(strings.ToLower(u.Path), "readme") {
		return false
	}
	switch u.Hostname() {
	case "learn.microsoft.com":
		return strings.HasPrefix(u.Path, "/en-us/") && !strings.Contains(u.Path, "/answers/")
	case "raw.githubusercontent.com":
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		return len(parts) >= 4 && commitPattern.MatchString(parts[2]) && ((parts[0] == "PowerShell" && parts[1] == "PowerShell") || (len(parts) == 4 && parts[0] == "microsoft" && parts[1] == "SysmonCommon" && slices.Contains([]string{"manifest.xml", "manifest.tt", "LICENSE"}, parts[3])))
	case "api.github.com":
		return strings.HasPrefix(u.Path, "/repos/PowerShell/PowerShell/releases/tags/") || strings.HasPrefix(u.Path, "/repos/PowerShell/PowerShell/git/tags/") || strings.HasPrefix(u.Path, "/repos/PowerShell/PowerShell/git/ref/tags/")
	}
	return false
}
func (f WindowsHTTPFetcher) FetchWindows(ctx context.Context, location string) (WindowsDownload, error) {
	if !windowsURL(location) {
		return WindowsDownload{}, windowsFailure("source_url_rejected", "", "first-party HTTPS source required")
	}
	client := http.Client{Timeout: 120 * time.Second}
	if f.Client != nil {
		client = *f.Client
		client.Timeout = 120 * time.Second
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 || !windowsURL(req.URL.String()) {
			return errors.New("source redirect rejected")
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return WindowsDownload{}, err
	}
	req.Header.Set("User-Agent", "Cartulary-Windows-Event-Source/1")
	response, err := client.Do(req)
	if err != nil {
		return WindowsDownload{}, windowsFailure("acquisition_failed", "", "HTTP acquisition failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return WindowsDownload{}, windowsFailure("acquisition_failed", "", response.Status)
	}
	b, err := io.ReadAll(io.LimitReader(response.Body, windowsArtifactLimit+1))
	if err != nil || int64(len(b)) > windowsArtifactLimit {
		return WindowsDownload{}, windowsFailure("source_overflow", "", "artifact byte limit")
	}
	if len(b) == 0 || !utf8.Valid(b) {
		return WindowsDownload{}, windowsFailure("input_invalid", "", "nonempty UTF-8 source required")
	}
	return WindowsDownload{Bytes: b, URL: response.Request.URL.String(), ContentType: response.Header.Get("Content-Type"), RetrievedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}
func retainWindowsDownload(root *os.Root, s WindowsSource, d WindowsDownload) (WindowsAcquired, error) {
	if d.Method == "" {
		d.Method = "https"
		d.ReceivedAt = d.RetrievedAt
	}
	var retrieved *string
	if d.RetrievedAt != "" {
		retrieved = &d.RetrievedAt
	}
	a := WindowsAcquired{Source: s, ResolvedURL: d.URL, RetrievedAt: retrieved, ReceivedAt: d.ReceivedAt, Method: d.Method, OriginReceipt: d.OriginReceipt, ContentType: d.ContentType, Evidence: EvidenceFile{Path: "sources/" + Digest(d.Bytes) + "/source", SHA256: Digest(d.Bytes), Size: int64(len(d.Bytes))}}
	if !validWindowsAcquired(a) || d.ContentType == "" || (!windowsBinary(s.Kind) && !utf8.Valid(d.Bytes)) || len(d.Bytes) == 0 || int64(len(d.Bytes)) > windowsArtifactLimit {
		return a, windowsFailure("input_invalid", s.ID, "invalid acquisition response")
	}
	if s.ExpectedSHA256 != nil && *s.ExpectedSHA256 != a.Evidence.SHA256 {
		return a, windowsFailure("source_lock_mismatch", s.ID, "expected source digest differs")
	}
	if err := Retain(root, a.Evidence.Path, d.Bytes, 0600); err != nil {
		return a, err
	}
	return a, nil
}

// WindowsAcquire never refreshes an accepted receipt. A new acquisition uses a
// new root. Failed attempts retain diagnostics without publishing source-lock.json.
func WindowsAcquire(ctx context.Context, root *os.Root, inventoryBytes []byte, fetch WindowsFetcher) (WindowsAcquisitionReport, error) {
	inv, err := windowsInventory(inventoryBytes)
	if err != nil {
		return WindowsAcquisitionReport{}, err
	}
	r := WindowsAcquisitionReport{"cartulary.windows_event_acquisition.v2", "failed", Digest(inventoryBytes), nil, []WindowsAcquired{}, []WindowsError{}}
	if _, err = root.Lstat("source-lock.json"); err == nil {
		l, e := WindowsCheck(root, inventoryBytes)
		if e != nil {
			return r, e
		}
		b, e := root.ReadFile("source-lock.json")
		if e != nil {
			return r, e
		}
		d := Digest(b)
		r.Status = "complete"
		r.Sources = l.Sources
		r.LockSHA256 = &d
		return r, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return r, err
	}
	if err = Retain(root, "inventory.json", inventoryBytes, 0600); err != nil {
		return r, err
	}
	var total int64
	for _, s := range inv.Sources {
		if err = ctx.Err(); err != nil {
			r.Issues = append(r.Issues, WindowsError{"acquisition_failed", s.ID, "not acquired: context canceled"})
			continue
		}
		if total > windowsTotalLimit {
			r.Issues = append(r.Issues, WindowsError{"source_overflow", s.ID, "not acquired: total source byte limit"})
			continue
		}
		d, e := fetch.FetchWindows(ctx, s.URL)
		if e != nil {
			r.Issues = append(r.Issues, windowsIssue(e, s.ID))
			continue
		}
		total += int64(len(d.Bytes))
		if total > windowsTotalLimit {
			r.Issues = append(r.Issues, WindowsError{"source_overflow", s.ID, "total source byte limit"})
			continue
		}
		a, e := retainWindowsDownload(root, s, d)
		if e != nil {
			r.Issues = append(r.Issues, windowsIssue(e, s.ID))
			continue
		}
		r.Sources = append(r.Sources, a)
	}
	if err != nil {
		r.Issues = append(r.Issues, WindowsError{"acquisition_failed", "", "acquisition canceled"})
	}
	if len(r.Issues) == 0 && len(r.Sources) == len(inv.Sources) {
		lock := WindowsLock{"cartulary.windows_event_source_lock.v2", Digest(inventoryBytes), inv.Purpose, r.Sources}
		if inv.Purpose == "real" {
			if e := verifyWindowsReleasePins(root, lock); e != nil {
				r.Issues = append(r.Issues, windowsIssue(e, "release_pin"))
			}
		}
		if len(r.Issues) == 0 {
			b, e := canonicaljson.Marshal(lock)
			if e != nil {
				return r, e
			}
			if e = Retain(root, "source-lock.json", b, 0600); e != nil {
				return r, e
			}
			d := Digest(b)
			r.LockSHA256 = &d
			r.Status = "complete"
		}
	}
	b, e := canonicaljson.Marshal(r)
	if e != nil {
		return r, e
	}
	if e = Retain(root, "acquisitions/"+Digest(b)+".json", b, 0600); e != nil {
		return r, e
	}
	if r.Status != "complete" {
		return r, windowsFailure("acquisition_incomplete", "", "see retained acquisition report")
	}
	return r, nil
}
func WindowsCheck(root *os.Root, inventoryBytes []byte) (WindowsLock, error) {
	inv, err := windowsInventory(inventoryBytes)
	if err != nil {
		return WindowsLock{}, err
	}
	for _, name := range []string{"source-lock.json", "inventory.json"} {
		if err = noSymlinks(root, name, false); err != nil {
			return WindowsLock{}, err
		}
	}
	b, err := root.ReadFile("source-lock.json")
	if err != nil {
		return WindowsLock{}, err
	}
	var l WindowsLock
	if err = windowsDecode(b, "lock", &l); err != nil {
		return l, err
	}
	if l.InventorySHA256 != Digest(inventoryBytes) || l.Purpose != inv.Purpose || len(l.Sources) != len(inv.Sources) {
		return l, windowsFailure("source_lock_mismatch", "", "inventory differs")
	}
	retained, err := root.ReadFile("inventory.json")
	if err != nil || !bytes.Equal(retained, inventoryBytes) {
		return l, windowsFailure("source_lock_mismatch", "", "retained inventory differs")
	}
	var total int64
	for i, a := range l.Sources {
		want, _ := canonicaljson.Marshal(inv.Sources[i])
		got, _ := canonicaljson.Marshal(a.Source)
		if !bytes.Equal(want, got) || !validWindowsAcquired(a) {
			return l, windowsFailure("source_lock_mismatch", a.Source.ID, "source binding differs")
		}
		total += a.Evidence.Size
		if a.Evidence.Size < 1 || a.Evidence.Size > windowsArtifactLimit || total > windowsTotalLimit {
			return l, windowsFailure("source_overflow", a.Source.ID, "locked byte budget")
		}
		if a.Evidence.Path != "sources/"+a.Evidence.SHA256+"/source" {
			return l, windowsFailure("source_lock_mismatch", a.Source.ID, "content-addressed member required")
		}
		if _, err = ReadEvidence(root, a.Evidence); err != nil {
			return l, windowsFailure("source_lock_mismatch", a.Source.ID, "retained source bytes differ")
		}
		if a.Source.ExpectedSHA256 != nil && *a.Source.ExpectedSHA256 != a.Evidence.SHA256 {
			return l, windowsFailure("source_lock_mismatch", a.Source.ID, "expected digest differs")
		}
	}
	if l.Purpose == "real" {
		if err = verifyWindowsReleasePins(root, l); err != nil {
			return l, err
		}
	}
	return l, nil
}

// GitHub's release target_commitish may be a branch. Only the retained annotated
// tag's commit object establishes the release-to-manifest revision binding.
func verifyWindowsReleasePins(root *os.Root, lock WindowsLock) error {
	tags, releases := map[string]string{}, map[string]bool{}
	for _, a := range lock.Sources {
		if a.Source.Kind != "release_metadata" {
			continue
		}
		b, e := ReadEvidence(root, a.Evidence)
		if e != nil {
			return e
		}
		var v struct {
			SHA     string `json:"sha"`
			Tag     string `json:"tag"`
			TagName string `json:"tag_name"`
			Object  struct {
				SHA  string `json:"sha"`
				Type string `json:"type"`
			} `json:"object"`
		}
		if _, e = canonicaljson.DecodeStrict(b); e != nil {
			return e
		}
		if e = json.Unmarshal(b, &v); e != nil {
			return e
		}
		u, _ := url.Parse(a.Source.URL)
		if strings.Contains(u.Path, "/git/tags/") {
			if v.SHA != path.Base(u.Path) || v.Object.Type != "commit" || a.Source.Revision == nil || v.Object.SHA != *a.Source.Revision || v.Tag == "" {
				return windowsFailure("source_lock_mismatch", a.Source.ID, "annotated release tag does not bind the declared commit")
			}
			tags[v.Object.SHA] = v.Tag
		} else if strings.Contains(u.Path, "/releases/tags/") {
			if v.TagName != path.Base(u.Path) {
				return windowsFailure("source_lock_mismatch", a.Source.ID, "release tag differs")
			}
			releases[v.TagName] = true
		}
	}
	for _, a := range lock.Sources {
		if a.Source.Kind != "powershell_manifest" {
			continue
		}
		u, _ := url.Parse(a.Source.URL)
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if u.Hostname() != "raw.githubusercontent.com" || len(parts) < 4 || a.Source.Revision == nil || parts[2] != *a.Source.Revision || !releases[tags[parts[2]]] {
			return windowsFailure("source_lock_mismatch", a.Source.ID, "release-pinned manifest requires retained release and annotated commit tag")
		}
	}
	return nil
}

var auditEventPath = regexp.MustCompile(`^event-[0-9]+$`)

const auditDirectory = "/en-us/previous-versions/windows/it-pro/windows-10/security/threat-protection/auditing/"

// WindowsDiscover walks only the declared auditing index/subcategory pages.
// Every in-directory link is classified; event pages are acquired separately.
func WindowsDiscover(ctx context.Context, root *os.Root, seedBytes []byte, fetch WindowsFetcher) (WindowsInventory, error) {
	inv, err := windowsInventory(seedBytes)
	if err != nil {
		return inv, err
	}
	byURL := map[string]WindowsSource{}
	queue := []WindowsSource{}
	for _, s := range inv.Sources {
		byURL[s.URL] = s
		if s.Kind == "audit_index" {
			queue = append(queue, s)
		}
	}
	inv.Discovery = []WindowsDiscovery{}
	var total int64
	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]
		d, e := fetch.FetchWindows(ctx, s.URL)
		if e != nil {
			return inv, e
		}
		total += int64(len(d.Bytes))
		if total > windowsTotalLimit {
			return inv, windowsFailure("source_overflow", s.ID, "discovery byte budget")
		}
		a, e := retainWindowsDownload(root, s, d)
		if e != nil {
			return inv, e
		}
		s.ExpectedSHA256 = &a.Evidence.SHA256
		byURL[s.URL] = s
		doc, e := html.Parse(bytes.NewReader(d.Bytes))
		if e != nil {
			return inv, e
		}
		base, _ := url.Parse(d.URL)
		links := map[string]bool{}
		walkHTML(doc, func(n *html.Node) {
			if n.Type != html.ElementNode || n.Data != "a" {
				return
			}
			for _, v := range n.Attr {
				if v.Key == "href" {
					u, e := base.Parse(v.Val)
					if e != nil {
						return
					}
					u.Fragment = ""
					if u.Host == base.Host && strings.HasPrefix(u.Path, auditDirectory) {
						u.RawQuery = ""
						links[u.String()] = true
					}
				}
			}
		})
		urls := []string{}
		for u := range links {
			urls = append(urls, u)
		}
		slices.Sort(urls)
		for _, location := range urls {
			u, _ := url.Parse(location)
			name := path.Base(u.Path)
			kind, disposition := "", "excluded_non_definition_navigation"
			if auditEventPath.MatchString(name) {
				kind, disposition = "event_document", "included_event_document"
			} else if strings.HasPrefix(name, "audit-") {
				kind, disposition = "audit_index", "included_index"
			}
			inv.Discovery = append(inv.Discovery, WindowsDiscovery{s.ID, location, disposition})
			if kind == "" {
				continue
			}
			if _, ok := byURL[location]; ok {
				continue
			}
			added := WindowsSource{"discovered-" + Digest([]byte(location))[:16], location, kind, nil, nil}
			byURL[location] = added
			if len(byURL) > windowsSourceLimit {
				return inv, windowsFailure("source_overflow", s.ID, "discovery source limit")
			}
			if kind == "audit_index" {
				queue = append(queue, added)
			}
		}
	}
	inv.Sources = []WindowsSource{}
	for _, s := range byURL {
		inv.Sources = append(inv.Sources, s)
	}
	slices.SortFunc(inv.Sources, func(a, b WindowsSource) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(inv.Discovery, func(a, b WindowsDiscovery) int { return strings.Compare(a.Parent+"\x00"+a.URL, b.Parent+"\x00"+b.URL) })
	b, err := canonicaljson.Marshal(inv)
	if err != nil {
		return inv, err
	}
	if err = Retain(root, "inventories/"+Digest(b)+".json", b, 0600); err != nil {
		return inv, err
	}
	return inv, nil
}
func walkHTML(n *html.Node, f func(*html.Node)) {
	f(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkHTML(c, f)
	}
}
