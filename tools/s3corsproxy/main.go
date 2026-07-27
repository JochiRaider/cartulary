package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	defaultListenAddr = "127.0.0.1:8333"
	defaultUpstream   = "http://127.0.0.1:18333"
	defaultOrigin     = "http://localhost:5173"
	healthPath        = "/.cartulary/s3corsproxy/health"
)

var (
	allowedMethods = []string{http.MethodPut, http.MethodOptions}
	allowedHeaders = []string{"content-type", "x-amz-checksum-sha256"}
	exposedHeaders = []string{"etag"}
)

type proxyConfig struct {
	Listen            string `json:"listen"`
	UpstreamOrigin    string `json:"upstream_origin"`
	AllowedOrigin     string `json:"allowed_origin"`
	HealthPath        string `json:"health_path"`
	ConfigFingerprint string `json:"configuration_fingerprint"`
}

type processProof struct {
	BootID           string `json:"boot_id"`
	PID              int    `json:"pid"`
	StartTimeTicks   uint64 `json:"start_time_ticks"`
	EffectiveUID     int    `json:"effective_uid"`
	ExecutableDevice uint64 `json:"executable_device"`
	ExecutableInode  uint64 `json:"executable_inode"`
	ExecutableSHA256 string `json:"executable_sha256"`
}

type startAttempt struct {
	SchemaID   string        `json:"schema_id"`
	InstanceID string        `json:"instance_id"`
	State      string        `json:"state"`
	CreatedAt  string        `json:"created_at"`
	BoundAt    string        `json:"bound_at,omitempty"`
	Config     proxyConfig   `json:"configuration"`
	Process    *processProof `json:"process,omitempty"`
	LogPath    string        `json:"log_path"`
}

type proxyLease struct {
	SchemaID   string       `json:"schema_id"`
	InstanceID string       `json:"instance_id"`
	ReadyAt    string       `json:"ready_at"`
	Config     proxyConfig  `json:"configuration"`
	Process    processProof `json:"process"`
	LogPath    string       `json:"log_path"`
}

type proxyHealth struct {
	SchemaID   string       `json:"schema_id"`
	InstanceID string       `json:"instance_id"`
	EmittedAt  string       `json:"emitted_at"`
	Config     proxyConfig  `json:"configuration"`
	Process    processProof `json:"process"`
}

type commandOptions struct {
	listen      string
	upstream    string
	origin      string
	attemptFile string
	leaseFile   string
	stateFile   string
	instanceID  string
	logPath     string
}

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Print(err)
		var coded *exitError
		if errors.As(err, &coded) {
			os.Exit(coded.code)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		args = append([]string{"serve"}, args...)
	}
	command := args[0]
	options, err := parseOptions(command, args[1:])
	if err != nil {
		return &exitError{code: 2, err: err}
	}
	config, err := normalizeConfig(options.listen, options.upstream, options.origin)
	if err != nil {
		return &exitError{code: 2, err: err}
	}

	switch command {
	case "attempt":
		return createAttempt(options, config)
	case "serve":
		return serve(options, config)
	case "status":
		return status(options.stateFile, config)
	case "promote":
		return promote(options, config)
	case "stop":
		return stop(options.stateFile)
	case "discard":
		return secureRemove(options.stateFile)
	default:
		return &exitError{code: 2, err: fmt.Errorf("unknown command %q", command)}
	}
}

