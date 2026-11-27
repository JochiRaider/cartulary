//go:build linux

package artifactstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"
)

// The Go binary key memo (P6-5). A Go binary's key needs its import closure
// from `go list -deps`, the effective settings from `go env` and the Node
// version: three subprocesses, about 0.2 s, on every check. The memo keeps
// what they returned together with everything they read: the executables,
// the build environment, the Go configuration and workspace files, the
// listing of every closure directory and the digest of every file in them.
// While all of that is unchanged the commands return the same, so the key is
// computed from the memo; otherwise they run again. The key is the same
// either way.

const goFactsSchemaID = "cartulary.go_binary_facts_memo.v1"

// goEnvironmentNames are the variables that change what `go list` and
// `go env` report, or which C compiler and configuration file they find.
var goEnvironmentNames = []string{
	"CC", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_ENABLED", "CGO_LDFLAGS", "CXX",
	"GO111MODULE", "GO386", "GOAMD64", "GOARCH", "GOARM", "GOARM64", "GODEBUG", "GOENV", "GOEXPERIMENT", "GOFIPS140",
	"GOFLAGS", "GOMIPS", "GOMIPS64", "GOOS", "GOPATH", "GOPPC64", "GORISCV64", "GOROOT", "GOTOOLCHAIN", "GOWASM", "GOWORK",
	"HOME", "PATH", "XDG_CONFIG_HOME",
}

// goFacts is one Go binary's memo.
type goFacts struct {
	SchemaID string `json:"schema_id"`
	// Environment digests the executables' identities, the build
	// environment and the package selection.
	Environment string `json:"environment"`
	// Directories lists every directory whose contents the closure depends
	// on, and Files the digest of each file the commands read, "" when it is
	// absent.
	Directories map[string][]string `json:"directories"`
	Files       map[string]string   `json:"files"`
	Closure     []string            `json:"closure"`
	Settings    map[string]string   `json:"settings"`
	NodeVersion string              `json:"node_version"`
}

// MemoKey is Key, computed from the memo in directory when it still holds.
func (b GoBinary) MemoKey(directory string) (string, error) {
	environment, err := b.environmentIdentity()
	if err != nil {
		return b.Key()
	}
	// Each environment a binary is checked in, such as a launcher's and a
	// graph unit's search path, keeps its own memo.
	file := filepath.Join(directory, "go-binary", environment+".json")
	digests := map[string]string{}
	if facts, ok := readGoFacts(file); ok && facts.Environment == environment && facts.current(digests) {
		now := time.Now()
		_ = os.Chtimes(file, now, now)
		return b.keyFrom(facts.Closure, facts.Settings, facts.NodeVersion, digests)
	}
	if err := b.ready(); err != nil {
		return "", err
	}
	started := time.Now()
	facts, err := b.learnFacts(environment)
	if err != nil {
		return "", err
	}
	key, err := b.keyFrom(facts.Closure, facts.Settings, facts.NodeVersion, nil)
	if err != nil {
		return "", err
	}
	// A file or directory changed near the commands may have been read in
	// another state than the one recorded, so its memo is not kept.
	if facts.settledBefore(started.Add(-memoSettleMargin)) {
		if data, err := json.Marshal(facts); err == nil && os.MkdirAll(filepath.Dir(file), 0o700) == nil {
			_ = writeAtomic(file, data)
		}
	}
	return key, nil
}

// memoSettleMargin covers the kernel's coarse timestamp clock: a change after
// a memo is recorded always moves a change time past it.
var memoSettleMargin = 2 * time.Second

// settledBefore reports whether every recorded directory and present file
// last changed before limit.
func (f goFacts) settledBefore(limit time.Time) bool {
	names := slices.Collect(maps.Keys(f.Directories))
	for name, digest := range f.Files {
		if digest != "" {
			names = append(names, name)
		}
	}
	for _, name := range names {
		stat, err := statIdentity(name)
		if err != nil || stat[3] >= limit.UnixNano() || stat[4] >= limit.UnixNano() {
			return false
		}
	}
	return true
}

