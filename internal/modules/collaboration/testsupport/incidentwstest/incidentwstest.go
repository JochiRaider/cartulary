package incidentwstest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

const waitTimeout = 5 * time.Second

type ConnectOptions struct {
	SessionToken     string
	Cookies          []*http.Cookie
	Headers          http.Header
	Origin           string
	ClientInstanceID string
	Presence         platformws.PresenceInput
}

type HelloAck struct {
	ConnectionID        string `json:"connection_id"`
	ResumeToken         string `json:"resume_token"`
	ServerTime          string `json:"server_time"`
	HeartbeatIntervalMS int    `json:"heartbeat_interval_ms"`
	PresenceTTLMS       int    `json:"presence_ttl_ms"`
	ResumeWindowMS      int    `json:"resume_window_ms"`
}

type ResumeAck struct {
	Status                   string `json:"status"`
	ResumeToken              string `json:"resume_token"`
	ServerHighWaterStreamSeq int64  `json:"server_high_water_stream_seq"`
}

type Client struct {
	raw              *wstest.Client
	HelloAck         HelloAck
	ResumeAck        ResumeAck
	ReplayedMessages []platformws.Message
	PresenceSnapshot []platformws.PresenceRecord
	events           chan event
}

type event struct {
	message *platformws.Message
	err     error
}

type unexpectedMessageError struct {
	messageType string
}

func (e unexpectedMessageError) Error() string {
	return fmt.Sprintf("unexpected websocket message before close: got %q", e.messageType)
}

func ConnectAndHello(t testing.TB, serverURL string, incidentID string, options ConnectOptions) *Client {
	t.Helper()

	options = normalizeConnectOptions(incidentID, options)

	rawClient := wstest.ConnectWithHeaders(t, serverURL, incidentPath(incidentID), connectHeaders(options))

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()

	if err := rawClient.Send(ctx, platformws.Message{
		Type: "hello",
		Payload: platformws.RawPayload(map[string]any{
			"client_instance_id": options.ClientInstanceID,
			"presence":           options.Presence,
		}),
	}); err != nil {
		t.Fatalf("send canonical incident websocket hello: %v", err)
	}

	ackMessage, err := rawClient.Receive(ctx)
	if err != nil {
		t.Fatalf("receive canonical incident websocket hello_ack: %v", err)
	}
	wstest.RequireMessageType(t, ackMessage, "hello_ack")
	ack := requireHelloAck(t, ackMessage)

	snapshotMessage, err := rawClient.Receive(ctx)
	if err != nil {
		t.Fatalf("receive canonical incident websocket presence_snapshot: %v", err)
	}
	wstest.RequireMessageType(t, snapshotMessage, "presence_snapshot")
	snapshot := requirePresenceSnapshot(t, snapshotMessage)

	client := &Client{
		raw:              rawClient,
		HelloAck:         ack,
		PresenceSnapshot: snapshot,
		events:           make(chan event, 32),
	}
	go client.readLoop()
	return client
}

func ConnectAndResume(t testing.TB, serverURL string, incidentID string, options ConnectOptions, resumeToken string, lastSeenStreamSeq int64) *Client {
	t.Helper()

	options = normalizeConnectOptions(incidentID, options)

	rawClient := wstest.ConnectWithHeaders(t, serverURL, incidentPath(incidentID), connectHeaders(options))

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()

	if err := rawClient.Send(ctx, platformws.Message{
		Type: "resume",
		Payload: platformws.RawPayload(map[string]any{
			"client_instance_id":   options.ClientInstanceID,
			"resume_token":         resumeToken,
			"last_seen_stream_seq": lastSeenStreamSeq,
			"presence":             options.Presence,
		}),
	}); err != nil {
		t.Fatalf("send canonical incident websocket resume: %v", err)
	}

	ackMessage, err := rawClient.Receive(ctx)
	if err != nil {
		t.Fatalf("receive canonical incident websocket resume_ack: %v", err)
	}
	wstest.RequireMessageType(t, ackMessage, "resume_ack")
	ack := requireResumeAck(t, ackMessage)

	replayed := make([]platformws.Message, 0)
	var snapshot []platformws.PresenceRecord
	for {
		message, err := rawClient.Receive(ctx)
		if err != nil {
			t.Fatalf("receive canonical incident websocket resume follow-up message: %v", err)
		}
		if message.Type == "presence_snapshot" {
			snapshot = requirePresenceSnapshot(t, message)
			break
		}
		replayed = append(replayed, message)
	}

	client := &Client{
		raw:              rawClient,
		ResumeAck:        ack,
		ReplayedMessages: replayed,
		PresenceSnapshot: snapshot,
		events:           make(chan event, 32),
	}
	go client.readLoop()
	return client
}

