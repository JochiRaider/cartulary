package producer

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
)

func csvFixtureSources(t *testing.T) (WindowsSource, []byte, map[string]WindowsAcquired, map[string][]byte) {
	t.Helper()
	var fixture struct {
		Sources []struct {
			ID    string `json:"id"`
			Kind  string `json:"kind"`
			Bytes string `json:"bytes_base64"`
		}
	}
	if e := json.Unmarshal(windowsInput(t, "fixture-local.v2.json"), &fixture); e != nil {
		t.Fatal(e)
	}
	sources := map[string]WindowsAcquired{}
	blobs := map[string][]byte{}
	var source WindowsSource
	var csvBytes []byte
	for _, item := range fixture.Sources {
		b, e := base64.StdEncoding.DecodeString(item.Bytes)
		if e != nil {
			t.Fatal(e)
		}
		digest := Digest(b)
		s := WindowsSource{ID: item.ID, Kind: item.Kind, URL: "urn:cartulary:supplied:" + digest, ExpectedSHA256: &digest}
		sources[s.ID] = WindowsAcquired{Source: s}
		blobs[s.ID] = b
		if item.Kind == "combined_csv" {
			source = s
			csvBytes = b
		}
	}
	return source, csvBytes, sources, blobs
}
func rewriteCSV(t *testing.T, b []byte, change func(map[string]string)) []byte {
	t.Helper()
	r := csv.NewReader(bytes.NewReader(b))
	rows, e := r.ReadAll()
	if e != nil {
		t.Fatal(e)
	}
	m := map[string]string{}
	for i, h := range rows[0] {
		m[h] = rows[1][i]
	}
	change(m)
	for i, h := range rows[0] {
		rows[1][i] = m[h]
	}
	var out bytes.Buffer
	w := csv.NewWriter(&out)
	if e = w.WriteAll(rows); e != nil {
		t.Fatal(e)
	}
	return out.Bytes()
}
func testWindowsCSVReferences(t *testing.T) {
	s, b, sources, blobs := csvFixtureSources(t)
	out, e := extractWindowsCSV(s, b, sources, blobs)
	if e != nil || len(out.Issues) != 0 || len(out.Records) != 1 {
		t.Fatalf("CSV input: %v %#v", e, out.Issues)
	}
	r := out.Records[0]
	if r.Provider != "PowerShellCore" || r.Version == nil || *r.Version != 0 || r.Level == nil || *r.Level != 0 || r.Task == nil || *r.Task != 0 || r.Opcode == nil || *r.Opcode != 0 {
		t.Fatal("explicit identity and zero descriptors", r)
	}
	message := false
	for _, d := range r.Documentation {
		if d.Kind == "message_template" && d.Scope == "event_version" {
			message = d.Content["text"] == "Synthetic multiline %1.\n  Keep indentation.\n"
		}
	}
	if !message {
		t.Fatal("complete quoted message and whitespace lost")
	}
	t.Run("presentation omits trailing paragraphs without mutating evidence", func(t *testing.T) {
		original := map[string]any{"text": "  %1\n\n  %2\n \t\n\u00a0", "rows": []any{map[string]any{"text": "cell\n\n"}}, "empty": "\n\t"}
		projected := windowsPresentation(original).(map[string]any)
		if projected["text"] != "  %1\n\n  %2" || projected["empty"] != "" || projected["rows"].([]any)[0].(map[string]any)["text"] != "cell" {
			t.Fatal("presentation damaged indentation or retained trailing paragraphs", projected)
		}
		if original["text"] != "  %1\n\n  %2\n \t\n\u00a0" || original["rows"].([]any)[0].(map[string]any)["text"] != "cell\n\n" {
			t.Fatal("presentation rewrote original evidence")
		}
	})
	t.Run("unknown structured field rejected", func(t *testing.T) {
		raw := rewriteCSV(t, b, func(r map[string]string) { r["quality_notes_json"] = `[{"code":"test","unknown":true}]` })
		if _, e := parseWindowsCSV(raw); e == nil {
			t.Fatal("open JSON field")
		}
	})
	t.Run("duplicate JSON key rejected", func(t *testing.T) {
		raw := rewriteCSV(t, b, func(r map[string]string) { r["quality_notes_json"] = `[{"code":"a","code":"b"}]` })
		if _, e := parseWindowsCSV(raw); e == nil {
			t.Fatal("duplicate JSON key")
		}
	})
	t.Run("missing version never becomes zero", func(t *testing.T) {
		raw := rewriteCSV(t, b, func(r map[string]string) { r["event_version"] = "" })
		if _, e := parseWindowsCSV(raw); e == nil {
			t.Fatal("missing version became zero")
		}
	})
	t.Run("UTF8 and quoting", func(t *testing.T) {
		for _, bad := range [][]byte{append([]byte{0xff}, b...), append([]byte{0xef, 0xbb, 0xbf}, b...), []byte("x,y\n\"unterminated")} {
			if _, e := parseWindowsCSV(bad); e == nil {
				t.Fatal("malformed CSV accepted")
			}
		}
	})
	t.Run("malformed payload cannot rewrite system", func(t *testing.T) {
		x, e := windowsSampleSystem(`<Event><System><Provider Name="PowerShellCore"/><EventID>4104</EventID><Version>0</Version></System><Data><broken &</Data>`)
		if e != nil || xmlOne(x, "EventID").Text != "4104" {
			t.Fatal("independently valid System", e)
		}
		if _, e = windowsSampleSystem(`<System><EventID>1</EventID></System><System><EventID>2</EventID></System>`); e == nil {
			t.Fatal("ambiguous System accepted")
		}
	})
	t.Run("ZIP boundaries", func(t *testing.T) {
		for _, names := range [][]string{{"../escape"}, {"word/document.xml", "word/document.xml"}} {
			var b bytes.Buffer
			z := zip.NewWriter(&b)
			for _, name := range names {
				w, e := z.Create(name)
				if e != nil {
					t.Fatal(e)
				}
				_, _ = w.Write([]byte("<document/>"))
			}
			if e := z.Close(); e != nil {
				t.Fatal(e)
			}
			if _, e := windowsOfficeMembers(b.Bytes()); e == nil {
				t.Fatal("unsafe ZIP accepted", names)
			}
		}
	})
	t.Run("external XML entities", func(t *testing.T) {
		if _, e := parseWindowsXML([]byte(`<!DOCTYPE x [<!ENTITY e SYSTEM "https://example.test/e">]><x>&e;</x>`)); e == nil {
			t.Fatal("external entity admitted")
		}
	})
	t.Run("no provider inference", func(t *testing.T) {
		_, _, _, bad := csvFixtureSources(t)
		office, e := windowsOfficeMembers(bad["synthetic-docx"])
		if e != nil {
			t.Fatal(e)
		}
		raw := strings.ReplaceAll(string(office["word/document.xml"]), "PowerShellCore", "Unknown-Provider")
		var zipped bytes.Buffer
		z := zip.NewWriter(&zipped)
		w, _ := z.Create("word/document.xml")
		_, _ = w.Write([]byte(raw))
		if e = z.Close(); e != nil {
			t.Fatal(e)
		}
		bad["synthetic-docx"] = zipped.Bytes()
		got, e := extractWindowsCSV(s, b, sources, bad)
		if e != nil || len(got.Records) != 0 || len(got.Issues) == 0 {
			t.Fatal("unqualified provider inferred", e)
		}
	})
}
