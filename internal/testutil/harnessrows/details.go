package harnessrows

import (
	"fmt"
	"slices"
	"strings"
)

const (
	detailTailBytes   = 4096
	detailActiveTests = 128
	detailFailedTests = 8
	detailNameBytes   = 512
)

// failureDetails keeps, during the parse that settles outcomes, at most 128
// active 4 KiB output tails and the tails of the first eight failed selected
// tests. Later passing output cannot evict an observed failure. The excerpts
// are diagnostics, never accounting evidence.
type failureDetails struct {
	selectors []string
	active    map[string][]byte
	recent    []string // active tests, least recently written first
	failed    map[string][]byte
	failedSet []string
}

func newFailureDetails(rows []Row) failureDetails {
	var selectors []string
	for _, row := range rows {
		selectors = append(selectors, row.Selectors...)
	}
	return failureDetails{selectors: selectors, active: map[string][]byte{}, failed: map[string][]byte{}}
}

func (d *failureDetails) observe(test, action, output string, hasOutput bool) {
	if !within(test, d.selectors) {
		return
	}
	if hasOutput && len(test) <= detailNameBytes {
		before := d.active[test]
		d.forget(test)
		if len(d.recent) == detailActiveTests {
			d.forget(d.recent[0])
		}
		tail := append(slices.Clip(before), output...)
		d.active[test] = slices.Clone(tail[max(0, len(tail)-detailTailBytes):])
		d.recent = append(d.recent, test)
	}
	if action == "pass" || action == "fail" || action == "skip" {
		if action == "fail" && len(d.failedSet) < detailFailedTests {
			if _, ok := d.failed[test]; !ok {
				d.failedSet = append(d.failedSet, test)
			}
			d.failed[test] = d.active[test]
		}
		d.forget(test)
	}
}

func (d *failureDetails) forget(test string) {
	if _, ok := d.active[test]; !ok {
		return
	}
	delete(d.active, test)
	d.recent = slices.DeleteFunc(d.recent, func(name string) bool { return name == test })
}

func (d *failureDetails) render(order []string, terminal map[string]Observation) string {
	var names []string
	for _, test := range order {
		if terminal[test].Action == "fail" && within(test, d.selectors) {
			names = append(names, test)
		}
	}
	if len(names) == 0 {
		return ""
	}
	selected := names[:min(len(names), detailFailedTests)]
	var out strings.Builder
	for _, name := range selected {
		detail, ok := d.failed[name]
		text := string(detail)
		if !ok || detail == nil {
			text = "detail unavailable within active-test bound"
		}
		fmt.Fprintf(&out, "[GO FAILURE] %s\n%s\n", strings.ToValidUTF8(name[:min(len(name), detailNameBytes)], "�"), text)
	}
	fmt.Fprintf(&out, "[GO FAILURE DETAIL] selected=%d omitted=%d; at most 4 KiB per test\n", len(selected), len(names)-len(selected))
	return out.String()
}
