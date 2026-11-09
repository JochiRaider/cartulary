// The module archive helper has only standard-library dependencies. Public Make
// workflows own invocation; exported modules never include this repository tool.
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type file struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size_bytes"`
}
type descriptor struct {
	SchemaID      string `json:"schema_id"`
	Module        string `json:"module_path"`
	Version       string `json:"version"`
	SourceSHA256  string `json:"source_sha256"`
	ArchiveSHA256 string `json:"archive_sha256"`
	ArchivePath   string `json:"archive_path"`
	ProxyRoot     string `json:"proxy_root"`
	Files         []file `json:"files"`
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func immutable(name string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(name), ".partial-")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	_, err = temporary.Write(b)
	err = errors.Join(err, temporary.Sync(), temporary.Close())
	if err != nil {
		return err
	}
	if err = os.Link(temporary.Name(), name); errors.Is(err, os.ErrExist) {
		info, e := os.Lstat(name)
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return errors.New("nonregular module output")
		}
		old, e := os.ReadFile(name)
		if e != nil {
			return e
		}
		if !bytes.Equal(old, b) {
			return errors.New("immutable module archive collision")
		}
		return nil
	}
	return err
}
func escaped(module string) string {
	var out strings.Builder
	for _, c := range module {
		if c >= 'A' && c <= 'Z' {
			out.WriteByte('!')
			out.WriteRune(c - 'A' + 'a')
		} else {
			out.WriteRune(c)
		}
	}
	return out.String()
}
func export(root, module, base, out string) (descriptor, error) {
	d := descriptor{SchemaID: "cartulary.reference_pack_module_archive.v1", Module: module, Files: []file{}}
	if !regexp.MustCompile(`^github.com/JochiRaider/cartulary/referencepacks/[a-z][a-z0-9_-]*$`).MatchString(module) || !regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(base) {
		return d, errors.New("invalid module archive identity")
	}
	records := map[string][]byte{}
	var total int64
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		// Prose and VCS state are never executable archive inputs, even if present.
		if entry.IsDir() {
			if relative != "." && (entry.Name() == "docs" || strings.HasPrefix(entry.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(strings.ToLower(entry.Name()), "readme") || strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("symlinked module input")
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("nonregular module input")
		}
		if relative != "go.mod" && relative != "go.sum" && filepath.Ext(relative) != ".go" && filepath.Ext(relative) != ".json" && filepath.Ext(relative) != ".ndjson" && filepath.Ext(relative) != ".txt" {
			return fmt.Errorf("undeclared module archive input type: %s", relative)
		}
		if info.Size() > 64<<20 || info.Size() > 512<<20-total {
			return errors.New("module archive input too large")
		}
		total += info.Size()
		b, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		if int64(len(b)) != info.Size() {
			return errors.New("module source changed")
		}
		if filepath.Ext(relative) == ".go" {
			if err := sourceInputs(relative, b); err != nil {
				return err
			}
		}
		records[relative] = b
		d.Files = append(d.Files, file{relative, digest(b), int64(len(b))})
		return nil
	})
	if err != nil {
		return d, err
	}
	mod, ok := records["go.mod"]
	if !ok || !bytes.Contains(mod, []byte("module "+module+"\n")) || regexp.MustCompile(`(?m)^\s*replace\b`).Match(mod) {
		return d, errors.New("module manifest boundary mismatch")
	}
	if _, ok = records["go.sum"]; !ok {
		return d, errors.New("pinned module checksums required")
	}
	slices.SortFunc(d.Files, func(a, b file) int { return strings.Compare(a.Path, b.Path) })
	hash := sha256.New()
	hash.Write([]byte("cartulary.reference-pack-module-source.v1\x00"))
	for _, f := range d.Files {
		binary.Write(hash, binary.BigEndian, uint64(len(f.Path)))
		hash.Write([]byte(f.Path))
		binary.Write(hash, binary.BigEndian, uint64(f.Size))
		hash.Write(records[f.Path])
	}
	d.SourceSHA256 = hex.EncodeToString(hash.Sum(nil))
	d.Version = base + "-dev." + d.SourceSHA256
	d.ProxyRoot = filepath.Join(out, "proxy")
	prefix := module + "@" + d.Version + "/"
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for _, f := range d.Files {
		header := &zip.FileHeader{Name: prefix + f.Path, Method: zip.Deflate}
		header.SetMode(0644)
		header.SetModTime(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
		member, err := writer.CreateHeader(header)
		if err != nil {
			return d, err
		}
		if _, err = member.Write(records[f.Path]); err != nil {
			return d, err
		}
	}
	if err = writer.Close(); err != nil {
		return d, err
	}
	d.ArchiveSHA256 = digest(archive.Bytes())
	directory := filepath.Join(d.ProxyRoot, filepath.FromSlash(escaped(module)), "@v")
	d.ArchivePath = filepath.Join(directory, d.Version+".zip")
	if err = immutable(d.ArchivePath, archive.Bytes()); err != nil {
		return d, err
	}
	if err = immutable(filepath.Join(directory, d.Version+".mod"), mod); err != nil {
		return d, err
	}
	info, _ := json.Marshal(map[string]string{"Version": d.Version, "Time": "2000-01-01T00:00:00Z"})
	if err = immutable(filepath.Join(directory, d.Version+".info"), info); err != nil {
		return d, err
	}
	return d, nil
}

// Executable input discovery is explicit. Producers accept paths and pinned
// source profiles from their caller; neither the caller's repository, home,
// environment nor an auxiliary executable may silently supply input files.
// Compiler-recorded builder identity remains available through ReadBuildInfo.
func sourceInputs(name string, raw []byte) error {
	parsed, err := parser.ParseFile(token.NewFileSet(), name, raw, parser.ParseComments)
	if err != nil {
		return errors.New("invalid module Go source")
	}
	for _, group := range parsed.Comments {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, "//go:generate") {
				return errors.New("independent modules contain authored inputs only; implicit generation is forbidden")
			}
		}
	}
	if strings.HasSuffix(name, "_test.go") {
		return nil
	}
	imports := map[string]string{}
	for _, spec := range parsed.Imports {
		module, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return err
		}
		alias := path.Base(module)
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		if alias == "." || module == "os/exec" || module == "plugin" || module == "C" {
			return errors.New("unbounded executable input mechanism in independent module")
		}
		imports[alias] = module
	}
	ast.Inspect(parsed, func(node ast.Node) bool {
		selection, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		identifier, ok := selection.X.(*ast.Ident)
		if !ok {
			return true
		}
		module := imports[identifier.Name]
		if module == "os" && slices.Contains([]string{"Getwd", "Getenv", "LookupEnv", "Environ", "UserHomeDir", "UserConfigDir", "UserCacheDir", "Executable", "Chdir"}, selection.Sel.Name) ||
			module == "runtime" && slices.Contains([]string{"Caller", "Callers", "FuncForPC"}, selection.Sel.Name) {
			err = errors.New("implicit executable input discovery in independent module")
		}
		return true
	})
	return err
}
func extract(archive, destination, module, version string) error {
	z, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer z.Close()
	root, err := os.OpenRoot(destination)
	if err != nil {
		return err
	}
	defer root.Close()
	prefix := module + "@" + version + "/"
	seen := map[string]bool{}
	var total uint64
	for _, member := range z.File {
		if !strings.HasPrefix(member.Name, prefix) {
			return errors.New("unexpected module archive prefix")
		}
		name := strings.TrimPrefix(member.Name, prefix)
		if name == "" || name == "." || name == ".." || name != path.Clean(name) || strings.HasPrefix(name, "/") || strings.HasPrefix(name, "../") || strings.Contains(name, "\\") || seen[name] || !member.Mode().IsRegular() {
			return errors.New("unsafe module archive member")
		}
		seen[name] = true
		if member.UncompressedSize64 > 64<<20 || member.UncompressedSize64 > 512<<20-total {
			return errors.New("module archive size limit")
		}
		total += member.UncompressedSize64
		if err = root.MkdirAll(path.Dir(name), 0700); err != nil {
			return err
		}
		input, err := member.Open()
		if err != nil {
			return err
		}
		output, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			input.Close()
			return err
		}
		n, err := io.Copy(output, io.LimitReader(input, int64(member.UncompressedSize64)+1))
		err = errors.Join(err, input.Close(), output.Close())
		if err != nil {
			return err
		}
		if uint64(n) != member.UncompressedSize64 {
			return errors.New("module archive member size mismatch")
		}
	}
	return nil
}
func main() {
	action := flag.String("action", "export", "")
	root := flag.String("root", "", "")
	module := flag.String("module", "", "")
	version := flag.String("version", "", "")
	out := flag.String("out", "", "")
	archive := flag.String("archive", "", "")
	flag.Parse()
	var err error
	switch *action {
	case "export":
		var d descriptor
		d, err = export(*root, *module, *version, *out)
		if err == nil {
			err = json.NewEncoder(os.Stdout).Encode(d)
		}
	case "extract":
		err = extract(*archive, *out, *module, *version)
	default:
		err = errors.New("unknown module archive action")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
