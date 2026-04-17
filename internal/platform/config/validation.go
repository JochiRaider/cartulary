package config

import (
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

type filesystemRoot struct {
	Path       string
	ConfigPath string
}

func Validate(cfg Config) (Config, error) {
	normalized := cfg
	diagnostics := validateConfigStructure(&normalized)
	if len(diagnostics) > 0 {
		return Config{}, newDiagnosticsError(diagnostics)
	}
	return normalized, nil
}

func ValidateForStartup(cfg Config) (Config, error) {
	normalized, err := Validate(cfg)
	if err != nil {
		return Config{}, err
	}

	diagnostics := validateStartupFilesystemRoots(&normalized)
	if len(diagnostics) > 0 {
		return Config{}, newDiagnosticsError(diagnostics)
	}

	return normalized, nil
}

func validateConfigStructure(cfg *Config) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)

	switch {
	case cfg.ConfigSchemaID == "":
		diagnostics = append(diagnostics, Diagnostic{
			Path:       "config_schema_id",
			ReasonCode: "missing_required_key",
			Message:    "config_schema_id is required",
		})
	case cfg.ConfigSchemaID != expectedConfigSchemaID:
		diagnostics = append(diagnostics, Diagnostic{
			Path:       "config_schema_id",
			ReasonCode: "unsupported_config_schema_id",
			Message:    fmt.Sprintf("unsupported config_schema_id %q", cfg.ConfigSchemaID),
		})
	}

	if cfg.DeploymentProfile == "" {
		diagnostics = append(diagnostics, Diagnostic{
			Path:       "deployment_profile",
			ReasonCode: "missing_required_key",
			Message:    "deployment_profile is required",
		})
	} else if !isValidDeploymentProfile(cfg.DeploymentProfile) {
		diagnostics = append(diagnostics, Diagnostic{
			Path:       "deployment_profile",
			ReasonCode: "invalid_enum",
			Message:    fmt.Sprintf("deployment_profile %q is not supported", cfg.DeploymentProfile),
		})
	}

	validateRootBinding(&cfg.Roots.DatabaseStorage, "roots.database_storage", cfg.DeploymentProfile, true, true, &diagnostics)
	validateRootBinding(&cfg.Roots.ObjectStorage, "roots.object_storage", cfg.DeploymentProfile, true, true, &diagnostics)
	validateRootBinding(&cfg.Roots.BackupStorage, "roots.backup_storage", cfg.DeploymentProfile, true, true, &diagnostics)
	validateRootBinding(&cfg.Roots.ReferencePackStorage, "roots.reference_pack_storage", cfg.DeploymentProfile, false, false, &diagnostics)
	validateRootBinding(&cfg.Roots.TemporaryWork, "roots.temporary_work", cfg.DeploymentProfile, false, false, &diagnostics)
	validateRootBinding(&cfg.Roots.ExportOutputs, "roots.export_outputs", cfg.DeploymentProfile, false, false, &diagnostics)

	if len(diagnostics) > 0 {
		return diagnostics
	}

	return append(diagnostics, detectFilesystemRootOverlap(collectFilesystemRoots(*cfg))...)
}

func validateStartupFilesystemRoots(cfg *Config) []Diagnostic {
	roots := collectFilesystemRoots(*cfg)
	for i := range roots {
		canonicalPath, diagnostic := canonicalizeFilesystemRoot(roots[i].Path, roots[i].ConfigPath)
		if diagnostic != nil {
			return []Diagnostic{*diagnostic}
		}
		roots[i].Path = canonicalPath
	}

	diagnostics := detectFilesystemRootOverlap(roots)
	if len(diagnostics) > 0 {
		return diagnostics
	}

	return nil
}

