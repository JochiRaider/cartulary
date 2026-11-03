package serverprocess

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

var processEmbeddedAssetPathPattern = regexp.MustCompile(`(?:src|href)="(/assets/[^"]+)"`)

func TestServerEmbeddedFrontendAssets_ProcessSmoke(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "authentication-e-1-09")
	client := *server.Client
	client.Timeout = 2 * time.Second

	rootResp, err := client.Get(server.BaseURL + "/")
	if err != nil {
		t.Fatalf("request root: %v", err)
	}
	defer rootResp.Body.Close()

	if rootResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected root status: got %d want %d", rootResp.StatusCode, http.StatusOK)
	}
	if contentType := rootResp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("unexpected root content type: %q", contentType)
	}

	rootBody, err := io.ReadAll(rootResp.Body)
	if err != nil {
		t.Fatalf("read root body: %v", err)
	}
	if !strings.Contains(string(rootBody), `<div id="root"></div>`) {
		t.Fatalf("expected embedded app shell in root response, got %q", string(rootBody))
	}

	if !strings.Contains(string(rootBody), `id="cartulary-client-extension-support-registry"`) || !strings.Contains(string(rootBody), `"client_build_id":"cartulary.web.standard.sha256:`) {
		t.Fatal("final build support is absent from bootstrap")
	}
	for _, private := range []string{"/index.html", "/_contracts/extension-registry-integrity.json", "/assets/generated/client-extension-support-registry.json", "/assets/_contracts/client-bootstrap-template.json"} {
		response, err := client.Get(server.BaseURL + private)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusNotFound {
			t.Fatalf("private package entry is served at %s: %d", private, response.StatusCode)
		}
	}
	assetMatch := processEmbeddedAssetPathPattern.FindStringSubmatch(string(rootBody))
	if len(assetMatch) < 2 {
		t.Fatalf("expected embedded asset reference in root response, got %q", string(rootBody))
	}

	assetResp, err := client.Get(server.BaseURL + assetMatch[1])
	if err != nil {
		t.Fatalf("request embedded asset %s: %v", assetMatch[1], err)
	}
	defer assetResp.Body.Close()

	if assetResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected asset status for %s: got %d want %d", assetMatch[1], assetResp.StatusCode, http.StatusOK)
	}

	assetBody, err := io.ReadAll(assetResp.Body)
	if err != nil {
		t.Fatalf("read embedded asset body: %v", err)
	}
	if len(assetBody) == 0 {
		t.Fatalf("expected non-empty embedded asset body for %s", assetMatch[1])
	}
}
