package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	defaultListenAddr = "127.0.0.1:8333"
	defaultUpstream   = "http://127.0.0.1:18333"
	defaultOrigin     = "http://localhost:5173"
)

var (
	allowedMethods = []string{http.MethodPut, http.MethodOptions}
	allowedHeaders = []string{"content-type", "x-amz-checksum-sha256"}
	exposedHeaders = []string{"etag"}
)

func main() {
	var listenAddr string
	var upstreamRaw string
	var origin string
	flag.StringVar(&listenAddr, "listen", envDefault("OBJECT_STORE_CORS_PROXY_LISTEN", defaultListenAddr), "listen address")
	flag.StringVar(&upstreamRaw, "upstream", envDefault("OBJECT_STORE_CORS_PROXY_UPSTREAM", defaultUpstream), "upstream SeaweedFS S3 base URL")
	flag.StringVar(&origin, "origin", envDefault("OBJECT_STORE_CORS_ORIGIN", defaultOrigin), "allowed browser origin")
	flag.Parse()

	upstreamURL, err := url.Parse(upstreamRaw)
	if err != nil {
		log.Fatalf("parse upstream: %v", err)
	}
	if upstreamURL.Scheme != "http" && upstreamURL.Scheme != "https" {
		log.Fatalf("upstream must use http or https")
	}
	if upstreamURL.Host == "" {
		log.Fatalf("upstream host is required")
	}
	origin = strings.TrimSpace(origin)
	if origin == "" {
		log.Fatalf("origin is required")
	}

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           newProxyHandler(upstreamURL, origin),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-signalCh:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("shutdown after %s: %v", sig, err)
		}
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	}
}

func newProxyHandler(upstreamURL *url.URL, origin string) http.Handler {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			originalHost := req.In.Host
			req.SetURL(upstreamURL)
			req.Out.Host = originalHost
		},
		ModifyResponse: func(resp *http.Response) error {
			stripCORSHeaders(resp.Header)
			if resp.Request != nil && resp.Request.Method == http.MethodPut && resp.Request.Header.Get("Origin") == origin {
				setActualPUTCORSHeaders(resp.Header, origin)
			}
			return nil
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			handlePreflight(w, r, origin)
			return
		}
		proxy.ServeHTTP(w, r)
	})
}

func handlePreflight(w http.ResponseWriter, r *http.Request, origin string) {
	if r.Header.Get("Origin") != origin {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if strings.ToUpper(strings.TrimSpace(r.Header.Get("Access-Control-Request-Method"))) != http.MethodPut {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !requestedHeadersAllowed(r.Header.Values("Access-Control-Request-Headers"), allowedHeaders) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	header := w.Header()
	header.Set("Access-Control-Allow-Origin", origin)
	header.Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
	header.Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
	header.Set("Access-Control-Max-Age", "600")
	header.Set("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")
	w.WriteHeader(http.StatusNoContent)
}

func setActualPUTCORSHeaders(header http.Header, origin string) {
	header.Set("Access-Control-Allow-Origin", origin)
	header.Set("Access-Control-Expose-Headers", strings.Join(exposedHeaders, ", "))
	vary := header.Get("Vary")
	if vary == "" {
		header.Set("Vary", "Origin")
		return
	}
	if !headerTokenSetContains(header.Values("Vary"), "origin") {
		header.Set("Vary", fmt.Sprintf("%s, Origin", vary))
	}
}

func stripCORSHeaders(header http.Header) {
	for key := range header {
		if strings.HasPrefix(strings.ToLower(key), "access-control-") {
			header.Del(key)
		}
	}
}

func requestedHeadersAllowed(values []string, allowed []string) bool {
	requested := headerTokens(values)
	if len(requested) == 0 {
		return true
	}
	allowedSet := map[string]bool{}
	for _, header := range allowed {
		allowedSet[strings.ToLower(header)] = true
	}
	for header := range requested {
		if !allowedSet[header] {
			return false
		}
	}
	return true
}

func headerTokenSetContains(values []string, want string) bool {
	tokens := headerTokens(values)
	return tokens[strings.ToLower(want)]
}

func headerTokens(values []string) map[string]bool {
	tokens := map[string]bool{}
	for _, value := range values {
		for _, token := range strings.Split(value, ",") {
			normalized := strings.ToLower(strings.TrimSpace(token))
			if normalized != "" {
				tokens[normalized] = true
			}
		}
	}
	return tokens
}

func envDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
