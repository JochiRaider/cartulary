package querypage

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestPredecessorNullBoundary_Unit(t *testing.T) {
	for _, direction := range []string{"asc", "desc"} {
		for _, value := range []string{`null`, `"2026-01-01"`} {
			var sql sqlshape.Builder
			args := []any{}
			err := AppendBefore(&sql, &args, []viewschema.SortEntry{{FieldKey: "date", Direction: direction}, {FieldKey: "record_id", Direction: "asc"}}, map[string]Field{"date": {Expression: sqlshape.Literal("p.date")}, "record_id": {Expression: sqlshape.Literal("p.id"), Cast: sqlshape.Literal("uuid")}}, map[string]string{"date": value, "record_id": `"10000000-0000-4000-8000-000000000001"`})
			if err != nil {
				t.Fatal(err)
			}
			text := sqlshape.Text(sql.Fragment())
			if strings.Contains(text, " OR p.date IS NULL") {
				t.Fatalf("null cannot precede non-null: %s", text)
			}
			if value == "null" && !strings.Contains(text, "p.date IS NOT NULL") {
				t.Fatalf("non-null must precede null: %s", text)
			}
			if !strings.Contains(text, "p.date IS NOT DISTINCT FROM $1") || !strings.Contains(text, "p.id < $2 :: uuid") {
				t.Fatalf("tie-break missing: %s", text)
			}
		}
	}
}
