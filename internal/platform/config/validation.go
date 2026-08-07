package config

import (
	"fmt"
	"net/url"
	"os"
	pathpkg "path"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/sys/unix"
)

type filesystemRoot struct {
	Path       string
	ConfigPath string
}

const limitRegistryMaxInt64 = int64(^uint64(0) >> 1)

func validate(cfg document) (document, error) {
	return validateWithExtensionPolicy(cfg, nil)
}

func validateWithExtensionPolicy(cfg document, policy ExtensionPolicy) (document, error) {
	normalized := cfg
	diagnostics := validateConfigStructure(&normalized, configPresence{}, policy)
	if len(diagnostics) > 0 {
		return document{}, newDiagnosticsError(diagnostics)
	}
	return normalized, nil
}

func validateForStartup(cfg document) (document, error) {
	return validateForStartupWithExtensionPolicy(cfg, nil)
}

func validateForStartupWithExtensionPolicy(cfg document, policy ExtensionPolicy) (document, error) {
	normalized, err := validateWithExtensionPolicy(cfg, policy)
	if err != nil {
		return document{}, err
	}

	diagnostics := validateStartupFilesystemRoots(&normalized)
	if len(diagnostics) > 0 {
		return document{}, newDiagnosticsError(diagnostics)
	}

	return normalized, nil
}

func validateConfigStructure(cfg *document, presence configPresence, inactivePolicy ExtensionPolicy) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)
	cfg.presence = presence

	applyDefaultLimitValues(cfg, presence)
	diagnostics = append(diagnostics, validateAndDiscardInactiveExtensionConfiguration(cfg, presence, inactivePolicy)...)

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

	validateApplication(&cfg.Application, &diagnostics)
	validateRootBinding(&cfg.Roots.DatabaseStorage, "roots.database_storage", cfg.DeploymentProfile, true, true, &diagnostics)
	validateRootBinding(&cfg.Roots.ObjectStorage, "roots.object_storage", cfg.DeploymentProfile, true, true, &diagnostics)
	validateRootBinding(&cfg.Roots.BackupStorage, "roots.backup_storage", cfg.DeploymentProfile, true, true, &diagnostics)
	validateRootBinding(&cfg.Roots.ReferencePackStorage, "roots.reference_pack_storage", cfg.DeploymentProfile, false, false, &diagnostics)
	validateRootBinding(&cfg.Roots.TemporaryWork, "roots.temporary_work", cfg.DeploymentProfile, false, false, &diagnostics)
	validateRootBinding(&cfg.Roots.ExportOutputs, "roots.export_outputs", cfg.DeploymentProfile, false, false, &diagnostics)
	validateBootstrapManifestPath(&cfg.Bootstrap, presence, &diagnostics)
	applyDefaultExtensionRuntimeValues(cfg, presence)
	validateLimitRegistry(cfg.Limits, &diagnostics)
	validateExtensionRuntimeValues(*cfg, &diagnostics)
	if len(diagnostics) > 0 {
		return diagnostics
	}

	roots := collectFilesystemRoots(*cfg)
	diagnostics = append(diagnostics, detectBackupStorageRootSubstitution(roots)...)
	diagnostics = append(diagnostics, detectFilesystemRootOverlap(roots)...)
	return diagnostics
}

func validateApplication(application *ApplicationConfig, diagnostics *[]Diagnostic) {
	if strings.TrimSpace(application.PublicOrigin) == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       "application.public_origin",
			ReasonCode: "missing_required_key",
			Message:    "application public origin is required",
		})
		return
	}

	parsed, err := url.Parse(application.PublicOrigin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       "application.public_origin",
			ReasonCode: "invalid_origin",
			Message:    "application public origin must be an absolute http or https origin",
		})
		return
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       "application.public_origin",
			ReasonCode: "invalid_origin",
			Message:    "application public origin scheme must be http or https",
		})
		return
	}
	if parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Path:       "application.public_origin",
			ReasonCode: "invalid_origin",
			Message:    "application public origin must not include userinfo, path, query, or fragment",
		})
		return
	}
	application.PublicOrigin = parsed.Scheme + "://" + parsed.Host
}

