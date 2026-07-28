package ws

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/coder/websocket"
)

var ErrMessageTooLarge = errors.New("websocket message exceeds configured read limit")

type MessageKind uint8

const (
	MessageText MessageKind = iota + 1
	MessageBinary
)

// Connection is a transport-only WebSocket. It owns frame I/O, bounded reads,
// serialized writes, and close mechanics without knowing application messages.
type Connection struct {
	conn      *websocket.Conn
	writeMu   sync.Mutex
	readLimit int64
}

// Accept upgrades a request using the configured public application origin as
// the only allowed cross-origin browser source.
func Accept(w http.ResponseWriter, r *http.Request, publicOrigin string) (*Connection, error) {
	options := &websocket.AcceptOptions{}
	if publicOrigin != "" {
		parsed, err := url.Parse(publicOrigin)
		if err != nil {
			http.Error(w, "invalid websocket origin configuration", http.StatusInternalServerError)
			return nil, err
		}
		options.OriginPatterns = []string{parsed.Scheme + "://" + parsed.Host}
	}
	conn, err := websocket.Accept(w, r, options)
	if err != nil {
		return nil, err
	}
	// The wrapper enforces its configured bound while reading at most one byte
	// beyond it. Disabling the library limit lets the application choose a
	// stable close reason without permitting an unbounded allocation.
	conn.SetReadLimit(-1)
	return &Connection{conn: conn}, nil
}

func RejectUntrustedBrowserOrigin(w http.ResponseWriter, r *http.Request, publicOrigin string) bool {
	if _, err := r.Cookie("cartulary_session"); err != nil {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	if origin != publicOrigin {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return true
	}
	return false
}

func (c *Connection) SetReadLimit(limit int64) {
	if c == nil {
		return
	}
	c.readLimit = limit
}

func (c *Connection) Read(ctx context.Context) (MessageKind, []byte, error) {
	messageType, reader, err := c.conn.Reader(ctx)
	if err != nil {
		return 0, nil, err
	}
	readerLimit := c.readLimit
	if readerLimit < 1 {
		readerLimit = 32_768
	}
	payload, err := io.ReadAll(io.LimitReader(reader, readerLimit+1))
	if err != nil {
		return 0, nil, err
	}
	if int64(len(payload)) > readerLimit {
		return 0, nil, ErrMessageTooLarge
	}
	switch messageType {
	case websocket.MessageText:
		return MessageText, payload, nil
	case websocket.MessageBinary:
		return MessageBinary, payload, nil
	default:
		return 0, nil, errors.New("unsupported websocket message kind")
	}
}

func (c *Connection) Write(ctx context.Context, kind MessageKind, payload []byte) error {
	var messageType websocket.MessageType
	switch kind {
	case MessageText:
		messageType = websocket.MessageText
	case MessageBinary:
		messageType = websocket.MessageBinary
	default:
		return errors.New("unsupported websocket message kind")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.Write(ctx, messageType, payload)
}

func (c *Connection) Close(code uint16, reason string) error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close(websocket.StatusCode(code), reason)
}

func (c *Connection) CloseNow() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.CloseNow()
}
