//go:build linux

package harnessreview

import (
	"bufio"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
)

const commandSchema = "cartulary.harness_review_command.v1"
const channelSchema = "cartulary.harness_review_channel.v1"

type commandRequest struct {
	Command     string         `json:"command"`
	Mode        string         `json:"mode"`
	Origin      *string        `json:"origin"`
	Profile     *string        `json:"profile"`
	ToolProfile map[string]any `json:"tool_profile"`
	Results     string         `json:"results_root"`
	RunID       string         `json:"run_id"`
	Policy      string         `json:"preparation_policy"`
	Locator     string         `json:"locator"`
	Request     jsonRequest    `json:"request"`
	Bundle      *string        `json:"bundle_id"`
}
type controlRequest struct {
	Schema  string      `json:"schema_id"`
	Session string      `json:"session_id"`
	Token   string      `json:"token"`
	Command string      `json:"command"`
	Request jsonRequest `json:"request"`
	Bundle  *string     `json:"bundle_id"`
}

func validateCommand(command string, request jsonRequest, bundle *string) error {
	kind := map[string]string{"ui-browser": "cartulary.ui_review_action.v3", "ui-capture": "cartulary.ui_review_capture_request.v1", "ui-analyze": "cartulary.ui_review_analysis_request.v1"}[command]
	if kind != "" {
		data, _ := json.Marshal(request)
		if bundle != nil || len(data) > LimitRequest {
			return failure("invalid_request")
		}
		value, err := harnessschema.Decode(data)
		if err != nil || harnessschema.Validate(kind, value) != nil {
			return failure("invalid_request")
		}
		return nil
	}
	if request != nil || (command == "ui-review-report") != (bundle != nil) {
		return failure("invalid_request")
	}
	return nil
}
func serveConnection(connection net.Conn, record *registryRecord, session *Session) {
	defer connection.Close()
	_ = connection.SetReadDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReaderSize(connection, harnesssession.MaximumFrame)
	bytes, err := reader.ReadSlice('\n')
	if err != nil {
		return
	}
	value, err := harnessschema.Decode(bytes)
	if err != nil || harnessschema.Validate(channelSchema, value) != nil {
		return
	}
	var q controlRequest
	if json.Unmarshal(bytes, &q) != nil || q.Session != record.Session || subtle.ConstantTimeCompare([]byte(q.Token), []byte(record.Token)) != 1 {
		return
	}
	_ = connection.SetReadDeadline(time.Time{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// A lost accepted caller cancels the session; an invalid/unauthenticated peer
	// never obtains cancellation authority. Extra bytes are rejected too.
	answered := make(chan struct{})
	go func() {
		var b [1]byte
		n, err := reader.Read(b[:])
		select {
		case <-answered:
			return
		default:
		}
		if err != nil || n > 0 {
			cancel()
			session.Stop(failure("interrupted"))
		}
	}()
	var result Result
	if err = validateCommand(q.Command, q.Request, q.Bundle); err != nil {
		result = session.fail(q.Command, err)
	} else {
		result = session.Handle(ctx, q.Command, q.Request, q.Bundle)
	}
	close(answered)
	data, err := jsonDocument(result.Schema, result)
	if err != nil || len(data) > LimitRequest {
		session.Stop(failure("unsafe_artifact"))
		return
	}
	_ = connection.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err = connection.Write(data); err != nil {
		session.Stop(failure("interrupted"))
	}
}
func control(ctx context.Context, repository string, q commandRequest, environment map[string]string) (Result, error) {
	file, locator, err := readLocator(repository, q.Locator)
	if err != nil {
		return result(q.Command), err
	}
	terminal := locator.State == "closed" || locator.State == "failed"
	if terminal {
		if q.Command != "ui-review-status" && q.Command != "ui-review-stop" {
			return result(q.Command), failure("session_mismatch")
		}
		prior, terminalErr := terminalResult(file, locator, q.Command)
		if q.Command != "ui-review-stop" || locator.State != "failed" {
			return prior, terminalErr
		}
		record, err := resolveRecord(repository, file, locator)
		if errors.Is(err, os.ErrNotExist) {
			return prior, terminalErr
		}
		if err != nil {
			return prior, err
		}
		alive, err := harnessadmission.Alive(record.Process)
		if err != nil || alive {
			return prior, errors.Join(terminalErr, err)
		}
		return recoverSession(ctx, repository, file, locator, q.Command, environment)
	}
	record, err := resolveRecord(repository, file, locator)
	if err != nil {
		return result(q.Command), failure("session_mismatch")
	}
	alive, err := harnessadmission.Alive(record.Process)
	if err != nil {
		return result(q.Command), err
	}
	if !alive {
		if q.Command != "ui-review-stop" {
			return result(q.Command), failure("session_mismatch")
		}
		return recoverSession(ctx, repository, file, locator, q.Command, environment)
	}
	if err = socketProof(record); err != nil {
		return result(q.Command), failure("session_mismatch")
	}
	if err = validateCommand(q.Command, q.Request, q.Bundle); err != nil {
		return result(q.Command), err
	}
	timeout := time.Duration(LimitOperation+LimitLock+10000) * time.Millisecond
	if q.Command == "ui-review-stop" {
		timeout = 360 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	connection, err := (&net.Dialer{}).DialContext(ctx, "unix", record.Socket)
	if err != nil {
		return result(q.Command), failure("session_lost")
	}
	defer connection.Close()
	go func() { <-ctx.Done(); connection.Close() }()
	_ = connection.SetDeadline(time.Now().Add(timeout))
	message := controlRequest{Schema: channelSchema, Session: record.Session, Token: record.Token, Command: q.Command, Request: q.Request, Bundle: q.Bundle}
	data, err := jsonDocument(channelSchema, message)
	if err != nil || len(data) > harnesssession.MaximumFrame {
		return result(q.Command), failure("invalid_request")
	}
	if _, err = connection.Write(data); err != nil {
		return result(q.Command), failure("session_lost")
	}
	data, err = bufio.NewReaderSize(connection, LimitRequest).ReadSlice('\n')
	if err != nil {
		return result(q.Command), failure("session_lost")
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate("cartulary.ui_review_command_result.v2", value) != nil {
		return result(q.Command), failure("invalid_artifact")
	}
	var response Result
	if json.Unmarshal(data, &response) != nil || response.Session == nil || *response.Session != record.Session || response.Command != commandID(q.Command) {
		return result(q.Command), failure("invalid_artifact")
	}
	return response, nil
}
func start(ctx context.Context, repository string, q commandRequest, environment map[string]string, output io.Writer, interrupted func() error) (Result, error) {
	if q.Mode == "seeded" && q.Profile == nil || q.Mode != "seeded" && q.Profile != nil || q.Mode == "dev" && q.Origin == nil || q.Mode != "dev" && q.Origin != nil {
		return result("ui-review"), failure("invalid_request")
	}
	runRoot, err := newRunRoot(repository, q.Results, q.RunID)
	if err != nil {
		return result("ui-review"), err
	}
	runtime, err := harnesssession.NewRuntime(repository, runRoot, q.RunID, "")
	if err != nil {
		return result("ui-review"), err
	}
	defer runtime.ReleaseHandles()
	adopted := false
	defer func() {
		if !adopted {
			_ = runtime.Close()
		}
	}()
	if err = runtime.RegisterEnvironment(environment); err != nil {
		return result("ui-review"), err
	}
	sessionID := "uireview-" + nonce(16)
	directory, err := harnesslocal.PrivateDirectory(registryRoot())
	if err != nil {
		_ = runtime.Close()
		return result("ui-review"), err
	}
	directory.Close()
	socket := filepath.Join(registryRoot(), strings.TrimPrefix(sessionID, "uireview-")+".sock")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		_ = runtime.Close()
		return result("ui-review"), err
	}
	listener.SetUnlinkOnClose(false)
	defer listener.Close()
	if err = os.Chmod(socket, 0600); err != nil {
		return result("ui-review"), err
	}
	var info unix.Stat_t
	if err = unix.Lstat(socket, &info); err != nil {
		return result("ui-review"), err
	}
	defer func() {
		if !adopted {
			var current unix.Stat_t
			if unix.Lstat(socket, &current) == nil && current.Dev == info.Dev && current.Ino == info.Ino {
				_ = os.Remove(socket)
			}
		}
	}()
	proof, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return result("ui-review"), err
	}
	record := &registryRecord{Schema: registrySchema, Session: sessionID, Locator: filepath.Join(runRoot, "ui-review/session.json"), Repository: repository, Process: proof, StartedTick: strconv.FormatInt(monotonic(), 10), Token: nonce(32), Socket: socket, Device: info.Dev, Inode: info.Ino, Runtime: runtimeRef{runtime.Root, runtime.LeaseID, runtime.RunID}}
	record.Presentation = "ui"
	if q.Command != "ui-review" {
		record.Presentation = "design"
	}
	if q.Profile != nil {
		record.Profile = *q.Profile
	}
	if err = writeRegistry(record, false); err != nil {
		return result("ui-review"), err
	}
	if err = runtime.RegisterSecret(record.Token); err != nil {
		_ = removeRegistry(record)
		return result("ui-review"), err
	}
	publisher := &diskPublisher{runtime: runtime, record: record, runRoot: runRoot}
	driver := &productionDriver{root: repository, runRoot: runRoot, session: sessionID, input: reviewInput{Mode: q.Mode, Origin: q.Origin, Profile: q.Profile, Policy: q.Policy, Presentation: "ui"}, profile: q.ToolProfile, runtime: runtime, environment: environment}
	var owner Publisher = publisher
	var design *designPublisher
	if q.Command != "ui-review" {
		driver.input.Presentation = "design"
		driver.input.VerifySamples = q.Command == "browser-design-review-smoke"
		design = &designPublisher{diskPublisher: publisher, profile: *q.Profile, driver: driver}
		owner = design
	}
	session := NewSession(sessionID, q.RunID, q.Mode, driver, owner)
	adopted = true
	driver.onLost = session.Stop
	driver.isReady = session.Ready
	var outputMu sync.Mutex
	emit := func(kind string, value any) {
		outputMu.Lock()
		defer outputMu.Unlock()
		if json.NewEncoder(output).Encode(map[string]any{"kind": kind, "value": value}) != nil {
			session.Stop(failure("session_lost"))
		}
	}
	// Publish the locator before announcing it; serving starts before preparation.
	session.mu.Lock()
	err = session.publishLocked()
	session.mu.Unlock()
	if err != nil {
		close(session.prepareDone)
		session.Stop(failure("unsafe_artifact"))
		<-session.Done()
		return session.Value("ui-review"), nil
	}
	emit("preparing", record.Locator)
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go serveConnection(connection, record, session)
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
			session.Stop(interrupted())
		case <-session.Done():
		}
	}()
	if err = session.Prepare(); err == nil {
		if design != nil {
			session.mu.Lock()
			design.ready = driver.facts.Review
			record.Review = design.ready
			err = errors.Join(writeRegistry(record, true), design.details(session.state, nil))
			session.mu.Unlock()
			if err != nil {
				session.Stop(failure("unsafe_artifact"))
			} else {
				emit("design_ready", design.ready)
			}
			if driver.input.VerifySamples {
				session.Stop(nil)
			}
		} else {
			emit("ready", record.Locator)
		}
	}
	<-session.Done()
	return session.Value("ui-review"), nil
}
