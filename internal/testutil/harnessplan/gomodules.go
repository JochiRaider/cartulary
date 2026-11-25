package harnessplan

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GoModules is the registry of independent Go modules: reference-pack
// modules that build against their own requirements, never the repository's.
type GoModules struct {
	Modules []GoModule `json:"modules"`
}

type GoModule struct {
	Root       string `json:"root"`
	ModulePath string `json:"module_path"`
	Version    string `json:"version"`
}

var (
	goModuleRoot    = regexp.MustCompile(`^referencepacks/[a-z][a-z0-9_-]*$`)
	goModuleVersion = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	goModuleReplace = regexp.MustCompile(`(?m)^\s*replace\b`)
)

// validate checks each module's boundary: a registered root holding a real
// go.mod that declares the registered path and overrides no dependency, so a
// module never acquires a sibling merely because its directory is present.
func (m GoModules) validate(root string) error {
	seen := map[string]bool{}
	for _, module := range m.Modules {
		if !goModuleRoot.MatchString(module.Root) || seen[module.Root] || module.ModulePath != "github.com/JochiRaider/cartulary/"+module.Root || !goModuleVersion.MatchString(module.Version) {
			return fmt.Errorf("invalid independent Go module boundary %q", module.Root)
		}
		seen[module.Root] = true
		for _, relative := range []string{"referencepacks", module.Root, module.Root + "/go.mod"} {
			info, err := os.Lstat(filepath.Join(root, relative))
			if err != nil || info.Mode()&os.ModeSymlink != 0 || strings.HasSuffix(relative, "go.mod") && !info.Mode().IsRegular() {
				return fmt.Errorf("unsafe independent Go module input %s", relative)
			}
		}
		manifest, err := readBounded(filepath.Join(root, module.Root, "go.mod"))
		if err != nil {
			return err
		}
		lines := strings.Split(strings.ReplaceAll(string(manifest), "\r\n", "\n"), "\n")
		declared := false
		for _, line := range lines {
			declared = declared || line == "module "+module.ModulePath
		}
		if !declared || goModuleReplace.Match(manifest) {
			return fmt.Errorf("independent module %s identity or dependency override mismatch", module.Root)
		}
	}
	return nil
}

var errUnregisteredModule = errors.New("unregistered independent Go module selection")

// goModule resolves a catalog package (./path) to the module directory to
// run in and the package path within it. Packages of the repository's own
// module run from the root under their catalog path.
func (in *Inputs) goModule(pkg string) (directory, within string, err error) {
	relative := strings.TrimPrefix(pkg, "./")
	for _, module := range in.GoModules.Modules {
		if relative == module.Root {
			return module.Root, ".", nil
		}
		if rest, ok := strings.CutPrefix(relative, module.Root+"/"); ok {
			return module.Root, "./" + rest, nil
		}
	}
	if strings.HasPrefix(relative, "referencepacks/") {
		return "", "", fmt.Errorf("%w: %w %s", ErrSelection, errUnregisteredModule, pkg)
	}
	return "", pkg, nil
}