func TryConnect(serverURL string, incidentID string, options ConnectOptions) (*websocket.Conn, *http.Response, error) {
	return wstest.TryConnect(serverURL, incidentPath(incidentID), connectHeaders(options))
}

func RequireBootstrapTokenRejected(t testing.TB, serverURL string, incidentID string, bootstrapToken string) {
	t.Helper()

	_, resp, err := TryConnect(serverURL, incidentID, ConnectOptions{SessionToken: bootstrapToken})
	if err == nil {
		t.Fatal("expected bootstrap-token canonical incident websocket dial to fail")
	}
	if resp == nil {
		t.Fatalf("expected HTTP rejection response for bootstrap-token canonical incident websocket dial, err=%v", err)
		return
	}
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "credential_bootstrap_rejected")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "not_allowed_for_route" {
		t.Fatalf("unexpected websocket bootstrap rejection: %#v", details)
	}
}

func RequireDialErrorEnvelope(t testing.TB, serverURL string, incidentID string, options ConnectOptions, status int, code string) {
	t.Helper()

	_, resp, err := TryConnect(serverURL, incidentID, options)
	if err == nil {
		t.Fatalf("expected canonical incident websocket dial to fail with %s", code)
	}
	if resp == nil {
		t.Fatalf("expected HTTP rejection response for canonical incident websocket dial, err=%v", err)
		return
	}
	httptestx.RequireErrorEnvelope(t, resp, status, code)
}

func RequireDialRejectedStatus(t testing.TB, serverURL string, incidentID string, options ConnectOptions, status int) {
	t.Helper()

	_, resp, err := TryConnect(serverURL, incidentID, options)
	if err == nil {
		t.Fatalf("expected canonical incident websocket dial to fail with status %d", status)
	}
	if resp == nil {
		t.Fatalf("expected HTTP rejection response for canonical incident websocket dial, err=%v", err)
		return
	}
	if resp.StatusCode != status {
		t.Fatalf("unexpected canonical incident websocket rejection status: got %d want %d", resp.StatusCode, status)
	}
}

func (c *Client) Send(ctx context.Context, message platformws.Message) error {
	return c.raw.Send(ctx, message)
}

func (c *Client) AwaitNextMessage(timeout time.Duration) (platformws.Message, error) {
	current, err := c.awaitEvent(timeout)
	if err != nil {
		return platformws.Message{}, err
	}
	if current.err != nil {
		return platformws.Message{}, current.err
	}
	return *current.message, nil
}

func (c *Client) AwaitClose(timeout time.Duration) error {
	current, err := c.awaitEvent(timeout)
	if err != nil {
		return err
	}
	if current.err == nil {
		return unexpectedMessageError{messageType: current.message.Type}
	}
	return current.err
}

func (c *Client) Close(code websocket.StatusCode, reason string) {
	if c == nil || c.raw == nil {
		return
	}
	c.raw.Close(code, reason)
}

