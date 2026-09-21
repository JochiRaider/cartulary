// Package historycontract defines the closed public History value boundary and
// pure projection mechanics. Source owners select fields and semantic kinds;
// this package has no source-family dispatch or persistence dependencies.
package historycontract

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"
)

const SchemaID = "cartulary.history_diff.v1"

const RepresentationGeneration = "cartulary.history.1"

var ErrInvalidFacts = errors.New("history: invalid retained facts")

var publicFieldKey = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

type Value struct {
	State string `json:"state"`
	Value any    `json:"value,omitempty"`
}

type Change struct {
	FieldKey string `json:"field_key"`
	Before   Value  `json:"before"`
	After    Value  `json:"after"`
}

type Unit struct {
	UnitRef   string   `json:"unit_ref"`
	Kind      string   `json:"kind"`
	Operation string   `json:"operation"`
	RecordIDs []string `json:"record_ids"`
	Changes   []Change `json:"changes"`
}

type Summary struct {
	SchemaID string `json:"schema_id"`
	Summary  string `json:"summary"`
	Units    []Unit `json:"units"`
}

// Facts contains canonical values already admitted by the Revisions catalog.
// Row values are canonical envelopes; non-row values are source-owned facts.
type Facts struct {
	RecordID  string
	Operation string
	Before    map[string]any
	After     map[string]any
}

type Projector func(Facts) ([]Unit, error)

type Field struct {
	Key      string
	Member   string
	Type     string
	Kind     string
	Required bool
}

// Fields declares same-named scalar members using a public field namespace.
func Fields(prefix, valueType, members string) []Field {
	fields := []Field{}
	for _, member := range strings.Fields(members) {
		fields = append(fields, Field{Key: prefix + "." + member, Member: member, Type: valueType, Kind: "field"})
	}
	return fields
}

func Source(snapshot map[string]any) map[string]any {
	if snapshot == nil {
		return nil
	}
	value, _ := snapshot["source"].(map[string]any)
	return value
}

func Record(snapshot map[string]any) map[string]any {
	if snapshot == nil {
		return nil
	}
	value, _ := snapshot["record"].(map[string]any)
	return value
}

func Changes(before, after map[string]any, fields []Field, includeUnchanged bool) ([]Change, error) {
	changes := []Change{}
	seen := map[string]bool{}
	for _, field := range fields {
		if field.Key == "" || field.Member == "" || seen[field.Key] {
			return nil, ErrInvalidFacts
		}
		seen[field.Key] = true
		left, err := fieldValue(before, field)
		if err != nil {
			return nil, err
		}
		right, err := fieldValue(after, field)
		if err != nil {
			return nil, err
		}
		if !includeUnchanged && reflect.DeepEqual(left, right) {
			continue
		}
		if left.State == "absent" && right.State == "absent" {
			continue
		}
		changes = append(changes, Change{FieldKey: field.Key, Before: left, After: right})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].FieldKey < changes[j].FieldKey })
	return changes, nil
}

func fieldValue(object map[string]any, field Field) (Value, error) {
	raw, present := object[field.Member]
	if object != nil && field.Required && (!present || raw == nil) {
		return Value{}, fmt.Errorf("%w: required %s", ErrInvalidFacts, field.Key)
	}
	if !present {
		return Value{State: "absent"}, nil
	}
	if raw == nil {
		return Value{State: "null"}, nil
	}
	valid := false
	switch field.Type {
	case "text":
		_, valid = raw.(string)
	case "boolean":
		_, valid = raw.(bool)
	case "number":
		switch value := raw.(type) {
		case json.Number:
			_, err := value.Float64()
			valid = err == nil
		case float64, int, int64:
			valid = true
		}
	case "strings":
		encoded, err := json.Marshal(raw)
		var values []string
		valid = err == nil && json.Unmarshal(encoded, &values) == nil && values != nil
		if valid {
			raw = values
		}
	case "uuid":
		value, ok := raw.(string)
		id, err := uuid.Parse(value)
		valid = ok && err == nil && id != uuid.Nil && id.String() == value
	}
	if !valid {
		return Value{}, fmt.Errorf("%w: value type for %s", ErrInvalidFacts, field.Key)
	}
	return Value{State: "present", Value: raw}, nil
}

// Operation describes the observed transition, independent of storage command
// spelling (including reversal commands). It never controls action admission.
func Operation(before, after map[string]any) string {
	if before == nil {
		return "create"
	}
	if after == nil {
		return "delete"
	}
	if before["deleted_at"] == nil && after["deleted_at"] != nil {
		return "delete"
	}
	if before["deleted_at"] != nil && after["deleted_at"] == nil {
		return "restore"
	}
	return "update"
}

func NewUnit(kind, operation, identity string, recordIDs []string, changes []Change) Unit {
	digest := sha256.Sum256([]byte(kind + ":" + identity))
	ids := append([]string{}, recordIDs...)
	sort.Strings(ids)
	unique := ids[:0]
	for _, id := range ids {
		if len(unique) == 0 || unique[len(unique)-1] != id {
			unique = append(unique, id)
		}
	}
	if changes == nil {
		changes = []Change{}
	}
	return Unit{UnitRef: "hunit_" + base64.RawURLEncoding.EncodeToString(digest[:]), Kind: kind, Operation: operation, RecordIDs: unique, Changes: changes}
}

