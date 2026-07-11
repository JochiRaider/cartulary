package timeline

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestTimelinePageSQLIsKeysetBounded(t *testing.T) {
	positionID := "00000000-0000-0000-0000-000000000931"
	sqlText, args, err := buildTimelineQueryPageSQL(
		uuid.MustParse("00000000-0000-0000-0000-000000000930"),
		viewschema.QueryMeta{Sort: []viewschema.SortEntry{{FieldKey: "record_id", Direction: "asc"}}},
		querypage.Window{Limit: 25, Position: map[string]string{"record_id": `"` + positionID + `"`}},
	)
	if err != nil {
		t.Fatalf("build timeline page SQL: %v", err)
	}
	if !strings.Contains(sqlText, "record_id >") || !strings.Contains(sqlText, " LIMIT $") || strings.Contains(strings.ToUpper(sqlText), "OFFSET") {
		t.Fatalf("timeline page SQL is not bounded keyset retrieval: %s", sqlText)
	}
	if got := args[len(args)-1]; got != 26 {
		t.Fatalf("timeline page SQL limit argument = %#v, want 26", got)
	}
}
