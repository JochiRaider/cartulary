// Package maintenancequery exposes closed Recovery and harness maintenance SQL.
// Relation sets and commands are generated from canonical machine catalogs;
// source-owner and projection read paths cannot import this capability.
package maintenancequery

import (
	"slices"
	"sort"
)

func sameTables(actual, expected []string) bool {
	actual = slices.Clone(actual)
	sort.Strings(actual)
	return slices.Equal(actual, expected)
}
