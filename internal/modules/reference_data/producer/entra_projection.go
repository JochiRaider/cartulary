package producer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	entrainputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/entra-app-ids"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"golang.org/x/text/unicode/norm"
)

type entraQuarantine struct {
	SchemaID   string `json:"schema_id"`
	Assertions []struct {
		SourceID     string `json:"source_id"`
		SourceSHA256 string `json:"source_sha256"`
		Locator      string `json:"locator"`
		RawRow       string `json:"raw_row"`
		Reason       string `json:"reason"`
	} `json:"assertions"`
}
type entraBaselineEntry struct {
	ID       string   `json:"app_id"`
	Label    string   `json:"display_label"`
	Locators []string `json:"source_locators"`
	Aliases  []string `json:"aliases"`
	Status   string   `json:"source_status"`
}
type entraEditorial struct {
	SchemaID       string `json:"schema_id"`
	HistoricalNote string `json:"historical_note"`
	Summaries      []struct {
		ID       string   `json:"app_id"`
		Text     string   `json:"text"`
		Locators []string `json:"source_locators"`
	} `json:"summaries"`
}
type EntraAssertion struct {
	SourceID     string   `json:"source_id"`
	SourceSHA256 string   `json:"source_sha256"`
	Locator      string   `json:"locator"`
	RawRow       string   `json:"raw_row"`
	RawCells     []string `json:"raw_cells"`
	AppID        *string  `json:"app_id"`
	Disposition  string   `json:"disposition"`
	Reason       string   `json:"reason"`
}
type EntraFieldAccounting struct {
	ID     string            `json:"app_id"`
	Status string            `json:"source_status"`
	Fields map[string]string `json:"fields"`
}
type EntraAudit struct {
	SchemaID            string                 `json:"schema_id"`
	SourceLockSHA256    string                 `json:"source_lock_sha256"`
	SourceProfileSHA256 string                 `json:"source_profile_sha256"`
	Builder             Builder                `json:"builder"`
	Counts              map[string]int         `json:"counts"`
	Assertions          []EntraAssertion       `json:"assertions"`
	Unselected          []EntraAssertion       `json:"unselected_rows"`
	AliasOmissions      map[string][]string    `json:"alias_omissions"`
	Fields              []EntraFieldAccounting `json:"field_accounting"`
	PayloadSHA256       string                 `json:"entries_sha256"`
	PayloadBytes        int                    `json:"entries_bytes"`
	MaximumLabelScalars int                    `json:"maximum_label_scalars"`
	MaximumAliases      int                    `json:"maximum_aliases"`
	MaximumLineBytes    int                    `json:"maximum_line_bytes"`
}
type EntraProjection struct {
	Payload   []byte
	Artifacts []packformat.SourceArtifact
	Report    EntraAudit
	Research  map[string][]byte
}
type entraRow struct {
	line            int
	raw             string
	cells, rawCells []string
}

var entraUUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
var entraSeparator = regexp.MustCompile(`^:?-{2,}:?$`)

func entraASCII(value string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, value)
}

