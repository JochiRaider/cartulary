package cryptography

import (
	"bytes"
	"crypto/fips140"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractcryptography"
)

func testKey(t *testing.T) MasterKey {
	t.Helper()
	key, err := AdmitKey(bytes.Repeat([]byte{0x42}, MasterKeyBytes))
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func projection(t *testing.T, name string, dst any) {
	t.Helper()
	for _, artifact := range contractcryptography.Artifacts {
		if artifact.Path == "contracts/cryptography/"+name {
			if err := json.Unmarshal([]byte(artifact.JSON), dst); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatalf("missing projection %s", name)
}

func TestPrimitiveContractAndIndependentVector(t *testing.T) {
	if os.Getenv("GOFIPS140") == ModuleSelector && (!fips140.Enabled() || fips140.Version() != ModuleVersion) {
		t.Fatal("pinned primitive execution did not use the enabled selected module")
	}
	var limits map[string]any
	projection(t, "primitives.v1.json", &limits)
	for name, want := range map[string]int{"master_key_bytes": MasterKeyBytes, "salt_bytes": SaltBytes, "envelope_version": int(EnvelopeVersion), "envelope_overhead_bytes": EnvelopeOverhead, "plaintext_max_bytes": MaxPlaintextBytes, "context_max_bytes": MaxContextBytes, "artifact_max_seals": MaxArtifactSeals} {
		if limits[name] != float64(want) {
			t.Fatalf("contract %s = %v, runtime %d", name, limits[name], want)
		}
	}
	var vector struct {
		Master             string   `json:"master_hex"`
		Purpose            string   `json:"purpose"`
		Binding            []string `json:"binding"`
		AAD                string   `json:"aad_hex"`
		Plaintext          string   `json:"plaintext_hex"`
		Envelope           string   `json:"envelope_hex"`
		PurposeKey         string   `json:"purpose_key_hex"`
		MessageKey         string   `json:"message_key_hex"`
		ArtifactSalt       string   `json:"artifact_salt_hex"`
		ArtifactCiphertext string   `json:"artifact_ciphertext_hex"`
	}
	projection(t, "vectors.v1.json", &vector)
	decode := func(raw string) []byte {
		value, err := hex.DecodeString(raw)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	key, err := AdmitKey(decode(vector.Master))
	if err != nil {
		t.Fatal(err)
	}
	derived, err := key.Derive(vector.Purpose, vector.Binding...)
	if err != nil || !bytes.Equal(derived[:], decode(vector.PurposeKey)) {
		t.Fatalf("independent derivation: %x %v", derived, err)
	}
	envelope := decode(vector.Envelope)
	info, err := encodeContext("cartulary.message.v1", vector.Purpose, vector.Binding, len(decode(vector.AAD)))
	if err != nil {
		t.Fatal(err)
	}
	derived, err = key.derive(envelope[1:1+SaltBytes], info)
	if err != nil || !bytes.Equal(derived[:], decode(vector.MessageKey)) {
		t.Fatalf("independent message derivation: %x %v", derived, err)
	}
	plaintext, err := key.Open(vector.Purpose, vector.Binding, envelope, decode(vector.AAD))
	if err != nil || !bytes.Equal(plaintext, decode(vector.Plaintext)) {
		t.Fatalf("independent opening: %x %v", plaintext, err)
	}
	var salt [SaltBytes]byte
	copy(salt[:], decode(vector.ArtifactSalt))
	reader, err := key.ArtifactOpener(vector.Purpose, vector.Binding, salt)
	if err != nil {
		t.Fatal(err)
	}
	plaintext, err = reader.Open(decode(vector.ArtifactCiphertext), decode(vector.AAD))
	if err != nil || !bytes.Equal(plaintext, decode(vector.Plaintext)) {
		t.Fatalf("independent artifact opening: %x %v", plaintext, err)
	}
}

func TestKeyAdmissionAndDomainSeparation(t *testing.T) {
	for _, size := range []int{0, 1, 16, 31, 33, 64} {
		if _, err := AdmitKey(make([]byte, size)); !errors.Is(err, ErrKey) {
			t.Fatalf("key size %d: %v", size, err)
		}
	}
	if _, err := (MasterKey{}).Derive("purpose"); !errors.Is(err, ErrKey) {
		t.Fatal("unadmitted zero key")
	}
	raw := bytes.Repeat([]byte{0x42}, 32)
	key, err := AdmitKey(raw)
	if err != nil {
		t.Fatal(err)
	}
	clear(raw)
	want, _ := testKey(t).Derive("purpose")
	got, _ := key.Derive("purpose")
	if got != want {
		t.Fatal("admitted key aliases input")
	}
	seen := map[[32]byte]bool{}
	for _, components := range [][]string{{"purpose"}, {"purpose", ""}, {"purpose", "ab", "c"}, {"purpose", "a", "bc"}, {"purpose", "a", "b", "c"}, {"purpose\x00", "ab", "c"}, {"other", "ab", "c"}} {
		derived, err := key.Derive(components[0], components[1:]...)
		if err != nil {
			t.Fatal(err)
		}
		if seen[derived] {
			t.Fatal("context collision")
		}
		seen[derived] = true
	}
}

func TestSealingBoundsAndAdversarialInputs(t *testing.T) {
	key := testKey(t)
	binding := []string{"record", "owner"}
	aad := []byte("metadata")
	plaintext := []byte("secret")
	envelope, err := key.Seal("purpose", binding, plaintext, aad)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := key.Open("purpose", binding, envelope, aad)
	if err != nil || !bytes.Equal(opened, plaintext) {
		t.Fatalf("roundtrip: %v", err)
	}
	second, err := key.Seal("purpose", binding, plaintext, aad)
	if err != nil || bytes.Equal(envelope[:1+SaltBytes], second[:1+SaltBytes]) {
		t.Fatal("message salt reused")
	}
	for index := range envelope {
		changed := bytes.Clone(envelope)
		changed[index] ^= 1
		if value, err := key.Open("purpose", binding, changed, aad); !errors.Is(err, ErrOpen) || value != nil {
			t.Fatalf("tamper byte %d: %x %v", index, value, err)
		}
	}
	for _, invalid := range [][]byte{nil, {}, envelope[:EnvelopeOverhead-1], make([]byte, MaxPlaintextBytes+EnvelopeOverhead+1)} {
		if _, err := key.Open("purpose", binding, invalid, aad); !errors.Is(err, ErrOpen) {
			t.Fatalf("framing: %v", err)
		}
	}
	for _, change := range []struct {
		purpose string
		binding []string
		aad     []byte
	}{{"other", binding, aad}, {"purpose", []string{"owner", "record"}, aad}, {"purpose", binding, []byte("changed")}} {
		if _, err := key.Open(change.purpose, change.binding, envelope, change.aad); !errors.Is(err, ErrOpen) {
			t.Fatalf("substitution: %v", err)
		}
	}
	wrong, _ := AdmitKey(make([]byte, 32))
	if _, err := wrong.Open("purpose", binding, envelope, aad); !errors.Is(err, ErrOpen) {
		t.Fatal("wrong key accepted")
	}
	if _, err := key.Seal("purpose", binding, make([]byte, MaxPlaintextBytes+1), aad); !errors.Is(err, ErrInput) {
		t.Fatal("oversized plaintext")
	}
	for _, invalid := range []struct {
		purpose string
		binding []string
		aad     []byte
	}{{"", nil, nil}, {strings.Repeat("x", MaxContextBytes+1), nil, nil}, {"p", []string{strings.Repeat("x", MaxContextBytes)}, nil}, {"p", nil, make([]byte, MaxContextBytes)}, {"p", make([]string, MaxContextBytes/4+1), nil}} {
		if _, err := key.Seal(invalid.purpose, invalid.binding, nil, invalid.aad); !errors.Is(err, ErrInput) {
			t.Fatalf("oversized context: %v", err)
		}
	}
	maxPlaintext := make([]byte, MaxPlaintextBytes)
	maxAAD := make([]byte, MaxContextBytes-1)
	maxEnvelope, err := key.Seal("p", nil, maxPlaintext, maxAAD)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := key.Open("p", nil, maxEnvelope, maxAAD); err != nil || !bytes.Equal(got, maxPlaintext) {
		t.Fatalf("inclusive bounds: %v", err)
	}
}

func TestArtifactCapabilityUseBoundAndIsolation(t *testing.T) {
	key := testKey(t)
	writer, err := key.NewArtifactSealer("recovery.artifact", "artifact-1")
	if err != nil {
		t.Fatal(err)
	}
	reader, err := key.ArtifactOpener("recovery.artifact", []string{"artifact-1"}, writer.Salt())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reader.(ArtifactSealer); ok {
		t.Fatal("opening capability can resume writing")
	}
	sealed, err := writer.Seal([]byte("chunk"), []byte("index=0"))
	if err != nil {
		t.Fatal(err)
	}
	if plain, err := reader.Open(sealed, []byte("index=0")); err != nil || string(plain) != "chunk" {
		t.Fatalf("artifact roundtrip: %v", err)
	}
	if _, err := reader.Open(sealed, []byte("index=1")); !errors.Is(err, ErrOpen) {
		t.Fatal("chunk AAD substitution")
	}
	for _, invalid := range [][]byte{nil, sealed[:len(sealed)-1], append(bytes.Clone(sealed), 0), make([]byte, MaxPlaintextBytes+29)} {
		if plain, err := reader.Open(invalid, []byte("index=0")); !errors.Is(err, ErrOpen) || plain != nil {
			t.Fatal("invalid artifact ciphertext accepted")
		}
	}
	if _, err := writer.Seal(make([]byte, MaxPlaintextBytes+1), nil); !errors.Is(err, ErrInput) {
		t.Fatal("artifact plaintext bound")
	}
	if _, err := writer.Seal(nil, make([]byte, MaxContextBytes)); !errors.Is(err, ErrInput) {
		t.Fatal("artifact context bound")
	}
	other, err := key.NewArtifactSealer("recovery.artifact", "artifact-1")
	if err != nil || writer.Salt() == other.Salt() {
		t.Fatal("artifact salt reused")
	}
	for _, tc := range []struct {
		purpose string
		binding []string
		salt    [SaltBytes]byte
	}{{"different", []string{"artifact-1"}, writer.Salt()}, {"recovery.artifact", []string{"artifact-2"}, writer.Salt()}, {"recovery.artifact", []string{"artifact-1"}, other.Salt()}} {
		wrong, err := key.ArtifactOpener(tc.purpose, tc.binding, tc.salt)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := wrong.Open(sealed, []byte("index=0")); !errors.Is(err, ErrOpen) {
			t.Fatal("artifact substitution")
		}
	}
	private := writer.(*artifactSealer)
	private.used = MaxArtifactSeals - 16
	var passed atomic.Int32
	var wg sync.WaitGroup
	for range 64 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := writer.Seal(nil, nil)
			if err == nil {
				passed.Add(1)
			} else if !errors.Is(err, ErrArtifactLimit) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if passed.Load() != 16 {
		t.Fatalf("seal budget %d", passed.Load())
	}
	if _, err := writer.Seal(nil, nil); !errors.Is(err, ErrArtifactLimit) {
		t.Fatal("exhausted capability reused")
	}
	if _, err := reader.Open(sealed, []byte("index=0")); err != nil {
		t.Fatal("exhaustion invalidated opening")
	}
}
