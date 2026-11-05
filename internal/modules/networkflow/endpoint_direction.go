package networkflow

import (
	"cmp"
	"encoding/json"
	"errors"
	"net/netip"
	"strings"
)

const (
	deriveCiscoOrientation      = "cisco_flow_search_orientation_v1"
	deriveDeclaredDirection     = "declared_source_destination_v1"
	directionClientServer       = "client_to_server"
	directionSourceDestination  = "source_to_destination"
	directionUnknown            = "unknown"
	fieldInputFirstIP           = "network_flow.input_first_ip"
	fieldInputSecondIP          = "network_flow.input_second_ip"
	fieldInputFirstPort         = "network_flow.input_first_port"
	fieldInputSecondPort        = "network_flow.input_second_port"
	fieldInputFirstOrientation  = "network_flow.input_first_orientation"
	fieldInputSecondOrientation = "network_flow.input_second_orientation"
	fieldDirection              = "network_flow.direction"
	fieldDirectionEvidence      = "network_flow.direction_evidence"
	fieldFromIP                 = "network_flow.from_ip"
	fieldToIP                   = "network_flow.to_ip"
	fieldFromPort               = "network_flow.from_port"
	fieldToPort                 = "network_flow.to_port"
	fieldDirectionKind          = "network_flow.direction_kind"
	fieldDirectionReason        = "network_flow.direction_reason"
)

func directionalQueryField(field string) bool {
	return field == fieldFromIP || field == fieldToIP || field == fieldFromPort || field == fieldToPort
}

// Endpoint position is an identity convention, never a traffic role.
type flowEndpoint struct {
	IP   string `json:"ip"`
	Port *int32 `json:"port"`
}

