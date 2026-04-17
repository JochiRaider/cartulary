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

const limitRegistryMaxInt64 = int64(^uint64(0) >> 1)

func Validate(cfg Config) (Config, error) {
	normalized := cfg
	diagnostics := validateConfigStructure(&normalized, configPresence{})
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

func validateConfigStructure(cfg *Config, presence configPresence) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)

	applyDefaultLimitValues(cfg, presence)

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
	validateBootstrapManifestPath(&cfg.Bootstrap, presence, &diagnostics)
	validateLimitRegistry(cfg.Limits, &diagnostics)

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
	return validateConfiguredAbsolutePOSIXPath(raw, path, "filesystem root paths")
}

func validateConfiguredManifestPath(raw string, path string) (string, *Diagnostic) {
	return validateConfiguredAbsolutePOSIXPath(raw, path, "bootstrap manifest path")
}

func validateConfiguredAbsolutePOSIXPath(raw string, path string, subject string) (string, *Diagnostic) {
	if !isPOSIXAbsolutePath(raw) {
		return "", &Diagnostic{
			Path:       path,
			ReasonCode: "path_not_absolute",
			Message:    subject + " must be an absolute POSIX path",
		}
	}

	if strings.ContainsRune(raw, '\x00') {
		return "", &Diagnostic{
			Path:       path,
			ReasonCode: "path_forbidden_segment",
			Message:    subject + " must not contain NUL",
		}
	}

	if strings.HasPrefix(raw, "~") || strings.Contains(raw, "$") {
		return "", &Diagnostic{
			Path:       path,
			ReasonCode: "path_forbidden_segment",
			Message:    subject + " must not use shell expansion segments",
		}
	}

	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return "", &Diagnostic{
				Path:       path,
				ReasonCode: "path_forbidden_segment",
				Message:    subject + " must not contain . or .. segments",
			}
		}
	}

	return cleanPOSIXPath(raw), nil
}

func validateBootstrapManifestPath(bootstrap *BootstrapConfig, presence configPresence, diagnostics *[]Diagnostic) {
	if bootstrap.FirstAdminManifestPath == "" {
		if presence.isDefined("bootstrap", "first_admin_manifest_path") {
			*diagnostics = append(*diagnostics, Diagnostic{
				Path:       "bootstrap.first_admin_manifest_path",
				ReasonCode: "path_not_absolute",
				Message:    "bootstrap manifest path must be an absolute POSIX path when configured",
			})
		}
		return
	}

	if normalized, diagnostic := validateConfiguredManifestPath(bootstrap.FirstAdminManifestPath, "bootstrap.first_admin_manifest_path"); diagnostic != nil {
		*diagnostics = append(*diagnostics, *diagnostic)
	} else {
		bootstrap.FirstAdminManifestPath = normalized
	}
}

func applyDefaultLimitValues(cfg *Config, presence configPresence) {
	applyDefaultInt64(&cfg.Limits.ObjectBlobs.MaxDeclaredByteSize, DefaultObjectBlobMaxDeclaredByteSize, presence, "limits", "object_blobs", "max_declared_byte_size")
	applyDefaultInt64(&cfg.Limits.Imports.MaxCSVSourceBytes, DefaultImportMaxCSVSourceBytes, presence, "limits", "imports", "max_csv_source_bytes")
	applyDefaultInt64(&cfg.Limits.Imports.MaxXLSXSourceBytes, DefaultImportMaxXLSXSourceBytes, presence, "limits", "imports", "max_xlsx_source_bytes")
	applyDefaultInt64(&cfg.Limits.Imports.MaxRows, DefaultImportMaxRows, presence, "limits", "imports", "max_rows")
	applyDefaultInt64(&cfg.Limits.Imports.MaxColumns, DefaultImportMaxColumns, presence, "limits", "imports", "max_columns")
	applyDefaultInt64(&cfg.Limits.Imports.MaxCells, DefaultImportMaxCells, presence, "limits", "imports", "max_cells")
	applyDefaultInt64(&cfg.Limits.Archives.DefaultMaxExtractedBytes, DefaultArchiveMaxExtractedBytes, presence, "limits", "archives", "default_max_extracted_bytes")
	applyDefaultInt64(&cfg.Limits.Archives.MaxCompressionRatio, DefaultArchiveMaxCompressionRatio, presence, "limits", "archives", "max_compression_ratio")
	applyDefaultInt64(&cfg.Limits.Archives.MaxMembers, DefaultArchiveMaxMembers, presence, "limits", "archives", "max_members")
	applyDefaultInt64(&cfg.Limits.ReferencePacks.MaxExtractedBytes, DefaultReferencePackMaxExtractedBytes, presence, "limits", "reference_packs", "max_extracted_bytes")
	applyDefaultInt64(&cfg.Limits.IncidentBundles.MaxExtractedBytes, DefaultIncidentBundleMaxExtractedBytes, presence, "limits", "incident_bundles", "max_extracted_bytes")
	applyDefaultInt64(&cfg.Limits.Previews.MaxPreviewablePayloadBytes, DefaultPreviewMaxPreviewablePayloadBytes, presence, "limits", "previews", "max_previewable_payload_bytes")
	applyDefaultInt64(&cfg.Limits.Previews.MaxTextInlineBytes, DefaultPreviewMaxTextInlineBytes, presence, "limits", "previews", "max_text_inline_bytes")
}

func applyDefaultInt64(target *int64, defaultValue int64, presence configPresence, path ...string) {
	if *target != 0 || presence.isDefined(path...) {
		return
	}
	*target = defaultValue
}

func validateLimitRegistry(limits LimitConfig, diagnostics *[]Diagnostic) {
	validateLimitValue(limits.ObjectBlobs.MaxDeclaredByteSize, "limits.object_blobs.max_declared_byte_size", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxCSVSourceBytes, "limits.imports.max_csv_source_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxXLSXSourceBytes, "limits.imports.max_xlsx_source_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxRows, "limits.imports.max_rows", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxColumns, "limits.imports.max_columns", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Imports.MaxCells, "limits.imports.max_cells", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Archives.DefaultMaxExtractedBytes, "limits.archives.default_max_extracted_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Archives.MaxCompressionRatio, "limits.archives.max_compression_ratio", 1, 1000, diagnostics)
	validateLimitValue(limits.Archives.MaxMembers, "limits.archives.max_members", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.ReferencePacks.MaxExtractedBytes, "limits.reference_packs.max_extracted_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.IncidentBundles.MaxExtractedBytes, "limits.incident_bundles.max_extracted_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Previews.MaxPreviewablePayloadBytes, "limits.previews.max_previewable_payload_bytes", 1, limitRegistryMaxInt64, diagnostics)
	validateLimitValue(limits.Previews.MaxTextInlineBytes, "limits.previews.max_text_inline_bytes", 1, limitRegistryMaxInt64, diagnostics)
}

func validateLimitValue(value int64, path string, min int64, max int64, diagnostics *[]Diagnostic) {
	switch {
	case value < min:
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path,
			ReasonCode: "value_below_minimum",
			Message:    fmt.Sprintf("value must be at least %d", min),
		})
	case value > max:
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       path,
			ReasonCode: "value_above_maximum",
			Message:    fmt.Sprintf("value must be at most %d", max),
		})
	}
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
