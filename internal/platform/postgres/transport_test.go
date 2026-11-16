package postgres

import (
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sys/unix"

	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
)

func TestPostgresCertificateDSNPolicy(t *testing.T) {
	dsn := certificatePolicyFixture(t)
	if _, err := parseCertificateDSN(dsn); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*url.URL)
	}{
		{"password", func(u *url.URL) { u.User = url.UserPassword("login", "secret") }},
		{"empty password", func(u *url.URL) { u.User = url.UserPassword("login", "") }},
		{"missing login", func(u *url.URL) { u.User = nil }},
		{"missing database", func(u *url.URL) { u.Path = "" }},
		{"fallback hosts", func(u *url.URL) { u.Host = "localhost,other:5432" }},
		{"unix socket", func(u *url.URL) { u.Host = "" }},
		{"fragment", func(u *url.URL) { u.Fragment = "forbidden" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			u, _ := url.Parse(dsn)
			tc.mutate(u)
			if _, err := parseCertificateDSN(u.String()); err == nil {
				t.Fatal("unsupported DSN admitted")
			}
		})
	}
	for _, pair := range [][2]string{
		{"sslmode", "disable"}, {"sslmode", "require"}, {"require_auth", "scram-sha-256"},
		{"sslrootcert", "system"}, {"sslcert", "../client.pem"}, {"sslkey", "/tmp/../client.key"},
		{"passfile", "/unread/secret"}, {"service", "ignored"}, {"servicefile", "/unread/secret"},
		{"password", "secret"}, {"host", "elsewhere"}, {"options", "-c role=postgres"},
		{"pool_max_conns", "1000"}, {"sslpassword", "secret"},
	} {
		t.Run(pair[0]+"="+pair[1], func(t *testing.T) {
			u, _ := url.Parse(dsn)
			q := u.Query()
			q.Set(pair[0], pair[1])
			u.RawQuery = q.Encode()
			if _, err := parseCertificateDSN(u.String()); err == nil {
				t.Fatal("unsupported DSN parameter admitted")
			}
		})
	}
	for _, raw := range []string{"host=localhost user=login database=cartulary", dsn + "&sslmode=verify-full", dsn + "&sslmode=disable", dsn + "\n", paddedCertificateDSN(t, dsn, 65537)} {
		if _, err := parseCertificateDSN(raw); err == nil {
			t.Fatal("malformed or duplicate DSN admitted")
		}
	}
	// The 65,537-byte DSN above is otherwise valid; at 65,536 it is admitted.
	if _, err := parseCertificateDSN(paddedCertificateDSN(t, dsn, 65536)); err != nil {
		t.Fatalf("DSN at the size bound rejected: %v", err)
	}
	t.Run("home-directory defaults never read", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		if err := unix.Mkfifo(filepath.Join(home, ".pgpass"), 0o600); err != nil {
			t.Fatal(err)
		}
		driver, tlsConfig, err := connectionInputs(dsn)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := pgx.ParseConfig(driver)
		if err != nil {
			t.Fatal(err)
		}
		pooled, err := pgxpool.ParseConfig(driver)
		if err != nil {
			t.Fatal(err)
		}
		for _, config := range []*pgx.ConnConfig{parsed, pooled.ConnConfig} {
			if config.Password != "" || config.RequireAuth != "none" || config.TLSConfig == nil || config.TLSConfig.InsecureSkipVerify || len(config.Fallbacks) != 0 {
				t.Fatal("unsafe driver construction")
			}
		}
		if tlsConfig.MinVersion != tls.VersionTLS13 || tlsConfig.MaxVersion != tls.VersionTLS13 || tlsConfig.VerifyConnection == nil || tlsConfig.ServerName != "localhost" || len(tlsConfig.Certificates) != 1 || tlsConfig.RootCAs == nil {
			t.Fatal("missing qualified TLS configuration")
		}
	})
	for _, name := range []string{"PGPASSWORD", "PGPASSFILE", "PGSERVICE", "PGSERVICEFILE", "PGSSLMODE", "PGUNKNOWN"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv(name, "/must-not-read")
			if _, _, err := connectionInputs(dsn); err == nil {
				t.Fatal("inherited PostgreSQL configuration admitted")
			}
			if !inheritedPostgresSettings(map[string]string{name: ""}) {
				t.Fatal("empty inherited PostgreSQL setting admitted")
			}
		})
	}
}

func certificatePolicyFixture(t *testing.T) string {
	t.Helper()
	ca, err := tlstest.NewAuthority()
	if err != nil {
		t.Fatal(err)
	}
	identity, err := ca.Issue("login", nil, x509.ExtKeyUsageClientAuth)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	write := func(name string, data []byte) string {
		t.Helper()
		p, err := tlstest.WriteFile(dir, name, data)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	u := url.URL{Scheme: "postgres", User: url.User("login"), Host: "localhost:5432", Path: "/cartulary"}
	u.RawQuery = url.Values{"sslmode": {"verify-full"}, "require_auth": {"none"}, "sslrootcert": {write("root.pem", ca.CertificatePEM)}, "sslcert": {write("client.pem", identity.CertificatePEM)}, "sslkey": {write("client.key", identity.PrivateKeyPEM)}}.Encode()
	// Test-specific inherited settings are restored by testing.T. The actual
	// harness must not pass production connection defaults into this process.
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(name, "PG") {
			t.Fatalf("unexpected inherited PostgreSQL setting %s", name)
		}
	}
	return u.String()
}

// paddedCertificateDSN lengthens dsn's sslrootcert path so the DSN is
// exactly size bytes and otherwise valid.
func paddedCertificateDSN(t testing.TB, dsn string, size int) string {
	t.Helper()
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	withPad := func(n int) string {
		query := u.Query()
		query.Set("sslrootcert", "/pad/"+strings.Repeat("a", n)+"/root.pem")
		u.RawQuery = query.Encode()
		return u.String()
	}
	padded := withPad(size - len(withPad(0)))
	if len(padded) != size {
		t.Fatalf("padded DSN is %d bytes, want %d", len(padded), size)
	}
	return padded
}
