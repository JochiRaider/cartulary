package networkflow

import (
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/google/uuid"
	"strings"
	"testing"
)

func TestQueryAuthoringAdmission_Unit(t *testing.T) {
	t.Run("direct table admission", assertDirectTableReadAdmission)
	t.Run("production facade", assertProductionSurface)
	t.Run("semantic dependency and error boundary", assertSemanticBoundary)
	for _, body := range []string{
		`{"field_keys":["unknown"]}`, `{"error_codes":["network_flow_unknown"]}`,
		`{"field_keys":["network_flow.src_ip","network_flow.src_ip"]}`,
		`{"source_row_range":{}}`, `{"source_row_range":{"gte":null,"lte":null}}`,
		`{"source_row_range":{"gte":3,"lte":2}}`, `{"source_row_range":{"gte":1.5}}`,
		`{"source_row_range":null}`, `{"field_keys":null}`,
	} {
		request := `{"schema_id":"cartulary.network_flow.rejected_rows_query_request.v1",` + body[1:]
		_, err := decodeRejectedRowsQueryRequest(strings.NewReader(request), defaultLimits())
		if err == nil || string(err.kind) != "network_flow_invalid_filter" {
			t.Fatalf("%s: %v", request, err)
		}
	}
	for _, body := range []string{
		`[{"field_key":"network_flow.ip_protocol","op":"in","value":[6,17]}]`,
		`[{"field_key":"network_flow.bytes_count","op":"eq","value":"18446744073709551615"}]`,
		`[{"field_key":"network_flow.bytes_count","op":"range","value":{"gte":null,"lte":"18446744073709551615"}}]`,
		`[{"field_key":"network_flow.bytes_count","op":"range","value":{"gte":"0","lte":"18446744073709551615"}}]`,
		`[{"field_key":"network_flow.input_interface","op":"in","value":["01","1","a,b"]}]`,
		`[{"field_key":"network_flow.src_port","op":"is_null"}]`,
		`[{"field_key":"network_flow.flow_start_utc","op":"range","value":{"gte":"2026-07-10T12:00:00.000001Z","lt":"2026-07-10T12:00:00.000002Z"}}]`,
	} {
		if _, err := decodeFilters(json.RawMessage(body), defaultLimits()); err != nil {
			t.Fatalf("%s: %v", body, err)
		}
	}
	for _, body := range []string{
		`[{"field_key":"network_flow.ip_protocol","op":"in","value":[6,6]}]`,
		`[{"field_key":"network_flow.bytes_count","op":"eq","value":"18446744073709551616"}]`,
		`[{"field_key":"network_flow.src_ip","op":"in","value":["2001:db8::1","2001:0db8:0:0:0:0:0:1"]}]`,
		`[{"field_key":"network_flow.flow_start_utc","op":"eq","value":"2026-07-10T12:00:00Z"}]`,
		`[{"field_key":"network_flow.flow_start_utc","op":"range","value":{"gte":"2026-07-10T12:00:00Z","lt":"2026-07-10T12:00:00Z"}}]`,
		`[{"field_key":"network_flow.src_port","op":"is_null","value":null}]`,
	} {
		_, err := decodeFilters(json.RawMessage(body), defaultLimits())
		if err == nil || string(err.kind) != "network_flow_invalid_filter" || (err.details.FilterIndex == nil || *err.details.FilterIndex != 0) || err.details.FieldKey == nil || err.details.Op == nil {
			t.Fatalf("%s: %v", body, err)
		}
	}
	if compareFilterValues(fieldInputInterface, "01", "1") == 0 {
		t.Fatal("text was compared as an integer")
	}
	request, err := decodeRejectedRowsQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.rejected_rows_query_request.v1","field_keys":["network_flow.src_ip","network_flow.dst_ip"],"source_row_range":{"gte":1,"lte":1}}`), defaultLimits())
	if err != nil || len(request.FieldKeys) != 2 {
		t.Fatalf("diagnostic round trip: %+v %v", request, err)
	}
}

func assertDirectTableReadAdmission(t *testing.T) {
	access := &authorizationAccess{err: &admission.Denied{Code: admission.DenialNotVisible}}
	app := &tableQueryApplication{incidentAccess: access}
	who := readIdentity{ActorID: uuid.New(), IncidentID: uuid.New(), SessionID: uuid.New()}
	_, a := app.profiles(context.Background(), who)
	_, b := app.list(context.Background(), who)
	_, c := app.get(context.Background(), who, "invalid")
	_, d := app.rows(context.Background(), who, "invalid", rowQueryRequest{Continuation: true, CursorToken: "invalid"})
	_, e := app.diagnostics(context.Background(), who, "invalid", rejectedRowsQueryRequest{})
	for _, f := range []*semanticFailure{a, b, c, d, e} {
		if f == nil || f.kind != failureAdmission {
			t.Fatalf("direct read bypassed current authority: %v", f)
		}
	}
}