func validateRootBinding(binding *RootBinding, path string, deploymentProfile string, allowManagedService bool, managedServiceDependsOnProfile bool, diagnostics *[]Diagnostic) {
	if binding.BindingKind == "" && binding.Path == "" && binding.ServiceRef == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path,
			ReasonCode: "missing_required_key",
			Message:    "required runtime root is missing",
		})
		return
	}

	if binding.BindingKind == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path + ".binding_kind",
			ReasonCode: "missing_required_key",
			Message:    "binding_kind is required",
		})
		return
	}

	switch binding.BindingKind {
	case "filesystem_root":
		if binding.Path == "" {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".path",
				ReasonCode: "missing_required_key",
				Message:    "path is required for filesystem_root bindings",
			})
		} else if normalized, diagnostic := validateConfiguredPath(binding.Path, path+".path"); diagnostic != nil {
			*diagnostics = append(*diagnostics, *diagnostic)
		} else {
			binding.Path = normalized
		}

		if binding.ServiceRef != "" {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".service_ref",
				ReasonCode: "type_mismatch",
				Message:    "service_ref is not allowed for filesystem_root bindings",
			})
		}

	case "managed_service":
		if !allowManagedService || (managedServiceDependsOnProfile && deploymentProfile == "disconnected") {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".binding_kind",
				ReasonCode: "profile_incompatible_binding",
				Message:    fmt.Sprintf("binding_kind %q is not allowed for deployment_profile %q", binding.BindingKind, deploymentProfile),
			})
		}

		if binding.ServiceRef == "" {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".service_ref",
				ReasonCode: "missing_required_key",
				Message:    "service_ref is required for managed_service bindings",
			})
		}

		if binding.Path != "" {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       path + ".path",
				ReasonCode: "type_mismatch",
				Message:    "path is not allowed for managed_service bindings",
			})
		}

	default:
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path + ".binding_kind",
			ReasonCode: "invalid_enum",
			Message:    fmt.Sprintf("binding_kind %q is not supported", binding.BindingKind),
		})
	}
}

func validateConfiguredPath(raw string, path string) (string, *Diagnostic) {
	if !isPOSIXAbsolutePath(raw) {
		return "", &Diagnostic{
			Path:       path,
			ReasonCode: "path_not_absolute",
			Message:    "filesystem root paths must be absolute POSIX paths",
		}
	}

	if strings.ContainsRune(raw, '\x00') {
		return "", &Diagnostic{
			Path:       path,
			ReasonCode: "path_forbidden_segment",
			Message:    "filesystem root paths must not contain NUL",
		}
	}

	if strings.HasPrefix(raw, "~") || strings.Contains(raw, "$") {
		return "", &Diagnostic{
			Path:       path,
			ReasonCode: "path_forbidden_segment",
			Message:    "filesystem root paths must not use shell expansion segments",
		}
	}

	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return "", &Diagnostic{
				Path:       path,
				ReasonCode: "path_forbidden_segment",
				Message:    "filesystem root paths must not contain . or .. segments",
			}
		}
	}

	return cleanPOSIXPath(raw), nil
}

func detectFilesystemRootOverlap(roots []filesystemRoot) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)
	for i := 0; i < len(roots); i++ {
		for j := i + 1; j < len(roots); j++ {
			switch {
			case roots[i].Path == roots[j].Path:
				diagnostics = append(diagnostics, Diagnostic{
					Path:       roots[j].ConfigPath,
					ReasonCode: "path_overlap",
					Message:    fmt.Sprintf("filesystem root %q overlaps %q", roots[j].Path, roots[i].ConfigPath),
				})
			case pathWithinRoot(roots[i].Path, roots[j].Path):
				diagnostics = append(diagnostics, Diagnostic{
					Path:       roots[j].ConfigPath,
					ReasonCode: "path_overlap",
					Message:    fmt.Sprintf("filesystem root %q overlaps %q", roots[j].Path, roots[i].ConfigPath),
				})
			case pathWithinRoot(roots[j].Path, roots[i].Path):
				diagnostics = append(diagnostics, Diagnostic{
					Path:       roots[i].ConfigPath,
					ReasonCode: "path_overlap",
					Message:    fmt.Sprintf("filesystem root %q overlaps %q", roots[i].Path, roots[j].ConfigPath),
				})
			}
		}
	}

	sortDiagnostics(diagnostics)
	return diagnostics
}

