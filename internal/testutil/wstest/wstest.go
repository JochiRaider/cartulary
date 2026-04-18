package wstest

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/coder/websocket"

	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type Client struct {
	Conn     *websocket.Conn
	Response *http.Response
}

func Connect(t testing.TB, serverURL string, path string) *Client {
	return ConnectWithHeaders(t, serverURL, path, nil)
}

func ConnectWithHeaders(t testing.TB, serverURL string, path string, headers http.Header) *Client {
	t.Helper()

	target, err := websocketURL(serverURL, path)
	if err != nil {
		t.Fatalf("build websocket url: %v", err)
	}

	conn, resp, err := websocket.Dial(context.Background(), target, &websocket.DialOptions{HTTPHeader: headers})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}

	client := &Client{
		Conn:     conn,
		Response: resp,
	}
	t.Cleanup(func() {
		client.Close(websocket.StatusNormalClosure, "test_cleanup")
	})

	return client
}

func (c *Client) Send(ctx context.Context, message platformws.Message) error {
	return platformws.WriteJSON(ctx, c.Conn, message)
}

func (c *Client) Receive(ctx context.Context) (platformws.Message, error) {
	var message platformws.Message
	err := platformws.ReadJSON(ctx, c.Conn, &message)
	return message, err
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
