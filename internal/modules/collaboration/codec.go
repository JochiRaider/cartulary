package collaboration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"unicode/utf8"
)

const MaximumMessageBytes = 32_768

type MessageKind uint8

const (
	MessageText MessageKind = iota + 1
	MessageBinary
)

type Socket interface {
	Read(context.Context) (MessageKind, []byte, error)
	Write(context.Context, MessageKind, []byte) error
	Close(code uint16, reason string) error
}

type AcceptSocket func(http.ResponseWriter, *http.Request) (Socket, error)
type CheckBrowserOrigin func(http.ResponseWriter, *http.Request) bool

var ErrMessageTooLarge = errors.New("collaboration websocket message too large")

type DecodeFailureKind uint8

const (
	DecodeFailureInvalidJSON DecodeFailureKind = iota + 1
	DecodeFailureDuplicateMember
	DecodeFailureBinaryMessage
)

type DecodeFailure struct {
	Kind DecodeFailureKind
	err  error
}

func (e *DecodeFailure) Error() string {
	if e == nil || e.err == nil {
		return "invalid collaboration websocket message"
	}
	return e.err.Error()
}

func (e *DecodeFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

type Codec struct{}

func (Codec) Decode(kind MessageKind, data []byte) (Message, error) {
	if kind == MessageBinary {
		return Message{}, &DecodeFailure{
			Kind: DecodeFailureBinaryMessage,
			err:  errors.New("binary application messages are unsupported"),
		}
	}
	if kind != MessageText || !utf8.Valid(data) {
		return Message{}, &DecodeFailure{
			Kind: DecodeFailureInvalidJSON,
			err:  errors.New("application message is not valid UTF-8 JSON"),
		}
	}
	if len(data) > MaximumMessageBytes {
		return Message{}, ErrMessageTooLarge
	}
	if err := validateJSONMembers(data); err != nil {
		var duplicate *duplicateMemberError
		if errors.As(err, &duplicate) {
			return Message{}, &DecodeFailure{Kind: DecodeFailureDuplicateMember, err: err}
		}
		return Message{}, &DecodeFailure{Kind: DecodeFailureInvalidJSON, err: err}
	}
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return Message{}, &DecodeFailure{Kind: DecodeFailureInvalidJSON, err: err}
	}
	return message, nil
}

func (Codec) Encode(message Message) ([]byte, error) {
	if !IsServerMessageType(message.Type) {
		return nil, fmt.Errorf("unsupported Collaboration server message type %q", message.Type)
	}
	encoded, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("encode Collaboration server message: %w", err)
	}
	if len(encoded) > MaximumMessageBytes {
		return nil, ErrMessageTooLarge
	}
	return encoded, nil
}

func IsClientMessageType(messageType string) bool {
	switch messageType {
	case "hello", "resume", "pong", "presence_update":
		return true
	default:
		return false
	}
}

func IsServerMessageType(messageType string) bool {
	switch messageType {
	case "hello_ack",
		"resume_ack",
		"presence_snapshot",
		"presence_delta",
		"record_changed",
		"extension_resource_changed",
		"job_progress",
		"ping",
		"error",
		"session_revoked":
		return true
	default:
		return false
	}
}

type duplicateMemberError struct {
	Name string
}

func (e *duplicateMemberError) Error() string {
	return fmt.Sprintf("duplicate JSON object member %q", e.Name)
}

func validateJSONMembers(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := validateJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("application message contains multiple JSON values")
		}
		return err
	}
	return nil
}

func validateJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			nameToken, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := nameToken.(string)
			if !ok {
				return errors.New("JSON object member name is invalid")
			}
			if _, duplicate := seen[name]; duplicate {
				return &duplicateMemberError{Name: name}
			}
			seen[name] = struct{}{}
			if err := validateJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return errors.New("JSON object is not terminated")
		}
	case '[':
		for decoder.More() {
			if err := validateJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return errors.New("JSON array is not terminated")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}
