// windowsprobe is a Make-built qualification client. Trust and credentials arrive
// over stdin and are never installed in Windows or written to evidence.
package main

import (
	"context"
	"crypto/fips140"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/coder/websocket"
)

type input struct {
	Origin   string `json:"origin"`
	CA       string `json:"ca"`
	Cookie   string `json:"cookie"`
	Incident string `json:"incident"`
}

func main() {
	if err := run(); err != nil {
		// Transport errors may contain request details. Retain only this bounded
		// classification, never credentials or a raw command/HTTP response.
		fmt.Fprintln(os.Stderr, "Windows boundary probe failed: "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	if runtime.GOOS != "windows" || !fips140.Enabled() {
		return errors.New("execution identity")
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("build identity")
	}
	selectors := 0
	selector := ""
	for _, s := range info.Settings {
		if s.Key == "GOFIPS140" {
			selectors++
			selector = s.Value
		}
	}
	if selectors != 1 {
		return errors.New("module identity")
	}
	var in input
	d := json.NewDecoder(io.LimitReader(os.Stdin, 128*1024))
	d.DisallowUnknownFields()
	if d.Decode(&in) != nil {
		return errors.New("input")
	}
	u, err := url.Parse(in.Origin)
	if err != nil || u.Scheme != "https" || u.Hostname() != "127.0.0.1" || u.Port() == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("origin")
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(in.CA)) {
		return errors.New("isolated trust")
	}
	cfg := &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: roots, ServerName: u.Hostname(), CurvePreferences: []tls.CurveID{tls.CurveP256, tls.CurveP384}}
	client := httpClient(cfg)
	defer client.CloseIdleConnections()
	r, err := client.Get(in.Origin + "/readyz")
	if err != nil {
		return errors.New("verified HTTPS")
	}
	defer r.Body.Close()
	if r.StatusCode != 200 || r.TLS == nil || r.TLS.Version != tls.VersionTLS13 || len(r.TLS.VerifiedChains) == 0 {
		return errors.New("HTTPS readiness")
	}
	leaf := sha256.Sum256(r.TLS.PeerCertificates[0].Raw)
	report := map[string]any{"status": "pass", "client_os": runtime.GOOS, "client_arch": runtime.GOARCH, "toolchain": runtime.Version(), "module_version": fips140.Version(), "module_enabled": fips140.Enabled(), "module_selector": selector, "https": true, "unix_ms": time.Now().UnixMilli(), "leaf_sha256": hex.EncodeToString(leaf[:]), "trust_scope": "stdin_private_pool", "global_trust_modified": false}
	for _, negative := range []string{"untrusted_authority", "wrong_name", "expired_at_client", "tls12"} {
		bad := cfg.Clone()
		switch negative {
		case "untrusted_authority":
			bad.RootCAs = x509.NewCertPool()
		case "wrong_name":
			bad.ServerName = "untrusted.invalid"
		case "expired_at_client":
			bad.Time = func() time.Time { return time.Now().Add(72 * time.Hour) }
		case "tls12":
			bad.MinVersion = tls.VersionTLS12
			bad.MaxVersion = tls.VersionTLS12
		}
		c := httpClient(bad)
		response, e := c.Get(in.Origin + "/readyz")
		if response != nil {
			response.Body.Close()
		}
		c.CloseIdleConnections()
		if e == nil {
			return errors.New("negative TLS: " + negative)
		}
		report[negative+"_rejected"] = true
	}
	if in.Incident != "" {
		if len(in.Incident) > 128 || strings.ContainsAny(in.Incident, "/?#\\\r\n") || in.Cookie == "" {
			return errors.New("session input")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		target := "wss://" + u.Host + "/ws/v1/incidents/" + in.Incident
		header := http.Header{"Origin": {in.Origin}, "Cookie": {in.Cookie}}
		conn, _, e := websocket.Dial(ctx, target, &websocket.DialOptions{HTTPClient: client, HTTPHeader: header})
		if e != nil {
			return errors.New("verified authenticated WSS")
		}
		defer conn.CloseNow()
		conn.SetReadLimit(1 << 20)
		if e = conn.Write(ctx, websocket.MessageText, []byte(`{"type":"hello","payload":{"client_instance_id":"windows-package-probe","presence":{"sheet_ref":{"kind":"view_schema","id":"cartulary.view.timeline.v2"},"mode":"viewing"}}}`)); e != nil {
			return errors.New("WSS hello write")
		}
		_, body, e := conn.Read(ctx)
		var message struct {
			Type string `json:"type"`
		}
		if e != nil || json.Unmarshal(body, &message) != nil || message.Type != "hello_ack" {
			return errors.New("WSS hello acknowledgement")
		}
		conn.Close(websocket.StatusNormalClosure, "")
		header.Set("Origin", "https://untrusted.invalid")
		bad, response, e := websocket.Dial(ctx, target, &websocket.DialOptions{HTTPClient: client, HTTPHeader: header})
		if bad != nil {
			bad.CloseNow()
		}
		if e == nil || response == nil || response.StatusCode != 403 {
			return errors.New("WSS origin rejection")
		}
		report["authenticated_wss"] = true
		report["untrusted_wss_origin_rejected"] = true
	}
	return json.NewEncoder(os.Stdout).Encode(report)
}

func httpClient(cfg *tls.Config) *http.Client {
	return &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{TLSClientConfig: cfg}, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirect forbidden") }}
}
