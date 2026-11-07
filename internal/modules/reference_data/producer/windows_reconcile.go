package producer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	windowsinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/windows-event-ids"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"golang.org/x/text/unicode/norm"
)

type WindowsProjection struct {
	Report                          WindowsReconciliation
	Evidence                        WindowsEvidence
	Payload, Archive, SourceProfile []byte
}

func WindowsReconcile(ctx context.Context, root *os.Root, inventoryBytes, editorialBytes []byte, builder Builder) (WindowsProjection, error) {
	var out WindowsProjection
	if norm.Version != "17.0.0" {
		return out, windowsFailure("input_invalid", "source_profile", "Unicode table change requires a new reviewed profile")
	}
	lock, err := WindowsCheck(root, inventoryBytes)
	if err != nil {
		return out, err
	}
	var editorial WindowsEditorial
	if err = windowsDecode(editorialBytes, "editorial", &editorial); err != nil {
		return out, err
	}
	profile, err := windowsinputs.Files.ReadFile("source-profile.v2.json")
	if err != nil {
		return out, err
	}
	out.SourceProfile = profile
	summaryBytes, err := windowsinputs.Files.ReadFile("original-summaries.v2.json")
	if err != nil {
		return out, err
	}
	var summaries windowsOriginalSummaries
	if err = windowsDecode(summaryBytes, "summaries", &summaries); err != nil {
		return out, err
	}
	lockBytes, err := root.ReadFile("source-lock.json")
	if err != nil {
		return out, err
	}
	r := WindowsReconciliation{SchemaID: "cartulary.windows_event_reconciliation.v2", Status: "blocked", SourceLockSHA256: Digest(lockBytes), SourceProfileSHA256: Digest(profile), EditorialSHA256: Digest(editorialBytes), Builder: builder, PayloadFileSHA256: Digest(nil), EvidenceSHA256: Digest(nil), Coverage: []WindowsCoverage{}, Dispositions: []WindowsDisposition{}, Issues: []WindowsError{}, Limitations: []string{"Coverage is the frozen documentary inventory, not installed Windows builds or observed occurrences.", "Unspecified event versions are distinct null-version identities, not wildcard versions.", "External Winmeta declarations are not guessed; unresolved numeric descriptors remain null.", "No complete event decoder or event-payload schema is supplied.", "Distribution contains factual definitions and original summaries under the project license; upstream research material stays local."}}
	evidence := WindowsEvidence{"cartulary.windows_event_evidence.v2", []WindowsRecord{}, editorial.Notes, []WindowsAssertion{}}
	sourceBytes := map[string][]byte{}
	sources := map[string]WindowsAcquired{}
	archiveMembers := map[string][]byte{}
	for _, a := range lock.Sources {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		b, e := ReadEvidence(root, a.Evidence)
		if e != nil {
			return out, e
		}
		sourceBytes[a.Source.ID] = b
		sources[a.Source.ID] = a
		archiveMembers[a.Evidence.Path] = b
	}
	if err = validateWindowsProviderBindings(editorial.ProviderBindings, sources, sourceBytes); err != nil {
		return out, err
	}
	for _, a := range lock.Sources {
		b := sourceBytes[a.Source.ID]
		var rows []WindowsRecord
		var dispositions []WindowsDisposition
		var e error
		switch a.Source.Kind {
		case "powershell_manifest":
			rows, dispositions, e = extractWindowsManifest(a.Source, b)
		case "sysmon_manifest":
			rows, dispositions, e = extractWindowsSysmonManifest(a.Source, b, sources, sourceBytes)
		case "sysmon_document":
			rows, dispositions, e = extractWindowsSysmonDocument(a.Source, b, sources, sourceBytes)
		case "event_catalog":
			rows, dispositions, e = extractWindowsCatalog(a.Source, b)
		case "event_document", "provider_article":
			rows, dispositions, e = extractWindowsDocument(a.Source, b)
		case "combined_csv":
			var csv windowsCSVResult
			csv, e = extractWindowsCSV(a.Source, b, sources, sourceBytes, editorial.ProviderBindings...)
			rows, dispositions = csv.Records, csv.Dispositions
			evidence.Assertions = append(evidence.Assertions, csv.Assertions...)
			r.Issues = append(r.Issues, csv.Issues...)
		default:
			dispositions = []WindowsDisposition{{a.Source.ID, "document", "excluded", "supporting_" + a.Source.Kind, nil}}
		}
		if e != nil {
			r.Issues = append(r.Issues, WindowsError{"source_parse_failed", a.Source.ID, e.Error()})
			r.Dispositions = append(r.Dispositions, WindowsDisposition{a.Source.ID, "document", "blocking", "source_parse_failed", nil})
			continue
		}
		for i := range rows {
			if rows[i].Documentation == nil {
				rows[i].Documentation = []WindowsDocumentation{}
			}
		}
		evidence.Records = append(evidence.Records, rows...)
		r.Dispositions = append(r.Dispositions, dispositions...)
	}
	if err = applyWindowsProviderExclusions(editorial.ProviderExclusions, sources, sourceBytes, &r); err != nil {
		return out, err
	}
	slices.SortFunc(evidence.Records, func(a, b WindowsRecord) int {
		aa, _ := canonicaljson.Marshal(a)
		bb, _ := canonicaljson.Marshal(b)
		return bytes.Compare(aa, bb)
	})
	slices.SortFunc(evidence.Notes, func(a, b WindowsNote) int { return strings.Compare(a.ID, b.ID) })
	groups := map[string][]int{}
	csvFragments := map[string]map[string]string{}
	for id, a := range sources {
		if a.Source.Kind == "combined_csv" {
			rows, e := parseWindowsCSV(sourceBytes[id])
			if e == nil {
				csvFragments[id] = map[string]string{}
				for _, row := range rows {
					csvFragments[id][fmt.Sprintf("csv/record/%d", row.Number)] = row.Raw
				}
			}
		}
	}
	fragmentFor := func(sourceID, locator string) (string, error) {
		if fragments, ok := csvFragments[sourceID]; ok {
			if raw, ok := fragments[locator]; ok {
				return raw, nil
			}
			return "", fmt.Errorf("unresolved CSV locator")
		}
		return windowsFragment(sources[sourceID].Source, sourceBytes[sourceID], locator)
	}
	guids := map[string]string{}
	coverage := map[string]*WindowsCoverage{}
	for _, p := range windowsProviders {
		coverage[p] = &WindowsCoverage{Provider: p, NumericVersions: []int64{}}
	}
	for i, record := range evidence.Records {
		provider, _ := windowsProvider(record.Provider)
		id := packformat.EventIdentity(record.Provider, record.EventID, record.Version)
		fragment, e := fragmentFor(record.SourceID, record.Locator)
		if e != nil || Digest([]byte(fragment)) != record.FragmentSHA256 {
			r.Issues = append(r.Issues, WindowsError{"source_accounting_failed", record.SourceID, "fragment does not resolve"})
		}
		if record.GUID != nil {
			old, ok := guids[provider]
			if ok && old != *record.GUID {
				r.Issues = append(r.Issues, WindowsError{"identity_conflict", id, "known provider GUIDs disagree"})
			}
			guids[provider] = *record.GUID
		}
		groups[id] = append(groups[id], i)
		coverage[provider].Assertions++
	}
	notes := map[string][]WindowsNote{}
	seenNotes := map[string]bool{}
	for _, note := range evidence.Notes {
		_, ok := sources[note.SourceID]
		fragment, e := fragmentFor(note.SourceID, note.Locator)
		if !ok || seenNotes[note.ID] || groups[note.EntryID] == nil || e != nil || Digest([]byte(fragment)) != note.FragmentSHA256 {
			r.Issues = append(r.Issues, WindowsError{"source_accounting_failed", note.ID, "unbound, duplicate or changed editorial note"})
			continue
		}
		seenNotes[note.ID] = true
		notes[note.EntryID] = append(notes[note.EntryID], note)
	}
	ids := []string{}
	for id := range groups {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	var payload bytes.Buffer
	for _, id := range ids {
		indices := groups[id]
		first := evidence.Records[indices[0]]
		level, task, opcode := first.Level, first.Task, first.Opcode
		var rowGUID *string
		documentation := []map[string]any{}
		references := []map[string]any{}
		conflict := false
		for _, i := range indices {
			record := evidence.Records[i]
			if record.GUID != nil {
				rowGUID = record.GUID
			}
			for _, field := range []struct {
				dst   **int64
				value *int64
			}{{&level, record.Level}, {&task, record.Task}, {&opcode, record.Opcode}} {
				if field.value != nil {
					if *field.dst != nil && **field.dst != *field.value {
						conflict = true
					}
					*field.dst = field.value
				}
			}
			ref := map[string]any{"artifact_index": 0, "locator": fmt.Sprintf("evidence.json#/records/%d", i)}
			references = append(references, ref)
			fragment, _ := fragmentFor(record.SourceID, record.Locator)
			for _, doc := range windowsPublicFacts(record, fragment) {
				documentation = append(documentation, map[string]any{"kind": doc.Kind, "scope": doc.Scope, "source_refs": []map[string]any{ref}, "content": doc.Content})
			}
		}
		if conflict {
			r.Issues = append(r.Issues, WindowsError{"projection_conflict", id, "known native descriptors disagree"})
		}
		description := windowsOriginalDescription(first, summaries)
		documentation = append(documentation, map[string]any{"kind": "explanation", "scope": "event", "source_refs": references, "content": map[string]any{"text": description}})
		if rowGUID != nil {
			documentation = append(documentation, map[string]any{"kind": "explanation", "scope": "source_context", "source_refs": references, "content": map[string]any{"text": "Provider GUID: " + *rowGUID + "."}})
		}
		for _, note := range notes[id] {
			for i, n := range evidence.Notes {
				if n.ID == note.ID {
					references = append(references, map[string]any{"artifact_index": 0, "locator": fmt.Sprintf("evidence.json#/notes/%d", i)})
					documentation = append(documentation, map[string]any{"kind": "explanation", "scope": "source_context", "source_refs": []map[string]any{{"artifact_index": 0, "locator": fmt.Sprintf("evidence.json#/notes/%d", i)}}, "content": map[string]any{"text": note.Text}})
				}
			}
		}
		slices.SortFunc(references, func(a, b map[string]any) int { return strings.Compare(a["locator"].(string), b["locator"].(string)) })
		description = windowsText(description)
		for _, block := range documentation {
			block["content"] = windowsPresentation(block["content"])
		}
		documentation, err = windowsCoalesceDocumentation(documentation)
		if err != nil {
			return out, err
		}
		slices.SortFunc(documentation, func(a, b map[string]any) int {
			aa, _ := canonicaljson.Marshal(a)
			bb, _ := canonicaljson.Marshal(b)
			return bytes.Compare(aa, bb)
		})
		length := utf8.RuneCountInString(description)
		r.MaximumDescriptionScalars = max(r.MaximumDescriptionScalars, length)
		versionLabel := "version unspecified"
		if first.Version != nil {
			versionLabel = "v" + strconv.FormatInt(*first.Version, 10)
		}
		row := map[string]any{"schema_id": "cartulary.reference_pack.enrichment.windows_event_ids.entry.v2", "entry_id": id, "display_label": first.Provider + " / " + strconv.FormatInt(first.EventID, 10) + " / " + versionLabel, "description": description, "aliases": []any{}, "deprecated": false, "replacement_entry_id": nil, "source_refs": references, "extensions": map[string]any{}, "provider_name": first.Provider, "event_id": first.EventID, "event_version": first.Version, "level": decimalPointer(level), "task": decimalPointer(task), "opcode": decimalPointer(opcode), "documentation": documentation}
		line, e := canonicaljson.Marshal(row)
		if e != nil {
			return out, e
		}
		r.MaximumLineBytes = max(r.MaximumLineBytes, len(line)+1)
		overflow := length > 8192 || len(references) > 64 || len(documentation) > 512 || len(line)+1 > 1048576
		if overflow {
			r.Issues = append(r.Issues, WindowsError{"text_reference_overflow", id, fmt.Sprintf("description=%d refs=%d line_bytes=%d", length, len(references), len(line)+1)})
		}
		for j, i := range indices {
			disposition, reason := "coalesced", "compatible_definition_evidence"
			if j == 0 {
				disposition, reason = "emitted", "canonical_identity"
			}
			if conflict || overflow {
				disposition, reason = "blocking", "projection_not_publishable"
			}
			record := evidence.Records[i]
			r.Dispositions = append(r.Dispositions, WindowsDisposition{record.SourceID, record.Locator, disposition, reason, &id})
		}
		payload.Write(line)
		payload.WriteByte('\n')
		provider, _ := windowsProvider(first.Provider)
		c := coverage[provider]
		c.Entries++
		if first.Version == nil {
			c.Unspecified++
		} else if !slices.Contains(c.NumericVersions, *first.Version) {
			c.NumericVersions = append(c.NumericVersions, *first.Version)
		}
	}
	for _, p := range windowsProviders {
		c := coverage[p]
		slices.Sort(c.NumericVersions)
		r.Coverage = append(r.Coverage, *c)
		if c.Entries > 10000 {
			r.Issues = append(r.Issues, WindowsError{"source_overflow", p, "provider definition budget exceeded"})
		}
		if c.Entries == 0 {
			r.Limitations = append(r.Limitations, "No admissible definition for scoped provider "+p)
		}
	}
	if payload.Len() > 268435456 {
		r.Issues = append(r.Issues, WindowsError{"source_overflow", "payload", "payload byte limit"})
	}
	// Content acceptance reuses the registered runtime validator without
	// inventing a license or pretending this candidate is a publishable manifest.
	count := int64(len(ids))
	candidate := packformat.Manifest{Key: WindowsKey, ProfileID: "cartulary.reference_pack.enrichment.windows_event_ids.v2", Artifacts: []packformat.SourceArtifact{{}}, Files: logicalMemberInventory(map[string][]byte{"payload/entries.ndjson": payload.Bytes()})}
	candidate.Summary.Kind = "entries"
	candidate.Summary.Entries = &count
	if e := packformat.ValidateContent(ctx, candidate, packageSource{"payload/entries.ndjson": payload.Bytes()}, nil); e != nil {
		r.Issues = append(r.Issues, WindowsError{"projection_conflict", "payload", "registered content validator rejected candidate: " + e.Error()})
	}
	evidenceBytes, err := canonicaljson.Marshal(evidence)
	if err != nil {
		return out, err
	}
	// The archive is producer evidence, never a runtime payload member.
	archiveMembers["evidence.json"] = evidenceBytes
	archiveMembers["inventory.json"] = inventoryBytes
	archiveMembers["source-lock.json"] = lockBytes
	archiveMembers["editorial.json"] = editorialBytes
	archiveMembers["original-summaries.json"] = summaryBytes
	var archive bytes.Buffer
	if err = packformat.WriteUSTAR(ctx, &archive, archiveMembers); err != nil {
		return out, err
	}
	r.EntryCount = len(ids)
	r.PayloadFileSHA256 = Digest(payload.Bytes())
	r.EvidenceSHA256 = Digest(archive.Bytes())
	slices.SortFunc(r.Dispositions, func(a, b WindowsDisposition) int {
		aa, _ := canonicaljson.Marshal(a)
		bb, _ := canonicaljson.Marshal(b)
		return bytes.Compare(aa, bb)
	})
	slices.SortFunc(r.Issues, func(a, b WindowsError) int {
		return strings.Compare(a.Code+"\x00"+a.Source+"\x00"+a.Detail, b.Code+"\x00"+b.Source+"\x00"+b.Detail)
	})
	if len(r.Issues) == 0 && len(ids) > 0 {
		r.Status = "accepted"
	}
	out = WindowsProjection{r, evidence, payload.Bytes(), archive.Bytes(), profile}
	if r.Status != "accepted" {
		return out, windowsFailure("projection_conflict", "", "see reconciliation report")
	}
	return out, nil
}

// Presentation removes trailing empty paragraphs/whitespace from text leaves.
// Evidence records and source fragments stay byte-faithful; internal newlines
// and indentation are not collapsed or trimmed from the beginning.
func windowsPresentation(value any) any {
	switch v := value.(type) {
	case string:
		return strings.TrimRightFunc(v, windowsTrim)
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, x := range v {
			out[k] = windowsPresentation(x)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, x := range v {
			out[i] = windowsPresentation(x)
		}
		return out
	default:
		return value
	}
}
func decimalPointer(n *int64) any {
	if n == nil {
		return nil
	}
	return strconv.FormatInt(*n, 10)
}

func WindowsAudit(ctx context.Context, root *os.Root, inventoryBytes, editorialBytes []byte, builder Builder) (WindowsReconciliation, error) {
	out, err := WindowsReconcile(ctx, root, inventoryBytes, editorialBytes, builder)
	if out.Report.SchemaID == "" {
		return out.Report, err
	}
	b, e := canonicaljson.Marshal(out.Report)
	if e != nil {
		return out.Report, e
	}
	base := "audits/" + Digest(b)
	// Publish report last; failed audits contain no successful logical manifest.
	for name, data := range map[string][]byte{"evidence.tar": out.Archive, "entries.candidate.ndjson": out.Payload, "source-profile.json": out.SourceProfile} {
		if e = Retain(root, base+"/"+name, data, 0600); e != nil {
			return out.Report, e
		}
	}
	if e = Retain(root, base+"/reconciliation.json", b, 0600); e != nil {
		return out.Report, e
	}
	return out.Report, err
}
