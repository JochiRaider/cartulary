package harnessruntime

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"unicode"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const testNetworkFlowRandomnessSchemaID = "cartulary.test.network_flow_randomness_control.v1"

const (
	NetworkFlowRandomStreamTableID           = "network_flow.table_id"
	NetworkFlowRandomStreamRowID             = "network_flow.row_id"
	NetworkFlowRandomStreamDiagnosticID      = "network_flow.diagnostic_id"
	NetworkFlowRandomStreamImportJobID       = "network_flow.import_job_id"
	NetworkFlowRandomStreamImportSourceRef   = "network_flow.import_source_ref"
	NetworkFlowRandomStreamCursorNonce       = "network_flow.cursor_nonce"
	NetworkFlowRandomStreamSafeDigestNonce   = "network_flow.safe_digest_nonce"
	NetworkFlowRandomStreamGraphInvocationID = "network_flow.graph_invocation_id"
)

const (
	NetworkFlowRandomValueKindUUID     = "uuid"
	NetworkFlowRandomValueKindToken    = "token"
	NetworkFlowRandomValueKindHexBytes = "hex_bytes"
)

const networkFlowRandomnessExhaustionFailClosed = "fail_closed"

var (
	ErrNetworkFlowRandomnessExhausted    = errors.New("network flow deterministic randomness stream exhausted")
	ErrNetworkFlowRandomnessKindMismatch = errors.New("network flow deterministic randomness value kind mismatch")
)

var (
	networkFlowRandomStreams = map[string]struct{}{
		NetworkFlowRandomStreamTableID:           {},
		NetworkFlowRandomStreamRowID:             {},
		NetworkFlowRandomStreamDiagnosticID:      {},
		NetworkFlowRandomStreamImportJobID:       {},
		NetworkFlowRandomStreamImportSourceRef:   {},
		NetworkFlowRandomStreamCursorNonce:       {},
		NetworkFlowRandomStreamSafeDigestNonce:   {},
		NetworkFlowRandomStreamGraphInvocationID: {},
	}

	networkFlowRandomValueKinds = map[string]struct{}{
		NetworkFlowRandomValueKindUUID:     {},
		NetworkFlowRandomValueKindToken:    {},
		NetworkFlowRandomValueKindHexBytes: {},
	}
)

type NetworkFlowRandomnessRegistry struct {
	mu      sync.Mutex
	streams map[string]*networkFlowRandomnessStream
}

type NetworkFlowRandomnessState struct {
	ControlID      string
	Stream         string
	ValueKind      string
	ValueCount     int
	RemainingCount int
}

type networkFlowRandomnessStream struct {
	controlID string
	stream    string
	valueKind string
	values    []string
	next      int
}

type networkFlowRandomnessService struct {
	guard  httpapi.TestRouteGuard
	random *NetworkFlowRandomnessRegistry
}

type networkFlowRandomnessRequest struct {
	Stream      string   `json:"stream"`
	ValueKind   string   `json:"value_kind"`
	Values      []string `json:"values"`
	ConsumeOnce bool     `json:"consume_once"`
	Exhaustion  string   `json:"exhaustion"`
}

type networkFlowRandomnessResult struct {
	SchemaID       string `json:"schema_id"`
	ControlID      string `json:"control_id"`
	Stream         string `json:"stream"`
	ValueKind      string `json:"value_kind"`
	ValueCount     int    `json:"value_count"`
	RemainingCount int    `json:"remaining_count"`
	ConsumeOnce    bool   `json:"consume_once"`
	Exhaustion     string `json:"exhaustion"`
}

func NewNetworkFlowRandomnessRegistry() *NetworkFlowRandomnessRegistry {
	return &NetworkFlowRandomnessRegistry{streams: map[string]*networkFlowRandomnessStream{}}
}