// Extraction checks table shape before interpreting any ID. Quarantine matches
// exact raw rows; it is not an option to skip malformed input.
func extractEntra(source EntraSource, b []byte) ([]entraRow, []EntraAssertion, error) {
	if !utf8.Valid(b) || bytes.HasPrefix(b, []byte{0xef, 0xbb, 0xbf}) {
		return nil, nil, errors.New("invalid Entra source encoding")
	}
	text := strings.ReplaceAll(string(b), "\r\n", "\n")
	if strings.ContainsRune(text, '\r') {
		return nil, nil, errors.New("unsupported Entra source line ending")
	}
	if source.Heading == nil {
		return nil, nil, errors.New("entra dataset heading required")
	}
	rows := []entraRow{}
	unselected := []EntraAssertion{}
	selected, found, tableStage := false, 0, 0
	heading := ""
	for i, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			heading = strings.TrimLeft(line, "# ")
			selected = heading == *source.Heading
			if selected {
				found++
				if found > 1 {
					return nil, nil, errors.New("duplicate Entra section")
				}
			}
			continue
		}
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "|") {
			continue
		}
		if !strings.HasSuffix(trimmed, "|") {
			if selected {
				return nil, nil, errors.New("invalid Entra table framing")
			}
			continue
		}
		rawCells := strings.Split(trimmed[1:len(trimmed)-1], "|")
		cells := make([]string, len(rawCells))
		separator := true
		for j, c := range rawCells {
			cells[j] = strings.TrimSpace(c)
			separator = separator && entraSeparator.MatchString(cells[j])
		}
		if !selected {
			if !separator && len(cells) > 1 && !strings.HasPrefix(cells[0], "Application") {
				unselected = append(unselected, EntraAssertion{source.ID, source.SHA256, "L" + strconv.Itoa(i+1), line, rawCells, nil, "unselected_section", heading})
			}
			continue
		}
		if len(cells) != len(source.Header) {
			return nil, nil, errors.New("unexpected Entra column count")
		}
		if tableStage == 0 {
			if !slices.Equal(cells, source.Header) {
				return nil, nil, errors.New("unexpected Entra table header")
			}
			tableStage++
			continue
		}
		if tableStage == 1 {
			if !separator {
				return nil, nil, errors.New("missing Entra table separator")
			}
			tableStage++
			continue
		}
		if separator || slices.Equal(cells, source.Header) {
			return nil, nil, errors.New("unexpected additional Entra table")
		}
		rows = append(rows, entraRow{i + 1, line, cells, rawCells})
	}
	if found != 1 || tableStage != 2 || len(rows) != source.Rows {
		return nil, nil, errors.New("entra section row count mismatch")
	}
	return rows, unselected, nil
}

func entraCell(value string, description bool) (string, error) {
	if description {
		value = strings.ReplaceAll(value, "<br/>", "\n")
	}
	if strings.ContainsAny(value, "<>`") || strings.Contains(value, "](") || strings.Contains(value, "**") || strings.Contains(value, "![") {
		return "", errors.New("unsupported Entra inline markup")
	}
	value = norm.NFC.String(html.UnescapeString(value))
	for _, r := range value {
		if unicode.IsControl(r) && !(description && r == '\n') {
			return "", errors.New("entra source control character")
		}
	}
	if value == "" {
		return "", errors.New("empty Entra source value")
	}
	return value, nil
}

