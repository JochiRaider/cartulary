package wstest

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
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
	return ConnectWithClient(t, http.DefaultClient, serverURL, path, headers)
}

func ConnectWithClient(t testing.TB, httpClient *http.Client, serverURL string, path string, headers http.Header) *Client {
	t.Helper()

	client, _, err := TryConnectWithClient(httpClient, serverURL, path, headers)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() {
		client.Close(websocket.StatusNormalClosure, "test_cleanup")
	})

	return client
}

func TryConnect(serverURL string, path string, headers http.Header) (*Client, *http.Response, error) {
	return TryConnectWithClient(http.DefaultClient, serverURL, path, headers)
}

func TryConnectWithClient(httpClient *http.Client, serverURL string, path string, headers http.Header) (*Client, *http.Response, error) {
	target, err := websocketURL(serverURL, path)
	if err != nil {
		return nil, nil, err
	}

	conn, response, err := websocket.Dial(context.Background(), target, &websocket.DialOptions{HTTPClient: httpClient, HTTPHeader: headers})
	if err != nil {
		return nil, response, err
	}
	return &Client{Conn: conn, Response: response}, response, nil
}

func (c *Client) WriteJSON(ctx context.Context, value any) error {
	return wsjson.Write(ctx, c.Conn, value)
}

func (c *Client) ReadJSON(ctx context.Context, value any) error {
	return wsjson.Read(ctx, c.Conn, value)
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