func RegisterNetworkFlowRandomnessRoutes(random *NetworkFlowRandomnessRegistry) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		if random == nil {
			return fmt.Errorf("register network flow randomness route: randomness registry is required")
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return fmt.Errorf("register network flow randomness route: %w", err)
		}
		service := &networkFlowRandomnessService{
			guard:  guard,
			random: random,
		}
		mux.HandleFunc("POST /api/v1/test/runtime/network-flow-randomness", service.handleArm)
		return nil
	}
}

func (r *NetworkFlowRandomnessRegistry) ConsumeNetworkFlowRandomString(stream string) (string, bool, error) {
	return r.consume(stream, NetworkFlowRandomValueKindToken)
}

func (r *NetworkFlowRandomnessRegistry) ConsumeNetworkFlowRandomUUID(stream string) (uuid.UUID, bool, error) {
	value, ok, err := r.consume(stream, NetworkFlowRandomValueKindUUID)
	if err != nil || !ok {
		return uuid.UUID{}, ok, err
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, false, fmt.Errorf("parse deterministic UUID: %w", err)
	}
	return parsed, true, nil
}

func (r *NetworkFlowRandomnessRegistry) ConsumeNetworkFlowRandomHexBytes(stream string) ([]byte, bool, error) {
	value, ok, err := r.consume(stream, NetworkFlowRandomValueKindHexBytes)
	if err != nil || !ok {
		return nil, ok, err
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, false, fmt.Errorf("decode deterministic hex bytes: %w", err)
	}
	return decoded, true, nil
}

func (r *NetworkFlowRandomnessRegistry) NetworkFlowRandomnessState(stream string) (NetworkFlowRandomnessState, bool) {
	if r == nil {
		return NetworkFlowRandomnessState{}, false
	}
	stream = strings.TrimSpace(stream)
	r.mu.Lock()
	defer r.mu.Unlock()
	configured, ok := r.streams[stream]
	if !ok {
		return NetworkFlowRandomnessState{}, false
	}
	return configured.state(), true
}

func (r *NetworkFlowRandomnessRegistry) Clear() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.streams = map[string]*networkFlowRandomnessStream{}
}

func (r *NetworkFlowRandomnessRegistry) arm(stream networkFlowRandomnessStream) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.streams == nil {
		r.streams = map[string]*networkFlowRandomnessStream{}
	}
	if _, exists := r.streams[stream.stream]; exists {
		return false
	}
	copy := stream
	copy.values = append([]string{}, stream.values...)
	r.streams[stream.stream] = &copy
	return true
}

func (r *NetworkFlowRandomnessRegistry) consume(stream string, valueKind string) (string, bool, error) {
	if r == nil {
		return "", false, nil
	}
	stream = strings.TrimSpace(stream)
	valueKind = strings.TrimSpace(valueKind)
	r.mu.Lock()
	defer r.mu.Unlock()
	configured, ok := r.streams[stream]
	if !ok {
		return "", false, nil
	}
	if configured.valueKind != valueKind {
		return "", false, ErrNetworkFlowRandomnessKindMismatch
	}
	if configured.next >= len(configured.values) {
		return "", false, ErrNetworkFlowRandomnessExhausted
	}
	value := configured.values[configured.next]
	configured.next++
	return value, true, nil
}

func (s networkFlowRandomnessStream) state() NetworkFlowRandomnessState {
	remaining := len(s.values) - s.next
	if remaining < 0 {
		remaining = 0
	}
	return NetworkFlowRandomnessState{
		ControlID:      s.controlID,
		Stream:         s.stream,
		ValueKind:      s.valueKind,
		ValueCount:     len(s.values),
		RemainingCount: remaining,
	}
}

