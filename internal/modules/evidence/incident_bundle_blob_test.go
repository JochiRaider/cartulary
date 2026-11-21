package evidence_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
)

const portableBlobTestPath = "data/object_blobs.ndjson"

var portableBlobTestBytes = []byte("portable blob\n")

func portableBlobTestRow(incident uuid.UUID, index int, state string) map[string]any {
	row := map[string]any{
		"object_blob_id": fmt.Sprintf("20000000-0000-4000-8000-%012d", index), "incident_id": incident.String(), "created_by_user_id": evidencePortableActor.String(),
		"upload_state": state, "byte_size": json.Number("14"), "filename_hint": "blob.txt", "content_type_hint": "text/plain", "expected_sha256_hex": nil,
		"observed_size": nil, "observed_content_type": nil, "observed_sha256_hex": nil, "target_expires_at": "2026-01-01T01:00:00+00:00", "pending_expires_at": "2026-01-02T00:00:00+00:00",
		"finalized_at": nil, "terminal_reason": nil, "failed_at": nil, "cleanup_due_at": nil, "cleaned_up_at": nil, "created_at": "2026-01-01T00:00:00+00:00", "updated_at": "2026-01-01T00:00:00+00:00",
	}
	if state == "available" || state == "quarantined" {
		sum := sha256.Sum256(portableBlobTestBytes)
		digest := hex.EncodeToString(sum[:])
		row["byte_size"], row["observed_size"] = json.Number(fmt.Sprint(len(portableBlobTestBytes))), json.Number(fmt.Sprint(len(portableBlobTestBytes)))
		row["observed_sha256_hex"], row["expected_sha256_hex"], row["observed_content_type"] = digest, digest, "text/plain"
		row["finalized_at"], row["updated_at"] = "2026-01-01T00:30:00+00:00", "2026-01-01T00:30:00+00:00"
	}
	if state == "failed" {
		row["terminal_reason"], row["failed_at"], row["cleanup_due_at"], row["updated_at"] = "pending_timeout", "2026-01-02T00:00:00+00:00", "2026-01-02T00:45:00+00:00", "2026-01-02T01:00:00+00:00"
	}
	return row
}
func portableBlobTestBundle(t testing.TB, rows ...map[string]any) portabilitysupport.MapBundle {
	t.Helper()
	bundle := evidencePortableBundle(t)
	var out bytes.Buffer
	for _, row := range rows {
		if err := json.NewEncoder(&out).Encode(row); err != nil {
			t.Fatal(err)
		}
		if digest, ok := row["observed_sha256_hex"].(string); ok {
			bundle["blobs/sha256/"+digest] = bytes.Clone(portableBlobTestBytes)
		}
	}
	bundle[portableBlobTestPath] = out.Bytes()
	return bundle
}
func TestEvidencePortableBlobAdmission_Unit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	incident, local := uuid.New(), uuid.New()
	input := evidencePortableInput(t, incident, local)
	port := evidence.NewIncidentBundleSourcePort()
	prepare := func(rows ...map[string]any) error {
		b := portableBlobTestBundle(t, rows...)
		before := maps.Clone(b)
		for k, v := range before {
			before[k] = bytes.Clone(v)
		}
		_, err := port.PrepareImport(ctx, b, input)
		if !reflect.DeepEqual(b, before) {
			t.Fatal("Prepare mutated source bytes")
		}
		return err
	}
	if err := prepare(); err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"pending", "failed", "available", "quarantined"} {
		t.Run(state, func(t *testing.T) {
			if err := prepare(portableBlobTestRow(incident, 1, state)); err != nil {
				t.Fatal(err)
			}
		})
	}
	base := portableBlobTestRow(incident, 1, "pending")
	nonnull := strings.Fields("object_blob_id incident_id created_by_user_id upload_state byte_size target_expires_at pending_expires_at created_at updated_at")
	required := map[string]bool{}
	for _, f := range nonnull {
		required[f] = true
	}
	for field := range base {
		t.Run(field, func(t *testing.T) {
			invariant := "row_shape_exact"
			if field == "object_blob_id" {
				invariant = "source_identity_admitted"
			}
			row := maps.Clone(base)
			delete(row, field)
			requireEvidencePortableFailure(t, prepare(row), invariant)
			row = maps.Clone(base)
			row[field] = true
			requireEvidencePortableFailure(t, prepare(row), invariant)
			row = maps.Clone(base)
			row[field] = nil
			if required[field] {
				requireEvidencePortableFailure(t, prepare(row), invariant)
			} else if err := prepare(row); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, tc := range []struct {
		name, field, invariant string
		value                  any
	}{
		{"extra", "extra", "row_shape_exact", true}, {"physical null", "storage_key", "storage_reference_legal", nil},
		{"foreign incident", "incident_id", "envelope_type_scope", uuid.NewString()}, {"missing actor", "created_by_user_id", "attribution_exact", uuid.NewString()},
		{"unknown state", "upload_state", "lifecycle_legal", "unknown"}, {"trimmed hint", "filename_hint", "row_shape_exact", " padded"}, {"empty hint", "content_type_hint", "row_shape_exact", ""},
		{"NUL", "filename_hint", "row_shape_exact", "bad\x00name"}, {"fractional size", "byte_size", "row_shape_exact", json.Number("1.0")}, {"negative size", "byte_size", "row_shape_exact", json.Number("-1")},
		{"minus zero", "byte_size", "row_shape_exact", json.Number("-0")}, {"exponent size", "observed_size", "row_shape_exact", json.Number("1e1")}, {"size overflow", "byte_size", "row_shape_exact", json.Number("9223372036854775808")},
		{"uppercase digest", "expected_sha256_hex", "row_shape_exact", strings.Repeat("A", 64)},
		{"pending observation", "observed_size", "lifecycle_legal", json.Number("0")}, {"pending finalization", "finalized_at", "lifecycle_legal", "2026-01-01T00:00:00Z"},
		{"pending reason", "terminal_reason", "lifecycle_legal", "pending_timeout"}, {"target expiry", "target_expires_at", "lifecycle_legal", "2026-01-01T02:00:00Z"},
		{"pending expiry", "pending_expires_at", "lifecycle_legal", "2026-01-01T01:00:00Z"}, {"updated before creation", "updated_at", "lifecycle_legal", "2025-12-31T23:59:59Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := maps.Clone(base)
			row[tc.field] = tc.value
			requireEvidencePortableFailure(t, prepare(row), tc.invariant)
		})
	}
	for _, field := range []string{"object_blob_id", "incident_id", "created_by_user_id"} {
		for _, value := range []string{uuid.Nil.String(), "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", "10000000000040008000000000000001", "bad"} {
			row := maps.Clone(base)
			row[field] = value
			invariant := "row_shape_exact"
			if field == "object_blob_id" {
				invariant = "source_identity_admitted"
			}
			requireEvidencePortableFailure(t, prepare(row), invariant)
		}
	}
	for _, field := range strings.Fields("created_at updated_at target_expires_at pending_expires_at finalized_at failed_at cleanup_due_at cleaned_up_at") {
		for _, value := range []string{"0000-01-01T00:00:00Z", "2026-01-01T00:00:00.1234567Z", "2026-01-01T00:00:00.0000000Z", "2026-01-01T01:00:00+01:00", "2026-01-01"} {
			row := maps.Clone(base)
			row[field] = value
			requireEvidencePortableFailure(t, prepare(row), "row_shape_exact")
		}
	}
	for _, tc := range []struct {
		state, field string
		value        any
	}{
		{"available", "finalized_at", nil}, {"available", "finalized_at", "2026-01-02T00:00:00Z"}, {"available", "failed_at", "2026-01-01T00:00:00Z"},
		{"quarantined", "cleaned_up_at", "2026-01-01T00:00:00Z"}, {"failed", "terminal_reason", nil}, {"failed", "terminal_reason", "unknown"},
		{"failed", "failed_at", nil}, {"failed", "cleanup_due_at", nil}, {"failed", "cleanup_due_at", "2026-01-02T00:44:59Z"},
		{"failed", "failed_at", "2026-01-01T00:00:00Z"}, {"failed", "cleaned_up_at", "2026-01-02T00:44:59Z"}, {"failed", "cleaned_up_at", "2026-01-02T01:00:01Z"},
		{"failed", "observed_content_type", "text/plain"}, {"available", "updated_at", "2026-01-01T00:29:59Z"},
		{"available", "terminal_reason", "pending_timeout"},
	} {
		row := portableBlobTestRow(incident, 1, tc.state)
		row[tc.field] = tc.value
		requireEvidencePortableFailure(t, prepare(row), "lifecycle_legal")
	}
	// The archive never supplies a physical key, in any state.
	for _, state := range []string{"pending", "failed", "available", "quarantined"} {
		for _, key := range []any{nil, "", "incidents/source/object-blobs/source/objects/source", map[string]any{"key": "untrusted"}} {
			row := portableBlobTestRow(incident, 1, state)
			row["storage_key"] = key
			requireEvidencePortableFailure(t, prepare(row), "storage_reference_legal")
		}
	}
	// A finalized row's sizes and digests must describe its member's bytes
	// before any shape or lifecycle phase.
	finalized := portableBlobTestRow(incident, 1, "available")
	member := "blobs/sha256/" + finalized["observed_sha256_hex"].(string)
	observation := func(mutate func(map[string]any)) error {
		row := maps.Clone(finalized)
		mutate(row)
		b := portableBlobTestBundle(t, row)
		for path := range b {
			if strings.HasPrefix(path, "blobs/sha256/") {
				delete(b, path)
			}
		}
		b[member] = bytes.Clone(portableBlobTestBytes)
		_, err := port.PrepareImport(ctx, b, input)
		return err
	}
	observations := map[string]func(map[string]any){
		"both sizes disagree with bytes": func(row map[string]any) {
			row["byte_size"], row["observed_size"] = json.Number("123"), json.Number("123")
		},
		"missing observed digest with expected digest": func(row map[string]any) { delete(row, "observed_sha256_hex") },
		"null observed digest with expected digest":    func(row map[string]any) { row["observed_sha256_hex"] = nil },
		"empty observed digest":                        func(row map[string]any) { row["observed_sha256_hex"] = "" },
		"noncanonical observed digest":                 func(row map[string]any) { row["observed_sha256_hex"] = strings.Repeat("A", 64) },
		"nonhex observed digest":                       func(row map[string]any) { row["observed_sha256_hex"] = strings.Repeat("g", 64) },
		"mistyped observed digest":                     func(row map[string]any) { row["observed_sha256_hex"] = true },
		"expected digest disagrees":                    func(row map[string]any) { row["expected_sha256_hex"] = strings.Repeat("0", 64) },
		"missing expected digest member":               func(row map[string]any) { delete(row, "expected_sha256_hex") },
		"mistyped expected digest":                     func(row map[string]any) { row["expected_sha256_hex"] = false },
	}
	for _, field := range []string{"byte_size", "observed_size"} {
		observations["missing "+field] = func(row map[string]any) { delete(row, field) }
		for _, value := range []any{nil, "14", json.Number("-1"), json.Number("-0"), json.Number("14.0"), json.Number("14e0"), json.Number("9223372036854775808"), json.Number("123")} {
			observations[fmt.Sprintf("%s=%v", field, value)] = func(row map[string]any) { row[field] = value }
		}
	}
	for name, mutate := range observations {
		t.Run("observation/"+name, func(t *testing.T) {
			requireEvidencePortableFailure(t, observation(mutate), "byte_size_digest_agree")
		})
	}
	if err := observation(func(map[string]any) {}); err != nil {
		t.Fatalf("agreeing observation: %v", err)
	}
	for _, reason := range []string{"pending_timeout", "declared_size_mismatch", "expected_sha256_mismatch"} {
		row := portableBlobTestRow(incident, 1, "failed")
		row["terminal_reason"] = reason
		row["cleaned_up_at"] = "2026-01-02T00:45:00Z"
		if err := prepare(row); err != nil {
			t.Fatal(err)
		}
	}
	max := maps.Clone(base)
	max["byte_size"] = json.Number("9223372036854775807")
	if err := prepare(max); err != nil {
		t.Fatal(err)
	}
	second := portableBlobTestRow(incident, 2, "pending")
	if err := prepare(base, second); err != nil {
		t.Fatal(err)
	}
	requireEvidencePortableFailure(t, prepare(base, base), "source_identity_admitted")
	requireEvidencePortableFailure(t, prepare(second, base), "source_identity_admitted")
	for _, payload := range []string{"{} {}\n", "[]\n", "\n", "{\"object_blob_id\":1,\"object_blob_id\":2}\n", "{\"x\":\"\\ud800\"}\n", string([]byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'})} {
		b := portableBlobTestBundle(t)
		b[portableBlobTestPath] = []byte(payload)
		_, err := port.PrepareImport(ctx, b, input)
		requireEvidencePortableFailure(t, err, "row_shape_exact")
	}
	bad := maps.Clone(base)
	bad["object_blob_id"] = uuid.Nil.String()
	b := portableBlobTestBundle(t, bad)
	b[portableBlobTestPath] = append(b[portableBlobTestPath], []byte("{\n")...)
	_, err := port.PrepareImport(ctx, b, input)
	requireEvidencePortableFailure(t, err, "row_shape_exact")
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = port.PrepareImport(cancelled, portableBlobTestBundle(t, base), input)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel=%v", err)
	}
}