func parseOptions(command string, args []string) (commandOptions, error) {
	options := commandOptions{}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.listen, "listen", envDefault("OBJECT_STORE_CORS_PROXY_LISTEN", defaultListenAddr), "listen address")
	flags.StringVar(&options.upstream, "upstream", envDefault("OBJECT_STORE_CORS_PROXY_UPSTREAM", defaultUpstream), "upstream origin")
	flags.StringVar(&options.origin, "origin", envDefault("OBJECT_STORE_CORS_ORIGIN", defaultOrigin), "allowed browser origin")
	flags.StringVar(&options.attemptFile, "attempt-file", "", "startup attempt state")
	flags.StringVar(&options.leaseFile, "lease-file", "", "ready lease state")
	flags.StringVar(&options.stateFile, "state-file", "", "attempt or lease state")
	flags.StringVar(&options.instanceID, "instance-id", "", "proxy instance identity")
	flags.StringVar(&options.logPath, "log-path", "", "per-instance log path")
	if err := flags.Parse(args); err != nil {
		return options, err
	}
	if flags.NArg() != 0 {
		return options, errors.New("unexpected positional arguments")
	}
	switch command {
	case "attempt", "serve":
		if options.attemptFile == "" || options.instanceID == "" || options.logPath == "" {
			return options, fmt.Errorf("%s requires --attempt-file, --instance-id, and --log-path", command)
		}
	case "status", "stop", "discard":
		if options.stateFile == "" {
			return options, fmt.Errorf("%s requires --state-file", command)
		}
	case "promote":
		if options.attemptFile == "" || options.leaseFile == "" {
			return options, errors.New("promote requires --attempt-file and --lease-file")
		}
	}
	return options, nil
}

func normalizeConfig(listenRaw, upstreamRaw, originRaw string) (proxyConfig, error) {
	host, port, err := net.SplitHostPort(strings.TrimSpace(listenRaw))
	if err != nil {
		return proxyConfig{}, fmt.Errorf("parse listener: %w", err)
	}
	ip := net.ParseIP(host)
	portNumber, err := strconv.Atoi(port)
	if ip == nil || !ip.Equal(net.ParseIP("127.0.0.1")) || portNumber < 1 || portNumber > 65535 {
		return proxyConfig{}, errors.New("listener must use 127.0.0.1 and a valid port")
	}
	listen := net.JoinHostPort(ip.String(), strconv.Itoa(portNumber))
	upstream, err := normalizeOrigin(upstreamRaw, false)
	if err != nil {
		return proxyConfig{}, fmt.Errorf("normalize upstream: %w", err)
	}
	origin, err := normalizeOrigin(originRaw, true)
	if err != nil {
		return proxyConfig{}, fmt.Errorf("normalize allowed origin: %w", err)
	}
	config := proxyConfig{
		Listen:         listen,
		UpstreamOrigin: upstream,
		AllowedOrigin:  origin,
		HealthPath:     healthPath,
	}
	semantic, _ := json.Marshal(config)
	sum := sha256.Sum256(semantic)
	config.ConfigFingerprint = "sha256:" + hex.EncodeToString(sum[:])
	return config, nil
}

func normalizeOrigin(raw string, requireOriginOnly bool) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", errors.New("origin must use http or https and include a host")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("origin must not contain userinfo, query, or fragment")
	}
	if requireOriginOnly && parsed.Path != "" && parsed.Path != "/" {
		return "", errors.New("allowed origin must not contain a path")
	}
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	parsed.RawPath = strings.TrimSuffix(parsed.RawPath, "/")
	return strings.TrimSuffix(parsed.String(), "/"), nil
}

func createAttempt(options commandOptions, config proxyConfig) error {
	attempt := startAttempt{
		SchemaID:   "cartulary.local_object_store_proxy_start_attempt.v1",
		InstanceID: options.instanceID,
		State:      "launching",
		CreatedAt:  time.Now().UTC().Format(time.RFC3339Nano),
		Config:     config,
		LogPath:    options.logPath,
	}
	return atomicWriteJSON(options.attemptFile, attempt)
}

