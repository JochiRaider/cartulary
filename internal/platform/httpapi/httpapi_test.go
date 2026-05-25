package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi/webassets"
)

var embeddedAssetPathPattern = regexp.MustCompile(`(?:src|href)="(/assets/[^"]+)"`)

func TestNewHandler_ServesEmbeddedRootAndAssets(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	expectedRoot, err := webassets.ReadIndexHTML()
	if err != nil {
		t.Fatalf("ReadIndexHTML(): %v", err)
	}

	rootRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	rootRecorder := httptest.NewRecorder()
	handler.ServeHTTP(rootRecorder, rootRequest)

	rootResponse := rootRecorder.Result()
	defer rootResponse.Body.Close()

	if rootResponse.StatusCode != http.StatusOK {
		t.Fatalf("unexpected root status: got %d want %d", rootResponse.StatusCode, http.StatusOK)
	}
	if contentType := rootResponse.Header.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("unexpected root content type: %q", contentType)
	}

	rootBody, err := io.ReadAll(rootResponse.Body)
	if err != nil {
		t.Fatalf("read root body: %v", err)
	}
	if string(rootBody) != string(expectedRoot) {
		t.Fatalf("unexpected root body: got %q want %q", string(rootBody), string(expectedRoot))
	}

	assetMatch := embeddedAssetPathPattern.FindSubmatch(rootBody)
	if len(assetMatch) < 2 {
		return
	}

	assetRequest := httptest.NewRequest(http.MethodGet, string(assetMatch[1]), nil)
	assetRecorder := httptest.NewRecorder()
	handler.ServeHTTP(assetRecorder, assetRequest)

	assetResponse := assetRecorder.Result()
	defer assetResponse.Body.Close()

	if assetResponse.StatusCode != http.StatusOK {
		t.Fatalf("unexpected asset status for %s: got %d want %d", string(assetMatch[1]), assetResponse.StatusCode, http.StatusOK)
	}
	assetBody, err := io.ReadAll(assetResponse.Body)
	if err != nil {
		t.Fatalf("read asset body: %v", err)
	}
	if len(assetBody) == 0 {
		t.Fatalf("expected non-empty asset body for %s", string(assetMatch[1]))
	}
}

func TestNewHandler_KeepsReservedExtensionRouting(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/providers", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected reserved-route status: got %d want %d", response.StatusCode, http.StatusNotFound)
	}

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode reserved-route payload: %v", err)
	}

	errorPayload, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("missing error payload: %#v", payload)
	}
	if errorPayload["code"] != "extension_profile_not_claimed" {
		t.Fatalf("unexpected reserved-route code: %#v", errorPayload)
	}
}

func TestNewHandler_KeepsUnclaimedReservedExtensionRootsUnavailable(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	for _, profile := range CurrentExtensionProfiles() {
		if profile.Claimed {
			continue
		}
		for _, routeFamily := range profile.RouteFamilies {
			path := strings.ReplaceAll(routeFamily, "{user_id}", "11111111-1111-1111-1111-111111111111")
			request := httptest.NewRequest(http.MethodGet, path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			response := recorder.Result()
			body, err := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if err != nil {
				t.Fatalf("read reserved-route response for %s: %v", path, err)
			}
			if response.StatusCode != http.StatusNotFound {
				t.Fatalf("reserved root %s returned status %d body %q", path, response.StatusCode, string(body))
			}

			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("decode reserved-route payload for %s: %v body=%q", path, err, string(body))
			}
			errorPayload, ok := payload["error"].(map[string]any)
			if !ok {
				t.Fatalf("missing error payload for %s: %#v", path, payload)
			}
			if errorPayload["code"] != "extension_profile_not_claimed" {
				t.Fatalf("reserved root %s returned unexpected error: %#v", path, errorPayload)
			}
			details, ok := errorPayload["details"].(map[string]any)
			if !ok {
				t.Fatalf("missing reserved-route details for %s: %#v", path, errorPayload)
			}
			if details["profile_id"] != profile.ProfileID || details["route_family"] != routeFamily {
				t.Fatalf("unexpected reserved-route details for %s: %#v", path, details)
			}
		}
	}
}
