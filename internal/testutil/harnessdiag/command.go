package harnessdiag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var errConfiguration = errors.New("invalid diagnostic configuration")

type inputPolicy struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Min      uint64 `json:"min"`
	Max      uint64 `json:"max"`
}

func makeInputs() (Inputs, bool, error) {
	var policy struct {
		Inputs  []inputPolicy `json:"inputs"`
		Globals []inputPolicy `json:"global_inputs"`
	}
	if json.Unmarshal([]byte(statusInputContractJSON), &policy) != nil {
		return Inputs{}, false, errUsage
	}
	declared := map[string]inputPolicy{}
	for _, v := range append(policy.Inputs, policy.Globals...) {
		declared[v.Name] = v
	}
	seen := map[string]bool{}
	sources := map[string]string{}
	for _, token := range strings.Fields(os.Getenv("CARTULARY_MAKE_INPUT_SOURCES")) {
		parts := strings.Split(token, "=")
		if len(parts) != 2 || seen[parts[0]] {
			return Inputs{}, false, errUsage
		}
		seen[parts[0]] = true
		sources[parts[0]] = parts[1]
		if parts[1] != "cli" && parts[1] != "env" && parts[1] != "file" && parts[1] != "unset" {
			return Inputs{}, false, errUsage
		}
		if parts[1] == "cli" {
			if _, ok := declared[parts[0]]; !ok {
				return Inputs{}, false, errUsage
			}
		}
	}
	if err := validateExplicitCacheInputs(policy.Globals, sources); err != nil {
		return Inputs{}, false, err
	}
	values := map[string]string{}
	for _, p := range policy.Inputs {
		raw := strings.TrimSpace(os.Getenv(p.Name))
		if raw == "" {
			if p.Required {
				return Inputs{}, false, errUsage
			}
			continue
		}
		switch p.Type {
		case "result_selector":
		case "run_id":
			if len(raw) > 96 || !regexpRunID(raw) {
				return Inputs{}, false, errUsage
			}
		case "target_name":
			if !validToken(raw, 255, true) {
				return Inputs{}, false, errUsage
			}
		case "exact_1_bool":
			if raw != "1" {
				return Inputs{}, false, errUsage
			}
		case "positive_integer":
			v, err := numberInput(raw, p.Max)
			if err != nil || v == nil || *v < float64(p.Min) {
				return Inputs{}, false, errUsage
			}
		default:
			return Inputs{}, false, errUsage
		}
		values[p.Name] = raw
	}
	mode := strings.TrimSpace(os.Getenv("CARTULARY_OUTPUT_MODE"))
	if mode != "" && mode != "summary" && mode != "verbose" && mode != "ci" && mode != "quiet" {
		return Inputs{}, false, errUsage
	}
	return Inputs{values["RESULTS_DIR"], values["RUN_ID"], values["TARGET"], values["AFTER_REVISION"], values["WAIT_SECONDS"]}, values["JSON"] == "1", nil
}

