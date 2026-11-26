//go:build linux

package artifactstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/runcontext"
)

// RegistryPath is the authored producer registry, relative to the repository
// root. Each declaration names a producer, the binding its output is
// published at, and what it builds; the store decides whether to build.
const RegistryPath = "tools/harness_producers.json"

const registrySchemaID = "cartulary.harness_producers.v1"

var (
	producerIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	bindingPattern    = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

// Declaration is one registered producer. It publishes at its default path,
// a file under the layout directory Directory. Binding names the variable
// consumers read that path from; a caller that sets it to another path
// provides that executable itself, and the producer neither builds nor
// publishes over it.
type Declaration struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Binding   string `json:"binding"`
	Directory string `json:"directory"`
	// Label and Target name the run step a build runs as.
	Label  string `json:"label,omitempty"`
	Target string `json:"target,omitempty"`
	// go_binary: the package, its build tags and the binary's file name.
	Package string `json:"package,omitempty"`
	Tags    string `json:"tags,omitempty"`
	Name    string `json:"name,omitempty"`
	// go_tool: the pinned tool's key in the toolchain pins, and the module
	// directory whose go.mod selects its dependencies.
	Tool      string `json:"tool,omitempty"`
	ModuleDir string `json:"module_dir,omitempty"`
	// shellcheck: the committed SHA-256 of each release archive by name.
	Archives map[string]string `json:"archives,omitempty"`
	// web_bundle: the Vite mode and the repository paths the build reads. A
	// bundle publishes no view: its root is copied into the run's frontend
	// artifact, which publishes the conventional output.
	Mode   string   `json:"mode,omitempty"`
	Inputs []string `json:"inputs,omitempty"`
	// embedded_assets: the frontend artifact target whose bundle is packaged,
	// and the authored client support source.
	Source        string `json:"source,omitempty"`
	ClientSupport string `json:"client_support,omitempty"`
}

// Registry is the producer registry.
type Registry struct {
	SchemaID  string        `json:"schema_id"`
	Producers []Declaration `json:"producers"`
}

// ReadRegistry reads and validates the registry under root.
func ReadRegistry(root string) (Registry, error) {
	var r Registry
	data, err := os.ReadFile(filepath.Join(root, RegistryPath))
	if err != nil {
		return r, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&r); err != nil {
		return r, fmt.Errorf("%s: %w", RegistryPath, err)
	}
	if r.SchemaID != registrySchemaID {
		return r, fmt.Errorf("%s: schema %q, want %s", RegistryPath, r.SchemaID, registrySchemaID)
	}
	ids, bindings := map[string]bool{}, map[string]bool{}
	for _, d := range r.Producers {
		if err := d.validate(); err != nil {
			return r, fmt.Errorf("%s: producer %q: %w", RegistryPath, d.ID, err)
		}
		if ids[d.ID] || d.Binding != "" && bindings[d.Binding] {
			return r, fmt.Errorf("%s: producer %q repeats an id or binding", RegistryPath, d.ID)
		}
		ids[d.ID], bindings[d.Binding] = true, true
	}
	return r, nil
}

