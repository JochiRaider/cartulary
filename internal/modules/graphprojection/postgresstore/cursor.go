package postgresstore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

const graphCursorVersion = "graphprojection.cursor.v1"

type listCursor struct {
	Operation             string    `json:"operation"`
	AfterGraphViewID      string    `json:"after_graph_view_id"`
	IssuedAt              time.Time `json:"issued_at"`
	QueryShapeDigest      string    `json:"query_shape_digest"`
	VisibilityScopeDigest string    `json:"visibility_scope_digest"`
}

func cursorInvalid(reason string) error {
	return graphprojection.NewQueryError("cursor_invalid", reason, map[string]any{"reason_code": reason}, graphprojection.ErrCursorInvalid)
}

type graphCursorCodec struct {
	aead cipher.AEAD
}

func newGraphCursorCodec(key []byte) (*graphCursorCodec, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create graph cursor cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create graph cursor AEAD: %w", err)
	}
	return &graphCursorCodec{aead: aead}, nil
}

func (c *graphCursorCodec) encode(cursor listCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nonce, nonce, payload, []byte(graphCursorVersion))
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (c *graphCursorCodec) decode(token string) (listCursor, error) {
	sealed, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(sealed) < c.aead.NonceSize() {
		return listCursor{}, graphprojection.ErrCursorInvalid
	}
	if base64.RawURLEncoding.EncodeToString(sealed) != token {
		return listCursor{}, graphprojection.ErrCursorInvalid
	}
	nonce := sealed[:c.aead.NonceSize()]
	payload, err := c.aead.Open(nil, nonce, sealed[c.aead.NonceSize():], []byte(graphCursorVersion))
	if err != nil {
		return listCursor{}, graphprojection.ErrCursorInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var cursor listCursor
	if err := decoder.Decode(&cursor); err != nil {
		return listCursor{}, graphprojection.ErrCursorInvalid
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return listCursor{}, graphprojection.ErrCursorInvalid
	}
	return cursor, nil
}
