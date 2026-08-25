package wstest

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

type Client struct {
	Conn     *websocket.Conn
	Response *http.Response
}

type StatusCode = websocket.StatusCode
type MessageType = websocket.MessageType

const (
	MessageText                   = websocket.MessageText
	MessageBinary                 = websocket.MessageBinary
	StatusNormalClosure           = websocket.StatusNormalClosure
	StatusUnsupportedData         = websocket.StatusUnsupportedData
	StatusInvalidFramePayloadData = websocket.StatusInvalidFramePayloadData
	StatusPolicyViolation         = websocket.StatusPolicyViolation
	StatusMessageTooBig           = websocket.StatusMessageTooBig
)

func Connect(t testing.TB, serverURL string, path string) *Client {
	return ConnectWithHeaders(t, serverURL, path, nil)
}

func ConnectWithHeaders(t testing.TB, serverURL string, path string, headers http.Header) *Client {
	t.Helper()

	client, _, err := TryConnect(serverURL, path, headers)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() {
		client.Close(websocket.StatusNormalClosure, "test_cleanup")
	})

	return client
}

func TryConnect(serverURL string, path string, headers http.Header) (*Client, *http.Response, error) {
	target, err := websocketURL(serverURL, path)
	if err != nil {
		return nil, nil, err
	}

	conn, response, err := websocket.Dial(context.Background(), target, &websocket.DialOptions{HTTPHeader: headers})
	if err != nil {
		return nil, response, err
	}
	return &Client{Conn: conn, Response: response}, response, nil
}

func (c *Client) Send(ctx context.Context, message platformws.Message) error {
	return wsjson.Write(ctx, c.Conn, message)
}

func (c *Client) Receive(ctx context.Context) (platformws.Message, error) {
	var message platformws.Message
	err := wsjson.Read(ctx, c.Conn, &message)
	return message, err
}

func (c *Client) WriteJSON(ctx context.Context, value any) error {
	return wsjson.Write(ctx, c.Conn, value)
}

func (c *Client) ReadJSON(ctx context.Context, value any) error {
	return wsjson.Read(ctx, c.Conn, value)
}

func (c *Client) Handshake(ctx context.Context) (platformws.Message, error) {
	if err := c.Send(ctx, platformws.Message{
		Type:    "handshake",
		Payload: platformws.RawPayload(map[string]any{"mode": "replay"}),
	}); err != nil {
		return platformws.Message{}, err
	}
	return c.Receive(ctx)
}

func (c *Client) Close(code websocket.StatusCode, reason string) {
	if c == nil || c.Conn == nil {
		return
	}
	_ = c.Conn.Close(code, reason)
}

func (c *Client) CloseNow() error {
	if c == nil || c.Conn == nil {
		return nil
	}
	return c.Conn.CloseNow()
}

func (c *Client) Write(ctx context.Context, kind websocket.MessageType, payload []byte) error {
	return c.Conn.Write(ctx, kind, payload)
}

func (c *Client) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	return c.Conn.Read(ctx)
}

func CloseStatus(err error) StatusCode {
	return websocket.CloseStatus(err)
}

func RequireMessageType(t testing.TB, message platformws.Message, want string) {
	t.Helper()
	if message.Type != want {
		t.Fatalf("unexpected websocket message type: got %q want %q", message.Type, want)
	}
}

func RequireSessionRevoked(t testing.TB, message platformws.Message) {
	t.Helper()
	RequireMessageType(t, message, "session_revoked")
}

func RequireClose(t testing.TB, err error, wantStatus websocket.StatusCode, wantReason string) {
	t.Helper()
	closeStatus := websocket.CloseStatus(err)
	if closeStatus != wantStatus {
		t.Fatalf("unexpected websocket close status: got %d want %d", closeStatus, wantStatus)
	}
	if wantReason != "" && !strings.Contains(err.Error(), wantReason) {
		t.Fatalf("unexpected websocket close error: %v", err)
	}
}

func RequireConnectionRefused(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected websocket dial to fail")
	}
	message := strings.ToLower(err.Error())
	if !strings.Contains(message, "connection refused") && !strings.Contains(message, "connect: cannot assign requested address") {
		t.Fatalf("expected websocket dial to fail before listener startup, got %v", err)
	}
}

type ReplayLifecycle struct {
	ResumeToken string
	Mode        string
}

func RequireReplayResetScaffold(t testing.TB, lifecycle ReplayLifecycle) {
	t.Helper()
	if lifecycle.Mode == "" {
		t.Fatal("expected replay lifecycle mode")
	}
}

type LifecycleEvent struct {
	MessageType string
	CloseStatus websocket.StatusCode
}

func RequireLifecycleEvent(t testing.TB, event LifecycleEvent) {
	t.Helper()
	if event.MessageType == "" && event.CloseStatus == 0 {
		t.Fatal("expected websocket lifecycle event to describe a message or close status")
	}
}

func websocketURL(serverURL string, path string) (string, error) {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return "", err
	}
	switch parsed.Scheme {
	case "http":
		parsed.Scheme = "ws"
	case "https":
		parsed.Scheme = "wss"
	}
	parsed.Path = path
	return parsed.String(), nil
}