// environmentIdentity digests what the commands run with: the resolved Go
// and Node executables' stat identities, the build environment and the
// package selection. It also names the memo.
func (b GoBinary) environmentIdentity() (string, error) {
	identity := map[string]any{"schema_id": goFactsSchemaID, "module": b.Module, "package": b.Package, "tags": b.Tags}
	for name, program := range map[string]string{"go": b.Go, "node": b.Node} {
		resolved, err := exec.LookPath(program)
		if err != nil {
			return "", err
		}
		stat, err := statIdentity(resolved)
		if err != nil {
			return "", err
		}
		identity[name] = map[string]any{"path": resolved, "stat": stat}
	}
	values := map[string]string{}
	for _, entry := range b.Env {
		if name, value, ok := strings.Cut(entry, "="); ok && slices.Contains(goEnvironmentNames, name) {
			values[name] = value
		}
	}
	identity["environment"] = values
	data, err := json.Marshal(identity)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func statIdentity(name string) ([]int64, error) {
	info, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, errors.New("no stat identity")
	}
	return []int64{int64(stat.Dev), int64(stat.Ino), stat.Size, stat.Mtim.Nano(), stat.Ctim.Nano()}, nil
}

func readGoFacts(file string) (goFacts, bool) {
	data, err := os.ReadFile(file)
	if err != nil {
		return goFacts{}, false
	}
	var facts goFacts
	if json.Unmarshal(data, &facts) != nil || facts.SchemaID != goFactsSchemaID || facts.Directories == nil || facts.Files == nil {
		return goFacts{}, false
	}
	return facts, true
}

// current reports whether every directory listing and file digest is as the
// memo records it. Digests it computes for present files are kept in digests,
// so the key does not read those files again.
func (f goFacts) current(digests map[string]string) bool {
	for directory, names := range f.Directories {
		if listed, err := listing(directory); err != nil || !slices.Equal(listed, names) {
			return false
		}
	}
	for name, want := range f.Files {
		got, err := fileDigest(name)
		if errors.Is(err, fs.ErrNotExist) {
			got = ""
		} else if err != nil {
			return false
		}
		if got != want {
			return false
		}
		if got != "" {
			digests[name] = got
		}
	}
	return true
}

// listing names a directory's entries other than test files, which a build
// without tests never reads.
func listing(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), "_test.go") {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

// learnFacts runs the commands and records what they read.
func (b GoBinary) learnFacts(environment string) (goFacts, error) {
	facts := goFacts{SchemaID: goFactsSchemaID, Environment: environment, Directories: map[string][]string{}, Files: map[string]string{}}
	module, err := filepath.Abs(b.Module)
	if err != nil {
		return facts, err
	}
	detail, err := goClosureDetail(b.Go, b.Module, b.Package, b.Tags, b.Env)
	if err != nil {
		return facts, err
	}
	facts.Closure = detail.files
	if facts.Settings, err = effectiveGoSettings(b.Go, b.Module, b.Env); err != nil {
		return facts, err
	}
	node, err := b.run(b.Node, []string{"--version"}, b.Env)
	if err != nil {
		return facts, err
	}
	facts.NodeVersion = strings.TrimSpace(string(node))
	locations, err := goConfiguration(b.Go, b.Module, b.Env)
	if err != nil {
		return facts, err
	}
	// The directories the closure reads: each package directory, and every
	// directory from a package down to an embedded file, so a new file that
	// a package or an embedding pattern would pick up changes a listing.
	directories := map[string]bool{}
	for _, directory := range detail.directories {
		directories[directory] = true
	}
	for _, embedded := range detail.embedded {
		for directory := filepath.Dir(embedded.file); directory != embedded.pkg && strings.HasPrefix(directory, embedded.pkg+string(filepath.Separator)); directory = filepath.Dir(directory) {
			directories[directory] = true
		}
		facts.Files[embedded.file] = ""
	}
	for directory := range directories {
		names, err := listing(directory)
		if err != nil {
			return facts, err
		}
		facts.Directories[directory] = names
		for _, name := range names {
			if info, err := os.Lstat(filepath.Join(directory, name)); err == nil && info.Mode().IsRegular() {
				facts.Files[filepath.Join(directory, name)] = ""
			}
		}
	}
	// The module and workspace files, and the Go configuration files.
	for _, name := range []string{"go.mod", "go.sum", filepath.Join("vendor", "modules.txt")} {
		facts.Files[filepath.Join(module, name)] = ""
	}
	for directory := module; ; directory = filepath.Dir(directory) {
		facts.Files[filepath.Join(directory, "go.work")] = ""
		if directory == filepath.Dir(directory) {
			break
		}
	}
	for _, name := range locations {
		facts.Files[name] = ""
	}
	for name := range facts.Files {
		digest, err := fileDigest(name)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return facts, err
		}
		facts.Files[name] = digest
	}
	return facts, nil
}