func (d Declaration) validate() error {
	if !producerIDPattern.MatchString(d.ID) {
		return errors.New("needs an id")
	}
	// A web bundle publishes no view: its root is copied into the run's
	// frontend artifact. Every other kind publishes at a binding.
	if d.Kind == "web_bundle" {
		if d.Binding != "" || d.Directory != "" || d.Label != "" {
			return errors.New("a web bundle publishes no view and runs no step of its own")
		}
	} else if !bindingPattern.MatchString(d.Binding) || d.Directory == "" {
		return errors.New("needs a binding and a layout directory")
	} else if d.Target != "" && d.Label == "" {
		return errors.New("a step target needs a label")
	}
	allowed, known := map[string][]string{
		"go_binary":       {"package", "tags", "name"},
		"go_tool":         {"tool", "module_dir"},
		"shellcheck":      {"archives"},
		"web_bundle":      {"mode", "inputs"},
		"embedded_assets": {"source", "client_support"},
	}[d.Kind]
	if !known {
		return fmt.Errorf("unknown kind %q", d.Kind)
	}
	for field, present := range map[string]bool{
		"package": d.Package != "", "tags": d.Tags != "", "name": d.Name != "", "tool": d.Tool != "", "module_dir": d.ModuleDir != "",
		"archives": d.Archives != nil, "mode": d.Mode != "", "inputs": d.Inputs != nil, "source": d.Source != "", "client_support": d.ClientSupport != "",
	} {
		if present && !slices.Contains(allowed, field) {
			return fmt.Errorf("a %s declares no %s", d.Kind, field)
		}
	}
	switch {
	case d.Kind == "go_binary" && (!strings.HasPrefix(d.Package, "./") || d.Name == "" || strings.Contains(d.Name, "/")):
		return errors.New("a go_binary declares a ./package and a file name")
	case d.Kind == "go_tool" && d.Tool == "":
		return errors.New("a go_tool declares a pinned tool")
	case d.Kind == "shellcheck" && len(d.Archives) == 0:
		return errors.New("shellcheck declares its archive digests")
	case d.Kind == "web_bundle" && (d.Mode != "production" && d.Mode != "measurement" || len(d.Inputs) == 0 || d.Target == ""):
		return errors.New("a web_bundle declares its mode, its inputs and its frontend artifact target")
	case d.Kind == "embedded_assets" && (d.Source == "" || d.ClientSupport == ""):
		return errors.New("embedded_assets declares its frontend artifact source and client support source")
	}
	return nil
}

// Lookup returns the declaration named id.
func (r Registry) Lookup(id string) (Declaration, bool) {
	index := slices.IndexFunc(r.Producers, func(d Declaration) bool { return d.ID == id })
	if index < 0 {
		return Declaration{}, false
	}
	return r.Producers[index], true
}

// Workspace is what producers read from a checkout and an environment: the
// layout's directories and the toolchain pins.
type Workspace struct {
	Root   string
	Env    []string
	layout map[string]string
	pins   struct {
		GoToolchain       string            `json:"go_toolchain"`
		Tools             map[string]string `json:"tools"`
		ShellCheckVersion string            `json:"shellcheck_version"`
		PNPMVersion       string            `json:"pnpm_version"`
	}
}

// OpenWorkspace reads the layout and pins of the checkout at root.
func OpenWorkspace(root string, env []string) (*Workspace, error) {
	w := &Workspace{Root: root, Env: slices.Clone(env), layout: map[string]string{}}
	var layout struct {
		Directories []struct {
			ID   string `json:"id"`
			Path string `json:"path"`
		} `json:"directories"`
	}
	for name, target := range map[string]any{"tools/workspace_layout.json": &layout, "tools/toolchain_pins.json": &w.pins} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, target); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	for _, directory := range layout.Directories {
		if directory.Path != "" && !filepath.IsAbs(directory.Path) && !strings.Contains(directory.Path, "*") {
			w.layout[directory.ID] = filepath.Join(root, filepath.FromSlash(directory.Path))
		}
	}
	return w, nil
}

// Getenv returns the environment's last value for name.
func (w *Workspace) Getenv(name string) string { return environmentValue(w.Env, name) }

// Directory returns the absolute path of a layout directory.
func (w *Workspace) Directory(id string) (string, error) {
	directory, ok := w.layout[id]
	if !ok {
		return "", fmt.Errorf("the workspace layout declares no %s directory", id)
	}
	return directory, nil
}

// DefaultName is the file name d publishes at when its binding is unset.
func (w *Workspace) DefaultName(d Declaration) (string, error) {
	switch d.Kind {
	case "go_binary":
		return d.Name, nil
	case "go_tool":
		tool, err := w.goTool(d)
		if err != nil {
			return "", err
		}
		return tool.Name() + "-" + tool.Version(), nil
	case "embedded_assets":
		return EmbeddedAssetNames[0], nil
	case "web_bundle":
		return "", fmt.Errorf("producer %s publishes no view", d.ID)
	default:
		if w.pins.ShellCheckVersion == "" {
			return "", errors.New("the toolchain pins name no shellcheck_version")
		}
		return "shellcheck-v" + w.pins.ShellCheckVersion, nil
	}
}

