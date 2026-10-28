package httptestx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// ReadResponseBody consumes and closes resp.Body, then checks the transport status.
// Headers remain available. The caller must not read or close the body again.
func ReadResponseBody(t testing.TB, resp *http.Response, wantStatus int) []byte {
	t.Helper()
	body, err := readResponseBody(resp, wantStatus)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func readResponseBody(resp *http.Response, wantStatus int) ([]byte, error) {
	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	var statusErr error
	if resp.StatusCode != wantStatus {
		statusErr = fmt.Errorf("unexpected HTTP status: got %d want %d", resp.StatusCode, wantStatus)
	}
	if readErr != nil {
		readErr = fmt.Errorf("read response body: %w", readErr)
	}
	if closeErr != nil {
		closeErr = fmt.Errorf("close response body: %w", closeErr)
	}
	return body, errors.Join(statusErr, readErr, closeErr)
}

// RequireNoContent consumes the response and requires HTTP 204 with an empty body.
func RequireNoContent(t testing.TB, resp *http.Response) {
	t.Helper()
	if err := readNoContent(resp); err != nil {
		t.Fatal(err)
	}
}

func readNoContent(resp *http.Response) error {
	body, err := readResponseBody(resp, http.StatusNoContent)
	if err != nil {
		return err
	}
	if len(body) != 0 {
		return errors.New("expected empty HTTP 204 response body")
	}
	return nil
}

// ReadJSONBody consumes and closes the body and requires one complete JSON object.
// It does not assert a transport status or an envelope contract.
func ReadJSONBody(t testing.TB, resp *http.Response) map[string]any {
	t.Helper()
	body, err := readJSONResponse(resp, resp.StatusCode)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func readJSONResponse(resp *http.Response, wantStatus int) (map[string]any, error) {
	payload, err := readResponseBody(resp, wantStatus)
	if err != nil {
		return nil, err
	}
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, fmt.Errorf("decode response object: %w", err)
	}
	if body == nil {
		return nil, errors.New("expected response object, got null")
	}
	return body, nil
}

// RequireRequestID checks only the header; it does not consume the response.
func RequireRequestID(t testing.TB, resp *http.Response) string {
	t.Helper()
	requestID := resp.Header.Get(httpapi.RequestIDHeader)
	if requestID == "" {
		t.Fatal("missing request id header")
	}
	return requestID
}

// RequireSuccessEnvelope consumes the response and checks the shared JSON contract.
// Resource shape, paging and other route-specific guarantees belong to the caller.
func RequireSuccessEnvelope(t testing.TB, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()
	body, err := readSuccessEnvelope(resp, wantStatus)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// RequireSuccessData consumes a success envelope whose route returns an object.
func RequireSuccessData(t testing.TB, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()
	data, err := readSuccessData(resp, wantStatus)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readSuccessData(resp *http.Response, wantStatus int) (map[string]any, error) {
	body, err := readSuccessEnvelope(resp, wantStatus)
	if err != nil {
		return nil, err
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		return nil, errors.New("expected route success data object")
	}
	return data, nil
}

func readSuccessEnvelope(resp *http.Response, wantStatus int) (map[string]any, error) {
	body, err := readJSONResponse(resp, wantStatus)
	if err != nil {
		return nil, err
	}
	requestID := resp.Header.Get(httpapi.RequestIDHeader)
	if requestID == "" {
		return nil, errors.New("missing request id header")
	}
	meta, ok := body["meta"].(map[string]any)
	if !ok || meta["request_id"] != requestID {
		return nil, errors.New("expected success meta object with matching request_id")
	}
	if _, ok := body["data"]; !ok {
		return nil, errors.New("missing success data member")
	}
	return body, nil
}

// RequireErrorEnvelope consumes the response and checks the shared JSON contract.
// Optional details and code-specific policy remain assertions at their route owner.
func RequireErrorEnvelope(t testing.TB, resp *http.Response, wantStatus int, wantCode string) map[string]any {
	t.Helper()
	body, err := readErrorEnvelope(resp, wantStatus, wantCode)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func readErrorEnvelope(resp *http.Response, wantStatus int, wantCode string) (map[string]any, error) {
	body, err := readJSONResponse(resp, wantStatus)
	if err != nil {
		return nil, err
	}
	requestID := resp.Header.Get(httpapi.RequestIDHeader)
	if requestID == "" {
		return nil, errors.New("missing request id header")
	}
	value, ok := body["error"].(map[string]any)
	if !ok {
		return nil, errors.New("expected error object")
	}
	if value["request_id"] != requestID {
		return nil, errors.New("error.request_id does not match header")
	}
	if value["code"] != wantCode {
		return nil, fmt.Errorf("unexpected error code: got %v want %s", value["code"], wantCode)
	}
	if message, ok := value["message"].(string); !ok || strings.TrimSpace(message) == "" {
		return nil, errors.New("expected nonempty error.message string")
	}
	if value["status"] != float64(wantStatus) {
		return nil, errors.New("error.status does not match transport status")
	}
	if _, ok := value["retryable"].(bool); !ok {
		return nil, errors.New("expected error.retryable boolean")
	}
	return body, nil
}
