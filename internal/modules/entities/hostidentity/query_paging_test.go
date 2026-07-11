package hostidentity

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestHostAndIdentityPageSQLIsKeysetBounded(t *testing.T) {
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000000910")
	positionID := "00000000-0000-0000-0000-000000000911"
	window := querypage.Window{Limit: 25, Position: map[string]string{"record_id": `"` + positionID + `"`}}
	query := viewschema.QueryMeta{Sort: []viewschema.SortEntry{{FieldKey: "record_id", Direction: "asc"}}}
	for name, build := range map[string]func(uuid.UUID, viewschema.QueryMeta, querypage.Window) (string, []any, error){
		"host":     buildHostQueryPageSQL,
		"identity": buildIdentityQueryPageSQL,
	} {
		t.Run(name, func(t *testing.T) {
			sqlText, args, err := build(incidentID, query, window)
			if err != nil {
				t.Fatalf("build page SQL: %v", err)
			}
			if !strings.Contains(sqlText, "record_id >") || !strings.Contains(sqlText, " LIMIT $") || strings.Contains(strings.ToUpper(sqlText), "OFFSET") {
				t.Fatalf("page SQL is not bounded keyset retrieval: %s", sqlText)
			}
			if got := args[len(args)-1]; got != 26 {
				t.Fatalf("page SQL limit argument = %#v, want 26", got)
			}
		})
	}
}