func AwaitSessionRevoked(client *Client, wantReasonCode string) error {
	revoked, err := client.AwaitNextMessage(waitTimeout)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf("timed out waiting for session_revoked message")
		case websocket.CloseStatus(err) >= 0:
			return fmt.Errorf("websocket closed before session_revoked message: %w", err)
		default:
			return fmt.Errorf("read session_revoked message: %w", err)
		}
	}
	if revoked.Type != "session_revoked" {
		return fmt.Errorf("unexpected websocket message before close: got %q want %q", revoked.Type, "session_revoked")
	}

	var payload map[string]any
	if err := json.Unmarshal(revoked.Payload, &payload); err != nil {
		return fmt.Errorf("decode session_revoked payload: %w", err)
	}
	if payload["reason_code"] != wantReasonCode {
		return fmt.Errorf("unexpected session_revoked payload reason_code: got %v want %q", payload["reason_code"], wantReasonCode)
	}

	closeErr := client.AwaitClose(waitTimeout)
	switch {
	case closeErr == nil:
		return nil
	case errors.Is(closeErr, context.DeadlineExceeded):
		return fmt.Errorf("timed out waiting for websocket close after session_revoked")
	default:
		var unexpected unexpectedMessageError
		if errors.As(closeErr, &unexpected) {
			return closeErr
		}
		closeStatus := websocket.CloseStatus(closeErr)
		if closeStatus != websocket.StatusPolicyViolation {
			return fmt.Errorf("unexpected websocket close status: got %d want %d: %w", closeStatus, websocket.StatusPolicyViolation, closeErr)
		}
		if !strings.Contains(closeErr.Error(), "session_revoked") {
			return fmt.Errorf("unexpected websocket close error: %v", closeErr)
		}
		return nil
	}
}

func ExpectSessionRevoked(t testing.TB, client *Client, wantReasonCode string) {
	t.Helper()

	if err := AwaitSessionRevoked(client, wantReasonCode); err != nil {
		t.Fatal(err)
	}
}

func AwaitIncidentClosed(client *Client) error {
	message, err := client.AwaitNextMessage(waitTimeout)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf("timed out waiting for incident_closed terminal error")
		case websocket.CloseStatus(err) >= 0:
			return fmt.Errorf("websocket closed before incident_closed terminal error: %w", err)
		default:
			return fmt.Errorf("read incident_closed terminal error: %w", err)
		}
	}
	if message.Type != "error" {
		return fmt.Errorf("unexpected websocket message before close: got %q want %q", message.Type, "error")
	}

	var payload map[string]any
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return fmt.Errorf("decode incident_closed payload: %w", err)
	}
	if payload["code"] != platformws.IncidentTerminalClosed {
		return fmt.Errorf("unexpected terminal error code: got %v want %q", payload["code"], platformws.IncidentTerminalClosed)
	}
	if payload["retryable"] != false {
		return fmt.Errorf("incident_closed retryable must be false: %#v", payload)
	}

	closeErr := client.AwaitClose(waitTimeout)
	switch {
	case closeErr == nil:
		return nil
	case errors.Is(closeErr, context.DeadlineExceeded):
		return fmt.Errorf("timed out waiting for websocket close after incident_closed")
	default:
		var unexpected unexpectedMessageError
		if errors.As(closeErr, &unexpected) {
			return closeErr
		}
		closeStatus := websocket.CloseStatus(closeErr)
		if closeStatus != websocket.StatusPolicyViolation {
			return fmt.Errorf("unexpected websocket close status: got %d want %d: %w", closeStatus, websocket.StatusPolicyViolation, closeErr)
		}
		if !strings.Contains(closeErr.Error(), "incident_closed") {
			return fmt.Errorf("unexpected websocket close error: %v", closeErr)
		}
		return nil
	}
}

func ExpectIncidentClosed(t testing.TB, client *Client) {
	t.Helper()

	if err := AwaitIncidentClosed(client); err != nil {
		t.Fatal(err)
	}
}

func (c *Client) readLoop() {
	defer close(c.events)

	for {
		message, err := c.raw.Receive(context.Background())
		if err != nil {
			c.events <- event{err: err}
			return
		}
		current := message
		c.events <- event{message: &current}
	}
}

func (c *Client) awaitEvent(timeout time.Duration) (event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case current, ok := <-c.events:
		if !ok {
			return event{}, fmt.Errorf("incident websocket event stream closed")
		}
		return current, nil
	case <-ctx.Done():
		return event{}, ctx.Err()
	}
}