func validateStartupFilesystemRoots(cfg *document) []Diagnostic {
	roots := collectFilesystemRoots(*cfg)
	for i := range roots {
		canonicalPath, diagnostic := canonicalizeFilesystemRoot(roots[i].Path, roots[i].ConfigPath)
		if diagnostic != nil {
			return []Diagnostic{*diagnostic}
		}
		roots[i].Path = canonicalPath
	}

	diagnostics := detectBackupStorageRootSubstitution(roots)
	diagnostics = append(diagnostics, detectFilesystemRootOverlap(roots)...)
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

func validateAndDiscardInactiveExtensionConfiguration(cfg *document, presence configPresence, policy ExtensionPolicy) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)
	if policy == nil {
		return diagnostics
	}
	for _, key := range policy.Keys() {
		claimKey, known := policy.ClaimKey(key)
		claimed, claimExists := booleanFieldAtPath(cfg, claimKey)
		if !known || !claimExists || claimed {
			continue
		}
		field, exists := configFieldAtPath(cfg, key)
		if !exists {
			continue
		}
		segments := strings.Split(key, ".")
		if field.IsZero() && !presence.isDefined(segments...) {
			continue
		}
		findings := policy.ValidateAndDiscard(map[string]any{key: field.Interface()})
		if len(findings) > 0 {
			for _, finding := range findings {
				diagnostics = append(diagnostics, Diagnostic{
					Path:       finding[0],
					ReasonCode: finding[1],
					Message:    "Extension configuration is present while the profile is inactive.",
					Details:    inactiveDiagnosticDetails(finding[0]),
				})
			}
			continue
		}
		field.Set(reflect.Zero(field.Type()))
	}
	return diagnostics
}

func booleanFieldAtPath(cfg *document, path string) (bool, bool) {
	if cfg == nil {
		return false, false
	}
	if claim, present := cfg.claims[path]; present {
		return claim.value, true
	}
	field, present := configFieldAtPath(cfg, path)
	if !present || field.Kind() != reflect.Bool {
		return false, false
	}
	return field.Bool(), true
}

func configFieldAtPath(cfg *document, path string) (reflect.Value, bool) {
	segments := strings.Split(path, ".")
	if cfg == nil || len(segments) == 0 {
		return reflect.Value{}, false
	}
	value := reflect.ValueOf(cfg).Elem()
	if namespace, present := cfg.namespaces[segments[0]]; present {
		value = reflect.ValueOf(namespace)
		for value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return reflect.Value{}, false
			}
			value = value.Elem()
		}
		segments = segments[1:]
	}
	for _, segment := range segments {
		field, ok := findTaggedField(value, segment)
		if !ok {
			return reflect.Value{}, false
		}
		value = field
	}
	return value, value.CanSet()
}

func applyDefaultLimitValues(cfg *document, presence configPresence) {
	applyDefaultInt64(&cfg.Limits.ObjectBlobs.MaxDeclaredByteSize, defaultObjectBlobMaxDeclaredByteSize, presence, "limits", "object_blobs", "max_declared_byte_size")
	applyDefaultInt64(&cfg.Limits.Imports.MaxCSVSourceBytes, defaultImportMaxCSVSourceBytes, presence, "limits", "imports", "max_csv_source_bytes")
	applyDefaultInt64(&cfg.Limits.Imports.MaxXLSXSourceBytes, defaultImportMaxXLSXSourceBytes, presence, "limits", "imports", "max_xlsx_source_bytes")
	applyDefaultInt64(&cfg.Limits.Imports.MaxRows, defaultImportMaxRows, presence, "limits", "imports", "max_rows")
	applyDefaultInt64(&cfg.Limits.Imports.MaxColumns, defaultImportMaxColumns, presence, "limits", "imports", "max_columns")
	applyDefaultInt64(&cfg.Limits.Imports.MaxCells, defaultImportMaxCells, presence, "limits", "imports", "max_cells")
	applyDefaultInt64(&cfg.Limits.Archives.DefaultMaxExtractedBytes, defaultArchiveMaxExtractedBytes, presence, "limits", "archives", "default_max_extracted_bytes")
	applyDefaultInt64(&cfg.Limits.Archives.MaxCompressionRatio, defaultArchiveMaxCompressionRatio, presence, "limits", "archives", "max_compression_ratio")
	applyDefaultInt64(&cfg.Limits.Archives.MaxMembers, defaultArchiveMaxMembers, presence, "limits", "archives", "max_members")
	applyDefaultInt64(&cfg.Limits.ReferencePacks.MaxExtractedBytes, defaultReferencePackMaxExtractedBytes, presence, "limits", "reference_packs", "max_extracted_bytes")
	applyDefaultInt64(&cfg.Limits.IncidentBundles.MaxExtractedBytes, defaultIncidentBundleMaxExtractedBytes, presence, "limits", "incident_bundles", "max_extracted_bytes")
	applyDefaultInt64(&cfg.Limits.Previews.MaxPreviewablePayloadBytes, defaultPreviewMaxPreviewablePayloadBytes, presence, "limits", "previews", "max_previewable_payload_bytes")
	applyDefaultInt64(&cfg.Limits.Previews.MaxTextInlineBytes, defaultPreviewMaxTextInlineBytes, presence, "limits", "previews", "max_text_inline_bytes")
}