func serve(options commandOptions, config proxyConfig) error {
	attempt, err := readAttempt(options.attemptFile)
	if err != nil {
		return err
	}
	if attempt.InstanceID != options.instanceID ||
		attempt.State != "launching" ||
		attempt.Config.ConfigFingerprint != config.ConfigFingerprint {
		return errors.New("startup attempt does not authorize this proxy instance")
	}
	upstreamURL, err := url.Parse(config.UpstreamOrigin)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return fmt.Errorf("bind proxy listener: %w", err)
	}
	defer listener.Close()
	proof, err := collectProcessProof(os.Getpid())
	if err != nil {
		return fmt.Errorf("collect proxy process proof: %w", err)
	}
	attempt.State = "bound"
	attempt.BoundAt = time.Now().UTC().Format(time.RFC3339Nano)
	attempt.Process = &proof
	if err := atomicWriteJSON(options.attemptFile, attempt); err != nil {
		return err
	}

	proxyHandler := newProxyHandler(upstreamURL, config.AllowedOrigin)
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == healthPath {
			if request.Method != http.MethodGet || request.Host != config.Listen {
				writer.WriteHeader(http.StatusForbidden)
				return
			}
			current, proofErr := collectProcessProof(os.Getpid())
			if proofErr != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(writer).Encode(proxyHealth{
				SchemaID:   "cartulary.local_object_store_proxy_health.v1",
				InstanceID: options.instanceID,
				EmittedAt:  time.Now().UTC().Format(time.RFC3339Nano),
				Config:     config,
				Process:    current,
			})
			return
		}
		proxyHandler.ServeHTTP(writer, request)
	})
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()
	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-signalCh:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown after %s: %w", sig, err)
		}
	case serveErr := <-errCh:
		if serveErr != nil && serveErr != http.ErrServerClosed {
			return fmt.Errorf("serve proxy: %w", serveErr)
		}
	}
	return nil
}

func status(stateFile string, expected proxyConfig) error {
	state, err := readState(stateFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &exitError{code: 4, err: err}
		}
		return &exitError{code: 1, err: err}
	}
	if err := verifyState(state); err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ESRCH) {
			return &exitError{code: 4, err: err}
		}
		return &exitError{code: 1, err: err}
	}
	if state.config.ConfigFingerprint != expected.ConfigFingerprint {
		return &exitError{code: 3, err: errors.New("proxy state is fully proven but configuration differs")}
	}
	fmt.Println("matching")
	return nil
}

func promote(options commandOptions, expected proxyConfig) error {
	attempt, err := readAttempt(options.attemptFile)
	if err != nil {
		return err
	}
	if attempt.State != "bound" || attempt.Process == nil {
		return errors.New("startup attempt is not bound and identity-proved")
	}
	state := provenState{
		instanceID: attempt.InstanceID,
		config:     attempt.Config,
		process:    *attempt.Process,
		logPath:    attempt.LogPath,
	}
	if state.config.ConfigFingerprint != expected.ConfigFingerprint {
		return errors.New("startup attempt configuration mismatch")
	}
	if err := verifyState(state); err != nil {
		return err
	}
	lease := proxyLease{
		SchemaID:   "cartulary.local_object_store_proxy_lease.v1",
		InstanceID: state.instanceID,
		ReadyAt:    time.Now().UTC().Format(time.RFC3339Nano),
		Config:     state.config,
		Process:    state.process,
		LogPath:    state.logPath,
	}
	if err := atomicWriteJSON(options.leaseFile, lease); err != nil {
		return err
	}
	return secureRemove(options.attemptFile)
}

func stop(stateFile string) error {
	state, err := readState(stateFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := verifyState(state); err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ESRCH) {
			return secureRemove(stateFile)
		}
		return fmt.Errorf("refusing to signal unproven proxy process: %w", err)
	}
	pidfd, err := unix.PidfdOpen(state.process.PID, 0)
	if err != nil {
		return fmt.Errorf("pidfd_open is required for proxy signaling: %w", err)
	}
	defer unix.Close(pidfd)
	if err := verifyProcessProof(state.process); err != nil {
		return fmt.Errorf("proxy proof changed after pidfd_open: %w", err)
	}
	if err := unix.PidfdSendSignal(pidfd, unix.SIGTERM, nil, 0); err != nil {
		return fmt.Errorf("pidfd_send_signal TERM: %w", err)
	}
	exited, err := waitPidfd(pidfd, 5*time.Second)
	if err != nil {
		return err
	}
	if !exited {
		if err := unix.PidfdSendSignal(pidfd, unix.SIGKILL, nil, 0); err != nil {
			return fmt.Errorf("pidfd_send_signal KILL: %w", err)
		}
		exited, err = waitPidfd(pidfd, 5*time.Second)
		if err != nil {
			return err
		}
	}
	if !exited {
		return errors.New("proxy did not terminate through pidfd")
	}
	return secureRemove(stateFile)
}

