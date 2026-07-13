package postgresstore

import . "github.com/JochiRaider/cartulary/internal/modules/graphprojection"

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const graphCursorVersion = "graphprojection.cursor.v1"

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

func randomCursorKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
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
		return listCursor{}, ErrCursorInvalid
	}
	if base64.RawURLEncoding.EncodeToString(sealed) != token {
		return listCursor{}, ErrCursorInvalid
	}
	nonce := sealed[:c.aead.NonceSize()]
	payload, err := c.aead.Open(nil, nonce, sealed[c.aead.NonceSize():], []byte(graphCursorVersion))
	if err != nil {
		return listCursor{}, ErrCursorInvalid
	}
	var cursor listCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return listCursor{}, ErrCursorInvalid
	}
	return cursor, nil
}