func (s *networkFlowRandomnessService) handleArm(w http.ResponseWriter, r *http.Request) {
	if !s.guard.Authorize(w, r) {
		return
	}
	request, err := decodeNetworkFlowRandomnessRequest(r)
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_network_flow_randomness_request", "invalid Network Flow randomness request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	stream, err := request.networkFlowRandomnessStream()
	if err != nil {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_network_flow_randomness_request", "invalid Network Flow randomness request", map[string]any{
			"reason": err.Error(),
		})
		return
	}
	if !s.random.arm(stream) {
		_ = httpapi.WriteError(w, r, http.StatusConflict, "test_network_flow_random_stream_already_armed", "Network Flow randomness stream is already armed", map[string]any{})
		return
	}
	state := stream.state()
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, networkFlowRandomnessResult{
		SchemaID:       testNetworkFlowRandomnessSchemaID,
		ControlID:      state.ControlID,
		Stream:         state.Stream,
		ValueKind:      state.ValueKind,
		ValueCount:     state.ValueCount,
		RemainingCount: state.RemainingCount,
		ConsumeOnce:    true,
		Exhaustion:     networkFlowRandomnessExhaustionFailClosed,
	})
}

func decodeNetworkFlowRandomnessRequest(r *http.Request) (networkFlowRandomnessRequest, error) {
	var request networkFlowRandomnessRequest
	if r.Body == nil {
		return request, errors.New("body is required")
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, fmt.Errorf("decode body: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return request, errors.New("body must contain a single JSON object")
	}
	return request, nil
}

func (r networkFlowRandomnessRequest) networkFlowRandomnessStream() (networkFlowRandomnessStream, error) {
	stream := strings.TrimSpace(r.Stream)
	valueKind := strings.TrimSpace(r.ValueKind)
	exhaustion := strings.TrimSpace(r.Exhaustion)
	if _, ok := networkFlowRandomStreams[stream]; !ok {
		return networkFlowRandomnessStream{}, errors.New("stream is not a supported Network Flow randomness stream")
	}
	if _, ok := networkFlowRandomValueKinds[valueKind]; !ok {
		return networkFlowRandomnessStream{}, errors.New("value_kind is not supported")
	}
	if len(r.Values) == 0 || len(r.Values) > 256 {
		return networkFlowRandomnessStream{}, errors.New("values must include 1..256 deterministic entries")
	}
	if !r.ConsumeOnce {
		return networkFlowRandomnessStream{}, errors.New("consume_once must be true")
	}
	if exhaustion != networkFlowRandomnessExhaustionFailClosed {
		return networkFlowRandomnessStream{}, errors.New("exhaustion must be fail_closed")
	}
	values := make([]string, len(r.Values))
	for i, value := range r.Values {
		normalized := strings.TrimSpace(value)
		if err := validateNetworkFlowRandomnessValue(valueKind, normalized); err != nil {
			return networkFlowRandomnessStream{}, fmt.Errorf("values[%d]: %w", i, err)
		}
		values[i] = normalized
	}
	return networkFlowRandomnessStream{
		controlID: uuid.NewString(),
		stream:    stream,
		valueKind: valueKind,
		values:    values,
	}, nil
}

func validateNetworkFlowRandomnessValue(valueKind string, value string) error {
	switch valueKind {
	case NetworkFlowRandomValueKindUUID:
		parsed, err := uuid.Parse(value)
		if err != nil {
			return errors.New("must be a UUID")
		}
		if parsed.String() != value {
			return errors.New("must be canonical lowercase UUID text")
		}
	case NetworkFlowRandomValueKindToken:
		if !isNetworkFlowRandomnessToken(value) {
			return errors.New("must be an ASCII token no longer than 128 characters")
		}
	case NetworkFlowRandomValueKindHexBytes:
		if !isNetworkFlowRandomnessHexBytes(value) {
			return errors.New("must be lowercase even-length hex bytes no longer than 512 characters")
		}
	default:
		return errors.New("unsupported value kind")
	}
	return nil
}

func isNetworkFlowRandomnessToken(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if r > unicode.MaxASCII {
			return false
		}
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '.', '_', ':', '-':
			continue
		default:
			return false
		}
	}
	return true
}

func isNetworkFlowRandomnessHexBytes(value string) bool {
	if value == "" || len(value) > 512 || len(value)%2 != 0 {
		return false
	}
	for _, r := range value {
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'a' && r <= 'f' {
			continue
		}
		return false
	}
	return true
}
