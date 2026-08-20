package valuecodec

import (
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCanonicalScalarAndCollectionValues(t *testing.T) {
	text := "value"
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	instant := time.Date(2026, 4, 10, 15, 4, 5, 123456789, time.FixedZone("offset", -5*60*60))

	cases := []struct {
		name  string
		value any
		want  string
	}{
		{name: "nil string", value: OptionalString(nil), want: `null`},
		{name: "string", value: OptionalString(&text), want: `"value"`},
		{name: "nil UUID", value: OptionalUUID(nil), want: `null`},
		{name: "UUID", value: OptionalUUID(&id), want: `"11111111-1111-4111-8111-111111111111"`},
		{name: "timestamp", value: Timestamp(instant), want: `"2026-04-10T20:04:05.123456789Z"`},
		{name: "optional timestamp", value: OptionalTimestamp(&instant), want: `"2026-04-10T20:04:05.123456789Z"`},
		{name: "nil timestamp", value: OptionalTimestamp(nil), want: `null`},
		{name: "date", value: OptionalDate(&instant), want: `"2026-04-10"`},
		{name: "nil date", value: OptionalDate(nil), want: `null`},
		{name: "nil collection", value: Collection(true, nil), want: `{"items":[],"kind":"collection_value_v1","ordered":true}`},
		{name: "empty collection", value: Collection(false, []map[string]any{}), want: `{"items":[],"kind":"collection_value_v1","ordered":false}`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			encoded, err := json.Marshal(testCase.value)
			if err != nil {
				t.Fatalf("marshal canonical value: %v", err)
			}
			if got := string(encoded); got != testCase.want {
				t.Fatalf("canonical value = %s, want %s", got, testCase.want)
			}
		})
	}
}

func TestCanonicalJSONSHA256(t *testing.T) {
	got := hex.EncodeToString(CanonicalJSONSHA256(map[string]any{
		"nullable": nil,
		"ordered":  Collection(true, nil),
	}))
	const want = "e547f31a0554351ae8cb2d6693a075cf93307bdae6fe32468bc9aa518fe93bc9"
	if got != want {
		t.Fatalf("canonical hash = %s, want %s", got, want)
	}
}
