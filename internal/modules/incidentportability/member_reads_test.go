package incidentportability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"testing"
)

type testMemberRead struct {
	member Member
	bytes  string
	closed bool
}

func (s *testMemberRead) Members() []Member { return []Member{s.member} }
func (s *testMemberRead) OpenMember(context.Context, string) (io.ReadCloser, Member, error) {
	return &closeTrackingRead{Reader: strings.NewReader(s.bytes), close: func() { s.closed = true }}, s.member, nil
}

type closeTrackingRead struct {
	io.Reader
	close func()
}

func (r *closeTrackingRead) Close() error { r.close(); return nil }

func TestPreparedMemberChecksMetadataAndActualBytes_Unit(t *testing.T) {
	raw := "{\"value\":1}\n"
	digest := sha256.Sum256([]byte(raw))
	expected := Member{Path: "data/source.ndjson", ByteSize: int64(len(raw)), SHA256: hex.EncodeToString(digest[:])}
	for _, name := range []string{"valid", "changed descriptor", "substituted bytes", "truncated", "overlong", "cancelled"} {
		t.Run(name, func(t *testing.T) {
			source := &testMemberRead{member: expected, bytes: raw}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			switch name {
			case "changed descriptor":
				source.member.ByteSize++
			case "substituted bytes":
				source.bytes = strings.Replace(raw, "1", "2", 1)
			case "truncated":
				source.bytes = ""
			case "overlong":
				source.bytes += "{}\n"
			case "cancelled":
				cancel()
			}
			calls := 0
			err := WalkMember(ctx, source, expected, func(map[string]any) error { calls++; return nil })
			if !source.closed {
				t.Fatal("member was not closed")
			}
			if name == "valid" {
				if err != nil || calls != 1 {
					t.Fatal(calls, err)
				}
				return
			}
			if name == "cancelled" {
				if !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
				return
			}
			var failure *VerificationFailure
			if !errors.As(err, &failure) || failure.ReasonCode != "checksum_mismatch" {
				t.Fatal(err)
			}
			if name == "changed descriptor" && calls != 0 {
				t.Fatal("changed descriptor reached owner")
			}
		})
	}
}
