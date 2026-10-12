//go:build linux

package harnessreview

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
)

const registrySchema = "cartulary.harness_review_registry.v1"

type runtimeRef struct {
	Root  string `json:"root"`
	Lease string `json:"lease_id"`
	RunID string `json:"run_id"`
}
type registryRecord struct {
	Schema       string                 `json:"schema_id"`
	Session      string                 `json:"session_id"`
	Locator      string                 `json:"locator"`
	Repository   string                 `json:"repository"`
	Process      harnessadmission.Proof `json:"process"`
	StartedTick  string                 `json:"started_tick_ns"`
	Token        string                 `json:"token"`
	Socket       string                 `json:"socket"`
	Device       uint64                 `json:"socket_device"`
	Inode        uint64                 `json:"socket_inode"`
	Runtime      runtimeRef             `json:"runtime"`
	Released     bool                   `json:"resources_released"`
	Presentation string                 `json:"presentation"`
	Profile      string                 `json:"profile"`
	Review       *designReady           `json:"review"`
	Outcome      *Receipt               `json:"outcome"`
}

func nonce(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func monotonic() int64 {
	var v unix.Timespec
	if unix.ClockGettime(unix.CLOCK_MONOTONIC, &v) != nil {
		return 0
	}
	return v.Nano()
}
func registryRoot() string {
	return filepath.Join(os.TempDir(), "cartulary-ui-review-"+strconv.Itoa(os.Getuid()))
}
func registryPath(session string) string {
	return filepath.Join(registryRoot(), strings.TrimPrefix(session, "uireview-")+".json")
}
func readDocument(file, schema string, maximum int64, out any) error {
	data, err := harnesslocal.ReadFile(file, maximum, false)
	if err != nil {
		return err
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate(schema, value) != nil {
		return harnessschema.ErrInvalid
	}
	return json.Unmarshal(data, out)
}
func publish(file, schema string, value any, replace bool) (*Reference, error) {
	data, err := jsonDocument(schema, value)
	if err != nil {
		return nil, err
	}
	directory, err := harnesslocal.PrivateDirectory(filepath.Dir(file))
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	if replace {
		err = harnesslocal.ReplaceFile(file, data)
	} else {
		err = harnesslocal.PublishNew(file, data)
	}
	if err != nil {
		return nil, err
	}
	return &Reference{Bytes: len(data), SHA256: fmt.Sprintf("%x", sha256.Sum256(data)), MediaType: "application/json"}, nil
}
func writeRegistry(r *registryRecord, replace bool) error {
	data, err := jsonDocument(registrySchema, r)
	if err != nil || len(data) > harnesssession.MaximumFrame {
		return harnessschema.ErrInvalid
	}
	_, err = publish(registryPath(r.Session), registrySchema, r, replace)
	return err
}
func socketProof(r *registryRecord) error {
	var info unix.Stat_t
	if err := unix.Lstat(r.Socket, &info); err != nil {
		return err
	}
	if info.Mode&unix.S_IFMT != unix.S_IFSOCK || info.Mode&0777 != 0600 || info.Uid != uint32(os.Getuid()) || info.Dev != r.Device || info.Ino != r.Inode {
		return harnesslocal.ErrUnsafe
	}
	return nil
}
func removeRegistry(r *registryRecord) error {
	var current registryRecord
	err := readDocument(registryPath(r.Session), registrySchema, harnesssession.MaximumFrame, &current)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || !reflect.DeepEqual(current, *r) {
		return harnesslocal.ErrUnsafe
	}
	if err = socketProof(r); err == nil {
		err = os.Remove(r.Socket)
	} else if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	if err != nil {
		return err
	}
	return os.Remove(registryPath(r.Session))
}
func resolveRecord(repository, file string, locator Locator) (*registryRecord, error) {
	var r registryRecord
	if err := readDocument(registryPath(locator.Session), registrySchema, harnesssession.MaximumFrame, &r); err != nil {
		return nil, err
	}
	if r.Repository != repository || r.Session != locator.Session || r.Locator != file || r.Runtime.RunID != locator.RunID || r.Socket != filepath.Join(registryRoot(), strings.TrimPrefix(r.Session, "uireview-")+".sock") {
		return nil, harnesslocal.ErrUnsafe
	}
	return &r, nil
}
func inputPath(repository, file string) (string, error) {
	if file == "" || strings.ContainsAny(file, "\\\x00\n\r") || strings.Contains(file, "://") {
		return "", failure("invalid_request")
	}
	for _, part := range strings.Split(file, "/") {
		if part == ".." {
			return "", failure("invalid_request")
		}
	}
	if !filepath.IsAbs(file) {
		file = filepath.Join(repository, file)
	}
	file = filepath.Clean(file)
	leaf := strings.ToLower(filepath.Base(file))
	ext := strings.ToLower(filepath.Ext(file))
	if leaf == "readme" || strings.HasPrefix(leaf, "readme.") || ext == ".md" || ext == ".markdown" || ext == ".mdown" {
		return "", failure("input_boundary")
	}
	// The restriction is an executable projection, never a Markdown dependency.
	data, err := os.ReadFile(filepath.Join(repository, "tools/executable_input_policy.json"))
	if err != nil {
		return "", err
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate("cartulary.executable_input_policy.v1", value) != nil {
		return "", harnessschema.ErrInvalid
	}
	for _, root := range value.(map[string]any)["restricted_roots"].([]any) {
		base := filepath.Join(repository, root.(string))
		relative, _ := filepath.Rel(base, file)
		if relative == "." || relative != ".." && !strings.HasPrefix(relative, "../") {
			return "", failure("input_boundary")
		}
	}
	if ext != ".json" {
		return "", failure("invalid_request")
	}
	return file, nil
}
func readLocator(repository, file string) (string, Locator, error) {
	var locator Locator
	absolute, err := inputPath(repository, file)
	if err != nil {
		return "", locator, err
	}
	if err = readDocument(absolute, "cartulary.ui_review_session.v1", LimitRequest, &locator); err != nil {
		return "", locator, normalized(err, "unsafe_artifact")
	}
	if filepath.Base(absolute) != "session.json" || filepath.Base(filepath.Dir(absolute)) != "ui-review" || filepath.Base(filepath.Dir(filepath.Dir(absolute))) != locator.RunID {
		return "", locator, failure("session_mismatch")
	}
	if locator.Receipt != nil && locator.State != "closed" && locator.State != "failed" {
		return "", locator, failure("invalid_artifact")
	}
	return absolute, locator, nil
}
func terminalResult(file string, locator Locator, command string) (Result, error) {
	v := result(command)
	ref := locator.Receipt
	if ref == nil {
		return v, failure("unsafe_artifact")
	}
	if ref.Path != "ui-review/terminal.json" || ref.MediaType != "application/json" {
		return v, failure("invalid_artifact")
	}
	data, err := harnesslocal.ReadFile(filepath.Join(filepath.Dir(filepath.Dir(file)), ref.Path), LimitRequest, false)
	if err != nil {
		return v, failure("unsafe_artifact")
	}
	if ref.Bytes != len(data) || ref.SHA256 != fmt.Sprintf("%x", sha256.Sum256(data)) {
		return v, failure("invalid_artifact")
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate("cartulary.ui_review_receipt.v2", value) != nil {
		return v, failure("invalid_artifact")
	}
	var receipt Receipt
	if json.Unmarshal(data, &receipt) != nil || receipt.Session != locator.Session || receipt.State != locator.State || receipt.Operation != nil || receipt.Command != commandID("ui-review") || receipt.Cleanup == "not_terminal" || (receipt.Status == "ok") != (len(receipt.Failures) == 0) || (receipt.Exit == 0) != (receipt.Status == "ok") {
		return v, failure("invalid_artifact")
	}
	for _, f := range receipt.Failures {
		if f.validate() != nil {
			return v, failure("invalid_artifact")
		}
	}
	v.Session = &locator.Session
	v.State = &locator.State
	v.Receipt = ref
	v.Status = receipt.Status
	v.Exit = receipt.Exit
	v.Failures = receipt.Failures
	return v, nil
}

type diskPublisher struct {
	runtime *harnesssession.Runtime
	record  *registryRecord
	runRoot string
	secrets []string
}

func (p *diskPublisher) Locator(v Locator) error {
	_, err := publish(p.record.Locator, v.Schema, v, true)
	return err
}
func (p *diskPublisher) Operation(v Receipt) (*Reference, error) {
	if v.Operation == nil {
		return nil, harnesslocal.ErrUnsafe
	}
	relative := fmt.Sprintf("ui-review/operations/%d/receipt.json", *v.Operation)
	ref, err := publish(filepath.Join(p.runRoot, relative), v.Schema, v, false)
	if ref != nil {
		ref.Path = relative
	}
	return ref, err
}
func (p *diskPublisher) Terminal(locator Locator, receipt Receipt, cleanupFailed bool) (*Reference, error) {
	var err error
	// Private intent preserves primary failure if filesystem publication fails.
	// It never substitutes for a scan-bound retained receipt.
	p.record.Outcome = &receipt
	if err = writeRegistry(p.record, true); err != nil {
		return nil, err
	}
	unresolved := false
	if p.runtime != nil {
		resources, e := p.runtime.Resources(4096)
		if e != nil {
			return nil, e
		}
		unresolved = len(resources) > 0
		if unresolved {
			p.secrets, err = p.runtime.ForbiddenValues()
		} else {
			p.secrets, err = p.runtime.BeginFinalization()
		}
		if err != nil {
			return nil, err
		}
	}
	if _, err = harnesssession.ScanRetained(p.runRoot, p.secrets); err != nil {
		return nil, err
	}
	if !unresolved && p.runtime != nil {
		p.record.Released = true
		if err = writeRegistry(p.record, true); err != nil {
			return nil, err
		}
		if err = p.runtime.Close(); err != nil {
			return nil, err
		}
	} else if p.runtime != nil {
		if err = p.runtime.PreserveRecovery(); err != nil {
			return nil, err
		}
	}
	ref, err := publish(filepath.Join(p.runRoot, "ui-review/terminal.json"), receipt.Schema, receipt, false)
	if err != nil {
		return nil, err
	}
	ref.Path = "ui-review/terminal.json"
	locator.Receipt = ref
	locator.Updated = timestamp(time.Now())
	if err = p.Locator(locator); err != nil {
		return nil, err
	}
	if _, err = harnesssession.ScanRetained(p.runRoot, p.secrets); err != nil {
		return nil, err
	}
	if !unresolved {
		if err = removeRegistry(p.record); err != nil {
			return nil, err
		}
	}
	return ref, nil
}

var runPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,96}$`)

func newRunRoot(repository, results, runID string) (string, error) {
	if !runPattern.MatchString(runID) || runID == "." || runID == ".." || !filepath.IsAbs(results) || results == "/" {
		return "", failure("invalid_request")
	}
	if _, err := inputPath(repository, filepath.Join(results, runID, "session.json")); err != nil {
		return "", err
	}
	// Validate the ancestor chain before the first directory or writer is created.
	parent, err := harnesslocal.PrivateDirectory(results)
	if err != nil {
		return "", err
	}
	defer parent.Close()
	if err = unix.Mkdirat(int(parent.Fd()), runID, 0700); err != nil {
		return "", failure("session_mismatch")
	}
	root := filepath.Join(results, runID)
	dir, err := harnesslocal.PrivateDirectory(filepath.Join(root, "ui-review"))
	if err == nil {
		err = dir.Close()
	}
	return root, err
}
