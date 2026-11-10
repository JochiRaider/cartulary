// Package sqlshape builds read-only queries from closed structural fragments. Literal and
// Format inputs are qualified by the repository's Go source analyzer. Runtime
// values enter only through parameters or Quote; relation names have no dynamic
// constructor. Token boundaries are retained between independently built parts.
package sqlshape

import (
	"errors"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/sqltemplate"
)

// Operator preserves a catalog-resolved PostgreSQL comparison operator without
// admitting SQL text. Names contain only operator characters; comment delimiters
// are forbidden, and the namespace is a single quoted identifier.
func Operator(namespace, name string) (Fragment, error) {
	if name == "" || strings.ContainsAny(namespace, "\x00") || strings.Contains(name, "--") || strings.Contains(name, "/*") || strings.Contains(name, "*/") {
		return Fragment{}, errors.New("invalid SQL operator")
	}
	for _, c := range name {
		if !strings.ContainsRune("+-*/<>=~!@#%^&|", c) {
			return Fragment{}, errors.New("invalid SQL operator")
		}
	}
	return Fragment{text: "OPERATOR(" + Text(Identifier(namespace)) + "." + name + ")"}, nil
}

type Fragment struct{ text string }

// Statement is a complete, statically qualified command template. Unlike read
// fragments, statements cannot be appended to a query or used as substitutions.
type Statement struct{ text string }

func Command(template string, parts ...Fragment) Statement {
	return Statement{text: Text(Format(template, parts...))}
}
func CommandText(statement Statement) string { return statement.text }

// Identifier admits a single quoted column name, never SQL structure. Relation
// positions still require literal names in the enclosing qualified template.
func Identifier(name string) Fragment {
	if strings.ContainsRune(name, 0) {
		panic("SQL identifier contains NUL")
	}
	return Fragment{text: `"` + strings.ReplaceAll(name, `"`, `""`) + `"`}
}

func Literal(text string) Fragment {
	if _, err := sqltemplate.Parse(text, false); err != nil {
		panic(err)
	}
	return Fragment{text: text}
}
func Quote(value string) Fragment {
	return Fragment{text: "E'" + strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "'", "''") + "'"}
}
func Number(value int) Fragment { return Fragment{text: strconv.Itoa(value)} }
func Parameter(index int) Fragment {
	if index < 1 {
		panic("SQL parameter index must be positive")
	}
	return Fragment{text: "$" + strconv.Itoa(index)}
}
func Text(fragment Fragment) string { return fragment.text }
func (f Fragment) Empty() bool      { return strings.TrimSpace(f.text) == "" }

func Concat(parts ...Fragment) Fragment {
	var builder Builder
	for _, part := range parts {
		builder.Write(part)
	}
	return builder.Fragment()
}

// Join retains structural boundaries for each independently qualified fragment.
func Join(parts []Fragment, separator Fragment) Fragment {
	var builder Builder
	for i, part := range parts {
		if i > 0 {
			builder.Write(separator)
		}
		builder.Write(part)
	}
	return builder.Fragment()
}

// Format substitutes complete fragments, never arbitrary strings. Separators
// prevent substitution from joining identifier, operator, or comment tokens.
func Format(template string, parts ...Fragment) Fragment {
	chunks, err := sqltemplate.Parse(template, true)
	if err != nil {
		panic(err)
	}
	if len(chunks) != len(parts)+1 {
		panic("SQL format argument count mismatch")
	}
	var builder Builder
	for i, chunk := range chunks {
		builder.Write(Fragment{text: chunk})
		if i < len(parts) {
			builder.Write(parts[i])
		}
	}
	return builder.Fragment()
}

type Builder struct{ buffer strings.Builder }

func (b *Builder) Write(fragment Fragment) {
	fragment.text = strings.TrimSpace(fragment.text)
	if fragment.text == "" {
		return
	}
	if b.buffer.Len() > 0 {
		b.buffer.WriteByte(' ')
	}
	b.buffer.WriteString(fragment.text)
}
func (b *Builder) Fragment() Fragment { return Fragment{text: b.buffer.String()} }