func EntraReconcile(ctx context.Context, root *os.Root, inventoryBytes []byte, builder Builder) (EntraProjection, error) {
	lock, err := EntraCheck(root, inventoryBytes)
	if err != nil {
		return EntraProjection{}, err
	}
	lockBytes, err := root.ReadFile("source-lock.json")
	if err != nil {
		return EntraProjection{}, err
	}
	profile, err := entrainputs.Files.ReadFile("source-profile.v2.json")
	if err != nil {
		return EntraProjection{}, err
	}
	p := EntraProjection{Research: map[string][]byte{"source-profile.json": profile, "source-lock.json": lockBytes, "inventory.json": inventoryBytes}, Report: EntraAudit{SchemaID: "cartulary.entra_reconciliation.v2", SourceLockSHA256: Digest(lockBytes), SourceProfileSHA256: Digest(profile), Builder: builder, Counts: map[string]int{}, Assertions: []EntraAssertion{}, Unselected: []EntraAssertion{}, AliasOmissions: map[string][]string{}, Fields: []EntraFieldAccounting{}}}
	var quarantine entraQuarantine
	var editorial entraEditorial
	var baseline struct {
		SchemaID string               `json:"schema_id"`
		Entries  []entraBaselineEntry `json:"entries"`
	}
	for _, input := range []struct {
		name   string
		target any
	}{{"quarantine.v2.json", &quarantine}, {"editorial.v2.json", &editorial}, {"baseline.v2.json", &baseline}} {
		b, e := entrainputs.Files.ReadFile(input.name)
		if e != nil {
			return p, e
		}
		if e = DecodeClosed(b, input.target); e != nil {
			return p, e
		}
		p.Research[input.name] = b
	}
	for _, name := range []string{"license.project.v2.json", "Apache-2.0.txt", "attribution.txt"} {
		b, e := entrainputs.Files.ReadFile(name)
		if e != nil {
			return p, e
		}
		p.Research[name] = b
	}
	// Every output-affecting policy artifact is identified before source indices
	// are assigned. Raw sources stay in research retention, never in the pack.
	for name, b := range p.Research {
		p.Artifacts = append(p.Artifacts, packformat.SourceArtifact{Ref: "urn:cartulary:entra-input:" + name, Version: EntraSourceProfile, SHA256: Digest(b)})
	}
	for _, source := range lock.Sources {
		p.Artifacts = append(p.Artifacts, packformat.SourceArtifact{Ref: source.Source.URL, Version: source.Source.Commit, SHA256: source.Source.SHA256})
	}
	sortEntraArtifacts(p.Artifacts)
	indices := map[string]int{}
	for i, a := range p.Artifacts {
		indices[a.Ref] = i
	}
	quarantines := map[string]int{}
	for i, q := range quarantine.Assertions {
		key := q.SourceID + ":" + q.Locator
		if _, exists := quarantines[key]; exists {
			return p, errors.New("duplicate Entra quarantine")
		}
		quarantines[key] = i
	}
	matched := map[int]bool{}
	entries := map[string]map[string]any{}
	expected := map[string]entraBaselineEntry{}
	locators := map[string][]string{}
	current := map[string]bool{}
	s1 := map[string]bool{}
	for _, entry := range baseline.Entries {
		if _, exists := expected[entry.ID]; exists {
			return p, errors.New("duplicate Entra baseline identity")
		}
		expected[entry.ID] = entry
	}
	for _, source := range lock.Sources {
		if source.Source.Kind != "dataset" {
			continue
		}
		if err = ctx.Err(); err != nil {
			return p, err
		}
		b, e := ReadEvidence(root, source.Evidence)
		if e != nil {
			return p, e
		}
		rows, unselected, e := extractEntra(source.Source, b)
		if e != nil {
			return p, e
		}
		p.Report.Unselected = append(p.Report.Unselected, unselected...)
		p.Report.Counts["rows"] += len(rows)
		assertions := 0
		for _, row := range rows {
			locator := "L" + strconv.Itoa(row.line)
			if qi, exists := quarantines[source.Source.ID+":"+locator]; exists {
				q := quarantine.Assertions[qi]
				if q.SourceSHA256 != source.Source.SHA256 || q.RawRow != row.raw || matched[qi] {
					return p, errors.New("entra quarantine byte/locator mismatch")
				}
				matched[qi] = true
				assertions++
				p.Report.Counts["quarantined"]++
				p.Report.Assertions = append(p.Report.Assertions, EntraAssertion{source.Source.ID, source.Source.SHA256, locator, row.raw, row.rawCells, nil, "quarantined", q.Reason})
				continue
			}
			label, e := entraCell(row.cells[0], false)
			if e != nil {
				return p, e
			}
			if len(row.cells) == 3 {
				if _, e = entraCell(row.cells[2], true); e != nil {
					return p, e
				}
			}
			ids := []string{row.cells[1]}
			if source.Source.ID == "S3" {
				ids = strings.Split(row.cells[1], "<br>")
			}
			for _, rawID := range ids {
				id := entraASCII(rawID)
				if !entraUUID.MatchString(id) {
					return p, errors.New("unquarantined malformed Entra ID")
				}
				loc := locator
				if len(ids) > 1 {
					loc += "#id=" + id
				}
				assertions++
				p.Report.Counts["admitted"]++
				rowDisposition := "included"
				entry, exists := entries[id]
				if exists {
					if entry["display_label"] != label {
						return p, errors.New("unresolved Entra identity conflict")
					}
					rowDisposition = "coalesced"
					p.Report.Counts["coalesced"]++
				} else {
					entry = map[string]any{"schema_id": "cartulary.reference_pack.enrichment.entra_app_ids.entry.v2", "entry_id": id, "app_id": id, "display_label": label, "description": nil, "aliases": []string{label}, "deprecated": false, "replacement_entry_id": nil, "source_refs": []map[string]any{}, "extensions": map[string]any{}, "publisher": "Microsoft", "service_principal_names": []string{}}
					entries[id] = entry
				}
				entry["source_refs"] = append(entry["source_refs"].([]map[string]any), map[string]any{"artifact_index": indices[source.Source.URL], "locator": loc})
				locators[id] = append(locators[id], source.Source.ID+":"+loc)
				current[id] = current[id] || source.Source.ID != "S3"
				s1[id] = s1[id] || source.Source.ID == "S1"
				p.Report.Assertions = append(p.Report.Assertions, EntraAssertion{source.Source.ID, source.Source.SHA256, loc, row.raw, row.rawCells, &id, rowDisposition, "source-backed identity and label"})
			}
		}
		if assertions != source.Source.Assertions {
			return p, errors.New("entra source assertion count mismatch")
		}
		p.Report.Counts["assertions"] += assertions
	}
	if len(matched) != len(quarantine.Assertions) {
		return p, errors.New("unmatched Entra quarantine decision")
	}
	for _, pair := range []struct {
		id      string
		aliases []string
	}{{"60ca1954-583c-4d1f-86de-39d835f3e452", []string{"Microsoft Defender for Identity", "Radius Aad Syncer"}}, {"00000005-0000-0ff1-ce00-000000000000", []string{"Viva Engage", "Yammer"}}} {
		entry, ok := entries[pair.id]
		if !ok {
			return p, errors.New("missing Entra former-name identity")
		}
		if entry["display_label"] != pair.aliases[0]+" (formerly "+pair.aliases[1]+")" {
			return p, errors.New("entra former-name evidence changed")
		}
		entry["aliases"] = append(entry["aliases"].([]string), pair.aliases...)
	}
	owners := map[string][]string{}
	for id, entry := range entries {
		unique := map[string]string{}
		for _, a := range entry["aliases"].([]string) {
			key := entraASCII(norm.NFC.String(a))
			if prior, ok := unique[key]; !ok || a < prior {
				unique[key] = a
			}
		}
		aliases := []string{}
		for key, a := range unique {
			owners[key] = append(owners[key], id)
			aliases = append(aliases, a)
		}
		entry["aliases"] = aliases
	}
	for key, ids := range owners {
		if len(ids) > 1 {
			slices.Sort(ids)
			p.Report.AliasOmissions[key] = ids
		}
	}
	summaries := map[string]string{}
	for _, summary := range editorial.Summaries {
		if !s1[summary.ID] || summaries[summary.ID] != "" || summary.Text == "" || len(summary.Locators) != 1 || !slices.Contains(locators[summary.ID], summary.Locators[0]) {
			return p, errors.New("unbound Entra editorial summary")
		}
		summaries[summary.ID] = summary.Text
	}
	ids := make([]string, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		entry := entries[id]
		aliases := []string{}
		for _, a := range entry["aliases"].([]string) {
			if len(owners[entraASCII(a)]) == 1 {
				aliases = append(aliases, a)
			}
		}
		slices.SortFunc(aliases, func(a, b string) int {
			if n := strings.Compare(entraASCII(a), entraASCII(b)); n != 0 {
				return n
			}
			return strings.Compare(a, b)
		})
		entry["aliases"] = aliases
		status := "current_documented"
		descriptionRule := "explicit null; selected ownership table has no description"
		if !current[id] {
			entry["description"] = editorial.HistoricalNote
			status = "historical_documented"
			descriptionRule = "editorial.v2.json historical_note; producer-authored historical qualification"
			p.Report.Counts["historical_only"]++
		} else if s1[id] {
			if summaries[id] == "" {
				return p, errors.New("missing Entra original summary")
			}
			entry["description"] = summaries[id]
			descriptionRule = "editorial.v2.json summaries[app_id]; original text supported by its S1 locator"
		}
		slices.Sort(locators[id])
		want, ok := expected[id]
		if !ok || want.Label != entry["display_label"] || want.Status != status || !slices.Equal(want.Aliases, aliases) || !slices.Equal(want.Locators, locators[id]) {
			return p, fmt.Errorf("entra baseline differs for %s", id)
		}
		refs := entry["source_refs"].([]map[string]any)
		slices.SortFunc(refs, func(a, b map[string]any) int {
			if a["artifact_index"].(int) != b["artifact_index"].(int) {
				return a["artifact_index"].(int) - b["artifact_index"].(int)
			}
			return strings.Compare(a["locator"].(string), b["locator"].(string))
		})
		p.Report.Fields = append(p.Report.Fields, EntraFieldAccounting{id, status, map[string]string{"schema_id": "registered Entra entry v2", "entry_id": "canonical app_id from " + strings.Join(locators[id], ", "), "app_id": "source application/client ID; lowercase ASCII hex only", "display_label": "exact factual source name; NFC and syntactic padding removal", "description": descriptionRule, "aliases": "source label and explicit former-name relationships; global collision removal", "deprecated": "false: no affirmative lifecycle evidence", "replacement_entry_id": "null: no replacement evidence", "source_refs": "all admitted assertions; indices into sorted source_artifacts", "extensions": "empty closed object", "publisher": "Microsoft section attribution; not Graph publisher verification", "service_principal_names": "empty: no source SPN property evidence"}})
		b, e := canonicaljson.Marshal(entry)
		if e != nil {
			return p, e
		}
		p.Payload = append(p.Payload, b...)
		p.Payload = append(p.Payload, '\n')
		p.Report.MaximumLabelScalars = max(p.Report.MaximumLabelScalars, utf8.RuneCountInString(entry["display_label"].(string)))
		p.Report.MaximumAliases = max(p.Report.MaximumAliases, len(aliases))
		p.Report.MaximumLineBytes = max(p.Report.MaximumLineBytes, len(b)+1)
		p.Report.Counts["aliases"] += len(aliases)
		if len(aliases) > 0 {
			p.Report.Counts["entries_with_aliases"]++
		}
	}
	p.Report.Counts["entries"] = len(entries)
	p.Report.Counts["spns"] = 0
	for key, want := range map[string]int{"rows": 145, "assertions": 147, "quarantined": 7, "admitted": 140, "coalesced": 1, "entries": 139, "historical_only": 119, "entries_with_aliases": 135, "aliases": 139, "spns": 0} {
		if p.Report.Counts[key] != want {
			return p, fmt.Errorf("entra %s accounting mismatch", key)
		}
	}
	if len(summaries) != 16 || len(expected) != 139 {
		return p, errors.New("entra editorial/baseline count mismatch")
	}
	p.Report.PayloadSHA256 = Digest(p.Payload)
	p.Report.PayloadBytes = len(p.Payload)
	return p, nil
}

func sortEntraArtifacts(artifacts []packformat.SourceArtifact) {
	slices.SortFunc(artifacts, func(a, b packformat.SourceArtifact) int {
		if n := strings.Compare(a.Ref, b.Ref); n != 0 {
			return n
		}
		if n := strings.Compare(a.Version, b.Version); n != 0 {
			return n
		}
		return strings.Compare(a.SHA256, b.SHA256)
	})
}

func EntraAuditSources(ctx context.Context, root *os.Root, inventory []byte, builder Builder) (EntraAudit, error) {
	p, err := EntraReconcile(ctx, root, inventory, builder)
	if err != nil {
		return p.Report, err
	}
	b, err := canonicaljson.Marshal(p.Report)
	if err != nil {
		return p.Report, err
	}
	if err = Retain(root, "audits/"+Digest(b)+".json", b, 0600); err != nil {
		return p.Report, err
	}
	if err = Retain(root, "candidates/"+p.Report.PayloadSHA256+".ndjson", p.Payload, 0600); err != nil {
		return p.Report, err
	}
	return p.Report, nil
}
