package producer

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

func applyWindowsProviderExclusions(exclusions []WindowsProviderExclusion, sources map[string]WindowsAcquired, blobs map[string][]byte, report *WindowsReconciliation) error {
	rowsBySource := map[string]map[string]windowsCSVRow{}
	excluded := map[string]bool{}
	for _, exclusion := range exclusions {
		key := exclusion.SourceID + "/" + exclusion.Locator
		if excluded[key] || sources[exclusion.SourceID].Source.Kind != "combined_csv" {
			return windowsFailure("input_invalid", key, "unique retained CSV exclusion required")
		}
		if rowsBySource[exclusion.SourceID] == nil {
			rows, err := parseWindowsCSV(blobs[exclusion.SourceID])
			if err != nil {
				return err
			}
			rowsBySource[exclusion.SourceID] = map[string]windowsCSVRow{}
			for _, row := range rows {
				rowsBySource[exclusion.SourceID][fmt.Sprintf("csv/record/%d", row.Number)] = row
			}
		}
		row, ok := rowsBySource[exclusion.SourceID][exclusion.Locator]
		if !ok || row.Fields["record_kind"] != "event_reference" || Digest([]byte(row.Raw)) != exclusion.FragmentSHA256 {
			return windowsFailure("input_invalid", key, "reviewed exclusion row changed or absent")
		}
		index := slices.IndexFunc(report.Dispositions, func(d WindowsDisposition) bool {
			return d.SourceID == exclusion.SourceID && d.Locator == exclusion.Locator && d.Disposition == "blocking" && d.Reason == "provider_evidence_missing"
		})
		if index < 0 {
			return windowsFailure("input_invalid", key, "reviewed exclusion no longer has an unresolved provider")
		}
		report.Dispositions[index].Disposition = "excluded"
		report.Dispositions[index].Reason = "user_excluded_provider_unresolved"
		excluded[key] = true
	}
	report.Issues = slices.DeleteFunc(report.Issues, func(issue WindowsError) bool {
		return issue.Code == "identity_conflict" && excluded[issue.Source]
	})
	if len(excluded) != 0 {
		report.Limitations = append(report.Limitations, fmt.Sprintf("%d reviewed CSV event/version rows are excluded at the user's direction because provider identity remains unresolved; retained source evidence is unchanged.", len(excluded)))
	}
	return nil
}

// The reviewed input is deliberately finite. Finding an ID elsewhere does not
// create an association, and a stale or unused association blocks reconciliation.
func validateWindowsProviderBindings(bindings []WindowsProviderBinding, sources map[string]WindowsAcquired, blobs map[string][]byte) error {
	seen := map[string]bool{}
	csvDigests := map[string]map[string]string{}
	for _, b := range bindings {
		key := b.CSVSourceID + "/" + b.CSVLocator
		if seen[key] || sources[b.CSVSourceID].Source.Kind != "combined_csv" || sources[b.SourceID].Source.Kind != "event_document" {
			return windowsFailure("input_invalid", key, "unique CSV to event-document provider binding required")
		}
		seen[key] = true
		if csvDigests[b.CSVSourceID] == nil {
			rows, err := parseWindowsCSV(blobs[b.CSVSourceID])
			if err != nil {
				return err
			}
			csvDigests[b.CSVSourceID] = map[string]string{}
			for _, row := range rows {
				csvDigests[b.CSVSourceID][fmt.Sprintf("csv/record/%d", row.Number)] = Digest([]byte(row.Raw))
			}
		}
		if csvDigests[b.CSVSourceID][b.CSVLocator] != b.CSVSHA256 || Digest(blobs[b.SourceID]) != b.SourceSHA256 {
			return windowsFailure("input_invalid", key, "reviewed provider binding bytes changed or locator absent")
		}
		fragment, err := windowsFragment(sources[b.SourceID].Source, blobs[b.SourceID], b.Locator)
		if err != nil || Digest([]byte(fragment)) != b.FragmentSHA256 {
			return windowsFailure("input_invalid", key, "reviewed provider System fragment changed or absent")
		}
	}
	return nil
}

var windowsProviderSampleLocator = regexp.MustCompile(`^html/pre/[0-9]+/text\(\)$`)

// A reviewed association selects one authoritative provider-qualified source.
// It does not require another original to repeat that source's facts.
func providerFromWindowsBinding(row windowsCSVRow, binding WindowsProviderBinding, sources map[string]WindowsAcquired, blobs map[string][]byte) (string, *string, error) {
	fail := func(s string) (string, *string, error) { return "", nil, errors.New(s) }
	if Digest([]byte(row.Raw)) != binding.CSVSHA256 || row.Fields["record_kind"] != "event_reference" {
		return fail("exact event row required")
	}
	source := sources[binding.SourceID].Source
	if source.Kind != "event_document" || Digest(blobs[binding.SourceID]) != binding.SourceSHA256 {
		return fail("reviewed event document bytes changed or absent")
	}
	fragment, err := windowsFragment(source, blobs[binding.SourceID], binding.Locator)
	if err != nil || !windowsProviderSampleLocator.MatchString(binding.Locator) || Digest([]byte(fragment)) != binding.FragmentSHA256 {
		return fail("explicit retained XML sample required")
	}
	system, err := windowsSampleSystem(fragment)
	if err != nil {
		return fail("independently parseable System required")
	}
	for _, field := range [][2]string{{"EventID", row.Fields["event_id"]}, {"Version", row.Fields["event_version"]}} {
		n := xmlOne(system, field[0])
		if n == nil || strings.TrimSpace(n.Text) != field[1] {
			return fail("sample event ID/version does not match original row")
		}
	}
	p := xmlOne(system, "Provider")
	if p == nil {
		return fail("sample provider absent")
	}
	provider, ok := windowsProvider(p.Attr["Name"])
	if !ok {
		return fail("sample provider outside reviewed family inventory")
	}
	guid, err := sourceGUID(p.Attr["Guid"])
	if err != nil {
		return fail("invalid sample provider GUID")
	}
	return provider, guid, nil
}