func collectFilesystemRoots(cfg Config) []filesystemRoot {
	roots := make([]filesystemRoot, 0, 6)
	if cfg.Roots.DatabaseStorage.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.DatabaseStorage.Path, ConfigPath: "roots.database_storage.path"})
	}
	if cfg.Roots.ObjectStorage.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.ObjectStorage.Path, ConfigPath: "roots.object_storage.path"})
	}
	if cfg.Roots.BackupStorage.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.BackupStorage.Path, ConfigPath: "roots.backup_storage.path"})
	}
	if cfg.Roots.ReferencePackStorage.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.ReferencePackStorage.Path, ConfigPath: "roots.reference_pack_storage.path"})
	}
	if cfg.Roots.TemporaryWork.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.TemporaryWork.Path, ConfigPath: "roots.temporary_work.path"})
	}
	if cfg.Roots.ExportOutputs.BindingKind == "filesystem_root" {
		roots = append(roots, filesystemRoot{Path: cfg.Roots.ExportOutputs.Path, ConfigPath: "roots.export_outputs.path"})
	}

	return roots
}

func canonicalizeFilesystemRoot(root string, configPath string) (string, *Diagnostic) {
	cleaned := cleanPOSIXPath(root)
	existingPrefix, exists, err := nearestExistingPath(cleaned)
	if err != nil {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    fmt.Sprintf("inspect filesystem root: %v", err),
		}
	}

	if !exists {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    "filesystem root must resolve under an existing writable parent",
		}
	}

	info, err := os.Stat(existingPrefix)
	if err != nil {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    fmt.Sprintf("stat filesystem root: %v", err),
		}
	}
	if !info.IsDir() {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "type_mismatch",
			Message:    "filesystem root must resolve to a directory",
		}
	}

	resolvedPrefix, err := filepath.EvalSymlinks(existingPrefix)
	if err != nil {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    fmt.Sprintf("resolve filesystem root symlinks: %v", err),
		}
	}

	resolvedPrefix = cleanPOSIXPath(resolvedPrefix)
	if !isWritablePath(resolvedPrefix) {
		return "", &Diagnostic{
			Path:       configPath,
			ReasonCode: "path_not_writable",
			Message:    "filesystem root is not writable at startup",
		}
	}

	if existingPrefix == cleaned {
		return resolvedPrefix, nil
	}

	suffix := strings.TrimPrefix(cleaned, existingPrefix)
	return cleanPOSIXPath(resolvedPrefix + suffix), nil
}

func nearestExistingPath(root string) (string, bool, error) {
	current := root
	for {
		_, err := os.Stat(current)
		if err == nil {
			return current, true, nil
		}
		if !os.IsNotExist(err) {
			return "", false, err
		}

		parent := pathpkg.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}

func pathEscapesRoot(root string, target string) bool {
	if !isPOSIXAbsolutePath(root) || !isPOSIXAbsolutePath(target) {
		return true
	}
	return !pathWithinRoot(cleanPOSIXPath(root), cleanPOSIXPath(target))
}

func pathWithinRoot(root string, target string) bool {
	if root == target {
		return true
	}
	if root == "/" {
		return strings.HasPrefix(target, "/")
	}
	return strings.HasPrefix(target, root+"/")
}

func cleanPOSIXPath(value string) string {
	cleaned := pathpkg.Clean(value)
	if cleaned == "." {
		return value
	}
	return cleaned
}

func isPOSIXAbsolutePath(value string) bool {
	return strings.HasPrefix(value, "/")
}

func isWritablePath(path string) bool {
	return unix.Access(path, unix.W_OK|unix.X_OK) == nil
}

func isValidDeploymentProfile(profile string) bool {
	switch profile {
	case "disconnected", "on_prem", "cloud":
		return true
	default:
		return false
	}
}
