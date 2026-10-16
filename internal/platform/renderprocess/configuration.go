package renderprocess

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (r Runtime) validate(outputs []Output) error {
	for _, root := range []string{r.Root, r.TemporaryRoot} {
		if !filepath.IsAbs(root) || filepath.Clean(root) != root || root == "/" {
			return ErrConfiguration
		}
		resolved, err := filepath.EvalSymlinks(root)
		if err != nil || resolved != root {
			return ErrConfiguration
		}
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			return ErrConfiguration
		}
	}
	if containsPath(r.Root, r.TemporaryRoot) || containsPath(r.TemporaryRoot, r.Root) {
		return ErrConfiguration
	}
	if len(r.SandboxCommand) == 0 || !filepath.IsAbs(r.SandboxCommand[0]) || len(r.Command) == 0 || !filepath.IsAbs(r.Command[0]) || len(r.Launcher) == 0 || !filepath.IsAbs(r.Launcher[0]) || r.LoopbackPort < 1024 {
		return ErrConfiguration
	}
	for _, command := range [][]string{r.SandboxCommand, r.Launcher, r.Command} {
		for _, argument := range command {
			if strings.ContainsRune(argument, '\x00') {
				return ErrConfiguration
			}
		}
	}
	seen := make(map[string]bool, len(outputs))
	for _, output := range outputs {
		if !safeRelativePath(output.Path) || output.MaxBytes < 1 || output.MaxBytes > 500<<20 || seen[output.Path] {
			return ErrConfiguration
		}
		seen[output.Path] = true
	}
	for key, value := range r.Environment {
		if key == "" || key == "CARTULARY_RENDER_LOOPBACK_PORT" || strings.ContainsAny(key, "=\x00") || strings.ContainsRune(value, '\x00') {
			return ErrConfiguration
		}
	}
	return nil
}

func containsPath(root, candidate string) bool {
	return candidate == root || strings.HasPrefix(candidate, root+string(filepath.Separator))
}

func safeRelativePath(value string) bool {
	if value == "" || len(value) > 512 || strings.ContainsAny(value, "\\:") {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func sortedEnvironmentKeys(environment map[string]string) []string {
	keys := make([]string, 0, len(environment))
	for key := range environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
