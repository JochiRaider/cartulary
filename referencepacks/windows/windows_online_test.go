package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/publication"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	packformat "github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	windowsinputs "github.com/JochiRaider/cartulary/referencepacks/windows/inputs"
)

type windowsFixtureFetcher struct {
	Values map[string][]byte
	Fail   string
	Calls  int
}

type windowsRoundTrip func(*http.Request) (*http.Response, error)

func (f windowsRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func (f *windowsFixtureFetcher) FetchWindows(_ context.Context, url string) (WindowsDownload, error) {
	f.Calls++
	if url == f.Fail {
		return WindowsDownload{}, errors.New("synthetic failure")
	}
	b, ok := f.Values[url]
	if !ok {
		return WindowsDownload{}, errors.New("unlisted synthetic source")
	}
	return WindowsDownload{Bytes: b, URL: url, ContentType: "text/plain", RetrievedAt: "2026-11-06T00:00:00Z"}, nil
}
func windowsInput(t *testing.T, name string) []byte {
	t.Helper()
	b, e := windowsinputs.Files.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func syntheticWindowsManifest() string {
	return `<assembly><instrumentation><events><provider name="PowerShellCore" guid="{f90714a8-5509-434a-bf6d-b1624c8a19a2}"><events><event value="4104" version="1" level="zero" task="specific" opcode="create" template="payload" message="$(string.message)"/></events><levels><level name="zero" value="0"/></levels><tasks><task name="specific" value="20"><opcodes><opcode name="create" value="10"/></opcodes></task></tasks><opcodes><opcode name="create" value="9"/></opcodes><templates><template tid="payload"><data name="Text" inType="win:UnicodeString"/></template></templates></provider></events></instrumentation><localization><resources culture="en-US"><stringTable><string id="message" value="Synthetic message %1."/></stringTable></resources></localization></assembly>`
}
func syntheticEventDocument(provider string, id, version int) string {
	return fmtEventDocument(provider, id, version)
}
func fmtEventDocument(provider string, id, version int) string {
	return "<main><h1>" + strconvI(id) + " Synthetic documented event</h1><pre>&lt;Event&gt;&lt;System&gt;&lt;Provider Name=\"" + provider + "\"/&gt;&lt;EventID&gt;" + strconvI(id) + "&lt;/EventID&gt;&lt;Version&gt;" + strconvI(version) + "&lt;/Version&gt;&lt;Level&gt;0&lt;/Level&gt;&lt;Task&gt;20&lt;/Task&gt;&lt;Opcode&gt;10&lt;/Opcode&gt;&lt;/System&gt;&lt;/Event&gt;</pre><p>Event Versions:</p><ul><li>0 - documented old variant</li><li>1 - documented middle variant</li><li>2 - documented example variant</li></ul><p>Field Descriptions:</p></main>"
}
func strconvI(n int) string { b, _ := json.Marshal(n); return string(b) }
func windowsFixture(t *testing.T) (*os.Root, []byte, []byte, *windowsFixtureFetcher, Builder) {
	t.Helper()
	root := testRoot(t)
	inv := WindowsInventory{SchemaID: "cartulary.windows_event_inventory.v2", Purpose: "synthetic", Sources: []WindowsSource{}, Discovery: []WindowsDiscovery{}}
	fetch := &windowsFixtureFetcher{Values: map[string][]byte{}}
	for _, item := range []struct{ id, kind, body string }{
		{"security", "event_document", syntheticEventDocument("Microsoft-Windows-Security-Auditing", 4624, 2)},
		{"core", "powershell_manifest", syntheticWindowsManifest()},
		{"legacy", "provider_article", "<main><p>Microsoft-Windows-PowerShell</p><p>Event ID: 4104</p></main>"},
		{"license", "license", "Synthetic project-authored fixture text; not a Microsoft licensing decision.\n"},
	} {
		url := "https://learn.microsoft.com/en-us/windows/synthetic-" + item.id
		inv.Sources = append(inv.Sources, WindowsSource{item.id, url, item.kind, nil, nil})
		fetch.Values[url] = []byte(item.body)
	}
	b := encoded(t, inv)
	if _, e := WindowsAcquire(t.Context(), root, b, fetch); e != nil {
		t.Fatal(e)
	}
	builder := Builder{"fixture.windows.producer", "synthetic-v1", Digest([]byte("synthetic builder")), 17, "synthetic"}
	return root, b, windowsInput(t, "editorial.v2.json"), fetch, builder
}

func TestWindowsOnlineSources(t *testing.T) {
	t.Run("CSV input and presentation", testWindowsCSVReferences)
	t.Run("redirect and encoding boundary", func(t *testing.T) {
		calls := 0
		client := &http.Client{Transport: windowsRoundTrip(func(r *http.Request) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"https://example.test/source"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		})}
		if _, e := (WindowsHTTPFetcher{Client: client}).FetchWindows(t.Context(), "https://learn.microsoft.com/en-us/windows/synthetic"); e == nil || calls != 1 {
			t.Fatal("redirect escaped allowlist", e, calls)
		}
		client.Transport = windowsRoundTrip(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(bytes.NewReader([]byte{255})), Request: r}, nil
		})
		if _, e := (WindowsHTTPFetcher{Client: client}).FetchWindows(t.Context(), "https://learn.microsoft.com/en-us/windows/synthetic"); e == nil {
			t.Fatal("invalid UTF-8 repaired")
		}
	})
	t.Run("incomplete acquisition", func(t *testing.T) {
		_, inventory, _, fetch, _ := windowsFixture(t)
		root := testRoot(t)
		for url := range fetch.Values {
			fetch.Fail = url
			break
		}
		r, e := WindowsAcquire(t.Context(), root, inventory, fetch)
		if e == nil || r.Status != "failed" || len(r.Issues) != 1 {
			t.Fatal("incomplete capture accepted", r, e)
		}
		if _, e = root.Stat("source-lock.json"); !errors.Is(e, os.ErrNotExist) {
			t.Fatal("failed acquisition published lock")
		}
	})
	t.Run("URL boundary", func(t *testing.T) {
		for _, url := range []string{"http://learn.microsoft.com/en-us/windows/event", "https://learn.microsoft.com.evil.example/en-us/event", "https://learn.microsoft.com/en-us/answers/question", "https://learn.microsoft.com/en-us/file.md", "https://raw.githubusercontent.com/PowerShell/PowerShell/master/file.man", "https://user@learn.microsoft.com/en-us/windows/event", "https://learn.microsoft.com:444/en-us/windows/event", "https://raw.githubusercontent.com/untrusted/repo/0123456789012345678901234567890123456789/data"} {
			if windowsURL(url) {
				t.Fatalf("unsafe URL %s", url)
			}
		}
	})
	t.Run("closed contracts", func(t *testing.T) {
		root, inventory, editorial, _, builder := windowsFixture(t)
		out, e := WindowsReconcile(t.Context(), root, inventory, editorial, builder)
		if e != nil {
			t.Fatal(e)
		}
		schemas, e := windowsSchemas()
		if e != nil {
			t.Fatal(e)
		}
		for name, value := range map[string]any{"reconciliation": out.Report, "evidence": out.Evidence} {
			var v any
			if e = json.Unmarshal(encoded(t, value), &v); e != nil {
				t.Fatal(e)
			}
			if e = schemas[name].Validate(v); e != nil {
				t.Fatal(name, e)
			}
		}
		var v map[string]any
		if e = json.Unmarshal(inventory, &v); e != nil {
			t.Fatal(e)
		}
		v["extra"] = true
		if _, e = windowsInventory(encoded(t, v)); e == nil {
			t.Fatal("unknown member accepted")
		}
		delete(v, "extra")
		delete(v, "discovery")
		if _, e = windowsInventory(encoded(t, v)); e == nil {
			t.Fatal("missing required array accepted")
		}
	})
}