// These global inputs are unused by observation, but explicitly supplied values
// still obey their authored path contract. Validation never creates cache state.
func validateExplicitCacheInputs(globals []inputPolicy, sources map[string]string) error {
	anySupplied := false
	for _, p := range globals {
		if sources[p.Name] != "" && sources[p.Name] != "unset" {
			anySupplied = true
		}
	}
	if !anySupplied {
		return nil
	}
	root := strings.TrimSpace(os.Getenv("CARTULARY_MACHINE_CACHE_DIR"))
	if sources["CARTULARY_MACHINE_CACHE_DIR"] == "unset" || sources["CARTULARY_MACHINE_CACHE_DIR"] == "" {
		if x := strings.TrimSpace(os.Getenv("XDG_CACHE_HOME")); filepath.IsAbs(x) {
			root = filepath.Join(x, "cartulary")
		} else if h := strings.TrimSpace(os.Getenv("HOME")); filepath.IsAbs(h) {
			root = filepath.Join(h, ".cache/cartulary")
		} else {
			return errConfiguration
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return errConfiguration
	}
	repository, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return errConfiguration
	}
	paths := map[string]string{}
	for _, p := range globals {
		raw := strings.TrimSpace(os.Getenv(p.Name))
		if sources[p.Name] == "unset" || sources[p.Name] == "" {
			switch p.Name {
			case "CARTULARY_MACHINE_CACHE_DIR":
				raw = root
			case "GO_CACHE_DIR":
				raw = filepath.Join(root, "go/build")
			case "GO_MOD_CACHE_DIR":
				raw = filepath.Join(root, "go/mod")
			case "GO_TMP_DIR":
				raw = filepath.Join(root, "go/tmp")
			default:
				return errConfiguration
			}
		}
		if !filepath.IsAbs(raw) || strings.ContainsRune(raw, 0) {
			return errConfiguration
		}
		existing := filepath.Clean(raw)
		suffix := []string{}
		for {
			_, err := os.Stat(existing)
			if err == nil {
				break
			}
			if !os.IsNotExist(err) {
				return errConfiguration
			}
			parent := filepath.Dir(existing)
			if parent == existing {
				return errConfiguration
			}
			suffix = append(suffix, filepath.Base(existing))
			existing = parent
		}
		canonical, err := filepath.EvalSymlinks(existing)
		if err != nil {
			return errConfiguration
		}
		for i := len(suffix) - 1; i >= 0; i-- {
			canonical = filepath.Join(canonical, suffix[i])
		}
		if canonical == repository || strings.HasPrefix(canonical, repository+string(os.PathSeparator)) {
			return errConfiguration
		}
		paths[p.Name] = canonical
	}
	names := []string{"GO_CACHE_DIR", "GO_MOD_CACHE_DIR", "GO_TMP_DIR"}
	for i, left := range names {
		for _, right := range names[i+1:] {
			a, b := paths[left], paths[right]
			if a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/") {
				return errConfiguration
			}
		}
	}
	return nil
}
func regexpRunID(value string) bool {
	if value == "." || value == ".." {
		return false
	}
	for _, c := range value {
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '.' || c == '-') {
			return false
		}
	}
	return true
}
func arguments(args []string) (Inputs, bool, error) {
	if len(args) == 1 && args[0] == "--make-inputs" {
		return makeInputs()
	}
	in := Inputs{}
	asJSON := false
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if seen[arg] {
			return in, asJSON, errUsage
		}
		seen[arg] = true
		if arg == "--json" {
			asJSON = true
			continue
		}
		if i+1 >= len(args) {
			return in, asJSON, errUsage
		}
		i++
		switch arg {
		case "--results-dir":
			in.ResultsDir = args[i]
		case "--run-id":
			in.RunID = args[i]
		case "--target":
			in.Target = args[i]
		case "--after-revision":
			in.AfterRevision = args[i]
		case "--wait-seconds":
			in.WaitSeconds = args[i]
		default:
			return in, asJSON, errUsage
		}
	}
	return in, asJSON, nil
}

// Main is the installed read-only diagnostic entrypoint. It never installs,
// starts observers, acquires leases, or modifies the retained invocation.
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 2 && args[0] == "--terminal-receipt" {
		if err := PublishTerminalReceipts(args[1]); err != nil {
			return 1
		}
		return 0
	}
	in, asJSON, err := arguments(args)
	if err != nil {
		reason := "usage_error"
		if errors.Is(err, errConfiguration) {
			reason = "configuration_error"
		}
		fmt.Fprintf(stderr, "[CONFIG] target=test-run-status failure_class=config failure_reason=%s invalid inputs\n", reason)
		return 2
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	done := make(chan struct{})
	defer close(done)
	signalCode := make(chan int, 1)
	go func() {
		select {
		case sig := <-signals:
			code := 130
			if sig == syscall.SIGTERM {
				code = 143
			}
			signalCode <- code
			cancel()
		case <-done:
		}
	}()
	result := Observe(ctx, in)
	select {
	case code := <-signalCode:
		result = failObservation(result, context.Canceled, code)
	default:
	}
	var output []byte
	if asJSON {
		output, err = json.Marshal(result)
		output = append(output, '\n')
	} else {
		output = []byte(Render(result))
	}
	if err != nil {
		fmt.Fprintln(stderr, "[ARTIFACT] invalid observation projection")
		return 11
	}
	if _, err = stdout.Write(output); err != nil {
		return 11
	}
	code, _ := strconv.Atoi(fmt.Sprint(result["operation_exit_code"]))
	return code
}
