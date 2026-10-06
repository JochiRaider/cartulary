// credentialassessment drives the real application's public credential routes.
// It is a qualification client, never part of the production server.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type outcome struct {
	Status     int            `json:"status"`
	Millis     float64        `json:"duration_ms"`
	Cancelled  bool           `json:"cancelled,omitempty"`
	Code       string         `json:"code,omitempty"`
	Details    map[string]any `json:"-"`
	RetryAfter string         `json:"-"`
	Retryable  bool           `json:"-"`
	Data       map[string]any `json:"-"`
	err        error
}

type client struct {
	origin, password, secret string
	http                     *http.Client
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--resource-report" {
		if err := resourceReport(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "unsupported assessment arguments")
		os.Exit(1)
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (result error) {
	report := map[string]any{"schema_id": "cartulary.credential_capacity_observations.v1", "status": "fail"}
	defer func() {
		if result == nil {
			report["status"] = "pass"
		}
		if err := json.NewEncoder(os.Stdout).Encode(report); result == nil {
			result = err
		}
	}()
	if err := cryptography.AdmitExecution(); err != nil {
		return err
	}
	origin := os.Getenv("CARTULARY_ASSESSMENT_ORIGIN")
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return errors.New("assessment requires an HTTPS origin")
	}
	tlsConfig, err := cryptography.TLSClient(cryptography.TLSClientOptions{ServerName: parsed.Hostname(), RootCertificatePath: os.Getenv("CARTULARY_ASSESSMENT_ROOT_CERTIFICATE_PATH")})
	if err != nil {
		return err
	}
	c := client{origin: origin, password: os.Getenv("CARTULARY_ASSESSMENT_PASSWORD"), http: &http.Client{Timeout: 8 * time.Second, Transport: &http.Transport{TLSClientConfig: tlsConfig, MaxIdleConnsPerHost: 64, MaxConnsPerHost: 64}, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("assessment redirects are forbidden") }}}
	defer c.http.CloseIdleConnections()
	if c.origin == "" || c.password == "" {
		return errors.New("assessment origin and credential are required")
	}
	login := c.request(context.Background(), "/api/v1/auth/login", c.loginBody())
	if login.err != nil || login.Status != 401 || login.Code != "mfa_setup_required" {
		return errors.New("assessment bootstrap login failed")
	}
	token, _ := login.Details["bootstrap_token"].(string)
	begin := c.authorized("/api/v1/auth/mfa/totp/begin", map[string]any{"client_txn_id": "capacity-enrollment"}, token)
	if begin.err != nil || begin.Status != 200 {
		return errors.New("assessment TOTP begin failed")
	}
	setup, _ := begin.Data["totp_setup"].(map[string]any)
	c.secret, _ = setup["secret_base32"].(string)
	if setup["algorithm"] != "SHA256" || len(c.secret) != 52 {
		return errors.New("assessment TOTP contract mismatch")
	}
	code, err := c.code()
	if err != nil {
		return err
	}
	complete := c.authorized("/api/v1/auth/mfa/totp/complete", map[string]any{"client_txn_id": "capacity-enrollment-complete", "enrollment_id": begin.Data["enrollment_id"], "code": code}, token)
	if complete.err != nil || complete.Status != 200 {
		return errors.New("assessment TOTP complete failed")
	}
	var latencies []float64
	bursts := make([][]outcome, 0, 20)
	for range 20 {
		results := c.burst(10, false)
		bursts = append(bursts, results)
		report["bursts"] = bursts
		for _, r := range results {
			if r.err != nil || r.Status != 200 {
				return errors.New("valid credential burst did not complete successfully")
			}
			latencies = append(latencies, r.Millis)
		}
	}
	sort.Float64s(latencies)
	p95 := latencies[int(math.Ceil(float64(len(latencies))*.95))-1]
	report["p95_ms"] = p95
	if p95 > 2000 {
		return fmt.Errorf("credential p95 %.2f ms exceeds 2000 ms", p95)
	}
	overload := c.burst(64, false)
	report["overload"] = overload
	accepted, rejected := 0, 0
	for _, r := range overload {
		if r.err != nil {
			return errors.New("overload request failed at transport")
		}
		switch r.Status {
		case 200:
			accepted++
		case 503:
			rejected++
			if r.Code != "authentication_capacity_exhausted" || len(r.Details) != 0 || !r.Retryable || r.RetryAfter != "1" || r.Millis > 250 {
				return errors.New("excess work violated the bounded public overload response")
			}
		default:
			return errors.New("overload returned an unexpected response")
		}
	}
	if accepted != 10 || rejected != 54 {
		return fmt.Errorf("synchronized overload admitted %d and rejected %d; expected two active and eight pending", accepted, rejected)
	}
	cancelled := c.burst(10, true)
	report["cancelled"] = cancelled
	for _, r := range cancelled {
		if !r.Cancelled {
			return errors.New("cancellation probe unexpectedly completed")
		}
	}
	// Allow the two already-running, non-interruptible derivations to finish.
	time.Sleep(500 * time.Millisecond)
	recovery := c.burst(10, false)
	report["recovery"] = recovery
	for _, r := range recovery {
		if r.err != nil || r.Status != 200 {
			return errors.New("credential capacity did not recover after cancellation")
		}
	}
	return nil
}