func TestWindowsOnlineProjection(t *testing.T) {
	t.Run("unqualified monitoring table", func(t *testing.T) {
		source := WindowsSource{ID: "monitoring", URL: "https://learn.microsoft.com/en-us/windows/synthetic-monitoring", Kind: "provider_article"}
		body := []byte(`<main><h1>Synthetic monitoring advice</h1><p>Event ID 1102 should be monitored.</p><table><tr><th>Current Windows Event ID</th><th>Legacy Windows Event ID</th><th>Potential Criticality</th></tr><tr><td>1102</td><td>517</td><td>High</td></tr><tr><td>4624</td><td>528,540</td><td>Low</td></tr></table><p>Example enumeration command: wevtutil gp Microsoft-Windows-Security-Auditing /ge</p></main>`)
		rows, dispositions, err := extractWindowsDocument(source, body)
		if err != nil || len(rows) != 0 || len(dispositions) != 2 {
			t.Fatal("monitoring table fabricated provider metadata or lost rows", rows, dispositions, err)
		}
		for _, d := range dispositions {
			if d.Disposition != "excluded" || d.Reason != "provider_not_stated_or_ambiguous_in_event_table" || d.EntryID != nil {
				t.Fatal("unqualified definition admitted", d)
			}
		}
	})
	root, inventory, editorial, _, builder := windowsFixture(t)
	out, err := WindowsReconcile(t.Context(), root, inventory, editorial, builder)
	if err != nil {
		t.Fatal(err)
	}
	again, err := WindowsReconcile(t.Context(), root, inventory, editorial, builder)
	if err != nil || !bytes.Equal(out.Payload, again.Payload) || !bytes.Equal(out.Archive, again.Archive) {
		t.Fatal("nondeterministic projection", err)
	}
	rows := map[string]map[string]any{}
	for _, line := range bytes.Split(bytes.TrimSuffix(out.Payload, []byte{'\n'}), []byte{'\n'}) {
		var v map[string]any
		if err = json.Unmarshal(line, &v); err != nil {
			t.Fatal(err)
		}
		rows[v["entry_id"].(string)] = v
	}
	expected := []string{"microsoft-windows-powershell:4104:*", "microsoft-windows-security-auditing:4624:0", "microsoft-windows-security-auditing:4624:1", "microsoft-windows-security-auditing:4624:2", "powershellcore:4104:1"}
	actual := []string{}
	for id := range rows {
		actual = append(actual, id)
	}
	slices.Sort(actual)
	if !slices.Equal(actual, expected) {
		t.Fatal("identity/version loss", actual)
	}
	for _, id := range expected[1:3] {
		if rows[id]["level"] != nil || rows[id]["task"] != nil || rows[id]["opcode"] != nil {
			t.Fatal("example fields leaked into another version")
		}
	}
	core := rows["powershellcore:4104:1"]
	if core["level"] != "0" || core["task"] != "20" || core["opcode"] != "10" {
		t.Fatal("numeric/scoped manifest projection", core)
	}
	t.Run("reordered source inventory", func(t *testing.T) {
		_, raw, ed, fetch, bld := windowsFixture(t)
		var inv WindowsInventory
		if e := DecodeClosed(raw, &inv); e != nil {
			t.Fatal(e)
		}
		slices.Reverse(inv.Sources)
		other := testRoot(t)
		if _, e := WindowsAcquire(t.Context(), other, encoded(t, inv), fetch); e != nil {
			t.Fatal(e)
		}
		reordered, e := WindowsReconcile(t.Context(), other, encoded(t, inv), ed, bld)
		if e != nil || !bytes.Equal(out.Payload, reordered.Payload) {
			t.Fatal("source order changed entries", e)
		}
	})
	t.Run("hostile XML", func(t *testing.T) {
		for _, input := range []string{`<!DOCTYPE assembly [<!ENTITY secret SYSTEM "file:///etc/passwd">]><assembly>&secret;</assembly>`, `<assembly>`, `<assembly/><assembly/>`} {
			if _, _, e := extractWindowsManifest(WindowsSource{ID: "hostile"}, []byte(input)); e == nil {
				t.Fatal("hostile XML accepted")
			}
		}
	})
	t.Run("descriptor range", func(t *testing.T) {
		for _, value := range []string{"-1", "65536", "1.5"} {
			b := strings.Replace(syntheticWindowsManifest(), `value="20"`, `value="`+value+`"`, 1)
			if _, _, e := extractWindowsManifest(WindowsSource{ID: "range"}, []byte(b)); e == nil {
				t.Fatal("bad task value accepted", value)
			}
		}
	})
	t.Run("source conflict", func(t *testing.T) {
		_, invBytes, ed, fetch, bld := windowsFixture(t)
		var inv WindowsInventory
		if e := DecodeClosed(invBytes, &inv); e != nil {
			t.Fatal(e)
		}
		added := WindowsSource{"other-core", "https://learn.microsoft.com/en-us/windows/synthetic-other", "powershell_manifest", nil, nil}
		inv.Sources = append(inv.Sources, added)
		fetch.Values[added.URL] = []byte(strings.Replace(syntheticWindowsManifest(), `value="10"`, `value="11"`, 1))
		other := testRoot(t)
		if _, e := WindowsAcquire(t.Context(), other, encoded(t, inv), fetch); e != nil {
			t.Fatal(e)
		}
		r, e := WindowsReconcile(t.Context(), other, encoded(t, inv), ed, bld)
		if e == nil || r.Report.Status != "blocked" {
			t.Fatal("conflict accepted")
		}
	})

}