type flowDirection struct {
	Kind         string `json:"kind"`
	FromEndpoint string `json:"from_endpoint,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

func (d flowDirection) valid() bool {
	switch d.Kind {
	case directionClientServer, directionSourceDestination:
		return (d.FromEndpoint == "a" || d.FromEndpoint == "b") && d.Reason == ""
	case directionUnknown:
		return d.FromEndpoint == "" && (d.Reason == "missing_orientation" || d.Reason == "incomplete_orientation")
	}
	return false
}

func (d flowDirection) resource() map[string]any {
	r := map[string]any{"kind": d.Kind}
	if d.Kind == directionUnknown {
		r["reason"] = d.Reason
	} else {
		r["from_endpoint"] = d.FromEndpoint
	}
	return r
}

func compareFlowEndpoints(a, b flowEndpoint) int {
	// Inputs have already passed ip_literal_v1; As16 deliberately preserves
	// mapped IPv6 as IPv6, while family is compared independently.
	aa, bb := netip.MustParseAddr(a.IP), netip.MustParseAddr(b.IP)
	if aa.Is4() != bb.Is4() {
		if aa.Is4() {
			return -1
		}
		return 1
	}
	if c := aa.Compare(bb); c != 0 {
		return c
	}
	if a.Port == nil {
		if b.Port == nil {
			return 0
		}
		return -1
	}
	if b.Port == nil {
		return 1
	}
	return cmp.Compare(*a.Port, *b.Port)
}

// Retained endpoints must already be canonical. Only import normalization may
// change source spellings; storage and claim admission never repair evidence.
func canonicalEndpointsValid(a, b flowEndpoint) bool {
	for _, endpoint := range []flowEndpoint{a, b} {
		ip, err := parseIPLiteral(endpoint.IP)
		if err != nil || ip != endpoint.IP || endpoint.Port != nil && (*endpoint.Port < 0 || *endpoint.Port > 65535) {
			return false
		}
	}
	return compareFlowEndpoints(a, b) <= 0
}

func orientationRolesConflict(first, second string) bool {
	return (first == "client" || first == "server") && first == second
}

// Scalar success is recorded by membership in values, including valid blanks.
// Unrelated field failures must not suppress a check whose operands are valid.
func orientationConflictDiagnostics(record csvRecord, mapping approvedMapping, values map[string]any, fields map[string]fieldMapping) []rejectedRowDiagnostic {
	first, firstValid := values[fieldInputFirstOrientation].(string)
	second, secondValid := values[fieldInputSecondOrientation].(string)
	if !firstValid || !secondValid || !orientationRolesConflict(first, second) {
		return nil
	}
	diagnostics := make([]rejectedRowDiagnostic, 0, 2)
	for _, key := range []string{fieldInputFirstOrientation, fieldInputSecondOrientation} {
		ordinal := fields[key].SourceColumnOrdinal
		diagnostics = append(diagnostics, diagnostic(record.SourceRowNumber, sourceColumnOrdinalPtr(ordinal), headerHashPtr(mapping, ordinal), stringPtr(key), "network_flow_orientation_conflict", "orientation_conflict", record.Fields[ordinal-1]))
	}
	return diagnostics
}

type endpointDerivationResult struct {
	A, B          flowEndpoint
	Direction     flowDirection
	FirstEndpoint string
}

// deriveEndpoints is shared by preview, both apply passes and invariant tests.
// Unsupported roles are rejected before this function; equal nonempty roles
// are a relationship conflict rather than invalid scalar tokens.
func deriveEndpoints(algorithm string, first, second flowEndpoint, firstRole, secondRole string) (endpointDerivationResult, error) {
	r := endpointDerivationResult{A: first, B: second, FirstEndpoint: "a"}
	if compareFlowEndpoints(first, second) > 0 {
		r.A, r.B, r.FirstEndpoint = second, first, "b"
	}
	fromFirst := true
	switch algorithm {
	case deriveDeclaredDirection:
		if firstRole != "" || secondRole != "" {
			return r, errors.New("unexpected orientation")
		}
		r.Direction.Kind = directionSourceDestination
	case deriveCiscoOrientation:
		for _, role := range []string{firstRole, secondRole} {
			if role != "" && role != "client" && role != "server" {
				return r, errors.New("invalid orientation")
			}
		}
		if firstRole == "" || secondRole == "" {
			r.Direction = flowDirection{Kind: directionUnknown, Reason: "incomplete_orientation"}
			if firstRole == "" && secondRole == "" {
				r.Direction.Reason = "missing_orientation"
			}
			return r, nil
		}
		if orientationRolesConflict(firstRole, secondRole) {
			return r, errors.New("orientation conflict")
		}
		r.Direction.Kind, fromFirst = directionClientServer, firstRole == "client"
	default:
		return r, errors.New("unsupported endpoint derivation")
	}
	r.Direction.FromEndpoint = r.FirstEndpoint
	if !fromFirst {
		if r.Direction.FromEndpoint == "a" {
			r.Direction.FromEndpoint = "b"
		} else {
			r.Direction.FromEndpoint = "a"
		}
	}
	if compareFlowEndpoints(first, second) == 0 {
		r.Direction.FromEndpoint = "a"
	}
	return r, nil
}

func normalizeOrientation(raw string) (string, bool) {
	value := strings.Trim(raw, " ")
	buf := []byte(value)
	for i, c := range buf {
		if c >= 'A' && c <= 'Z' {
			buf[i] = c + ('a' - 'A')
		}
	}
	value = string(buf)
	return value, value == "" || value == "client" || value == "server"
}

func counterSemanticsForProfile(id string) string {
	p, ok := mappingSourceProfile(id)
	if !ok {
		return ""
	}
	return p.CounterSemantics
}

func normalizeRecordEndpoints(record csvRecord, mapping approvedMapping, values map[string]any, fields map[string]fieldMapping) (endpointDerivationResult, map[string]any, error) {
	port := func(v any) *int32 {
		if v == nil {
			return nil
		}
		p := int32(v.(int))
		return &p
	}
	role := func(key string) string { s, _ := values[key].(string); return s }
	firstRole, secondRole := role(fieldInputFirstOrientation), role(fieldInputSecondOrientation)
	r, err := deriveEndpoints(mapping.EndpointDerivationID,
		flowEndpoint{values[fieldInputFirstIP].(string), port(values[fieldInputFirstPort])},
		flowEndpoint{values[fieldInputSecondIP].(string), port(values[fieldInputSecondPort])}, firstRole, secondRole)
	if err != nil {
		return r, nil, err
	}
	inputs := make([]any, 0, 6)
	for _, key := range []string{fieldInputFirstIP, fieldInputFirstPort, fieldInputFirstOrientation, fieldInputSecondIP, fieldInputSecondPort, fieldInputSecondOrientation} {
		binding, ok := fields[key]
		if !ok {
			continue
		}
		ordinal := binding.SourceColumnOrdinal
		inputs = append(inputs, map[string]any{"input_key": key, "source_column_ordinal": ordinal, "raw_header_sha256": mapping.SourceColumns[ordinal-1].RawHeaderSHA256, "raw_value_sha256": sha256Hex([]byte(record.Fields[ordinal-1]))})
	}
	evidence := map[string]any{"derivation_id": mapping.EndpointDerivationID, "first_endpoint": r.FirstEndpoint, "first_role": firstRole, "second_role": secondRole, "inputs": inputs}
	return r, evidence, nil
}

func (row flowRow) endpoints() (flowEndpoint, flowEndpoint) {
	return flowEndpoint{row.EndpointAIP, row.EndpointAPort}, flowEndpoint{row.EndpointBIP, row.EndpointBPort}
}

func (row flowRow) directedEndpoints() (flowEndpoint, flowEndpoint, bool) {
	a, b := row.endpoints()
	if !row.Direction.valid() || row.Direction.Kind == directionUnknown {
		return flowEndpoint{}, flowEndpoint{}, false
	}
	if row.Direction.FromEndpoint == "b" {
		return b, a, true
	}
	return a, b, true
}

// validateObservationDirection is the owner invariant at write and restore
// boundaries. Evidence is required even when unmapped source columns are ignored.
func validateObservationDirection(a, b flowEndpoint, direction flowDirection, raw json.RawMessage, profileID string) error {
	if !canonicalEndpointsValid(a, b) || !direction.valid() {
		return errInvalidNormalizedRow
	}
	var evidence struct {
		DerivationID  string            `json:"derivation_id"`
		FirstEndpoint string            `json:"first_endpoint"`
		FirstRole     string            `json:"first_role"`
		SecondRole    string            `json:"second_role"`
		Inputs        []json.RawMessage `json:"inputs"`
	}
	if _, err := decodeClosedImportObject(raw, "derivation_id first_endpoint first_role second_role inputs", "", "", &evidence); err != nil {
		return errInvalidNormalizedRow
	}
	profile, ok := mappingSourceProfile(profileID)
	if !ok || evidence.DerivationID != profile.EndpointDerivationID || evidence.FirstEndpoint != "a" && evidence.FirstEndpoint != "b" {
		return errInvalidNormalizedRow
	}
	first, second := a, b
	if evidence.FirstEndpoint == "b" {
		first, second = b, a
	}
	derived, err := deriveEndpoints(evidence.DerivationID, first, second, evidence.FirstRole, evidence.SecondRole)
	if err != nil || derived.Direction != direction || derived.FirstEndpoint != evidence.FirstEndpoint {
		return errInvalidNormalizedRow
	}
	required := []string{fieldInputFirstIP, fieldInputFirstPort, fieldInputSecondIP, fieldInputSecondPort}
	if evidence.DerivationID == deriveCiscoOrientation {
		required = []string{fieldInputFirstIP, fieldInputFirstPort, fieldInputFirstOrientation, fieldInputSecondIP, fieldInputSecondPort, fieldInputSecondOrientation}
	}
	if len(evidence.Inputs) != len(required) {
		return errInvalidNormalizedRow
	}
	ordinals := map[int64]bool{}
	for index, rawInput := range evidence.Inputs {
		var input struct {
			Key     string `json:"input_key"`
			Ordinal int64  `json:"source_column_ordinal"`
			Header  string `json:"raw_header_sha256"`
			Value   string `json:"raw_value_sha256"`
		}
		if _, err := decodeClosedImportObject(rawInput, "input_key source_column_ordinal raw_header_sha256 raw_value_sha256", "", "", &input); err != nil {
			return errInvalidNormalizedRow
		}
		if input.Key != required[index] || ordinals[input.Ordinal] || input.Ordinal < 1 || !validLowerSHA256(input.Header) || !validLowerSHA256(input.Value) {
			return errInvalidNormalizedRow
		}
		ordinals[input.Ordinal] = true
	}
	return nil
}
