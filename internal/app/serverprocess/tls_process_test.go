package serverprocess

import (
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestApplicationTLSAdmission_Process(t *testing.T) {
	for _, binding := range []string{"TLS_CERTIFICATE_PATH", "TLS_PRIVATE_KEY_PATH"} {
		t.Run(binding, func(t *testing.T) {
			env := newProcessEnv(t, processEnvOptions{ConfigPath: writeConfig(t, string(fixtures.MustRead("config", "valid.toml"))), Overrides: map[string]string{
				"CARTULARY__APPLICATION__" + binding: "/private/missing-identity.pem",
			}})
			server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
			t.Cleanup(func() { server.Stop(t) })
			if err := server.WaitForExit(t); err == nil {
				t.Fatal("missing TLS identity started")
			}
			server.RequireDiagnosticsField(t, "application", "application_tls_invalid")
			server.RequireConnectionRefused(t, "/healthz")
			server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/unknown")
			for key, path := range env {
				if !strings.HasPrefix(key, "CARTULARY__ROOTS__") || !strings.HasSuffix(key, "__PATH") {
					continue
				}
				entries, err := os.ReadDir(path)
				if err != nil || len(entries) != 0 {
					t.Fatalf("TLS rejection changed root %s: entries=%d err=%v", key, len(entries), err)
				}
			}
		})
	}
}

func TestApplicationHTTPSBoundary_Process(t *testing.T) {
	server := startServerProcess(t, "application_tls_boundary")
	response, err := server.Client.Get(server.BaseURL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK || response.TLS == nil || response.TLS.Version != tls.VersionTLS13 {
		t.Fatal("ready application did not use TLS 1.3")
	}
	for _, scenario := range []string{"untrusted", "tls12", "plaintext"} {
		t.Run(scenario, func(t *testing.T) {
			transport := &http.Transport{}
			if scenario == "tls12" {
				transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12, RootCAs: server.Client.Transport.(*http.Transport).TLSClientConfig.RootCAs}
			}
			defer transport.CloseIdleConnections()
			client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
			target := server.BaseURL + "/healthz"
			if scenario == "plaintext" {
				target = strings.Replace(target, "https://", "http://", 1)
			}
			response, err := client.Get(target)
			if response != nil {
				_ = response.Body.Close()
			}
			if scenario == "plaintext" {
				if err == nil && response.StatusCode != http.StatusBadRequest {
					t.Fatalf("plaintext reached application: %d", response.StatusCode)
				}
			} else if err == nil {
				t.Fatal("unqualified TLS reached application")
			}
		})
	}
}