func windowsBuildFixture(t *testing.T) (WindowsBuildInput, Builder, WindowsPackageReport) {
	t.Helper()
	root, inventory, editorial, _, builder := windowsFixture(t)
	dir := t.TempDir()
	for name, b := range map[string][]byte{"inventory.json": inventory, "editorial.json": editorial} {
		if e := os.WriteFile(filepath.Join(dir, name), b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	lock, e := root.ReadFile("source-lock.json")
	if e != nil {
		t.Fatal(e)
	}
	decision := WindowsLicenseDecision{"cartulary.windows_event_license_decision.v2", "synthetic", Digest(lock), "synthetic-unit-fixture-only", "test-author", packformat.License{Expression: "LicenseRef-Synthetic", ListVersion: "3.28.0", Redistribution: "restricted", Notices: []string{"notices/synthetic.txt"}, Bindings: []packformat.LicenseBinding{{Ref: "LicenseRef-Synthetic", Path: "notices/synthetic.txt"}}}, []WindowsNotice{{Path: "notices/synthetic.txt", SourceID: "license"}}}
	if e = os.WriteFile(filepath.Join(dir, "decision.json"), encoded(t, decision), 0600); e != nil {
		t.Fatal(e)
	}
	in := WindowsBuildInput{"cartulary.windows_event_build_input.v3", root.Name(), filepath.Join(dir, "inventory.json"), filepath.Join(dir, "editorial.json"), filepath.Join(dir, "public"), filepath.Join(dir, "ledger"), "2026-11-06T01:00:00Z", "synthetic.1", builder.SHA256, WindowsFixtureRepository, filepath.Join(dir, "decision.json")}
	report, e := AssembleWindows(t.Context(), encoded(t, in), builder)
	if e != nil {
		t.Fatal(e)
	}
	return in, builder, report
}
func TestWindowsOnlinePublication(t *testing.T) {
	in, builder, first := windowsBuildFixture(t)
	second, e := AssembleWindows(t.Context(), encoded(t, in), builder)
	if e != nil || first.ManifestSHA256 != second.ManifestSHA256 || first.PayloadSHA256 != second.PayloadSHA256 {
		t.Fatal("deterministic rebuild", e)
	}
	in.BuiltAt = "2026-11-06T02:00:00Z"
	if _, e = AssembleWindows(t.Context(), encoded(t, in), builder); e == nil {
		t.Fatal("same version accepts different explicit inputs")
	}
	in.PackVersion = "synthetic.2"
	in.LicenseDecision = filepath.Join(t.TempDir(), "absent.json")
	if _, e = AssembleWindows(t.Context(), encoded(t, in), builder); e == nil {
		t.Fatal("missing licensing decision accepted")
	}
	t.Run("publication is atomic", func(t *testing.T) {
		root := t.TempDir()
		if e := publishTree(root, "logical/complete", map[string][]byte{"good": []byte("a"), "../bad": []byte("b")}); e == nil {
			t.Fatal("invalid tree accepted")
		}
		if _, e := os.Stat(filepath.Join(root, "logical/complete")); !errors.Is(e, os.ErrNotExist) {
			t.Fatal("partial final tree exists")
		}
		if e := publishTree(root, "logical/complete", map[string][]byte{"good": []byte("a")}); e != nil {
			t.Fatal(e)
		}
		if e := publishTree(root, "logical/complete", map[string][]byte{"good": []byte("b")}); e == nil {
			t.Fatal("immutable tree overwritten")
		}
	})
	t.Run("ledger namespaces", func(t *testing.T) {
		dir := t.TempDir()
		if _, e := publication.ReserveRelease(dir, "1", Digest([]byte("a")), fixtureScope(t, WindowsFixtureRepository, WindowsKey)); e != nil {
			t.Fatal(e)
		}
		if _, e := publication.ReserveRelease(dir, "2", Digest([]byte("b")), fixtureScope(t, "fixture.unrelated", "enrichment.synthetic")); e == nil {
			t.Fatal("cross-repository ledger reused")
		}
	})
}

type fixtureClock struct{}

func (fixtureClock) Now() time.Time { return time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC) }
func fixtureScope(t *testing.T, repository, key string) publication.Scope {
	t.Helper()
	s, err := publication.FixtureScope(repository, key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func TestWindowsOnlineFixtureTrust(t *testing.T) {
	in, _, report := windowsBuildFixture(t)
	env := publication.EnvelopeInput{LogicalRoot: filepath.Join(in.OutputRoot, report.LogicalPath), OutputRoot: t.TempDir(), SigningContext: filepath.Join(t.TempDir(), "private.json"), CreateSigningContext: true, IssuedAt: "2026-11-08T00:00:00Z", RootExpires: "2027-11-08T00:00:00Z", OnlineExpires: "2026-12-08T00:00:00Z", MetadataVersion: 1, LedgerRoot: t.TempDir()}
	request := FixtureEnvelopeInput{SchemaID: "cartulary.windows_event_fixture_envelope_input.v3", Envelope: env}
	first, err := PublishFixture(t.Context(), encoded(t, request), fixtureClock{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := PublishFixture(t.Context(), encoded(t, request), fixtureClock{})
	if err != nil || first.ContainerSHA256 != second.ContainerSHA256 || first.BootstrapSHA256 != second.BootstrapSHA256 {
		t.Fatal("fixture envelope reproduction", err)
	}
	if _, err = publication.PublishFixture(t.Context(), publication.FixtureInput{Envelope: env}, fixtureScope(t, "fixture.unrelated", WindowsKey), fixtureClock{}); err == nil {
		t.Fatal("foreign authority accepted")
	}
	if _, err = PublishRelease(t.Context(), encoded(t, request), fixtureClock{}); err == nil {
		t.Fatal("fixture request entered production path")
	}
}
