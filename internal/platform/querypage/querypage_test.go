package querypage

import (
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestAppendKeysetBuildsNullsLastLexicographicPredicate(t *testing.T) {
	var builder sqlshape.Builder
	args := []any{"incident"}
	err := AppendKeyset(&builder, &args, []viewschema.SortEntry{
		{FieldKey: "state", Direction: "asc"},
		{FieldKey: "record_id", Direction: "desc"},
	}, map[string]Field{
		"state":     {Expression: sqlshape.Literal("p.state")},
		"record_id": {Expression: sqlshape.Literal("p.record_id"), Cast: sqlshape.Literal("uuid")},
	}, map[string]string{
		"state":     `"open"`,
		"record_id": `"00000000-0000-0000-0000-000000000001"`,
	})
	if err != nil {
		t.Fatalf("append keyset: %v", err)
	}
	want := `
   AND (((p.state > $2 OR p.state IS NULL)) OR (p.state IS NOT DISTINCT FROM $2 AND (p.record_id < $3::uuid OR p.record_id IS NULL)))`
	if strings.Join(strings.Fields(sqlshape.Text(builder.Fragment())), "") != strings.Join(strings.Fields(want), "") {
		t.Fatalf("keyset SQL\ngot  %q\nwant %q", sqlshape.Text(builder.Fragment()), want)
	}
	if !reflect.DeepEqual(args, []any{"incident", "open", "00000000-0000-0000-0000-000000000001"}) {
		t.Fatalf("keyset args = %#v", args)
	}
}

func TestAppendKeysetRejectsIncompletePositionAndFinishBoundsRows(t *testing.T) {
	var builder sqlshape.Builder
	err := AppendKeyset(&builder, &[]any{}, []viewschema.SortEntry{
		{FieldKey: "state", Direction: "asc"},
		{FieldKey: "record_id", Direction: "asc"},
	}, map[string]Field{
		"state":     {Expression: sqlshape.Literal("p.state")},
		"record_id": {Expression: sqlshape.Literal("p.record_id"), Cast: sqlshape.Literal("uuid")},
	}, map[string]string{"state": `"open"`})
	if err == nil {
		t.Fatal("incomplete cursor position was accepted")
	}

	builder = sqlshape.Builder{}
	err = AppendKeyset(&builder, &[]any{}, []viewschema.SortEntry{{FieldKey: "record_id", Direction: "asc"}}, map[string]Field{
		"record_id": {Expression: sqlshape.Literal("p.record_id"), Cast: sqlshape.Literal("uuid")},
	}, map[string]string{})
	if err != nil {
		t.Fatalf("empty position is the first page: %v", err)
	}

	rows := []map[string]any{{"record_id": "1"}, {"record_id": "2"}, {"record_id": "3"}}
	page := Finish(rows, 2)
	if !page.HasMore || len(page.Rows) != 2 || page.Rows[1]["record_id"] != "2" {
		t.Fatalf("bounded page = %#v", page)
	}
}
