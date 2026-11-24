//go:build linux

package harnessplan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// Files preparation reads besides the planning inputs.
const (
	toolchainPinsPath          = "tools/toolchain_pins.json"
	workspaceLayoutPath        = "tools/workspace_layout.json"
	redactionManifestPath      = "tools/harness_redaction_manifest.json"
	instrumentationPolicyPath  = "tools/harness_instrumentation_policy.json"
	defaultResultsDirectory    = ".cartulary/test-results"
	servicesModeInput          = "CARTULARY_TEST_SERVICES_MODE"
	servicesSessionFileInput   = "CARTULARY_TEST_SERVICES_SESSION_FILE"
	servicesBinaryEnvironment  = "CARTULARY_TEST_SERVICES_BIN"
	callerServiceStatePrefix   = "CARTULARY_TEST_SERVICES_"
	runIDEnvironment           = "CARTULARY_TEST_RUN_ID"
	hostAdmissionLease         = "CARTULARY_HOST_ADMISSION_LEASE"
	inheritedUnitCPUTokens     = "CARTULARY_UNIT_CPU_TOKENS"
	capacityOverrideInput      = "CARTULARY_HARNESS_CAPACITY_OVERRIDE"
	cacheModeInput             = "CARTULARY_HARNESS_CACHE_MODE"
	instrumentationModeInput   = "HARNESS_DIAGNOSTICS"
	machineCacheEnvironment    = "CARTULARY_MACHINE_CACHE_DIR"
	defaultServicesSessionFile = "test-services/session.json"
)

var runIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,96}$`)

// Prepare turns one graph invocation into a validated plan: it resolves the
// request against its public command, captures the host's capacity and the
// executable source, compiles the selection and binds the invocation's facts.
func Prepare(root string, request Request, now time.Time, pid int) (*Plan, error) {
	in, err := Load(root)
	if err != nil {
		return nil, err
	}
	resolved, err := in.resolve(request)
	if err != nil {
		return nil, err
	}
	inputs := resolved.inputs
	value := func(name, fallback string) string {
		if input, ok := inputs[name]; ok {
			return input.Value
		}
		return fallback
	}
	cacheMode := value(cacheModeInput, "normal")
	if request.CacheMode != "" {
		cacheMode = request.CacheMode
	}
	instrumentation := value(instrumentationModeInput, "off")
	if request.InstrumentationMode != "" {
		instrumentation = request.InstrumentationMode
	}
	mode, sessionFile, err := serviceSession(request, inputs)
	if err != nil {
		return nil, err
	}
	var override *CapacityOverride
	if path := value(capacityOverrideInput, ""); path != "" {
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		if override, err = ReadCapacityOverride(path); err != nil {
			return nil, err
		}
	}
	snapshot, err := Capture(DetectHost(root), in.Registry, override)
	if err != nil {
		return nil, err
	}
	budget, err := goWorkerBudget(snapshot.CPUTokens, request.Environment)
	if err != nil {
		return nil, err
	}
	lanes := in.Topology.BrowserLanePolicy.Default
	if text, ok := inputs["BROWSER_LANES"]; ok {
		if lanes, err = strconv.ParseInt(text.Value, 10, 64); err != nil {
			return nil, fmt.Errorf("%w: BROWSER_LANES: %v", ErrSelection, err)
		}
	}
	pins, err := os.ReadFile(filepath.Join(root, toolchainPinsPath))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrInput, toolchainPinsPath, err)
	}
	var toolchain struct {
		GoVersion string `json:"go_version"`
	}
	if err := json.Unmarshal(pins, &toolchain); err != nil || toolchain.GoVersion == "" {
		return nil, fmt.Errorf("%w: %s has no go_version", ErrInput, toolchainPinsPath)
	}
	options := Options{CPUTokens: budget, GoVersion: toolchain.GoVersion, GoShuffleSeed: value("GO_SHUFFLE_SEED", ""), AttachedServices: mode == "attach"}
	graph, err := Compile(in, resolved.selection, options)
	if err != nil {
		return nil, err
	}
	var projections map[string]Projection
	if request.Kind == "command" || request.Kind == "aggregate" {
		if projections, err = in.projections(graph, resolved.entry, options); err != nil {
			return nil, err
		}
	}
	bindings, err := launchBindings(root, in, request.Environment, mode, sessionFile)
	if err != nil {
		return nil, err
	}
	declared, environment, err := commandInputs(in, resolved, root, bindings)
	if err != nil {
		return nil, err
	}
	runID, err := runIdentity(request.Environment, now, pid)
	if err != nil {
		return nil, err
	}
	policy, err := os.ReadFile(filepath.Join(root, instrumentationPolicyPath))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrInput, instrumentationPolicyPath, err)
	}
	source, err := CaptureSource(root)
	if err != nil {
		return nil, err
	}
	toolchainSum := sha256.Sum256(pins)
	toolchainDigest := "sha256:" + hex.EncodeToString(toolchainSum[:])
	var cache []CacheFact
	var cacheInputs map[string][]int
	if cacheMode == "off" {
		reason := "mode_off"
		for _, unit := range graph.Sorted() {
			cache = append(cache, CacheFact{UnitID: unit.ID, Reason: &reason})
		}
	} else {
		environment := request.Environment
		cache, cacheInputs, err = CacheFacts(root, in, graph, source.Inputs, CacheContext{
			ToolchainDigest: toolchainDigest, Platform: runtime.GOOS + "/" + runtime.GOARCH,
			FIPSSelector: environment["GOFIPS140"], GoDebug: environment["GODEBUG"], GoToolchain: environment["GOTOOLCHAIN"],
		})
		if err != nil {
			return nil, err
		}
	}
	return Assemble(in, graph, Invocation{
		RunID: runID, CommandID: resolved.entry.CommandID, Target: request.Target,
		DeclaredInputs: declared, InputEnvironment: environment, BrowserLaneLimit: int(lanes),
		Snapshot: snapshot, GoWorkerBudget: budget,
		SourceInputs: source.Inputs, SourceCommit: source.Commit, SourceState: source.State,
		ToolchainDigest: toolchainDigest, CacheMode: cacheMode,
		InstrumentationMode: instrumentation, InstrumentationPolicy: policy,
		Host: DetectHostProfile(), Bindings: bindings, Cache: cache, CacheInputs: cacheInputs, StartedAt: now, Projections: projections,
	})
}

// commandInputs are the command's declared input values for the manifest
// (credential values redacted) and the environment projection every unit
// receives: all public inputs cleared, this command's runtime inputs set.
// Inputs that preparation turns into launch bindings (the service mode and
// session file) reach units only as bindings.
func commandInputs(in *Inputs, resolved command, root string, bindings map[string]string) (map[string]string, InputEnvironment, error) {
	var redaction struct {
		Replacement string `json:"replacement"`
	}
	data, err := readBounded(filepath.Join(root, redactionManifestPath))
	if err != nil || json.Unmarshal(data, &redaction) != nil || redaction.Replacement == "" {
		return nil, InputEnvironment{}, fmt.Errorf("%w: %s has no replacement marker", ErrInput, redactionManifestPath)
	}
	declared := map[string]string{}
	forward := map[string]string{}
	for name, input := range resolved.inputs {
		redacted := input.Definition.SummaryEmission == "redacted_value"
		declared[name] = input.Value
		if redacted {
			declared[name] = redaction.Replacement
		}
		if _, bound := bindings[name]; bound {
			continue
		}
		if forwarding := input.Definition.ChildForwarding; forwarding != "runtime_env" && forwarding != "argv_and_runtime_env" {
			continue
		}
		if redacted {
			return nil, InputEnvironment{}, fmt.Errorf("%w: credential input %s requires an owned private runtime binding", ErrSelection, name)
		}
		forward[name] = input.Value
		if input.Definition.Type == "exact_1_bool" {
			forward[name] = ""
			if input.Value == "1" || input.Value == "true" {
				forward[name] = "1"
			}
		}
	}
	return declared, InputEnvironment{Clear: in.TaskSurface.publicInputNames(), Forward: forward}, nil
}

// runIdentity is the run ID Make assigned, or a fresh one.
func runIdentity(environment map[string]string, now time.Time, pid int) (string, error) {
	runID, ok := environment[runIDEnvironment]
	if !ok || runID == "" {
		runID = now.UTC().Format("20060102T150405Z") + "-p" + strconv.Itoa(pid)
	}
	if !runIDPattern.MatchString(runID) || runID == "." || runID == ".." {
		return "", fmt.Errorf("%w: %s must be 1-96 characters of A-Z a-z 0-9 _ . -", ErrSelection, runIDEnvironment)
	}
	return runID, nil
}

// goWorkerBudget is the CPU budget Go units may claim: the snapshot's tokens,
// bounded by an enclosing admission lease's unit tokens when nested.
func goWorkerBudget(cpuTokens int64, environment map[string]string) (int64, error) {
	if _, nested := environment[hostAdmissionLease]; !nested {
		return cpuTokens, nil
	}
	inherited, err := strconv.ParseInt(environment[inheritedUnitCPUTokens], 10, 64)
	if environment[hostAdmissionLease] == "" || err != nil || inherited < 1 || !digitsPattern.MatchString(environment[inheritedUnitCPUTokens]) {
		return 0, fmt.Errorf("%w: invalid inherited Go CPU budget", ErrSelection)
	}
	return min(cpuTokens, inherited), nil
}

// serviceSession resolves owned or attached test services. Attachment is
// permitted only for test-slice SCOPE=services and browser-e2e SUITE
// functional or stateful (NLSpec service-session clause); the session file
// is accepted only when attaching.
func serviceSession(request Request, inputs map[string]resolvedInput) (string, string, error) {
	fail := func(format string, args ...any) (string, string, error) {
		return "", "", fmt.Errorf("%w: %s", ErrSelection, fmt.Sprintf(format, args...))
	}
	var reserved []string
	for name := range request.Environment {
		if strings.HasPrefix(name, callerServiceStatePrefix) && !slices.Contains([]string{servicesBinaryEnvironment, servicesModeInput, servicesSessionFileInput}, name) {
			reserved = append(reserved, name)
		}
	}
	if len(reserved) != 0 {
		slices.Sort(reserved)
		return fail("%s is reserved internal service state", strings.Join(reserved, ","))
	}
	mode := "owned"
	if input, ok := inputs[servicesModeInput]; ok && input.Value != "" {
		mode = input.Value
	}
	explicit := inputs[servicesSessionFileInput].Value
	switch mode {
	case "owned":
		if explicit != "" {
			return fail("%s is accepted only in attach mode", servicesSessionFileInput)
		}
		return mode, "", nil
	case "attach":
	default:
		return fail("%s must be owned or attach", servicesModeInput)
	}
	suite := inputs["SUITE"].Value
	allowed := request.Target == "test-slice" && inputs["SCOPE"].Value == "services" ||
		request.Target == "browser-e2e" && (suite == "functional" || suite == "stateful")
	if !allowed {
		return fail("%s is owned-only and rejects local service attachment", request.Target)
	}
	file := explicit
	if file == "" && request.Environment[machineCacheEnvironment] != "" {
		file = filepath.Join(request.Environment[machineCacheEnvironment], defaultServicesSessionFile)
	}
	if !filepath.IsAbs(file) {
		return fail("%s requires an absolute path or the machine-cache default", servicesSessionFileInput)
	}
	file = filepath.Clean(file)
	for current := file; current != filepath.Dir(current); current = filepath.Dir(current) {
		var stat unix.Stat_t
		if err := unix.Lstat(current, &stat); err == nil && stat.Mode&unix.S_IFMT == unix.S_IFLNK {
			return fail("%s must not resolve through a symbolic link", servicesSessionFileInput)
		}
	}
	return mode, file, nil
}

// launchBindings are the paths and modes the runtime binds when it starts a
// unit: runtime binaries, the services and harness binaries, the Node
// runtime and the service session.
func launchBindings(root string, in *Inputs, environment map[string]string, mode, sessionFile string) (map[string]string, error) {
	var layout struct {
		Directories []struct {
			ID   string `json:"id"`
			Path string `json:"path"`
		} `json:"directories"`
	}
	data, err := readBounded(filepath.Join(root, workspaceLayoutPath))
	if err != nil || json.Unmarshal(data, &layout) != nil {
		return nil, fmt.Errorf("%w: cannot read %s", ErrInput, workspaceLayoutPath)
	}
	directory := map[string]string{}
	for _, entry := range layout.Directories {
		directory[entry.ID] = entry.Path
	}
	if directory["toolbin"] == "" || directory["node_runtime"] == "" {
		return nil, fmt.Errorf("%w: %s lacks toolbin or node_runtime", ErrInput, workspaceLayoutPath)
	}
	absolute := func(path string) string {
		if filepath.IsAbs(path) {
			return filepath.Clean(path)
		}
		return filepath.Join(root, path)
	}
	first := func(names []string, fallback string) string {
		for _, name := range names {
			if environment[name] != "" {
				return environment[name]
			}
		}
		return fallback
	}
	bindings := map[string]string{}
	for _, binary := range in.Topology.RuntimeBinaries {
		output := absolute(first([]string{binary.OutputMakeVariable}, binary.DefaultOutputPath))
		bindings[binary.ConsumerEnv], bindings[binary.OutputMakeVariable] = output, output
	}
	bindings[servicesBinaryEnvironment] = absolute(first([]string{servicesBinaryEnvironment, "TEST_SERVICES_BIN"}, directory["toolbin"]+"/cartulary-test-services"))
	bindings["CARTULARY_HARNESS_RUNTIME_BIN"] = absolute(first([]string{"CARTULARY_HARNESS_RUNTIME_BIN", "HARNESS_RUNTIME_BIN"}, directory["toolbin"]+"/cartulary-harness"))
	bindings["NODE_RUNTIME_DIR"] = absolute(first([]string{"NODE_RUNTIME_DIR"}, directory["node_runtime"]))
	node, selected := environment["NODE_BIN"]
	if !selected {
		node = filepath.Join(bindings["NODE_RUNTIME_DIR"], "bin/node")
	}
	if strings.TrimSpace(node) == "" {
		return nil, fmt.Errorf("%w: selected NODE_BIN is missing", ErrSelection)
	}
	node = absolute(node)
	if info, err := os.Stat(node); err != nil || !info.Mode().IsRegular() || unix.Access(node, unix.X_OK) != nil {
		return nil, fmt.Errorf("%w: selected NODE_BIN is not an executable file", ErrSelection)
	}
	bindings["NODE_BIN"] = node
	bindings["PNPM"] = first([]string{"PNPM"}, filepath.Join(bindings["NODE_RUNTIME_DIR"], "bin/pnpm"))
	bindings[servicesModeInput] = mode
	if sessionFile != "" {
		bindings[servicesSessionFileInput] = sessionFile
	}
	return bindings, nil
}
