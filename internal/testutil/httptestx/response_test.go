package httptestx

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type observedBody struct {
	io.Reader
	readErr, closeErr error
	closes            int
}

func (b *observedBody) Read(p []byte) (int, error) {
	if b.readErr != nil {
		return 0, b.readErr
	}
	return b.Reader.Read(p)
}

func (b *observedBody) Close() error {
	b.closes++
	return b.closeErr
}

func responseFixture(status int, payload string) (*http.Response, *observedBody) {
	body := &observedBody{Reader: strings.NewReader(payload)}
	return &http.Response{StatusCode: status, Header: http.Header{http.CanonicalHeaderKey(httpapi.RequestIDHeader): []string{"request-1"}}, Body: body}, body
}

func TestResponseBodyOwnership(t *testing.T) {
	readFailure, closeFailure := errors.New("read failed"), errors.New("close failed")
	for _, tc := range []struct {
		name              string
		status            int
		readErr, closeErr error
		wantErr           bool
	}{
		{"binary success", 200, nil, nil, false},
		{"wrong status", 503, nil, nil, true},
		{"read failure", 200, readFailure, nil, true},
		{"close failure", 200, nil, closeFailure, true},
		{"all failures", 503, readFailure, closeFailure, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, body := responseFixture(tc.status, "\x00\xffbytes")
			body.readErr, body.closeErr = tc.readErr, tc.closeErr
			got, err := readResponseBody(resp, http.StatusOK)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v", err)
			}
			if tc.readErr != nil && !errors.Is(err, tc.readErr) {
				t.Fatal("read failure lost")
			}
			if tc.closeErr != nil && !errors.Is(err, tc.closeErr) {
				t.Fatal("close failure lost")
			}
			if tc.readErr == nil && string(got) != "\x00\xffbytes" {
				t.Fatal("raw bytes changed")
			}
			if body.closes != 1 || resp.Body != body {
				t.Fatalf("body closes=%d; ownership replaced=%v", body.closes, resp.Body != body)
			}
			if resp.Header.Get(httpapi.RequestIDHeader) != "request-1" {
				t.Fatal("response header changed")
			}
		})
	}
	resp, body := responseFixture(200, "public reader")
	if string(ReadResponseBody(t, resp, 200)) != "public reader" || body.closes != 1 {
		t.Fatal("public reader lost ownership")
	}
}

func TestJSONResponseFraming(t *testing.T) {
	for _, payload := range []string{"", "null", "[]", `"string"`, `{`, `{} {}`, `{} trailing`} {
		t.Run(payload, func(t *testing.T) {
			resp, body := responseFixture(200, payload)
			if _, err := readJSONResponse(resp, 200); err == nil {
				t.Fatal("invalid object accepted")
			}
			if body.closes != 1 {
				t.Fatal("invalid JSON body not closed")
			}
		})
	}
	resp, body := responseFixture(200, " {\"number\": 2, \"added\": true} \n")
	value := ReadJSONBody(t, resp)
	if value["number"] != float64(2) || value["added"] != true || body.closes != 1 {
		t.Fatal("valid object representation or ownership changed")
	}
}

func TestNoContentResponse(t *testing.T) {
	for _, tc := range []struct {
		status  int
		payload string
		wantErr bool
	}{
		{204, "", false}, {200, "", true}, {204, "unexpected", true},
	} {
		resp, body := responseFixture(tc.status, tc.payload)
		if err := readNoContent(resp); (err != nil) != tc.wantErr {
			t.Fatalf("status=%d body=%q error=%v", tc.status, tc.payload, err)
		}
		if body.closes != 1 {
			t.Fatal("no-content body not closed")
		}
	}
	resp, body := responseFixture(204, "")
	RequireNoContent(t, resp)
	if RequireRequestID(t, resp) != "request-1" || body.closes != 1 {
		t.Fatal("no-content header/ownership changed")
	}
}

