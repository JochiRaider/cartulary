package querypage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var ErrInvalidPosition = errors.New("invalid query page position")

// Window is the validated database retrieval window derived from an opaque
// workbook cursor. Position values retain the cursor's canonical JSON scalar
// encoding so providers can bind them using their storage field types.
type Window struct {
	Limit    int
	Position map[string]string
}

type Result struct {
	Rows    []map[string]any
	HasMore bool
}

type Field struct {
	Expression string
	Cast       string
}

func Finish(rows []map[string]any, limit int) Result {
	if limit < 1 || len(rows) <= limit {
		return Result{Rows: rows, HasMore: false}
	}
	return Result{Rows: rows[:limit], HasMore: true}
}

func AppendKeyset(builder *strings.Builder, args *[]any, sort []viewschema.SortEntry, fields map[string]Field, position map[string]string) error {
	if len(position) == 0 {
		return nil
	}
	values := make([]any, len(sort))
	placeholders := make([]string, len(sort))
	for index, entry := range sort {
		field, ok := fields[entry.FieldKey]
		if !ok || strings.TrimSpace(field.Expression) == "" {
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
		placeholders[index] = fmt.Sprintf("$%d", len(*args))
		if field.Cast != "" {
			placeholders[index] += "::" + field.Cast
		}
	}

	builder.WriteString("\n   AND (")
	wroteBranch := false
	for index, entry := range sort {
		if values[index] == nil {
			continue
		}
		if wroteBranch {
			builder.WriteString(" OR ")
		}
		builder.WriteByte('(')
		for previous := 0; previous < index; previous++ {
			if previous > 0 {
				builder.WriteString(" AND ")
			}
			builder.WriteString(fields[sort[previous].FieldKey].Expression)
			builder.WriteString(" IS NOT DISTINCT FROM ")
			builder.WriteString(placeholders[previous])
		}
		if index > 0 {
			builder.WriteString(" AND ")
		}
		expr := fields[entry.FieldKey].Expression
		builder.WriteByte('(')
		builder.WriteString(expr)
		if entry.Direction == "desc" {
			builder.WriteString(" < ")
		} else {
			builder.WriteString(" > ")
		}
		builder.WriteString(placeholders[index])
		builder.WriteString(" OR ")
		builder.WriteString(expr)
		builder.WriteString(" IS NULL)")
		builder.WriteByte(')')
		wroteBranch = true
	}
	if !wroteBranch {
		builder.WriteString("FALSE")
	}
	builder.WriteByte(')')
	return nil
}

func AppendLimit(builder *strings.Builder, args *[]any, limit int) error {
	if limit < 1 {
		return fmt.Errorf("query page limit must be positive")
	}
	*args = append(*args, limit+1)
	builder.WriteString(fmt.Sprintf(" LIMIT $%d", len(*args)))
	return nil
}
