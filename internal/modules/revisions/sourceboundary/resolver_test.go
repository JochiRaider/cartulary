package sourceboundary

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCanonicalBoundaryV1BytesHashAndDefensiveCopy(t *testing.T) {
	testCases := []struct {
		name       string
		input      ResolveInput
		latestID   *uuid.UUID
		latestTime *time.Time
		wantJSON   string
		wantToken  string
	}{
		{
			name: "null latest change set",
			input: ResolveInput{
				IncidentID:      uuid.MustParse("11111111-1111-4111-8111-111111111111"),
				IncidentVersion: 1,
			},
			wantJSON:  `{"incident_id":"11111111-1111-4111-8111-111111111111","incident_version":1,"latest_change_set_id":null,"latest_change_set_created_at":null}`,
			wantToken: "cartulary.source_boundary.v1:7118b27f199de626a091552abc20fb026c5f25f04ca303e0edc790f86ee82478",
		},
		{
			name: "canonical UUID and UTC timestamp",
			input: ResolveInput{
				IncidentID:      uuid.MustParse("11111111-1111-4111-8111-111111111112"),
				IncidentVersion: 7,
			},
			latestID:   uuidPointer(uuid.MustParse("22222222-2222-4222-8222-222222222222")),
			latestTime: timePointer(time.Date(2026, 8, 18, 16, 1, 2, 345678000, time.FixedZone("EDT", -4*60*60))),
			wantJSON:   `{"incident_id":"11111111-1111-4111-8111-111111111112","incident_version":7,"latest_change_set_id":"22222222-2222-4222-8222-222222222222","latest_change_set_created_at":"2026-08-18T20:01:02.345678Z"}`,
			wantToken:  "cartulary.source_boundary.v1:71316dc602a4af569f4e887d5939c01344a2a5ad5f8d47249d5d7368a4555499",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			boundary, err := buildBoundary(testCase.input, testCase.latestID, testCase.latestTime)
			if err != nil {
				t.Fatalf("build boundary: %v", err)
			}
			if got := string(boundary.CanonicalJSON); got != testCase.wantJSON {
				t.Fatalf("canonical JSON = %q, want %q", got, testCase.wantJSON)
			}
			if boundary.Token != testCase.wantToken {
				t.Fatalf("token = %q, want %q", boundary.Token, testCase.wantToken)
			}
			if err := ValidateRetained(boundary, testCase.input.IncidentID); err != nil {
				t.Fatal("producer boundary failed retained validation", err)
			}
			var formatted bytes.Buffer
			if err := json.Indent(&formatted, boundary.CanonicalJSON, "", "  "); err != nil {
				t.Fatal(err)
			}
			if err := ValidateRetained(Boundary{Token: boundary.Token, CanonicalJSON: formatted.Bytes()}, testCase.input.IncidentID); err != nil {
				t.Fatal("storage formatting changed boundary semantics", err)
			}
			boundary.CanonicalJSON[0] = '['
			again, err := buildBoundary(testCase.input, testCase.latestID, testCase.latestTime)
			if err != nil {
				t.Fatalf("rebuild boundary: %v", err)
			}
			if got := string(again.CanonicalJSON); got != testCase.wantJSON {
				t.Fatalf("returned bytes alias resolver state: %q", got)
			}
		})
	}
}

func TestInvalidBoundaryInputsFailClosed(t *testing.T) {
	if _, err := buildBoundary(ResolveInput{}, nil, nil); err == nil {
		t.Fatal("zero input was accepted")
	}
	valid := ResolveInput{IncidentID: uuid.New(), IncidentVersion: 1}
	if _, err := buildBoundary(valid, uuidPointer(uuid.New()), nil); err == nil {
		t.Fatal("partial latest change-set state was accepted")
	}
	boundary, err := buildBoundary(valid, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(boundary.CanonicalJSON, &object); err != nil {
		t.Fatal(err)
	}
	for key := range object {
		changed := make(map[string]json.RawMessage, len(object))
		for k, v := range object {
			if k != key {
				changed[k] = v
			}
		}
		raw, _ := json.Marshal(changed)
		if err := ValidateRetained(Boundary{Token: boundary.Token, CanonicalJSON: raw}, valid.IncidentID); err == nil {
			t.Fatal("omitted member was repaired", key)
		}
	}
	for _, raw := range []string{
		`null`, `[]`, string(boundary.CanonicalJSON) + `{}`,
		strings.Replace(string(boundary.CanonicalJSON), `"incident_version":1`, `"incident_version":1,"future":true`, 1),
		strings.Replace(string(boundary.CanonicalJSON), `"incident_version":1`, `"incident_version":1,"incident_version":1`, 1),
		strings.Replace(string(boundary.CanonicalJSON), `"incident_version":1`, `"incident_version":1.0`, 1),
		strings.Replace(string(boundary.CanonicalJSON), `"incident_version":1`, `"incident_version":0`, 1),
		strings.Replace(string(boundary.CanonicalJSON), `"latest_change_set_id":null`, `"latest_change_set_id":"`+valid.IncidentID.String()+`"`, 1),
	} {
		if err := ValidateRetained(Boundary{Token: boundary.Token, CanonicalJSON: []byte(raw)}, valid.IncidentID); err == nil {
			t.Fatal("invalid retained source boundary admitted")
		}
	}
	if err := ValidateRetained(boundary, uuid.New()); err == nil {
		t.Fatal("foreign incident boundary admitted")
	}
	boundary.Token = tokenPrefix + strings.Repeat("0", 64)
	if err := ValidateRetained(boundary, valid.IncidentID); err == nil {
		t.Fatal("changed digest admitted")
	}
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

func timePointer(value time.Time) *time.Time {
	return &value
}