func applyDefaultExtensionRuntimeValues(cfg *document, presence configPresence) {
	timeouts := &cfg.Timeouts.Extensions
	applyDefaultInt64(&timeouts.MigrationLockSeconds, 30, presence, "timeouts", "extensions", "migration_lock_seconds")
	applyDefaultInt64(&timeouts.MigrationStepSeconds, 900, presence, "timeouts", "extensions", "migration_step_seconds")
	applyDefaultInt64(&timeouts.ProfileMigrationSeconds, 900, presence, "timeouts", "extensions", "profile_migration_seconds")
	applyDefaultInt64(&timeouts.ValidationSeconds, 30, presence, "timeouts", "extensions", "validation_seconds")
	applyDefaultInt64(&timeouts.ReconciliationSeconds, 300, presence, "timeouts", "extensions", "reconciliation_seconds")
	applyDefaultInt64(&timeouts.ShutdownDrainSeconds, 30, presence, "timeouts", "extensions", "shutdown_drain_seconds")
	applyDefaultInt64(&timeouts.ProcessLeaseAcquireSeconds, 30, presence, "timeouts", "extensions", "process_lease_acquire_seconds")
	applyDefaultInt64(&timeouts.ProcessLeaseLossDetectionSeconds, 5, presence, "timeouts", "extensions", "process_lease_loss_detection_seconds")
	applyDefaultInt64(&timeouts.PublicationSeconds, 30, presence, "timeouts", "extensions", "publication_seconds")
	applyDefaultInt64(&timeouts.TransactionParticipantSeconds, 30, presence, "timeouts", "extensions", "transaction_participant_seconds")
	applyDefaultInt64(&timeouts.CancellationGraceSeconds, 2, presence, "timeouts", "extensions", "cancellation_grace_seconds")
	applyDefaultInt64(&timeouts.StagedObjectCleanupSeconds, 300, presence, "timeouts", "extensions", "staged_object_cleanup_seconds")
	applyDefaultInt64(&timeouts.PortabilityParticipantSeconds, 300, presence, "timeouts", "extensions", "portability_participant_seconds")
	applyDefaultInt64(&timeouts.SnapshotReportingParticipantSeconds, 300, presence, "timeouts", "extensions", "snapshot_reporting_participant_seconds")
	applyDefaultInt64(&timeouts.BackupRestoreParticipantSeconds, 900, presence, "timeouts", "extensions", "backup_restore_participant_seconds")
	applyDefaultInt64(&cfg.Intervals.Extensions.StagedObjectSweepSeconds, 300, presence, "intervals", "extensions", "staged_object_sweep_seconds")
	applyDefaultInt64(&cfg.Limits.Extensions.StagedObjectCleanupBatch, defaultExtensionStagedObjectCleanupBatch, presence, "limits", "extensions", "staged_object_cleanup_batch")
	applyDefaultInt64(&cfg.Limits.Extensions.MaxNonterminalJobsPerProfile, defaultExtensionMaxNonterminalJobsPerProfile, presence, "limits", "extensions", "max_nonterminal_jobs_per_profile")
}