// View returns the path d publishes its output at: its default file under its
// layout directory.
func (w *Workspace) View(d Declaration) (string, error) {
	directory, err := w.Directory(d.Directory)
	if err != nil {
		return "", err
	}
	name, err := w.DefaultName(d)
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, name), nil
}

// Provided returns the file the caller provides for d: its binding's value
// when the environment sets it to a path other than d's view. A provided file
// is never built or published over and must exist: an executable for a
// program's producer.
func (w *Workspace) Provided(d Declaration) (string, bool, error) {
	value := ""
	if d.Binding != "" {
		value = w.Getenv(d.Binding)
	}
	if value == "" {
		return "", false, nil
	}
	if !filepath.IsAbs(value) {
		return "", false, fmt.Errorf("%s must be an absolute path: %s", d.Binding, value)
	}
	view, err := w.View(d)
	if err != nil || samePath(value, view) {
		return "", false, err
	}
	info, err := os.Stat(value)
	if err != nil || !info.Mode().IsRegular() || d.Kind != "embedded_assets" && info.Mode().Perm()&0o111 == 0 {
		return "", false, fmt.Errorf("%s names %s, which the caller provides, but it is not a usable file", d.Binding, value)
	}
	return value, true, nil
}

// samePath reports whether two absolute paths name the same location once
// symbolic links in their existing ancestors are resolved, so a checkout
// reached through a link matches its own default paths.
func samePath(left, right string) bool {
	return left == right || resolvedPath(left) == resolvedPath(right)
}

func resolvedPath(name string) string {
	name = filepath.Clean(name)
	rest := ""
	for {
		if resolved, err := filepath.EvalSymlinks(name); err == nil {
			return filepath.Join(resolved, rest)
		}
		parent := filepath.Dir(name)
		if parent == name {
			return filepath.Join(name, rest)
		}
		rest = filepath.Join(filepath.Base(name), rest)
		name = parent
	}
}

func (w *Workspace) goTool(d Declaration) (GoTool, error) {
	module, ok := w.pins.Tools[d.Tool]
	if !ok {
		return GoTool{}, fmt.Errorf("the toolchain pins name no tool %s", d.Tool)
	}
	return GoTool{Root: w.Root, ModuleDir: d.ModuleDir, Module: module}, nil
}

// goEnvironment is the environment Go producers run with: the workspace's
// environment with the build, module and temporary caches it names.
func (w *Workspace) goEnvironment() (string, []string, error) {
	env := slices.Clone(w.Env)
	for name, variable := range map[string]string{"GOCACHE": "GO_CACHE_DIR", "GOMODCACHE": "GO_MOD_CACHE_DIR", "GOTMPDIR": "GO_TMP_DIR"} {
		directory := w.Getenv(variable)
		if directory == "" {
			return "", nil, fmt.Errorf("%s is required", variable)
		}
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return "", nil, err
		}
		env = append(env, name+"="+directory)
	}
	goExecutable := w.Getenv("GO")
	if goExecutable == "" {
		var err error
		if goExecutable, err = exec.LookPath("go"); err != nil {
			return "", nil, err
		}
	}
	return goExecutable, env, nil
}

// step runs a build as d's run step when the workspace has a step runner.
func (w *Workspace) step(d Declaration, env []string) ([]string, func([]string) []string) {
	script := w.Getenv("RUN_STEP_SCRIPT")
	if d.Label == "" || script == "" {
		return env, nil
	}
	env = append(env, "CARTULARY_SUPPRESS_CHILD_SUCCESS=1")
	if d.Target != "" && w.Getenv("CARTULARY_TEST_TARGET") == "" {
		env = append(env, runcontext.Entries(runcontext.Context{Target: d.Target}.Environment())...)
	}
	return env, func(command []string) []string { return append([]string{"bash", script, d.Label, "--"}, command...) }
}