// goConfiguration returns the configuration files `go env` reads beyond the
// environment: the toolchain's go.env and the user's GOENV file.
func goConfiguration(goExecutable, dir string, env []string) ([]string, error) {
	command := exec.Command(goExecutable, "env", "-json", "GOROOT", "GOENV")
	command.Dir, command.Env = dir, env
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("go env: %w", err)
	}
	var values map[string]string
	if err := json.Unmarshal(output, &values); err != nil {
		return nil, err
	}
	files := []string{}
	if values["GOROOT"] != "" {
		files = append(files, filepath.Join(values["GOROOT"], "go.env"))
	}
	if values["GOENV"] != "" && values["GOENV"] != "off" {
		files = append(files, values["GOENV"])
	}
	return files, nil
}

// closureDetail is a package's import closure as goClosure lists it, with
// the absolute package directories and embedded files it came from.
type closureDetail struct {
	files       []string
	directories []string
	embedded    []embeddedFile
}

type embeddedFile struct{ pkg, file string }

func goClosureDetail(goExecutable, module, pkg, tags string, env []string) (closureDetail, error) {
	args := []string{"list", "-deps", "-json"}
	if tags != "" {
		args = append(args, "-tags", tags)
	}
	command := exec.Command(goExecutable, append(args, "--", pkg)...)
	command.Dir, command.Env = module, append(slices.Clone(env), "GOFLAGS=-mod=readonly")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return closureDetail{}, fmt.Errorf("go list %s: %w: %s", pkg, err, strings.TrimSpace(stderr.String()))
	}
	if module, err = filepath.Abs(module); err != nil {
		return closureDetail{}, err
	}
	detail := closureDetail{files: []string{}}
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var p struct {
			Dir                                                 string
			Standard                                            bool
			GoFiles, CgoFiles, CFiles, CXXFiles, MFiles, HFiles []string
			FFiles, SFiles, SwigFiles, SwigCXXFiles, SysoFiles  []string
			EmbedFiles                                          []string
		}
		if err := decoder.Decode(&p); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return closureDetail{}, err
		}
		if p.Standard || !strings.HasPrefix(p.Dir+string(filepath.Separator), module+string(filepath.Separator)) {
			continue
		}
		detail.directories = append(detail.directories, p.Dir)
		for _, group := range [][]string{p.GoFiles, p.CgoFiles, p.CFiles, p.CXXFiles, p.MFiles, p.HFiles, p.FFiles, p.SFiles, p.SwigFiles, p.SwigCXXFiles, p.SysoFiles, p.EmbedFiles} {
			for _, name := range group {
				relative, err := filepath.Rel(module, filepath.Join(p.Dir, name))
				if err != nil {
					return closureDetail{}, err
				}
				detail.files = append(detail.files, filepath.ToSlash(relative))
			}
		}
		for _, name := range p.EmbedFiles {
			detail.embedded = append(detail.embedded, embeddedFile{pkg: p.Dir, file: filepath.Join(p.Dir, name)})
		}
	}
	slices.Sort(detail.files)
	detail.files = slices.Compact(detail.files)
	return detail, nil
}
