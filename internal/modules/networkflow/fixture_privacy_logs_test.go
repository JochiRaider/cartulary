package networkflow_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry/testsupport/logcapture"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
)

// System trust is immutable after first use. As in the platform's qualified
// transport tests, a child isolates the fixture CA without changing host trust.
func runPrivacyFixtureChild(t *testing.T) bool {
	t.Helper()
	const marker = "CARTULARY_NF_PRIVACY_FIXTURE_CHILD"
	if os.Getenv(marker) == "1" {
		return false
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestNetworkFlowPrivacyFixtureComparison$", "-test.v")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "SSL_CERT_DIR=") && !strings.HasPrefix(entry, marker+"=") {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(command.Env, marker+"=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("isolated privacy fixture: %v\n%s", err, output)
	}
	return true
}
func (p *privacyFixtureCapture) receive(body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.exports) >= 128 {
		p.overflow = true
		return fmt.Errorf("collector export overflow")
	}
	raw, records, err := logcapture.DecodeExport(body, 1024-len(p.records))
	if err != nil {
		return err
	}
	p.exports = append(p.exports, raw)
	p.records = append(p.records, records...)
	return nil
}
func (p *privacyFixtureCapture) startCollector(t *testing.T) map[string]string {
	t.Helper()
	authority, err := tlstest.NewAuthority()
	if err != nil {
		t.Fatal(err)
	}
	identity, err := authority.Issue("collector", []string{"127.0.0.1"}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	write := func(name string, data []byte) string {
		path, err := tlstest.WriteFile(directory, name, data)
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	certificate, key := write("server.pem", identity.CertificatePEM), write("server.key", identity.PrivateKeyPEM)
	tlsConfig, err := cryptography.TLSServer("127.0.0.1", certificate, key)
	if err != nil {
		t.Fatal(err)
	}
	// Keep the managed service CA and the collector CA in one explicitly
	// selected process bundle. All clients then use the same immutable roots.
	existingRoots, err := os.ReadFile(os.Getenv("SSL_CERT_FILE"))
	if err != nil {
		t.Fatal(err)
	}
	roots := append(append(existingRoots, '\n'), authority.CertificatePEM...)
	t.Setenv("SSL_CERT_FILE", write("root.pem", roots))
	t.Setenv("SSL_CERT_DIR", "")
	collector := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fail := func(err error) {
			p.mu.Lock()
			p.collectorError = err
			p.mu.Unlock()
			http.Error(w, "fixture collector rejected export", http.StatusBadRequest)
		}
		if r.Method != http.MethodPost || r.URL.Path != "/v1/logs" || r.Header.Get("Content-Type") != "application/x-protobuf" || r.TLS == nil || r.TLS.Version != tls.VersionTLS13 {
			fail(fmt.Errorf("unexpected collector transport"))
			return
		}
		raw, err := io.ReadAll(io.LimitReader(r.Body, (1<<20)+1))
		if err != nil || len(raw) > 1<<20 {
			fail(fmt.Errorf("collector body limit/read failure"))
			return
		}
		if err := p.receive(raw); err != nil {
			fail(err)
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	collector.TLS = tlsConfig
	collector.StartTLS()
	t.Cleanup(collector.Close)
	return map[string]string{
		"CARTULARY__TELEMETRY__ENABLED": "true", "CARTULARY__TELEMETRY__EXPORTER__KIND": "otlp_http", "CARTULARY__TELEMETRY__EXPORTER__ENDPOINT": collector.URL, "CARTULARY__TELEMETRY__EXPORTER__COMPRESSION": "none", "CARTULARY__TELEMETRY__EXPORTER__RETRY__ENABLED": "false",
		"CARTULARY__TELEMETRY__TRACES__ENABLED": "false", "CARTULARY__TELEMETRY__METRICS__ENABLED": "false", "CARTULARY__TELEMETRY__LOGS__BRIDGE_ENABLED": "true",
		"CARTULARY__TELEMETRY__RESOURCE__SERVICE_VERSION": "0.0.0+fixture", "CARTULARY__TELEMETRY__RESOURCE__SERVICE_INSTANCE_ID": "00000000-0000-4000-8000-000000000088",
	}
}
