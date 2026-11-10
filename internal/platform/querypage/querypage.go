package querypage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var ErrInvalidPosition = errors.New("invalid query page position")

// Window is the validated database retrieval window derived from an opaque
// workbook cursor. Position values retain the cursor's canonical JSON scalar
// encoding so providers can bind them using their storage field types.
type Window struct {
	Limit    int
	Position map[string]string
	// TargetRecordID and Before are private bounded locator selections. They
	// are never admitted from an ordinary public query request.
	TargetRecordID string
	Before         bool
}

type Result struct {
	Rows    []map[string]any
	HasMore bool
}

type Field struct {
	Expression sqlshape.Fragment
	Cast       sqlshape.Fragment
}

func Finish(rows []map[string]any, limit int) Result {
	if limit < 1 || len(rows) <= limit {
		return Result{Rows: rows, HasMore: false}
	}
	return Result{Rows: rows[:limit], HasMore: true}
}

func AppendKeyset(builder *sqlshape.Builder, args *[]any, sort []viewschema.SortEntry, fields map[string]Field, position map[string]string) error {
	return appendKeyset(builder, args, sort, fields, position, false)
}

// AppendBefore selects positions strictly before the boundary in the ordinary
// NULLS LAST comparator. Callers reverse ORDER BY with NULLS FIRST.
func AppendBefore(builder *sqlshape.Builder, args *[]any, sort []viewschema.SortEntry, fields map[string]Field, position map[string]string) error {
	return appendKeyset(builder, args, sort, fields, position, true)
}

func appendKeyset(builder *sqlshape.Builder, args *[]any, sort []viewschema.SortEntry, fields map[string]Field, position map[string]string, before bool) error {
	if len(position) == 0 {
		return nil
	}
	values := make([]any, len(sort))
	placeholders := make([]sqlshape.Fragment, len(sort))
	for index, entry := range sort {
		field, ok := fields[entry.FieldKey]
		if !ok || field.Expression.Empty() {
			return fmt.Errorf("query page sort field %q not mapped", entry.FieldKey)
		}
		encoded, ok := position[entry.FieldKey]
		if !ok {
			return fmt.Errorf("%w: missing sort field %q", ErrInvalidPosition, entry.FieldKey)
		}
		decoder := json.NewDecoder(strings.NewReader(encoded))
		decoder.UseNumber()
		if err := decoder.Decode(&values[index]); err != nil {
			return fmt.Errorf("%w: field %q: %v", ErrInvalidPosition, entry.FieldKey, err)
		}
		*args = append(*args, values[index])
		placeholders[index] = sqlshape.Parameter(len(*args))
		if !field.Cast.Empty() {
			placeholders[index] = sqlshape.Concat(placeholders[index], sqlshape.Literal("::"), field.Cast)
		}
	}

	builder.Write(sqlshape.Literal("\n   AND ("))
	wroteBranch := false
	for index, entry := range sort {
		if values[index] == nil && !before {
			continue
		}
		if wroteBranch {
			builder.Write(sqlshape.Literal(" OR "))
		}
		builder.Write(sqlshape.Literal("("))
		for previous := 0; previous < index; previous++ {
			if previous > 0 {
				builder.Write(sqlshape.Literal(" AND "))
			}
			builder.Write(fields[sort[previous].FieldKey].Expression)
			builder.Write(sqlshape.Literal(" IS NOT DISTINCT FROM "))
			builder.Write(placeholders[previous])
		}
		if index > 0 {
			builder.Write(sqlshape.Literal(" AND "))
		}
		expr := fields[entry.FieldKey].Expression
		builder.Write(sqlshape.Literal("("))
		builder.Write(expr)
		if values[index] == nil {
			builder.Write(sqlshape.Literal(" IS NOT NULL"))
		} else {
			if (entry.Direction == "desc") != before {
				builder.Write(sqlshape.Literal(" < "))
			} else {
				builder.Write(sqlshape.Literal(" > "))
			}
			builder.Write(placeholders[index])
			if !before {
				builder.Write(sqlshape.Format(" OR %s IS NULL", expr))
			}
		}
		builder.Write(sqlshape.Literal(")"))
		builder.Write(sqlshape.Literal(")"))
		wroteBranch = true
	}
	if !wroteBranch {
		builder.Write(sqlshape.Literal("FALSE"))
	}
	builder.Write(sqlshape.Literal(")"))
	return nil
}

func AppendLimit(builder *sqlshape.Builder, args *[]any, limit int) error {
	if limit < 1 {
		return fmt.Errorf("query page limit must be positive")
	}
	*args = append(*args, limit+1)
	builder.Write(sqlshape.Format(" LIMIT %s", sqlshape.Parameter(len(*args))))
	return nil
}
