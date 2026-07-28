package collaboration

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCollaborationCodecSemanticContract(t *testing.T) {
	codec := Codec{}
	incidentID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	message := EphemeralMessage(incidentID, "resume_ack", map[string]any{
		"status":       ResumeStatusReplayed,
		"resume_token": "token",
	}, time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC))

	encoded, err := codec.Encode(message)
	if err != nil {
		t.Fatalf("encode message: %v", err)
	}
	if bytes.HasSuffix(encoded, []byte("\n")) {
		t.Fatalf("encoded message has a trailing LF: %q", encoded)
	}
	decoded, err := codec.Decode(MessageText, appendUnknownMembers(encoded))
	if err != nil {
		t.Fatalf("decode message with additive members: %v", err)
	}
	if decoded.Type != "resume_ack" {
		t.Fatalf("decoded type = %q", decoded.Type)
	}

	for name, testCase := range map[string]struct {
		kind     MessageKind
		payload  []byte
		wantKind DecodeFailureKind
		wantSize bool
	}{
		"binary": {
			kind:     MessageBinary,
			payload:  []byte(`{"type":"pong","payload":{}}`),
			wantKind: DecodeFailureBinaryMessage,
		},
		"malformed": {
			kind:     MessageText,
			payload:  []byte(`{"type":`),
			wantKind: DecodeFailureInvalidJSON,
		},
		"invalid UTF-8": {
			kind:     MessageText,
			payload:  []byte{0xff},
			wantKind: DecodeFailureInvalidJSON,
		},
		"duplicate envelope member": {
			kind:     MessageText,
			payload:  []byte(`{"type":"pong","type":"pong","payload":{}}`),
			wantKind: DecodeFailureDuplicateMember,
		},
		"duplicate nested member": {
			kind:     MessageText,
			payload:  []byte(`{"type":"pong","payload":{"value":1,"value":2}}`),
			wantKind: DecodeFailureDuplicateMember,
		},
		"oversized": {
			kind:     MessageText,
			payload:  bytes.Repeat([]byte("x"), MaximumMessageBytes+1),
			wantSize: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := codec.Decode(testCase.kind, testCase.payload)
			if testCase.wantSize {
				if !errors.Is(err, ErrMessageTooLarge) {
					t.Fatalf("error = %v want ErrMessageTooLarge", err)
				}
				return
			}
			var failure *DecodeFailure
			if !errors.As(err, &failure) || failure.Kind != testCase.wantKind {
				t.Fatalf("failure = %#v / %v want kind %d", failure, err, testCase.wantKind)
			}
		})
	}

	if got := safeWebSocketEventType("extension_resource_changed"); got != "extension_resource_changed" {
		t.Fatalf("extension event classification = %q", got)
	}
	if got := safeWebSocketEventType("resume_result"); got != "other" {
		t.Fatalf("non-public resume_result classification = %q", got)
	}
}

func appendUnknownMembers(encoded []byte) []byte {
	if len(encoded) == 0 || encoded[len(encoded)-1] != '}' {
		return encoded
	}
	output := append([]byte(nil), encoded[:len(encoded)-1]...)
	output = append(output, []byte(`,"future_envelope_member":{"nested":true}}`)...)
	return output
}