func requireHelloAck(t testing.TB, message platformws.Message) HelloAck {
	t.Helper()

	var ack HelloAck
	if err := json.Unmarshal(message.Payload, &ack); err != nil {
		t.Fatalf("decode hello_ack payload: %v", err)
	}
	if _, err := uuid.Parse(ack.ConnectionID); err != nil {
		t.Fatalf("hello_ack connection_id must be a UUID: %#v", ack)
	}
	if ack.ResumeToken == "" {
		t.Fatalf("hello_ack resume_token is required: %#v", ack)
	}
	if _, err := time.Parse(time.RFC3339Nano, ack.ServerTime); err != nil {
		t.Fatalf("hello_ack server_time must be RFC3339Nano: %#v", ack)
	}
	if ack.HeartbeatIntervalMS != int(platformws.HeartbeatInterval/time.Millisecond) {
		t.Fatalf("hello_ack heartbeat_interval_ms = %d want %d", ack.HeartbeatIntervalMS, int(platformws.HeartbeatInterval/time.Millisecond))
	}
	if ack.PresenceTTLMS != int(platformws.PresenceTTL/time.Millisecond) {
		t.Fatalf("hello_ack presence_ttl_ms = %d want %d", ack.PresenceTTLMS, int(platformws.PresenceTTL/time.Millisecond))
	}
	if ack.ResumeWindowMS < int(platformws.ResumeWindow/time.Millisecond) {
		t.Fatalf("hello_ack resume_window_ms = %d want >= %d", ack.ResumeWindowMS, int(platformws.ResumeWindow/time.Millisecond))
	}
	return ack
}

func requireResumeAck(t testing.TB, message platformws.Message) ResumeAck {
	t.Helper()

	var ack ResumeAck
	if err := json.Unmarshal(message.Payload, &ack); err != nil {
		t.Fatalf("decode resume_ack payload: %v", err)
	}
	switch ack.Status {
	case platformws.ResumeStatusReplayed, platformws.ResumeStatusResetNeeded:
	default:
		t.Fatalf("resume_ack status = %q", ack.Status)
	}
	if ack.ResumeToken == "" {
		t.Fatalf("resume_ack resume_token is required: %#v", ack)
	}
	if ack.ServerHighWaterStreamSeq < 0 {
		t.Fatalf("resume_ack server_high_water_stream_seq must be non-negative: %#v", ack)
	}
	return ack
}

func requirePresenceSnapshot(t testing.TB, message platformws.Message) []platformws.PresenceRecord {
	t.Helper()

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(message.Payload, &raw); err != nil {
		t.Fatalf("decode presence_snapshot payload: %v", err)
	}
	presencesRaw, ok := raw["presences"]
	if !ok {
		t.Fatalf("presence_snapshot must include presences[]: %#v", raw)
	}
	var presences []platformws.PresenceRecord
	if err := json.Unmarshal(presencesRaw, &presences); err != nil {
		t.Fatalf("decode presence_snapshot presences[]: %v", err)
	}
	seen := make(map[string]struct{}, len(presences))
	connectionIDs := make([]string, 0, len(presences))
	for _, presence := range presences {
		if presence.ConnectionID == "" {
			t.Fatalf("presence_snapshot entry missing connection_id: %#v", presence)
		}
		if _, ok := seen[presence.ConnectionID]; ok {
			t.Fatalf("presence_snapshot contains duplicate connection_id %s", presence.ConnectionID)
		}
		seen[presence.ConnectionID] = struct{}{}
		connectionIDs = append(connectionIDs, presence.ConnectionID)
	}
	if !sort.StringsAreSorted(connectionIDs) {
		t.Fatalf("presence_snapshot connection_ids not sorted ascending: %#v", connectionIDs)
	}
	return presences
}

func normalizeConnectOptions(incidentID string, options ConnectOptions) ConnectOptions {
	if options.ClientInstanceID == "" {
		options.ClientInstanceID = "incident-ws-test-" + incidentID
	}
	if options.Presence.SheetRef == nil {
		options.Presence = platformws.PresenceInput{
			SheetRef: map[string]string{
				"kind": "view_schema",
				"id":   "cartulary.view.timeline.v2",
			},
			Mode: "viewing",
		}
	}
	return options
}

func connectHeaders(options ConnectOptions) http.Header {
	headers := http.Header{}
	for key, values := range options.Headers {
		for _, value := range values {
			headers.Add(key, value)
		}
	}
	if options.SessionToken != "" {
		headers.Set("Authorization", "Bearer "+options.SessionToken)
	}
	if options.Origin != "" {
		headers.Set("Origin", options.Origin)
	}
	for _, cookie := range options.Cookies {
		if cookie == nil {
			continue
		}
		headers.Add("Cookie", cookie.String())
	}
	return headers
}

func incidentPath(incidentID string) string {
	return "/ws/v1/incidents/" + incidentID
}