// Row projects only the source owner's declared public fields and the generic
// record lifecycle. No raw snapshot or storage locator crosses this boundary.
func Row(facts Facts, fields []Field) ([]Unit, error) {
	before, after := Source(facts.Before), Source(facts.After)
	if before == nil && after == nil {
		return nil, ErrInvalidFacts
	}
	for _, source := range []map[string]any{before, after} {
		if source == nil {
			continue
		}
		for _, field := range fields {
			if _, present := source[field.Member]; !present {
				return nil, fmt.Errorf("%w: missing canonical field %s", ErrInvalidFacts, field.Key)
			}
		}
	}
	changes, err := Changes(before, after, fields, false)
	if err != nil {
		return nil, err
	}
	kinds := map[string]string{}
	for _, field := range fields {
		kinds[field.Key] = field.Kind
	}
	operation := Operation(Record(facts.Before), Record(facts.After))
	units := []Unit{}
	for _, change := range changes {
		kind := kinds[change.FieldKey]
		if kind == "" {
			kind = "field"
		}
		unitOperation := operation
		if kind == "record" && strings.HasSuffix(change.FieldKey, ".merged_into_record_id") && change.After.State == "present" {
			unitOperation = "merge"
		}
		units = append(units, NewUnit(kind, unitOperation, facts.RecordID+":"+change.FieldKey, []string{facts.RecordID}, []Change{change}))
	}
	lifecycle, err := Changes(Record(facts.Before), Record(facts.After), Fields("record", "text", "deleted_at deleted_by_user_id"), false)
	if err != nil {
		return nil, err
	}
	if operation != "update" || len(lifecycle) > 0 || len(units) == 0 {
		units = append(units, NewUnit("record", operation, facts.RecordID, []string{facts.RecordID}, lifecycle))
	}
	return units, nil
}

// Collection preserves complete relation context on updates as well as adds
// and removals. Identity fields are semantic public references, never target_id.
func Collection(facts Facts, kind string, fields []Field, identityMembers, recordMembers []string) ([]Unit, error) {
	if facts.Before == nil && facts.After == nil {
		return nil, ErrInvalidFacts
	}
	changes, err := Changes(facts.Before, facts.After, fields, true)
	if err != nil {
		return nil, err
	}
	identity := facts.After
	if identity == nil {
		identity = facts.Before
	}
	parts := []string{}
	for _, member := range identityMembers {
		value, ok := identity[member].(string)
		if !ok || value == "" {
			return nil, fmt.Errorf("%w: identity %s", ErrInvalidFacts, member)
		}
		parts = append(parts, value)
	}
	ids := []string{}
	for _, object := range []map[string]any{facts.Before, facts.After} {
		for _, member := range recordMembers {
			if object[member] == nil {
				continue
			}
			value, err := fieldValue(object, Field{Key: member, Member: member, Type: "uuid"})
			if err != nil {
				return nil, err
			}
			ids = append(ids, value.Value.(string))
		}
	}
	operation := Operation(facts.Before, facts.After)
	if operation == "create" {
		operation = "add"
	}
	if operation == "delete" {
		operation = "remove"
	}
	return []Unit{NewUnit(kind, operation, strings.Join(parts, ":"), ids, changes)}, nil
}

func Summarize(units []Unit) (Summary, error) {
	if len(units) == 0 {
		return Summary{}, ErrInvalidFacts
	}
	sort.Slice(units, func(i, j int) bool {
		if units[i].Kind != units[j].Kind {
			return units[i].Kind < units[j].Kind
		}
		if len(units[i].Changes) > 0 && len(units[j].Changes) > 0 && units[i].Changes[0].FieldKey != units[j].Changes[0].FieldKey {
			return units[i].Changes[0].FieldKey < units[j].Changes[0].FieldKey
		}
		return units[i].UnitRef < units[j].UnitRef
	})
	seen := map[string]bool{}
	for _, unit := range units {
		if unit.UnitRef == "" || seen[unit.UnitRef] || len(unit.RecordIDs) == 0 {
			return Summary{}, ErrInvalidFacts
		}
		seen[unit.UnitRef] = true
		fields := map[string]bool{}
		if unit.Kind != "record" && len(unit.Changes) == 0 {
			return Summary{}, ErrInvalidFacts
		}
		for _, change := range unit.Changes {
			if !publicFieldKey.MatchString(change.FieldKey) || fields[change.FieldKey] || !validPublicValue(change.Before) || !validPublicValue(change.After) {
				return Summary{}, ErrInvalidFacts
			}
			fields[change.FieldKey] = true
		}
		switch unit.Kind {
		case "field", "link", "mention", "tag", "evidence_association", "capture_state", "record", "entity_identifier", "indicator_observation", "indicator_interval":
		default:
			return Summary{}, ErrInvalidFacts
		}
		switch unit.Operation {
		case "create", "update", "delete", "restore", "merge", "add", "remove":
		default:
			return Summary{}, ErrInvalidFacts
		}
		for _, id := range unit.RecordIDs {
			parsed, err := uuid.Parse(id)
			if err != nil || parsed == uuid.Nil || parsed.String() != id {
				return Summary{}, ErrInvalidFacts
			}
		}
	}
	label := strings.ReplaceAll(units[0].Kind, "_", " ") + " " + units[0].Operation
	if len(units) > 1 {
		label = fmt.Sprintf("%d changes", len(units))
	}
	return Summary{SchemaID: SchemaID, Summary: label, Units: units}, nil
}

func validPublicValue(value Value) bool {
	switch value.State {
	case "absent", "null":
		return value.Value == nil
	case "present":
		switch typed := value.Value.(type) {
		case string, bool, int, int64:
			return true
		case []string:
			return typed != nil
		case json.Number:
			number, err := typed.Float64()
			return err == nil && !math.IsNaN(number) && !math.IsInf(number, 0)
		case float64:
			return !math.IsNaN(typed) && !math.IsInf(typed, 0)
		}
	}
	return false
}