// Plan returns d's producer and the views it publishes.
func (w *Workspace) Plan(d Declaration) (Producer, map[string]string, error) {
	if d.Kind == "web_bundle" {
		node, runtime, pnpm := w.Getenv("NODE_BIN"), w.Getenv("NODE_RUNTIME_DIR"), w.Getenv("PNPM")
		if node == "" || runtime == "" || pnpm == "" {
			return nil, nil, errors.New("NODE_BIN, NODE_RUNTIME_DIR and PNPM are required")
		}
		return WebBundle{Root: w.Root, Inputs: d.Inputs, Mode: d.Mode, Node: node, PNPM: pnpm, NodeRuntime: runtime, PNPMVersion: w.pins.PNPMVersion, Env: w.Env}, map[string]string{}, nil
	}
	view, err := w.View(d)
	if err != nil {
		return nil, nil, err
	}
	switch d.Kind {
	case "embedded_assets":
		source, err := w.frontendArtifact(d.Source)
		if err != nil {
			return nil, nil, err
		}
		goExecutable, env, err := w.goEnvironment()
		if err != nil {
			return nil, nil, err
		}
		views := map[string]string{}
		for _, name := range EmbeddedAssetNames {
			views[name] = filepath.Join(filepath.Dir(view), name)
		}
		return EmbeddedAssets{Root: w.Root, Source: source, ClientSupport: d.ClientSupport, Go: goExecutable, Env: env}, views, nil
	case "go_binary":
		goExecutable, env, err := w.goEnvironment()
		if err != nil {
			return nil, nil, err
		}
		node := w.Getenv("NODE_BIN")
		if node == "" {
			return nil, nil, errors.New("NODE_BIN is required")
		}
		env, wrap := w.step(d, env)
		return GoBinary{Module: w.Root, Package: d.Package, Tags: d.Tags, Name: d.Name, Go: goExecutable, Node: node, Env: env, Wrap: wrap},
			map[string]string{d.Name: view, d.Name + ".crypto.json": view + ".crypto.json"}, nil
	case "go_tool":
		tool, err := w.goTool(d)
		if err != nil {
			return nil, nil, err
		}
		if tool.Go, tool.Env, err = w.goEnvironment(); err != nil {
			return nil, nil, err
		}
		tool.Env, tool.Wrap = w.step(d, tool.Env)
		return tool, map[string]string{tool.Name(): view}, nil
	default:
		archive, err := ShellCheckArchive(w.pins.ShellCheckVersion)
		if err != nil {
			return nil, nil, err
		}
		digest, ok := d.Archives[archive]
		if !ok {
			return nil, nil, fmt.Errorf("producer %s has no committed digest for %s", d.ID, archive)
		}
		archives, err := w.Directory("shellcheck_archives")
		if err != nil {
			return nil, nil, err
		}
		return ShellCheck{Root: w.Root, Version: w.pins.ShellCheckVersion, Archive: archive, SHA256: digest, Archives: archives},
			map[string]string{"shellcheck": view}, nil
	}
}

// IsGo reports whether d builds with the Go toolchain.
func (d Declaration) IsGo() bool {
	return d.Kind == "go_binary" || d.Kind == "go_tool" || d.Kind == "embedded_assets"
}

// frontendArtifact resolves the run's sealed frontend artifact of target
// through its owner, which admits it by its receipt. Packaging reads that
// artifact and never rebuilds the bundle.
func (w *Workspace) frontendArtifact(target string) (string, error) {
	if w.Getenv("CARTULARY_HARNESS_SUITE_RUNTIME_ROOT") == "" {
		return "", fmt.Errorf("the embedded assets package the run's %s artifact, so they run in a graph", target)
	}
	node := w.Getenv("NODE_BIN")
	if node == "" {
		return "", errors.New("NODE_BIN is required")
	}
	command := exec.Command(node, "tools/harness/readiness/frontend-artifact.mjs", "resolve", target)
	command.Dir, command.Env, command.Stderr = w.Root, w.Env, os.Stderr
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("resolve the %s artifact: %w", target, err)
	}
	directory := strings.TrimSpace(string(output))
	if !filepath.IsAbs(directory) {
		return "", fmt.Errorf("the %s artifact resolved to %q", target, directory)
	}
	return directory, nil
}
