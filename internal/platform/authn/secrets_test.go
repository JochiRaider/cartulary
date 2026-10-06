package authn

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
)

func TestMasterKeyAdmissionAndHKDFPurposes(t *testing.T) {
	for _, size := range []int{0, 16, 24, 31, 33, 64} {
		env := map[string]string{AuthMasterKeyEnv: base64.RawStdEncoding.EncodeToString(make([]byte, size))}
		if _, err := LoadMasterKeys(env); err == nil {
			t.Fatalf("admitted %d bytes", size)
		}
		if err := RegisterMasterSecretPurpose(secretpurpose.NewRegistry(), env); err == nil {
			t.Fatalf("registered %d bytes", size)
		}
	}
	if _, err := LoadMasterKeys(map[string]string{}); err == nil {
		t.Fatal("missing key admitted")
	}
	material := make([]byte, 32)
	for i := range material {
		material[i] = byte(i)
	}
	raw := base64.RawStdEncoding.EncodeToString(material)
	for _, invalid := range []string{raw + "\n", "!" + raw, raw + "=" + "="} {
		if _, err := LoadMasterKeys(map[string]string{AuthMasterKeyEnv: invalid}); err == nil {
			t.Fatal("malformed key admitted")
		}
	}
	keys, err := LoadMasterKeys(map[string]string{AuthMasterKeyEnv: raw})
	if err != nil {
		t.Fatal(err)
	}
	// Independently derived with Python hashlib/hmac using RFC 5869.
	if got := hex.EncodeToString(keys.tokenFingerprintKey[:]); got != "5900e799374e7d6e9bd98236e4e167990b13a43ff7797752e331ff097797f9ae" {
		t.Fatalf("HKDF vector: %s", got)
	}
	other, err := DerivePurposeKey(keys, "token-fingerprint")
	if err != nil {
		t.Fatal(err)
	}
	if other == keys.tokenFingerprintKey || keys.csrfKey == keys.requestFingerprintKey {
		t.Fatal("purpose collision")
	}
	if _, err := DerivePurposeKey(MasterKeys{}, "consumer"); err == nil {
		t.Fatal("unadmitted key derived")
	}
}

func TestAuthSecretEnvelopeRecordAndPurposeBinding(t *testing.T) {
	keys, err := LoadMasterKeys(map[string]string{AuthMasterKeyEnv: base64.RawStdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))})
	if err != nil {
		t.Fatal(err)
	}
	user, record, provider := uuid.New(), uuid.New(), uuid.New()
	pending := SecretBinding{Purpose: PendingTOTPSecret, RecordID: record, SubjectID: user}
	plaintext := bytes.Repeat([]byte{3}, 32)
	envelope, err := SealSecret(keys, pending, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if opened, err := OpenSecret(keys, pending, envelope); err != nil || !bytes.Equal(opened, plaintext) {
		t.Fatal("pending open failed", err)
	}
	for _, binding := range []SecretBinding{
		{Purpose: ActiveTOTPSecret, RecordID: user, SubjectID: user},
		{Purpose: PendingTOTPSecret, RecordID: uuid.New(), SubjectID: user},
		{Purpose: PendingTOTPSecret, RecordID: record, SubjectID: uuid.New()},
		{Purpose: EnterprisePKCESecret, RecordID: record, SubjectID: provider},
		{Purpose: "unknown", RecordID: record, SubjectID: user}, {},
	} {
		if _, err := OpenSecret(keys, binding, envelope); !errors.Is(err, ErrSecret) {
			t.Fatal("substitution accepted")
		}
	}
	for _, bad := range [][]byte{nil, envelope[1:], append([]byte{0}, envelope[1:]...), append(append([]byte(nil), envelope...), 0)} {
		if _, err := OpenSecret(keys, pending, bad); !errors.Is(err, ErrSecret) {
			t.Fatal("old/malformed envelope accepted")
		}
	}
	for i := range envelope {
		bad := append([]byte(nil), envelope...)
		bad[i] ^= 1
		if _, err := OpenSecret(keys, pending, bad); err == nil {
			t.Fatalf("tamper %d", i)
		}
	}
	active := SecretBinding{Purpose: ActiveTOTPSecret, RecordID: user, SubjectID: user}
	resealed, err := SealSecret(keys, active, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(resealed, envelope) {
		t.Fatal("reseal reused envelope")
	}
	if _, err := OpenSecret(keys, pending, resealed); err == nil {
		t.Fatal("active secret used as pending")
	}
	enterprise := SecretBinding{Purpose: EnterprisePKCESecret, RecordID: record, SubjectID: provider}
	pkce, err := SealSecret(keys, enterprise, []byte(base64.RawURLEncoding.EncodeToString(plaintext)))
	if err != nil {
		t.Fatal(err)
	}
	enterprise.SubjectID = uuid.New()
	if _, err := OpenSecret(keys, enterprise, pkce); err == nil {
		t.Fatal("provider substitution accepted")
	}
	if _, err := SealSecret(keys, pending, make([]byte, 33)); err == nil {
		t.Fatal("invalid TOTP payload admitted")
	}
}
