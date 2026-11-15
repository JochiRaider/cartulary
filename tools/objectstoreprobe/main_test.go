package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const corsTestOrigin = "http://127.0.0.1:5173"

func TestCheckCORSPreflightRequiresExactDirectPutPolicy(t *testing.T) {
	t.Run("exact policy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestHeaders := strings.ToLower(r.Header.Get("Access-Control-Request-Headers"))
			if r.Header.Get("Origin") != corsTestOrigin ||
				r.Header.Get("Access-Control-Request-Method") != http.MethodPut ||
				strings.Contains(requestHeaders, "authorization") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", corsTestOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT")
			w.Header().Set("Access-Control-Allow-Headers", "x-amz-checksum-sha256, content-type")
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		if err := checkCORSPreflight(context.Background(), http.DefaultClient, server.URL, corsTestOrigin); err != nil {
			t.Fatalf("exact CORS policy rejected: %v", err)
		}
	})

	cases := []struct {
		name      string
		configure func(http.ResponseWriter)
	}{
		{
			name: "wildcard origin",
			configure: func(w http.ResponseWriter) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT")
				w.Header().Set("Access-Control-Allow-Headers", "x-amz-checksum-sha256, content-type")
			},
		},
		{
			name: "extra method",
			configure: func(w http.ResponseWriter) {
				w.Header().Set("Access-Control-Allow-Origin", corsTestOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, GET")
				w.Header().Set("Access-Control-Allow-Headers", "x-amz-checksum-sha256, content-type")
			},
		},
		{
			name: "extra header",
			configure: func(w http.ResponseWriter) {
				w.Header().Set("Access-Control-Allow-Origin", corsTestOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT")
				w.Header().Set("Access-Control-Allow-Headers", "x-amz-checksum-sha256, content-type, authorization")
			},
		},
		{
			name: "credentials",
			configure: func(w http.ResponseWriter) {
				w.Header().Set("Access-Control-Allow-Origin", corsTestOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT")
				w.Header().Set("Access-Control-Allow-Headers", "x-amz-checksum-sha256, content-type")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				tc.configure(w)
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			if err := checkCORSPreflight(context.Background(), http.DefaultClient, server.URL, corsTestOrigin); err == nil {
				t.Fatalf("%s CORS policy unexpectedly accepted", tc.name)
			}
		})
	}
}

func TestCheckCORSPreflightRejectsNullOriginAcceptance(t *testing.T) {
	// Each fixture is the exact policy except for one granted request, so the
	// probe must reach that stage to reject it.
	for _, tc := range []struct {
		name, grantOrigin, grantMethod, want string
	}{
		{name: "null origin", grantOrigin: "null", grantMethod: http.MethodPut, want: "CORS Origin null unexpectedly allowed"},
		{name: "disallowed method", grantOrigin: corsTestOrigin, grantMethod: http.MethodGet, want: "CORS disallowed preflight unexpectedly granted"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				origin := r.Header.Get("Origin")
				method := r.Header.Get("Access-Control-Request-Method")
				exact := origin == corsTestOrigin && method == http.MethodPut
				granted := origin == tc.grantOrigin && method == tc.grantMethod
				if (!exact && !granted) || strings.Contains(strings.ToLower(r.Header.Get("Access-Control-Request-Headers")), "authorization") {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT")
				w.Header().Set("Access-Control-Allow-Headers", "x-amz-checksum-sha256, content-type")
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			if err := checkCORSPreflight(context.Background(), http.DefaultClient, server.URL, corsTestOrigin); err == nil || err.Error() != tc.want {
				t.Fatalf("CORS preflight error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestUploadPresignedPUTRequiresExactCORSResponse(t *testing.T) {
	t.Run("exact response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Origin"); got != corsTestOrigin {
				t.Fatalf("PUT origin got %q want %q", got, corsTestOrigin)
			}
			w.Header().Set("Access-Control-Allow-Origin", corsTestOrigin)
			w.Header().Set("Access-Control-Expose-Headers", "etag")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		if err := uploadPresignedPUT(context.Background(), http.DefaultClient, server.URL, corsTestOrigin, []byte("payload")); err != nil {
			t.Fatalf("exact PUT CORS response rejected: %v", err)
		}
	})

	t.Run("extra exposed header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", corsTestOrigin)
			w.Header().Set("Access-Control-Expose-Headers", "etag, x-amz-request-id")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		if err := uploadPresignedPUT(context.Background(), http.DefaultClient, server.URL, corsTestOrigin, []byte("payload")); err == nil {
			t.Fatalf("extra exposed header unexpectedly accepted")
		}
	})
}
