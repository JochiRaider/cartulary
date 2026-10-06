package auth_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestTOTPEnvelopeSubstitutionAndActivation_Integration(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	server, db := startServer(t, runtime, "auth-envelope-substitution")
	defer db.Close()
	first := seedLocalUser(t, db, "envelope-first@example.test", "First", "EnvelopePassword1!", true)
	second := seedLocalUser(t, db, "envelope-second@example.test", "Second", "EnvelopePassword1!", true)
	firstToken := requireBootstrapLogin(t, server, "envelope-first@example.test", "EnvelopePassword1!")
	secondToken := requireBootstrapLogin(t, server, "envelope-second@example.test", "EnvelopePassword1!")
	one := beginTOTPEnrollment(t, server, firstToken, map[string]any{"client_txn_id": "first-enrollment"})
	two := beginTOTPEnrollment(t, server, secondToken, map[string]any{"client_txn_id": "second-enrollment"})
	var firstEnvelope, secondEnvelope []byte
	if err := db.QueryRowContext(context.Background(), `SELECT secret_envelope FROM pending_totp_enrollments WHERE id::text=$1`, one["enrollment_id"]).Scan(&firstEnvelope); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(context.Background(), `SELECT secret_envelope FROM pending_totp_enrollments WHERE id::text=$1`, two["enrollment_id"]).Scan(&secondEnvelope); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(context.Background(), `UPDATE pending_totp_enrollments SET secret_envelope=$2 WHERE id::text=$1`, one["enrollment_id"], secondEnvelope); err != nil {
		t.Fatal(err)
	}
	rejected := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{"client_txn_id": "first-enrollment"}, withHeader("Authorization", "Bearer "+firstToken))
	httptestx.RequireErrorEnvelope(t, rejected, http.StatusInternalServerError, "internal_error")
	if got := queryCount(t, db, `SELECT COUNT(*) FROM users WHERE id::text=$1 AND totp_enrolled_at IS NOT NULL`, first); got != 0 {
		t.Fatal("substitution activated a factor")
	}
	if _, err := db.ExecContext(context.Background(), `UPDATE pending_totp_enrollments SET secret_envelope=$2 WHERE id::text=$1`, one["enrollment_id"], firstEnvelope); err != nil {
		t.Fatal(err)
	}
	firstSecret := one["totp_setup"].(map[string]any)["secret_base32"].(string)
	secondSecret := two["totp_setup"].(map[string]any)["secret_base32"].(string)
	for _, item := range []struct{ token, id, secret string }{{firstToken, one["enrollment_id"].(string), firstSecret}, {secondToken, two["enrollment_id"].(string), secondSecret}} {
		completed := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{"client_txn_id": "complete-" + item.id, "enrollment_id": item.id, "code": generateTOTPCode(t, item.secret)}, withHeader("Authorization", "Bearer "+item.token))
		httptestx.RequireSuccessEnvelope(t, completed, http.StatusOK)
	}
	var active []byte
	if err := db.QueryRowContext(context.Background(), `SELECT totp_secret_envelope FROM users WHERE id::text=$1`, first).Scan(&active); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(active, firstEnvelope) {
		t.Fatal("activation copied pending ciphertext")
	}
	_ = loginLocalUserWithSecondFactor(t, server, "envelope-first@example.test", "EnvelopePassword1!", generateTOTPCode(t, firstSecret))
	before := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text=$1`, second)
	for _, envelope := range [][]byte{firstEnvelope, active} {
		if _, err := db.ExecContext(context.Background(), `UPDATE users SET totp_secret_envelope=$2 WHERE id::text=$1`, second, envelope); err != nil {
			t.Fatal(err)
		}
		login := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{"username": "envelope-second@example.test", "password": "EnvelopePassword1!", "second_factor": map[string]any{"kind": "totp", "assertion": map[string]any{"code": generateTOTPCode(t, firstSecret)}}})
		httptestx.RequireErrorEnvelope(t, login, http.StatusInternalServerError, "internal_error")
	}
	if after := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text=$1`, second); after != before {
		t.Fatal("substituted secret issued a session")
	}
}

func TestEnterprisePKCEEnvelopeSubstitution_Integration(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	server, db := startEnterpriseServer(t, runtime, "enterprise-envelope-substitution")
	defer db.Close()
	user := seedLocalUser(t, db, "enterprise-envelope@example.test", "Enterprise", "EnvelopePassword1!", false)
	provider := seedEnterpriseProvider(t, db, enterpriseProviderSeed{Key: "sealed-oidc", Type: "oidc", Name: "Sealed OIDC", Enabled: true, Interactive: true})
	seedEnterpriseBinding(t, db, user, provider, "BoundSubject", user)
	first := beginEnterpriseAuthWithBody(t, server, "sealed-oidc", map[string]any{"return_to": "/"})
	second := beginEnterpriseAuthWithBody(t, server, "sealed-oidc", map[string]any{"return_to": "/"})
	one, two := parseEnterpriseRedirect(t, first.redirectURL), parseEnterpriseRedirect(t, second.redirectURL)
	var original, other []byte
	if err := db.QueryRowContext(context.Background(), `SELECT pkce_verifier_envelope FROM enterprise_auth_transactions WHERE state=$1`, one.Query().Get("state")).Scan(&original); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(context.Background(), `SELECT pkce_verifier_envelope FROM enterprise_auth_transactions WHERE state=$1`, two.Query().Get("state")).Scan(&other); err != nil {
		t.Fatal(err)
	}
	tampered := append([]byte(nil), original...)
	tampered[len(tampered)-1] ^= 1
	for _, bad := range [][]byte{other, tampered} {
		if _, err := db.ExecContext(context.Background(), `UPDATE enterprise_auth_transactions SET pkce_verifier_envelope=$2 WHERE state=$1`, one.Query().Get("state"), bad); err != nil {
			t.Fatal(err)
		}
		callback := doNoRedirect(t, http.MethodGet, oidcCallbackURL(server, "sealed-oidc", one.Query().Get("state"), one.Query().Get("nonce"), "valid-code", "BoundSubject"), nil, withCookies(first.cookie))
		httptestx.RequireErrorEnvelope(t, callback, http.StatusInternalServerError, "internal_error")
	}
	if count := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text=$1`, user); count != 0 {
		t.Fatal("substituted PKCE issued session")
	}
	if count := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_transactions WHERE state=$1 AND consumed_at IS NOT NULL`, one.Query().Get("state")); count != 0 {
		t.Fatal("failed callback consumed transaction")
	}
	if _, err := db.ExecContext(context.Background(), `UPDATE enterprise_auth_transactions SET pkce_verifier_envelope=NULL WHERE state=$1`, one.Query().Get("state")); err == nil {
		t.Fatal("missing PKCE admitted")
	}
	if _, err := db.ExecContext(context.Background(), `UPDATE enterprise_auth_transactions SET pkce_verifier_envelope=$2 WHERE state=$1`, one.Query().Get("state"), original); err != nil {
		t.Fatal(err)
	}
	callback := doNoRedirect(t, http.MethodGet, oidcCallbackURL(server, "sealed-oidc", one.Query().Get("state"), one.Query().Get("nonce"), "valid-code", "BoundSubject"), nil, withCookies(first.cookie))
	defer callback.Body.Close()
	if callback.StatusCode != http.StatusSeeOther {
		t.Fatalf("restored own envelope: status %d", callback.StatusCode)
	}
}