func (c *client) code() (string, error) {
	return totp.GenerateCodeCustom(c.secret, time.Now(), totp.ValidateOpts{Period: 30, Skew: 1, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA256})
}
func (c *client) loginBody() map[string]any {
	body := map[string]any{"username": "admin@example.test", "password": c.password}
	if c.secret != "" {
		code, _ := c.code()
		body["second_factor"] = map[string]any{"kind": "totp", "assertion": map[string]any{"code": code}}
	}
	return body
}
func (c *client) authorized(path string, body map[string]any, token string) outcome {
	ctx := context.WithValue(context.Background(), authKey{}, token)
	return c.request(ctx, path, body)
}

type authKey struct{}

func (c *client) request(ctx context.Context, path string, body map[string]any) outcome {
	b, _ := json.Marshal(body)
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.origin+path, bytes.NewReader(b))
	if err != nil {
		return outcome{err: err}
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", c.origin)
	if token, ok := ctx.Value(authKey{}).(string); ok {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	start := time.Now()
	response, err := c.http.Do(r)
	if err != nil {
		return outcome{Millis: float64(time.Since(start)) / float64(time.Millisecond), Cancelled: errors.Is(err, context.Canceled), err: err}
	}
	defer response.Body.Close()
	var envelope struct {
		Data  map[string]any `json:"data"`
		Error struct {
			Code      string         `json:"code"`
			Details   map[string]any `json:"details"`
			Retryable bool           `json:"retryable"`
		} `json:"error"`
	}
	err = json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&envelope)
	return outcome{Status: response.StatusCode, Millis: float64(time.Since(start)) / float64(time.Millisecond), Code: envelope.Error.Code, Details: envelope.Error.Details, Retryable: envelope.Error.Retryable, RetryAfter: response.Header.Get("Retry-After"), Data: envelope.Data, err: err}
}
func (c *client) burst(n int, cancel bool) []outcome {
	ready := sync.WaitGroup{}
	ready.Add(n)
	release := make(chan struct{})
	results := make(chan outcome, n)
	ctx, stop := context.WithTimeout(context.Background(), 8*time.Second)
	defer stop()
	for range n {
		go func() {
			var once sync.Once
			trace := &httptrace.ClientTrace{GotConn: func(httptrace.GotConnInfo) {
				once.Do(func() {
					ready.Done()
					select {
					case <-release:
					case <-ctx.Done():
					}
				})
			}}
			result := c.request(httptrace.WithClientTrace(ctx, trace), "/api/v1/auth/login", c.loginBody())
			once.Do(ready.Done)
			results <- result
		}()
	}
	ready.Wait()
	close(release)
	if cancel {
		time.Sleep(5 * time.Millisecond)
		stop()
	}
	out := make([]outcome, 0, n)
	for range n {
		out = append(out, <-results)
	}
	return out
}

// Executed inside the measured application container through a read-only fixture
// bind. This records its effective cgroup, not the host or another container's.
func resourceReport() error {
	if err := cryptography.AdmitExecution(); err != nil {
		return err
	}
	report := map[string]string{}
	for _, name := range []string{"cpuset.cpus.effective", "cpu.max", "memory.max", "memory.swap.max"} {
		body, err := os.ReadFile(filepath.Join("/sys/fs/cgroup", name))
		if err != nil || len(body) > 4096 {
			return errors.New("effective cgroup observation failed")
		}
		report[name] = string(body)
	}
	return json.NewEncoder(os.Stdout).Encode(report)
}
