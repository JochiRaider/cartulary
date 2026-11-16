package queryengine

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestHostPageSQLIsKeysetBoundedAndProjectionOwned(t *testing.T) {
	t.Run("base-table detector", checkBaseTableDetector)
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000000910")
	positionID := "00000000-0000-0000-0000-000000000911"
	window := querypage.Window{
		Limit:    25,
		Position: map[string]string{"record_id": `"` + positionID + `"`},
	}
	query := viewschema.QueryMeta{
		Sort: []viewschema.SortEntry{{FieldKey: "record_id", Direction: "asc"}},
	}
	sqlText, args, err := buildHostQueryPageSQL(incidentID, query, window)
	if err != nil {
		t.Fatalf("build host page SQL: %v", err)
	}
	if !strings.Contains(sqlshape.Text(sqlText), "host_grid_projection") || referencesBaseTable(sqlshape.Text(sqlText), "hosts") {
		t.Fatalf("host page selection does not isolate projection storage: %s", sqlText)
	}
	if !strings.Contains(sqlshape.Text(sqlText), "record_id >") || !strings.Contains(sqlshape.Text(sqlText), " LIMIT $") || strings.Contains(strings.ToUpper(sqlshape.Text(sqlText)), "OFFSET") {
		t.Fatalf("host page SQL is not bounded keyset retrieval: %s", sqlText)
	}
	if !slices.ContainsFunc(args, func(arg any) bool { return fmt.Sprint(arg) == positionID }) {
		t.Fatalf("%s page SQL does not bind keyset position %s: %#v", "host", positionID, args)
	}
	if got := args[len(args)-1]; got != 26 {
		t.Fatalf("host page SQL limit argument = %#v, want 26", got)
	}
}

// referencesBaseTable reports whether sql reads table as a relation (FROM or
// JOIN, optionally schema-qualified or quoted) or qualifies a column with it.
func referencesBaseTable(sql, table string) bool {
	name := regexp.QuoteMeta(table)
	pattern := regexp.MustCompile(`(?i)\b(?:from|join)\s+(?:"?\w+"?\.)?"?` + name + `"?(?:\s|$|,|\))|(?:^|[^\w.])"?` + name + `"?\.`)
	return pattern.MatchString(sql)
}

// checkBaseTableDetector is the detector's own control, run inside the routed
// Host page test.
func checkBaseTableDetector(t *testing.T) {
	t.Helper()
	for _, sql := range []string{
		"SELECT 1 FROM hosts",
		"SELECT 1 FROM public.hosts h",
		`SELECT 1 FROM "hosts"`,
		"SELECT 1 FROM host_grid_projection p JOIN hosts h ON h.record_id = p.record_id",
		"SELECT hosts.display_name FROM host_grid_projection",
	} {
		if !referencesBaseTable(sql, "hosts") {
			t.Fatalf("detector missed base-table access: %s", sql)
		}
	}
	for _, sql := range []string{
		"SELECT 1 FROM host_grid_projection p",
		"SELECT p.hosts_count FROM host_grid_projection p",
	} {
		if referencesBaseTable(sql, "hosts") {
			t.Fatalf("detector flagged projection-only SQL: %s", sql)
		}
	}
}