func validateExtensionRuntimeValues(cfg document, diagnostics *[]Diagnostic) {
	timeouts := cfg.Timeouts.Extensions
	validateLimitValue(timeouts.MigrationLockSeconds, "timeouts.extensions.migration_lock_seconds", 1, 300, diagnostics)
	validateLimitValue(timeouts.MigrationStepSeconds, "timeouts.extensions.migration_step_seconds", 1, 3600, diagnostics)
	validateLimitValue(timeouts.ProfileMigrationSeconds, "timeouts.extensions.profile_migration_seconds", 1, 7200, diagnostics)
	validateLimitValue(timeouts.ValidationSeconds, "timeouts.extensions.validation_seconds", 1, 300, diagnostics)
	validateLimitValue(timeouts.ReconciliationSeconds, "timeouts.extensions.reconciliation_seconds", 1, 3600, diagnostics)
	validateLimitValue(timeouts.ShutdownDrainSeconds, "timeouts.extensions.shutdown_drain_seconds", 1, 300, diagnostics)
	validateLimitValue(timeouts.ProcessLeaseAcquireSeconds, "timeouts.extensions.process_lease_acquire_seconds", 1, 300, diagnostics)
	validateLimitValue(timeouts.ProcessLeaseLossDetectionSeconds, "timeouts.extensions.process_lease_loss_detection_seconds", 1, 30, diagnostics)
	validateLimitValue(timeouts.PublicationSeconds, "timeouts.extensions.publication_seconds", 1, 300, diagnostics)
	validateLimitValue(timeouts.TransactionParticipantSeconds, "timeouts.extensions.transaction_participant_seconds", 1, 300, diagnostics)
	validateLimitValue(timeouts.CancellationGraceSeconds, "timeouts.extensions.cancellation_grace_seconds", 0, 30, diagnostics)
	validateLimitValue(timeouts.StagedObjectCleanupSeconds, "timeouts.extensions.staged_object_cleanup_seconds", 30, 3600, diagnostics)
	validateLimitValue(timeouts.PortabilityParticipantSeconds, "timeouts.extensions.portability_participant_seconds", 1, 3600, diagnostics)
	validateLimitValue(timeouts.SnapshotReportingParticipantSeconds, "timeouts.extensions.snapshot_reporting_participant_seconds", 1, 3600, diagnostics)
	validateLimitValue(timeouts.BackupRestoreParticipantSeconds, "timeouts.extensions.backup_restore_participant_seconds", 1, 7200, diagnostics)
	validateLimitValue(cfg.Intervals.Extensions.StagedObjectSweepSeconds, "intervals.extensions.staged_object_sweep_seconds", 30, 3600, diagnostics)
	validateLimitValue(cfg.Limits.Extensions.StagedObjectCleanupBatch, "limits.extensions.staged_object_cleanup_batch", 1, 10000, diagnostics)
	validateLimitValue(cfg.Limits.Extensions.MaxNonterminalJobsPerProfile, "limits.extensions.max_nonterminal_jobs_per_profile", 1, 1000000, diagnostics)
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

func detectBackupStorageRootSubstitution(roots []filesystemRoot) []Diagnostic {
	var backupRoot *filesystemRoot
	for i := range roots {
		if roots[i].ConfigPath == "roots.backup_storage.path" {
			backupRoot = &roots[i]
			break
		}
	}
	if backupRoot == nil {
		return nil
	}

	diagnostics := make([]Diagnostic, 0, 2)
	for _, root := range roots {
		switch root.ConfigPath {
		case "roots.export_outputs.path", "roots.temporary_work.path":
			if filesystemRootsOverlap(backupRoot.Path, root.Path) {
				diagnostics = append(diagnostics, Diagnostic{
					Path:       backupRoot.ConfigPath,
					ReasonCode: "path_overlap",
					Message:    fmt.Sprintf("backup storage root %q must remain distinct from %s", backupRoot.Path, strings.TrimSuffix(root.ConfigPath, ".path")),
				})
			}
		}
	}

	sortDiagnostics(diagnostics)
	return diagnostics
}

func filesystemRootsOverlap(left string, right string) bool {
	return left == right || pathWithinRoot(left, right) || pathWithinRoot(right, left)
}

func collectFilesystemRoots(cfg document) []filesystemRoot {
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
