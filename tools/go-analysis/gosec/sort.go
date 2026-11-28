package main

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"github.com/securego/gosec/v2/issue"
)

// sortIssues orders issues as the gosec command does: by severity, rule,
// file and first line, each descending.
func sortIssues(issues []*issue.Issue) {
	slices.SortFunc(issues, func(i, j *issue.Issue) int {
		return -cmp.Or(
			cmp.Compare(i.Severity, j.Severity),
			cmp.Compare(i.What, j.What),
			cmp.Compare(i.File, j.File),
			cmp.Compare(firstLine(i.Line), firstLine(j.Line)),
		)
	})
}

// firstLine is the first line of an issue's line or line range.
func firstLine(lines string) int {
	line, _ := strconv.Atoi(strings.Split(lines, "-")[0])
	return line
}
