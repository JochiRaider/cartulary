package admission

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
)

func TestTimeProfileAdmission_Unit(t *testing.T) {
	t.Parallel()
	t.Run("label", requireTimeProfileLabelAdmission)
	t.Run("enabled", requireTimeProfileEnabledAdmission)
}

func requireTimeProfileEnabledAdmission(t *testing.T) {
	for _, value := range []string{"true", "false", "null", "0", "1", `"false"`, `"true"`, "[]", "{}"} {
		raw := `{"base_profile_version":1,"local_offset_minutes":60,"local_label":null,"enabled":` + value + `}`
		request, apiErr := DecodeTimelineTimeConversionProfilePutRequest(strings.NewReader(raw))
		if value == "true" || value == "false" {
			if apiErr != nil || request.Enabled != (value == "true") || request.BaseProfileVersion != 1 || request.LocalOffsetMinutes == nil || *request.LocalOffsetMinutes != 60 || request.LocalLabel != nil {
				t.Fatalf("boolean %s => %#v/%#v", value, request, apiErr)
			}
			continue
		}
		if apiErr == nil || apiErr.Code != "invalid_mutation_payload" || apiErr.Details["field"] != "enabled" || apiErr.Details["reason_code"] != "invalid_value" {
			t.Fatalf("non-boolean %s => %#v", value, apiErr)
		}
	}
	_, apiErr := DecodeTimelineTimeConversionProfilePutRequest(strings.NewReader(`{"base_profile_version":1,"local_offset_minutes":60,"local_label":null}`))
	if apiErr == nil || apiErr.Code != "invalid_mutation_payload" || apiErr.Details["field"] != "enabled" || apiErr.Details["reason_code"] != "missing_required_field" {
		t.Fatalf("missing boolean => %#v", apiErr)
	}
}

func requireTimeProfileLabelAdmission(t *testing.T) {
	decode := func(value any, present bool) (label *string, code, field, reason string) {
		t.Helper()
		body := map[string]any{"base_profile_version": 1, "enabled": false, "local_offset_minutes": nil}
		if present {
			body["local_label"] = value
		}
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		request, apiErr := DecodeTimelineTimeConversionProfilePutRequest(bytes.NewReader(raw))
		if apiErr != nil {
			return nil, apiErr.Code, apiErr.Details["field"].(string), apiErr.Details["reason_code"].(string)
		}
		return request.LocalLabel, "", "", ""
	}
	for _, test := range []struct{ raw, want string }{
		{" UTC ", "UTC"},
		{"\u2003Cafe\u0301\u00a0", "Café"},
		{"Area\u200dName", "Area\u200dName"},
		{"<script>label</script>", "<script>label</script>"},
		{strings.Repeat("界", 512), strings.Repeat("界", 512)},
		{strings.Repeat("e\u0301", 512), strings.Repeat("é", 512)},
	} {
		label, code, _, _ := decode(test.raw, true)
		if code != "" || label == nil || *label != test.want {
			t.Fatalf("label %q => %v/%s want %q", test.raw, label, code, test.want)
		}
	}
	for _, value := range []any{nil, "", " \u2003 "} {
		if label, code, _, _ := decode(value, true); label != nil || code != "" {
			t.Fatalf("cleared label %v=%v/%s", value, label, code)
		}
	}
	if _, code, field, reason := decode(nil, false); code != "invalid_mutation_payload" || field != "local_label" || reason != "missing_required_field" {
		t.Fatalf("missing label=%s/%s/%s", code, field, reason)
	}
	reject := func(value any) {
		t.Helper()
		if _, code, field, reason := decode(value, true); code != "invalid_mutation_payload" || field != "local_label" || reason != "invalid_value" {
			t.Fatalf("label %q => %s/%s/%s", value, code, field, reason)
		}
	}
	for _, value := range []any{strings.Repeat("界", 513), true, 1, []any{}, map[string]any{}} {
		reject(value)
	}
	for r := rune(0); r <= 159; r++ {
		if r > 31 && r < 127 {
			continue
		}
		for _, value := range []string{string(r) + "UTC", "U" + string(r) + "TC", "UTC" + string(r)} {
			reject(value)
		}
	}
	if _, ok := mutationpolicy.NormalizeProfileLabel(string([]byte{0xff})); ok {
		t.Fatal("profile label policy accepted invalid UTF-8")
	}
}