func TestSuccessEnvelopeContract(t *testing.T) {
	for _, tc := range []struct {
		name, payload     string
		noHeader, wantErr bool
	}{
		{"object", `{"data":{},"meta":{"request_id":"request-1"}}`, false, false},
		{"batch and additions", `{"data":[{"id":1}],"meta":{"request_id":"request-1","future":true},"extra":1}`, false, false},
		{"null resource", `{"data":null,"meta":{"request_id":"request-1"}}`, false, false},
		{"missing data", `{"meta":{"request_id":"request-1"}}`, false, true},
		{"missing meta", `{"data":{}}`, false, true},
		{"wrong meta type", `{"data":{},"meta":[]}`, false, true},
		{"missing id", `{"data":{},"meta":{}}`, false, true},
		{"wrong id", `{"data":{},"meta":{"request_id":"different"}}`, false, true},
		{"missing header", `{"data":{},"meta":{"request_id":"request-1"}}`, true, true},
		{"trailing", `{"data":{},"meta":{"request_id":"request-1"}} {}`, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, body := responseFixture(200, tc.payload)
			if tc.noHeader {
				resp.Header.Del(httpapi.RequestIDHeader)
			}
			if _, err := readSuccessEnvelope(resp, 200); (err != nil) != tc.wantErr {
				t.Fatalf("error=%v", err)
			}
			if body.closes != 1 {
				t.Fatal("success envelope body not closed")
			}
			if !tc.wantErr {
				resp, _ = responseFixture(200, tc.payload)
				RequireSuccessEnvelope(t, resp, 200)
			}
		})
	}
	resp, body := responseFixture(200, `{"data":{"id":1},"meta":{"request_id":"request-1"}}`)
	if RequireSuccessData(t, resp, 200)["id"] != float64(1) || body.closes != 1 {
		t.Fatal("route object data or ownership changed")
	}
	for _, payload := range []string{`{"data":[],"meta":{"request_id":"request-1"}}`, `{"data":null,"meta":{"request_id":"request-1"}}`} {
		resp, body := responseFixture(200, payload)
		if _, err := readSuccessData(resp, 200); err == nil || body.closes != 1 {
			t.Fatal("route object assertion accepted a non-object or leaked its body")
		}
	}

}

func TestErrorEnvelopeContract(t *testing.T) {
	valid := `{"error":{"code":"invalid_request","message":"Invalid request","status":400,"request_id":"request-1","retryable":false}}`
	for _, tc := range []struct {
		name, payload     string
		noHeader, wantErr bool
	}{
		{"optional details absent", valid, false, false},
		{"details and additions", strings.Replace(valid, `"retryable":false`, `"retryable":false,"details":{"field":"name"},"future":true`, 1), false, false},
		{"missing error", `{}`, false, true},
		{"wrong error type", `{"error":[]}`, false, true},
		{"missing code", strings.Replace(valid, `"code":"invalid_request",`, "", 1), false, true},
		{"wrong code", strings.Replace(valid, "invalid_request", "other", 1), false, true},
		{"missing message", strings.Replace(valid, `"message":"Invalid request",`, "", 1), false, true},
		{"empty message", strings.Replace(valid, "Invalid request", " ", 1), false, true},
		{"wrong message type", strings.Replace(valid, `"Invalid request"`, `null`, 1), false, true},
		{"missing status", strings.Replace(valid, `"status":400,`, "", 1), false, true},
		{"wrong status", strings.Replace(valid, `400`, `401`, 1), false, true},
		{"string status", strings.Replace(valid, `400`, `"400"`, 1), false, true},
		{"missing id", strings.Replace(valid, `"request_id":"request-1",`, "", 1), false, true},
		{"wrong id", strings.Replace(valid, "request-1", "other", 1), false, true},
		{"missing retryable", strings.Replace(valid, `,"retryable":false`, "", 1), false, true},
		{"wrong retryable type", strings.Replace(valid, `false`, `0`, 1), false, true},
		{"missing header", valid, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, body := responseFixture(400, tc.payload)
			if tc.noHeader {
				resp.Header.Del(httpapi.RequestIDHeader)
			}
			if _, err := readErrorEnvelope(resp, 400, "invalid_request"); (err != nil) != tc.wantErr {
				t.Fatalf("error=%v", err)
			}
			if body.closes != 1 {
				t.Fatal("error envelope body not closed")
			}
			if !tc.wantErr {
				resp, _ = responseFixture(400, tc.payload)
				RequireErrorEnvelope(t, resp, 400, "invalid_request")
			}
		})
	}
}
