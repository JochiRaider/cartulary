package queryengine

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestIdentityPageSQLIsKeysetBoundedAndProjectionOwned(t *testing.T) {
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000000920")
	positionID := "00000000-0000-0000-0000-000000000921"
	window := querypage.Window{
		Limit:    25,
		Position: map[string]string{"record_id": `"` + positionID + `"`},
	}
	query := viewschema.QueryMeta{
		Sort: []viewschema.SortEntry{{FieldKey: "record_id", Direction: "asc"}},
	}
	sqlText, args, err := buildIdentityQueryPageSQL(incidentID, query, window)
	if err != nil {
		t.Fatalf("build identity page SQL: %v", err)
	}
	if !strings.Contains(sqlshape.Text(sqlText), "identity_grid_projection") || referencesBaseTable(sqlshape.Text(sqlText), "identities") {
		t.Fatalf("identity page selection does not isolate projection storage: %s", sqlText)
	}
	if !strings.Contains(sqlshape.Text(sqlText), "record_id >") || !strings.Contains(sqlshape.Text(sqlText), " LIMIT $") || strings.Contains(strings.ToUpper(sqlshape.Text(sqlText)), "OFFSET") {
		t.Fatalf("identity page SQL is not bounded keyset retrieval: %s", sqlText)
	}
	if !slices.ContainsFunc(args, func(arg any) bool { return fmt.Sprint(arg) == positionID }) {
		t.Fatalf("%s page SQL does not bind keyset position %s: %#v", "identity", positionID, args)
	}
	if got := args[len(args)-1]; got != 26 {
		t.Fatalf("identity page SQL limit argument = %#v, want 26", got)
	}
}