type provenState struct {
	instanceID string
	config     proxyConfig
	process    processProof
	logPath    string
}

func readState(file string) (provenState, error) {
	raw, err := secureRead(file)
	if err != nil {
		return provenState{}, err
	}
	var identity struct {
		SchemaID string `json:"schema_id"`
	}
	if err := json.Unmarshal(raw, &identity); err != nil {
		return provenState{}, err
	}
	switch identity.SchemaID {
	case "cartulary.local_object_store_proxy_start_attempt.v1":
		attempt, err := decodeAttempt(raw)
		if err != nil {
			return provenState{}, err
		}
		if attempt.State != "bound" || attempt.Process == nil {
			return provenState{}, errors.New("startup attempt has no complete process proof")
		}
		return provenState{attempt.InstanceID, attempt.Config, *attempt.Process, attempt.LogPath}, nil
	case "cartulary.local_object_store_proxy_lease.v1":
		var lease proxyLease
		if err := decodeStrict(raw, &lease); err != nil {
			return provenState{}, err
		}
		return provenState{lease.InstanceID, lease.Config, lease.Process, lease.LogPath}, nil
	default:
		return provenState{}, fmt.Errorf("unsupported proxy state schema %q", identity.SchemaID)
	}
}

func readAttempt(file string) (startAttempt, error) {
	raw, err := secureRead(file)
	if err != nil {
		return startAttempt{}, err
	}
	return decodeAttempt(raw)
}

func decodeAttempt(raw []byte) (startAttempt, error) {
	var attempt startAttempt
	if err := decodeStrict(raw, &attempt); err != nil {
		return startAttempt{}, err
	}
	if attempt.SchemaID != "cartulary.local_object_store_proxy_start_attempt.v1" {
		return startAttempt{}, errors.New("unexpected startup attempt schema")
	}
	return attempt, nil
}

func decodeStrict(raw []byte, output any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(output); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return errors.New("proxy state contains trailing JSON")
	}
	return nil
}

func verifyState(state provenState) error {
	if err := verifyProcessProof(state.process); err != nil {
		return err
	}
	health, err := fetchHealth(state.config)
	if err != nil {
		return err
	}
	if health.SchemaID != "cartulary.local_object_store_proxy_health.v1" ||
		health.InstanceID != state.instanceID ||
		health.Config.ConfigFingerprint != state.config.ConfigFingerprint ||
		health.Process != state.process {
		return errors.New("proxy health identity does not match retained process proof")
	}
	return nil
}

