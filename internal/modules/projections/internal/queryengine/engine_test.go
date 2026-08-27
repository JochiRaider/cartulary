package queryengine

import (
	"strings"
	"testing"
)

func TestCompileSurfaceOwnsExecutablePlanValidation(t *testing.T) {
	valid := TimelinePlans()[0]
	tests := map[string]struct {
		mutate func(*Surface)
		want   string
	}{
		"empty view": {
			mutate: func(surface *Surface) { surface.ViewSchemaID = "" },
			want:   "empty view_schema_id",
		},
		"unknown view": {
			mutate: func(surface *Surface) { surface.ViewSchemaID = "cartulary.view.missing.v1" },
			want:   "no registered view schema",
		},
		"empty from": {
			mutate: func(surface *Surface) { surface.FromSQL = "" },
			want:   "empty from_sql",
		},
		"statement separator": {
			mutate: func(surface *Surface) { surface.WhereSQL = "TRUE; DROP TABLE projection" },
			want:   "forbidden SQL token",
		},
		"placeholder": {
			mutate: func(surface *Surface) { surface.RecordExpr = "$1" },
			want:   "forbidden SQL token",
		},
		"empty fields": {
			mutate: func(surface *Surface) { surface.Fields = nil },
			want:   "has no query fields",
		},
		"duplicate field": {
			mutate: func(surface *Surface) { surface.Fields = append(surface.Fields, surface.Fields[0]) },
			want:   "duplicate query field",
		},
		"unsupported kind": {
			mutate: func(surface *Surface) { surface.Fields[0].Kind = FieldKind("opaque") },
			want:   "unsupported field kind",
		},
		"missing schema field": {
			mutate: func(surface *Surface) { surface.Fields = surface.Fields[1:] },
			want:   "does not map schema field",
		},
		"unknown schema field": {
			mutate: func(surface *Surface) { surface.Fields[0].Key = "timeline.unknown" },
			want:   "maps unknown schema field",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			surface := cloneSurface(valid)
			test.mutate(&surface)
			_, err := CompileSurface(surface)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("CompileSurface error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestCompileSurfaceReturnsImmutableSchemaDerivedPlan(t *testing.T) {
	input := cloneSurface(TimelinePlans()[0])
	input.GroupingFields = []string{"caller-controlled"}
	compiled, err := CompileSurface(input)
	if err != nil {
		t.Fatalf("CompileSurface: %v", err)
	}
	if len(compiled.GroupingFields) == 0 || compiled.GroupingFields[0] == "caller-controlled" {
		t.Fatalf("grouping fields were not derived from the schema: %#v", compiled.GroupingFields)
	}
	originalKey := compiled.Fields[0].Key
	input.Fields[0].Key = "mutated-input"
	input.GroupingFields[0] = "mutated-input"
	if compiled.Fields[0].Key != originalKey || compiled.GroupingFields[0] == "mutated-input" {
		t.Fatalf("compiled plan aliases input slices: %#v", compiled)
	}
	compiled.Fields[0].Key = "mutated-output"
	again, err := CompileSurface(TimelinePlans()[0])
	if err != nil {
		t.Fatalf("CompileSurface again: %v", err)
	}
	if again.Fields[0].Key != originalKey {
		t.Fatalf("compiled plan mutation escaped: got %q want %q", again.Fields[0].Key, originalKey)
	}
}

func cloneSurface(surface Surface) Surface {
	surface.Fields = append([]Field(nil), surface.Fields...)
	surface.GroupingFields = append([]string(nil), surface.GroupingFields...)
	return surface
}