func fetchHealth(config proxyConfig) (proxyHealth, error) {
	client := &http.Client{Timeout: 750 * time.Millisecond}
	request, err := http.NewRequest(http.MethodGet, "http://"+config.Listen+healthPath, nil)
	if err != nil {
		return proxyHealth{}, err
	}
	request.Host = config.Listen
	response, err := client.Do(request)
	if err != nil {
		return proxyHealth{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return proxyHealth{}, fmt.Errorf("proxy health returned %s", response.Status)
	}
	var health proxyHealth
	decoder := json.NewDecoder(io.LimitReader(response.Body, 64*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&health); err != nil {
		return proxyHealth{}, err
	}
	return health, nil
}

func collectProcessProof(pid int) (processProof, error) {
	statBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return processProof{}, err
	}
	closeIndex := strings.LastIndex(string(statBytes), ")")
	if closeIndex < 0 {
		return processProof{}, errors.New("malformed proc stat")
	}
	fields := strings.Fields(string(statBytes)[closeIndex+1:])
	if len(fields) < 20 {
		return processProof{}, errors.New("proc stat omits start-time ticks")
	}
	startTicks, err := strconv.ParseUint(fields[19], 10, 64)
	if err != nil {
		return processProof{}, err
	}
	statusBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return processProof{}, err
	}
	effectiveUID := -1
	for _, line := range strings.Split(string(statusBytes), "\n") {
		if strings.HasPrefix(line, "Uid:") {
			uidFields := strings.Fields(line)
			if len(uidFields) >= 3 {
				effectiveUID, err = strconv.Atoi(uidFields[2])
			}
			break
		}
	}
	if err != nil || effectiveUID < 0 {
		return processProof{}, errors.New("proc status omits effective UID")
	}
	executable := fmt.Sprintf("/proc/%d/exe", pid)
	info, err := os.Stat(executable)
	if err != nil {
		return processProof{}, err
	}
	systemStat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return processProof{}, errors.New("executable stat lacks device/inode proof")
	}
	executableBytes, err := os.ReadFile(executable)
	if err != nil {
		return processProof{}, err
	}
	bootID, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return processProof{}, err
	}
	sum := sha256.Sum256(executableBytes)
	return processProof{
		BootID:           strings.TrimSpace(string(bootID)),
		PID:              pid,
		StartTimeTicks:   startTicks,
		EffectiveUID:     effectiveUID,
		ExecutableDevice: uint64(systemStat.Dev),
		ExecutableInode:  systemStat.Ino,
		ExecutableSHA256: "sha256:" + hex.EncodeToString(sum[:]),
	}, nil
}

func verifyProcessProof(expected processProof) error {
	actual, err := collectProcessProof(expected.PID)
	if err != nil {
		return err
	}
	if actual != expected {
		return errors.New("process proof mismatch")
	}
	return nil
}

func waitPidfd(pidfd int, timeout time.Duration) (bool, error) {
	poll := []unix.PollFd{{Fd: int32(pidfd), Events: unix.POLLIN}}
	result, err := unix.Poll(poll, int(timeout.Milliseconds()))
	if err != nil {
		return false, fmt.Errorf("wait for pidfd: %w", err)
	}
	return result > 0 && poll[0].Revents != 0, nil
}

func secureRead(file string) ([]byte, error) {
	info, err := os.Lstat(file)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("proxy state must be a non-symlink regular file")
	}
	systemStat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(systemStat.Uid) != os.Geteuid() {
		return nil, errors.New("proxy state must be owned by the effective user")
	}
	return os.ReadFile(file)
}

func atomicWriteJSON(file string, value any) error {
	parent := filepath.Dir(file)
	if err := ensureSecureDirectory(parent); err != nil {
		return err
	}
	if info, err := os.Lstat(file); err == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("refusing to replace non-regular proxy state")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	temporary, err := os.OpenFile(
		file+".tmp-"+strconv.Itoa(os.Getpid())+"-"+strconv.FormatInt(time.Now().UnixNano(), 10),
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o600,
	)
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if _, err := temporary.Write(payload); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, file); err != nil {
		return err
	}
	return syncDirectory(parent)
}

func ensureSecureDirectory(directory string) error {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return err
	}
	systemStat, ok := info.Sys().(*syscall.Stat_t)
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || !ok || int(systemStat.Uid) != os.Geteuid() {
		return errors.New("proxy runtime root must be an owner-controlled non-symlink directory")
	}
	if info.Mode().Perm()&0o077 != 0 {
		if err := os.Chmod(directory, 0o700); err != nil {
			return err
		}
	}
	return nil
}

func secureRemove(file string) error {
	if _, err := secureRead(file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := os.Remove(file); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(file))
}

func syncDirectory(directory string) error {
	handle, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer handle.Close()
	return handle.Sync()
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
