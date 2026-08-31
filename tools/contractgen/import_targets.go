package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	importTargetGeneratedPrefix       = "contracts/imports/generated/"
	importTargetSourceManifestPath    = importTargetGeneratedPrefix + "source-manifest.v1.json"
	importTargetRegistryPath          = importTargetGeneratedPrefix + "import-target-registry.v1.json"
	importTargetAdapterDescriptorPath = importTargetGeneratedPrefix + "adapter-descriptors.v1.json"
	importTargetVerificationPath      = importTargetGeneratedPrefix + "verification-targets.v1.json"
	importTargetIntegrityPath         = importTargetGeneratedPrefix + "integrity.v1.json"
)

var importTargetIdentifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_.:/@-]{0,255}$`)

type importTargetInputCatalog struct {
	Schema                 string                        `json:"$schema"`
	SchemaID               string                        `json:"schema_id"`
	CoreSchemaPath         string                        `json:"core_schema_path"`
	ViewTargetInputPath    string                        `json:"view_target_input_path"`
	ViewSchemaRegistryPath string                        `json:"view_schema_registry_path"`
	TargetOrder            []string                      `json:"target_order"`
	OwnerContracts         []string                      `json:"owner_contracts"`
	FacadeContracts        []importTargetFacadeContract  `json:"facade_contracts"`
	AnalyticalInputs       []importTargetAnalyticalInput `json:"analytical_inputs"`
	ErrorTranslationIDs    []string                      `json:"error_translation_ids"`
	CommitProtocolIDs      []string                      `json:"commit_protocol_ids"`
}

type importTargetFacadeContract struct {
	FacadeID         string `json:"facade_id"`
	FacadeKind       string `json:"facade_kind"`
	OwnerContractRef string `json:"owner_contract_ref"`
}

type importTargetAnalyticalInput struct {
	BindingPath                 string   `json:"binding_path"`
	OwnerSchemaPaths            []string `json:"owner_schema_paths"`
	ExtensionContributionPath   string   `json:"extension_contribution_path"`
	AvailabilityKind            string   `json:"availability_kind"`
	ActivationPolicy            string   `json:"activation_policy"`
	DefaultUnknownColumnPolicy  string   `json:"default_unknown_column_policy"`
	EntityBearingDefault        string   `json:"entity_bearing_default"`
	PublicProjectionDisposition string   `json:"public_projection_disposition"`
}

type importViewTargetInputSet struct {
	SchemaID string                  `json:"schema_id"`
	Targets  []importViewTargetInput `json:"targets"`
}

type importViewTargetInput struct {
	TargetViewSchemaID          string  `json:"target_view_schema_id"`
	OwnerContractRef            string  `json:"owner_contract_ref"`
	SourceResourceFamily        string  `json:"source_resource_family"`
	FacadeID                    *string `json:"facade_id"`
	AvailabilityKind            string  `json:"availability_kind"`
	ActivationPolicy            string  `json:"activation_policy"`
	MappingContractSchemaID     string  `json:"mapping_contract_schema_id"`
	DefaultUnknownColumnPolicy  string  `json:"default_unknown_column_policy"`
	EntityBearingDefault        string  `json:"entity_bearing_default"`
	PublicProjectionDisposition string  `json:"public_projection_disposition"`
}

type analyticalFacadeBinding struct {
	SchemaID               string `json:"schema_id"`
	TargetKind             string `json:"target_kind"`
	ExtensionProfileID     string `json:"extension_profile_id"`
	OwnerContractRef       string `json:"owner_contract_ref"`
	FacadeID               string `json:"facade_id"`
	ContractMajor          int    `json:"contract_major"`
	MappingSchemaID        string `json:"mapping_schema_id"`
	PreviewRequestSchemaID string `json:"preview_request_schema_id"`
	PreviewResultSchemaID  string `json:"preview_result_schema_id"`
	ApplyRequestSchemaID   string `json:"apply_request_schema_id"`
	ApplyResultSchemaID    string `json:"apply_result_schema_id"`
	ErrorSchemaID          string `json:"error_schema_id"`
	ErrorTranslationID     string `json:"error_translation_id"`
	CommitProtocolID       string `json:"commit_protocol_id"`
}

type importTargetRegistry struct {
	SchemaID     string                    `json:"schema_id"`
	SourceSHA256 string                    `json:"source_sha256"`
	Rows         []importTargetRegistryRow `json:"rows"`
}

type importTargetRegistryRow struct {
	RegistryOrder               int     `json:"registry_order"`
	TargetID                    string  `json:"target_id"`
	TargetKind                  string  `json:"target_kind"`
	TargetViewSchemaID          *string `json:"target_view_schema_id"`
	ExtensionProfileID          *string `json:"extension_profile_id"`
	OwnerContractRef            string  `json:"owner_contract_ref"`
	SourceResourceFamily        string  `json:"source_resource_family"`
	FacadeKind                  string  `json:"facade_kind"`
	FacadeBindingID             *string `json:"facade_binding_id"`
	FacadeID                    *string `json:"facade_id"`
	AvailabilityKind            string  `json:"availability_kind"`
	ActivationPolicy            string  `json:"activation_policy"`
	MappingContractSchemaID     string  `json:"mapping_contract_schema_id"`
	CreateRequestSchemaID       *string `json:"create_request_schema_id"`
	CreateResultSchemaID        *string `json:"create_result_schema_id"`
	PreviewRequestSchemaID      *string `json:"preview_request_schema_id"`
	PreviewResultSchemaID       *string `json:"preview_result_schema_id"`
	ApplyRequestSchemaID        *string `json:"apply_request_schema_id"`
	ApplyResultSchemaID         *string `json:"apply_result_schema_id"`
	ErrorSchemaID               string  `json:"error_schema_id"`
	ErrorTranslationID          *string `json:"error_translation_id"`
	CommitProtocolID            string  `json:"commit_protocol_id"`
	DefaultUnknownColumnPolicy  string  `json:"default_unknown_column_policy"`
	EntityBearingDefault        string  `json:"entity_bearing_default"`
	PublicProjectionDisposition string  `json:"public_projection_disposition"`
	RowSHA256                   string  `json:"row_sha256"`
}

type importTargetSourceManifest struct {
	SchemaID string                          `json:"schema_id"`
	Inputs   []importTargetSourceManifestRow `json:"inputs"`
}

type importTargetSourceManifestRow struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type importTargetAdapterDescriptorSet struct {
	SchemaID     string                          `json:"schema_id"`
	SourceSHA256 string                          `json:"source_sha256"`
	Adapters     []importTargetAdapterDescriptor `json:"adapters"`
}

type importTargetAdapterDescriptor struct {
	TargetID               string  `json:"target_id"`
	FacadeKind             string  `json:"facade_kind"`
	FacadeBindingID        *string `json:"facade_binding_id"`
	BindingSchemaID        *string `json:"binding_schema_id"`
	ContractMajor          *int    `json:"contract_major"`
	FacadeID               *string `json:"facade_id"`
	OwnerContractRef       string  `json:"owner_contract_ref"`
	CreateRequestSchemaID  *string `json:"create_request_schema_id"`
	CreateResultSchemaID   *string `json:"create_result_schema_id"`
	PreviewRequestSchemaID *string `json:"preview_request_schema_id"`
	PreviewResultSchemaID  *string `json:"preview_result_schema_id"`
	ApplyRequestSchemaID   *string `json:"apply_request_schema_id"`
	ApplyResultSchemaID    *string `json:"apply_result_schema_id"`
	ErrorSchemaID          string  `json:"error_schema_id"`
	ErrorTranslationID     *string `json:"error_translation_id"`
	CommitProtocolID       string  `json:"commit_protocol_id"`
}

type importTargetVerificationSet struct {
	SchemaID     string                           `json:"schema_id"`
	SourceSHA256 string                           `json:"source_sha256"`
	Targets      []importTargetVerificationTarget `json:"targets"`
}

type importTargetVerificationTarget struct {
	RegistryOrder    int    `json:"registry_order"`
	TargetID         string `json:"target_id"`
	VerificationID   string `json:"verification_id"`
	AvailabilityKind string `json:"availability_kind"`
}

type importTargetDerivedSet struct {
	SourceManifest importTargetSourceManifest
	Registry       importTargetRegistry
	Adapters       importTargetAdapterDescriptorSet
	Verification   importTargetVerificationSet
}

type declaredImportTargetInput[T any] struct {
	Path      string
	Canonical string
	Raw       any
	Value     T
}

func validateImportTargetAuthoredInput(relativePath string, value any) error {
	object, err := asObject(value, "contracts/imports/"+relativePath)
	if err != nil {
		return err
	}
	schemaID, err := requiredString(object, "schema_id", "contracts/imports/"+relativePath)
	if err != nil {
		return err
	}
	switch relativePath {
	case "index.json":
		if schemaID != "cartulary.import_target_registry_input_catalog.v1" {
			return fmt.Errorf("contracts/imports/index.json has unsupported schema_id %s", schemaID)
		}
	case "schemas.v1.json":
		if schemaID != "cartulary.imports_contract_schemas.v1" {
			return fmt.Errorf("contracts/imports/schemas.v1.json has unsupported schema_id %s", schemaID)
		}
	case "view-targets.v1.json":
		if schemaID != "cartulary.import_view_target_input_set.v1" {
			return fmt.Errorf("contracts/imports/view-targets.v1.json has unsupported schema_id %s", schemaID)
		}
	default:
		return fmt.Errorf("unexpected imports artifact %s", relativePath)
	}
	return nil
}

func validateImportTargetContractFamily(root string) error {
	_, err := materializeImportTargetArtifacts(root)
	return err
}

func deriveImportTargetArtifacts(root string) ([]artifact, error) {
	derived, err := materializeImportTargetArtifacts(root)
	if err != nil {
		return nil, err
	}
	sourceArtifact, err := importTargetArtifact(importTargetSourceManifestPath, derived.SourceManifest)
	if err != nil {
		return nil, err
	}
	registryArtifact, err := importTargetArtifact(importTargetRegistryPath, derived.Registry)
	if err != nil {
		return nil, err
	}
	adapterArtifact, err := importTargetArtifact(importTargetAdapterDescriptorPath, derived.Adapters)
	if err != nil {
		return nil, err
	}
	verificationArtifact, err := importTargetArtifact(importTargetVerificationPath, derived.Verification)
	if err != nil {
		return nil, err
	}
	integrity := map[string]any{
		"schema_id":                   "cartulary.import_target_registry_integrity.v1",
		"source_sha256":               derived.Registry.SourceSHA256,
		"registry_sha256":             registryArtifact.SHA256,
		"adapter_descriptors_sha256":  adapterArtifact.SHA256,
		"verification_targets_sha256": verificationArtifact.SHA256,
	}
	integrityArtifact, err := importTargetArtifact(importTargetIntegrityPath, integrity)
	if err != nil {
		return nil, err
	}
	return []artifact{
		sourceArtifact,
		registryArtifact,
		adapterArtifact,
		verificationArtifact,
		integrityArtifact,
	}, nil
}

func materializeImportTargetArtifacts(root string) (importTargetDerivedSet, error) {
	catalogInput, err := readDeclaredImportTargetInput[importTargetInputCatalog](
		root,
		"contracts/imports/index.json",
	)
	if err != nil {
		return importTargetDerivedSet{}, err
	}
	catalog := catalogInput.Value
	if err := validateImportTargetCatalog(catalog); err != nil {
		return importTargetDerivedSet{}, err
	}

	coreSchemaInput, err := readDeclaredImportTargetInput[map[string]any](root, catalog.CoreSchemaPath)
	if err != nil {
		return importTargetDerivedSet{}, err
	}
	viewTargetInput, err := readDeclaredImportTargetInput[importViewTargetInputSet](root, catalog.ViewTargetInputPath)
	if err != nil {
		return importTargetDerivedSet{}, err
	}
	viewSchemaInput, err := readDeclaredImportTargetInput[map[string]any](root, catalog.ViewSchemaRegistryPath)
	if err != nil {
		return importTargetDerivedSet{}, err
	}

	coreSchemaIDs := map[string]struct{}{}
	collectImportTargetSchemaIDs(coreSchemaInput.Raw, coreSchemaIDs)
	for _, required := range []string{
		"cartulary.imports.approved_view_mapping.v1",
		"cartulary.imports.owner_create_request.v1",
		"cartulary.imports.owner_create_result.v1",
		"cartulary.imports.owner_create_error.v1",
		"cartulary.imports.analytical_facade_binding.v1",
		"cartulary.import_target_registry.v1",
	} {
		if _, ok := coreSchemaIDs[required]; !ok {
			return importTargetDerivedSet{}, fmt.Errorf("core Imports schema input does not define %s", required)
		}
	}
	viewSchemaIDs, err := importTargetViewSchemaIDs(viewSchemaInput.Raw)
	if err != nil {
		return importTargetDerivedSet{}, err
	}

	ownerContracts := stringSliceSet(catalog.OwnerContracts)
	facades := make(map[string]importTargetFacadeContract, len(catalog.FacadeContracts))
	for _, facade := range catalog.FacadeContracts {
		facades[facade.FacadeID] = facade
	}
	commitProtocols := stringSliceSet(catalog.CommitProtocolIDs)
	errorTranslations := stringSliceSet(catalog.ErrorTranslationIDs)

	rowsByID := map[string]importTargetRegistryRow{}
	analyticalBindingsByTargetID := map[string]analyticalFacadeBinding{}
	if viewTargetInput.Value.SchemaID != "cartulary.import_view_target_input_set.v1" {
		return importTargetDerivedSet{}, fmt.Errorf("view-target input has unsupported schema_id %s", viewTargetInput.Value.SchemaID)
	}
	previousViewOrder := -1
	targetOrderIndex := make(map[string]int, len(catalog.TargetOrder))
	for index, targetID := range catalog.TargetOrder {
		targetOrderIndex[targetID] = index
	}
	for index, target := range viewTargetInput.Value.Targets {
		targetID := "view_schema:" + target.TargetViewSchemaID
		order, exists := targetOrderIndex[targetID]
		if !exists {
			return importTargetDerivedSet{}, fmt.Errorf("view target %s is absent from target_order", target.TargetViewSchemaID)
		}
		if order <= previousViewOrder {
			return importTargetDerivedSet{}, fmt.Errorf("view-target inputs must follow their relative target_order")
		}
		previousViewOrder = order
		if _, duplicate := rowsByID[targetID]; duplicate {
			return importTargetDerivedSet{}, fmt.Errorf("duplicate import target %s", targetID)
		}
		if _, exists := viewSchemaIDs[target.TargetViewSchemaID]; !exists {
			return importTargetDerivedSet{}, fmt.Errorf("view target %s is absent from the view-schema registry", target.TargetViewSchemaID)
		}
		if _, exists := ownerContracts[target.OwnerContractRef]; !exists {
			return importTargetDerivedSet{}, fmt.Errorf("view target %s references unknown owner %s", target.TargetViewSchemaID, target.OwnerContractRef)
		}
		if _, exists := coreSchemaIDs[target.MappingContractSchemaID]; !exists {
			return importTargetDerivedSet{}, fmt.Errorf("view target %s references unknown mapping schema %s", target.TargetViewSchemaID, target.MappingContractSchemaID)
		}
		if err := validateImportViewAvailability(target); err != nil {
			return importTargetDerivedSet{}, fmt.Errorf("view target %s: %w", target.TargetViewSchemaID, err)
		}
		if target.FacadeID != nil {
			facade, exists := facades[*target.FacadeID]
			if !exists || facade.FacadeKind != "owner_create" || facade.OwnerContractRef != target.OwnerContractRef {
				return importTargetDerivedSet{}, fmt.Errorf("view target %s has an unknown or mismatched owner facade", target.TargetViewSchemaID)
			}
		}
		targetSchemaID := target.TargetViewSchemaID
		createRequest := "cartulary.imports.owner_create_request.v1"
		createResult := "cartulary.imports.owner_create_result.v1"
		row := importTargetRegistryRow{
			RegistryOrder:               order,
			TargetID:                    targetID,
			TargetKind:                  "view_schema",
			TargetViewSchemaID:          &targetSchemaID,
			OwnerContractRef:            target.OwnerContractRef,
			SourceResourceFamily:        target.SourceResourceFamily,
			FacadeKind:                  "owner_create",
			FacadeBindingID:             cloneStringPointer(target.FacadeID),
			FacadeID:                    cloneStringPointer(target.FacadeID),
			AvailabilityKind:            target.AvailabilityKind,
			ActivationPolicy:            target.ActivationPolicy,
			MappingContractSchemaID:     target.MappingContractSchemaID,
			CreateRequestSchemaID:       &createRequest,
			CreateResultSchemaID:        &createResult,
			ErrorSchemaID:               "cartulary.imports.owner_create_error.v1",
			CommitProtocolID:            "cartulary.imports.unit_commit.v1",
			DefaultUnknownColumnPolicy:  target.DefaultUnknownColumnPolicy,
			EntityBearingDefault:        target.EntityBearingDefault,
			PublicProjectionDisposition: target.PublicProjectionDisposition,
		}
		if _, exists := commitProtocols[row.CommitProtocolID]; !exists {
			return importTargetDerivedSet{}, fmt.Errorf("view target %s references unknown commit protocol", target.TargetViewSchemaID)
		}
		rowsByID[targetID] = row
		_ = index
	}

	sourceInputs := map[string]declaredImportTargetSource{
		catalogInput.Path:    {Path: catalogInput.Path, Canonical: catalogInput.Canonical},
		coreSchemaInput.Path: {Path: coreSchemaInput.Path, Canonical: coreSchemaInput.Canonical},
		viewTargetInput.Path: {Path: viewTargetInput.Path, Canonical: viewTargetInput.Canonical},
		viewSchemaInput.Path: {Path: viewSchemaInput.Path, Canonical: viewSchemaInput.Canonical},
	}
	for _, input := range catalog.AnalyticalInputs {
		bindingInput, readErr := readDeclaredImportTargetInput[analyticalFacadeBinding](root, input.BindingPath)
		if readErr != nil {
			return importTargetDerivedSet{}, readErr
		}
		contributionInput, readErr := readDeclaredImportTargetInput[map[string]any](root, input.ExtensionContributionPath)
		if readErr != nil {
			return importTargetDerivedSet{}, readErr
		}
		for _, source := range []declaredImportTargetSource{
			{Path: bindingInput.Path, Canonical: bindingInput.Canonical},
			{Path: contributionInput.Path, Canonical: contributionInput.Canonical},
		} {
			sourceInputs[source.Path] = source
		}
		ownerSchemaIDs := map[string]struct{}{}
		for _, ownerSchemaPath := range input.OwnerSchemaPaths {
			ownerSchemaInput, readErr := readDeclaredImportTargetInput[map[string]any](
				root,
				ownerSchemaPath,
			)
			if readErr != nil {
				return importTargetDerivedSet{}, readErr
			}
			sourceInputs[ownerSchemaInput.Path] = declaredImportTargetSource{
				Path:      ownerSchemaInput.Path,
				Canonical: ownerSchemaInput.Canonical,
			}
			collectImportTargetSchemaIDs(ownerSchemaInput.Raw, ownerSchemaIDs)
		}
		binding := bindingInput.Value
		if err := validateAnalyticalFacadeBinding(
			binding,
			input,
			ownerContracts,
			facades,
			commitProtocols,
			errorTranslations,
			coreSchemaIDs,
			ownerSchemaIDs,
		); err != nil {
			return importTargetDerivedSet{}, fmt.Errorf("%s: %w", input.BindingPath, err)
		}
		facadeBindingID, err := resolveImportTargetContribution(
			contributionInput.Raw,
			binding.TargetKind,
			binding.ExtensionProfileID,
		)
		if err != nil {
			return importTargetDerivedSet{}, fmt.Errorf("%s: %w", input.ExtensionContributionPath, err)
		}
		targetID := binding.TargetKind + ":" + binding.ExtensionProfileID
		order, exists := targetOrderIndex[targetID]
		if !exists {
			return importTargetDerivedSet{}, fmt.Errorf("analytical target %s is absent from target_order", targetID)
		}
		if _, duplicate := rowsByID[targetID]; duplicate {
			return importTargetDerivedSet{}, fmt.Errorf("duplicate import target %s", targetID)
		}
		extensionProfileID := binding.ExtensionProfileID
		facadeID := binding.FacadeID
		previewRequest := binding.PreviewRequestSchemaID
		previewResult := binding.PreviewResultSchemaID
		applyRequest := binding.ApplyRequestSchemaID
		applyResult := binding.ApplyResultSchemaID
		errorTranslation := binding.ErrorTranslationID
		rowsByID[targetID] = importTargetRegistryRow{
			RegistryOrder:               order,
			TargetID:                    targetID,
			TargetKind:                  binding.TargetKind,
			ExtensionProfileID:          &extensionProfileID,
			OwnerContractRef:            binding.OwnerContractRef,
			SourceResourceFamily:        binding.TargetKind,
			FacadeKind:                  "owner_preview_apply",
			FacadeBindingID:             &facadeBindingID,
			FacadeID:                    &facadeID,
			AvailabilityKind:            input.AvailabilityKind,
			ActivationPolicy:            input.ActivationPolicy,
			MappingContractSchemaID:     binding.MappingSchemaID,
			PreviewRequestSchemaID:      &previewRequest,
			PreviewResultSchemaID:       &previewResult,
			ApplyRequestSchemaID:        &applyRequest,
			ApplyResultSchemaID:         &applyResult,
			ErrorSchemaID:               binding.ErrorSchemaID,
			ErrorTranslationID:          &errorTranslation,
			CommitProtocolID:            binding.CommitProtocolID,
			DefaultUnknownColumnPolicy:  input.DefaultUnknownColumnPolicy,
			EntityBearingDefault:        input.EntityBearingDefault,
			PublicProjectionDisposition: input.PublicProjectionDisposition,
		}
		analyticalBindingsByTargetID[targetID] = binding
	}

	if len(rowsByID) != len(catalog.TargetOrder) {
		return importTargetDerivedSet{}, fmt.Errorf(
			"target_order contains %d identities but inputs resolve %d",
			len(catalog.TargetOrder),
			len(rowsByID),
		)
	}
	rows := make([]importTargetRegistryRow, len(catalog.TargetOrder))
	for index, targetID := range catalog.TargetOrder {
		row, exists := rowsByID[targetID]
		if !exists {
			return importTargetDerivedSet{}, fmt.Errorf("target_order identity %s has no resolved input", targetID)
		}
		if row.RegistryOrder != index {
			return importTargetDerivedSet{}, fmt.Errorf("target %s registry order is not contiguous", targetID)
		}
		digest, err := importTargetRowDigest(row)
		if err != nil {
			return importTargetDerivedSet{}, err
		}
		row.RowSHA256 = digest
		rows[index] = row
	}

	sourcePaths := make([]string, 0, len(sourceInputs))
	for path := range sourceInputs {
		sourcePaths = append(sourcePaths, path)
	}
	sort.Strings(sourcePaths)
	sourceManifest := importTargetSourceManifest{
		SchemaID: "cartulary.import_target_registry_source_manifest.v1",
		Inputs:   make([]importTargetSourceManifestRow, 0, len(sourcePaths)),
	}
	for _, path := range sourcePaths {
		sum := sha256.Sum256([]byte(sourceInputs[path].Canonical))
		sourceManifest.Inputs = append(sourceManifest.Inputs, importTargetSourceManifestRow{
			Path:   path,
			SHA256: hex.EncodeToString(sum[:]),
		})
	}
	sourceCanonical, err := canonicalizeDecoded(sourceManifest)
	if err != nil {
		return importTargetDerivedSet{}, fmt.Errorf("canonicalize import target source manifest: %w", err)
	}
	sourceSum := sha256.Sum256([]byte(sourceCanonical))
	sourceSHA := hex.EncodeToString(sourceSum[:])
	registry := importTargetRegistry{
		SchemaID:     "cartulary.import_target_registry.v1",
		SourceSHA256: sourceSHA,
		Rows:         rows,
	}
	adapters := importTargetAdapterDescriptorSet{
		SchemaID:     "cartulary.import_target_adapter_descriptors.v1",
		SourceSHA256: sourceSHA,
		Adapters:     make([]importTargetAdapterDescriptor, 0, len(rows)),
	}
	verification := importTargetVerificationSet{
		SchemaID:     "cartulary.import_target_verification_projection.v1",
		SourceSHA256: sourceSHA,
		Targets:      make([]importTargetVerificationTarget, 0, len(rows)),
	}
	verificationIDs := map[string]string{}
	for _, row := range rows {
		var bindingSchemaID *string
		var contractMajor *int
		if binding, present := analyticalBindingsByTargetID[row.TargetID]; present {
			bindingSchemaID = cloneStringPointer(&binding.SchemaID)
			major := binding.ContractMajor
			contractMajor = &major
		}
		adapters.Adapters = append(adapters.Adapters, importTargetAdapterDescriptor{
			TargetID:               row.TargetID,
			FacadeKind:             row.FacadeKind,
			FacadeBindingID:        cloneStringPointer(row.FacadeBindingID),
			BindingSchemaID:        bindingSchemaID,
			ContractMajor:          contractMajor,
			FacadeID:               cloneStringPointer(row.FacadeID),
			OwnerContractRef:       row.OwnerContractRef,
			CreateRequestSchemaID:  cloneStringPointer(row.CreateRequestSchemaID),
			CreateResultSchemaID:   cloneStringPointer(row.CreateResultSchemaID),
			PreviewRequestSchemaID: cloneStringPointer(row.PreviewRequestSchemaID),
			PreviewResultSchemaID:  cloneStringPointer(row.PreviewResultSchemaID),
			ApplyRequestSchemaID:   cloneStringPointer(row.ApplyRequestSchemaID),
			ApplyResultSchemaID:    cloneStringPointer(row.ApplyResultSchemaID),
			ErrorSchemaID:          row.ErrorSchemaID,
			ErrorTranslationID:     cloneStringPointer(row.ErrorTranslationID),
			CommitProtocolID:       row.CommitProtocolID,
		})
		verificationID := importTargetVerificationID(row.TargetID)
		if prior, collision := verificationIDs[verificationID]; collision {
			return importTargetDerivedSet{}, fmt.Errorf("verification identity %s collides for %s and %s", verificationID, prior, row.TargetID)
		}
		verificationIDs[verificationID] = row.TargetID
		verification.Targets = append(verification.Targets, importTargetVerificationTarget{
			RegistryOrder:    row.RegistryOrder,
			TargetID:         row.TargetID,
			VerificationID:   verificationID,
			AvailabilityKind: row.AvailabilityKind,
		})
	}
	return importTargetDerivedSet{
		SourceManifest: sourceManifest,
		Registry:       registry,
		Adapters:       adapters,
		Verification:   verification,
	}, nil
}

type declaredImportTargetSource struct {
	Path      string
	Canonical string
}

func validateImportTargetCatalog(catalog importTargetInputCatalog) error {
	if catalog.Schema != contractDraft202012Schema {
		return fmt.Errorf("import target catalog must use JSON Schema draft 2020-12")
	}
	if catalog.SchemaID != "cartulary.import_target_registry_input_catalog.v1" {
		return fmt.Errorf("import target catalog has unsupported schema_id %s", catalog.SchemaID)
	}
	for label, path := range map[string]string{
		"core_schema_path":          catalog.CoreSchemaPath,
		"view_target_input_path":    catalog.ViewTargetInputPath,
		"view_schema_registry_path": catalog.ViewSchemaRegistryPath,
	} {
		if err := validateImportTargetDeclaredPath(path); err != nil {
			return fmt.Errorf("%s: %w", label, err)
		}
	}
	if len(catalog.TargetOrder) == 0 {
		return fmt.Errorf("target_order must not be empty")
	}
	if err := requireUniqueImportTargetStrings(catalog.TargetOrder, "target_order", false); err != nil {
		return err
	}
	if err := requireUniqueImportTargetStrings(catalog.OwnerContracts, "owner_contracts", true); err != nil {
		return err
	}
	ownerSet := stringSliceSet(catalog.OwnerContracts)
	previousFacade := ""
	for index, facade := range catalog.FacadeContracts {
		label := fmt.Sprintf("facade_contracts[%d]", index)
		if !importTargetIdentifierPattern.MatchString(facade.FacadeID) {
			return fmt.Errorf("%s.facade_id is invalid", label)
		}
		if previousFacade != "" && previousFacade >= facade.FacadeID {
			return fmt.Errorf("facade_contracts must be sorted and unique by facade_id")
		}
		previousFacade = facade.FacadeID
		if facade.FacadeKind != "owner_create" && facade.FacadeKind != "owner_preview_apply" {
			return fmt.Errorf("%s.facade_kind is invalid", label)
		}
		if _, exists := ownerSet[facade.OwnerContractRef]; !exists {
			return fmt.Errorf("%s references unknown owner %s", label, facade.OwnerContractRef)
		}
	}
	for index, input := range catalog.AnalyticalInputs {
		for label, path := range map[string]string{
			"binding_path":                input.BindingPath,
			"extension_contribution_path": input.ExtensionContributionPath,
		} {
			if err := validateImportTargetDeclaredPath(path); err != nil {
				return fmt.Errorf("analytical_inputs[%d].%s: %w", index, label, err)
			}
		}
		if len(input.OwnerSchemaPaths) == 0 {
			return fmt.Errorf("analytical_inputs[%d].owner_schema_paths must not be empty", index)
		}
		previousOwnerSchemaPath := ""
		for schemaIndex, path := range input.OwnerSchemaPaths {
			if err := validateImportTargetDeclaredPath(path); err != nil {
				return fmt.Errorf(
					"analytical_inputs[%d].owner_schema_paths[%d]: %w",
					index,
					schemaIndex,
					err,
				)
			}
			if previousOwnerSchemaPath != "" && previousOwnerSchemaPath >= path {
				return fmt.Errorf(
					"analytical_inputs[%d].owner_schema_paths must be sorted and unique",
					index,
				)
			}
			previousOwnerSchemaPath = path
		}
		if err := validateAnalyticalAvailability(input); err != nil {
			return fmt.Errorf("analytical_inputs[%d]: %w", index, err)
		}
	}
	if err := requireUniqueImportTargetStrings(catalog.ErrorTranslationIDs, "error_translation_ids", true); err != nil {
		return err
	}
	if err := requireUniqueImportTargetStrings(catalog.CommitProtocolIDs, "commit_protocol_ids", true); err != nil {
		return err
	}
	return nil
}

func validateImportViewAvailability(target importViewTargetInput) error {
	if target.TargetViewSchemaID == "" ||
		target.OwnerContractRef == "" ||
		target.SourceResourceFamily == "" ||
		target.MappingContractSchemaID == "" ||
		target.DefaultUnknownColumnPolicy == "" ||
		target.EntityBearingDefault == "" {
		return fmt.Errorf("required target metadata is empty")
	}
	switch target.AvailabilityKind {
	case "enabled":
		if target.FacadeID == nil || *target.FacadeID == "" ||
			target.ActivationPolicy != "always" ||
			target.PublicProjectionDisposition != "selectable" {
			return fmt.Errorf("enabled target requires an always-active facade and selectable projection")
		}
	case "reserved":
		if target.FacadeID != nil ||
			target.ActivationPolicy != "unavailable_reserved" ||
			target.PublicProjectionDisposition != "hidden_reserved" {
			return fmt.Errorf("reserved target must be unbound, unavailable, and hidden")
		}
	default:
		return fmt.Errorf("availability_kind is invalid")
	}
	return nil
}

func validateAnalyticalAvailability(input importTargetAnalyticalInput) error {
	if input.DefaultUnknownColumnPolicy == "" || input.EntityBearingDefault == "" {
		return fmt.Errorf("analytical semantic defaults are required")
	}
	switch input.AvailabilityKind {
	case "enabled":
		if input.ActivationPolicy != "always" || input.PublicProjectionDisposition != "selectable" {
			return fmt.Errorf("enabled analytical target must be always active and selectable")
		}
	case "claim_gated":
		if input.ActivationPolicy != "extension_claim_required" ||
			input.PublicProjectionDisposition != "extension_claim_gated" {
			return fmt.Errorf("claim-gated analytical target must require a claim and use claim-gated projection")
		}
	default:
		return fmt.Errorf("analytical availability_kind is invalid")
	}
	return nil
}

func validateAnalyticalFacadeBinding(
	binding analyticalFacadeBinding,
	input importTargetAnalyticalInput,
	ownerContracts map[string]struct{},
	facades map[string]importTargetFacadeContract,
	commitProtocols map[string]struct{},
	errorTranslations map[string]struct{},
	coreSchemaIDs map[string]struct{},
	ownerSchemaIDs map[string]struct{},
) error {
	if binding.SchemaID != "cartulary.imports.analytical_facade_binding.v1" {
		return fmt.Errorf("binding schema_id is invalid")
	}
	if _, exists := coreSchemaIDs[binding.SchemaID]; !exists {
		return fmt.Errorf("binding schema_id is not defined by Core Imports")
	}
	for label, value := range map[string]string{
		"target_kind":               binding.TargetKind,
		"extension_profile_id":      binding.ExtensionProfileID,
		"owner_contract_ref":        binding.OwnerContractRef,
		"facade_id":                 binding.FacadeID,
		"mapping_schema_id":         binding.MappingSchemaID,
		"preview_request_schema_id": binding.PreviewRequestSchemaID,
		"preview_result_schema_id":  binding.PreviewResultSchemaID,
		"apply_request_schema_id":   binding.ApplyRequestSchemaID,
		"apply_result_schema_id":    binding.ApplyResultSchemaID,
		"error_schema_id":           binding.ErrorSchemaID,
		"error_translation_id":      binding.ErrorTranslationID,
		"commit_protocol_id":        binding.CommitProtocolID,
	} {
		if !importTargetIdentifierPattern.MatchString(value) {
			return fmt.Errorf("%s is invalid", label)
		}
	}
	if binding.ContractMajor < 1 ||
		!strings.HasSuffix(binding.OwnerContractRef, "@"+strconv.Itoa(binding.ContractMajor)) {
		return fmt.Errorf("contract_major does not match owner_contract_ref")
	}
	if _, exists := ownerContracts[binding.OwnerContractRef]; !exists {
		return fmt.Errorf("binding references unknown owner %s", binding.OwnerContractRef)
	}
	facade, exists := facades[binding.FacadeID]
	if !exists || facade.FacadeKind != "owner_preview_apply" || facade.OwnerContractRef != binding.OwnerContractRef {
		return fmt.Errorf("binding facade is unknown or mismatched")
	}
	if _, exists := commitProtocols[binding.CommitProtocolID]; !exists {
		return fmt.Errorf("binding commit protocol is unknown")
	}
	if _, exists := errorTranslations[binding.ErrorTranslationID]; !exists {
		return fmt.Errorf("binding error translator is unknown")
	}
	for _, schemaID := range []string{
		binding.MappingSchemaID,
		binding.PreviewRequestSchemaID,
		binding.PreviewResultSchemaID,
		binding.ApplyRequestSchemaID,
		binding.ApplyResultSchemaID,
		binding.ErrorSchemaID,
	} {
		if _, exists := ownerSchemaIDs[schemaID]; !exists {
			return fmt.Errorf("binding references unresolved owner schema %s", schemaID)
		}
	}
	return validateAnalyticalAvailability(input)
}

func resolveImportTargetContribution(value any, targetKind, profileID string) (string, error) {
	object, err := asObject(value, "extension contribution input")
	if err != nil {
		return "", err
	}
	facts, err := objectArray(object["facts"], "extension contribution facts")
	if err != nil {
		return "", err
	}
	matches := []string{}
	for _, fact := range facts {
		if fact["fact_kind"] != "contribution" || fact["profile_id"] != profileID {
			continue
		}
		contribution, ok := fact["contribution"].(map[string]any)
		if !ok || contribution["kind"] != "import_target" || contribution["target_kind"] != targetKind {
			continue
		}
		bindingID, ok := contribution["facade_binding_id"].(string)
		if !ok || !importTargetIdentifierPattern.MatchString(bindingID) ||
			!strings.HasPrefix(bindingID, profileID+".") {
			return "", fmt.Errorf("import_target contribution has invalid facade_binding_id")
		}
		matches = append(matches, bindingID)
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("expected exactly one import_target contribution for %s/%s, got %d", profileID, targetKind, len(matches))
	}
	return matches[0], nil
}

func importTargetViewSchemaIDs(value any) (map[string]struct{}, error) {
	object, err := asObject(value, "view-schema registry")
	if err != nil {
		return nil, err
	}
	rows, err := objectArray(object["view_schemas"], "view-schema registry rows")
	if err != nil {
		return nil, err
	}
	ids := make(map[string]struct{}, len(rows))
	for index, row := range rows {
		id, err := requiredString(row, "view_schema_id", fmt.Sprintf("view_schemas[%d]", index))
		if err != nil {
			return nil, err
		}
		if _, duplicate := ids[id]; duplicate {
			return nil, fmt.Errorf("view-schema registry contains duplicate %s", id)
		}
		ids[id] = struct{}{}
	}
	return ids, nil
}

func collectImportTargetSchemaIDs(value any, ids map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if (key == "schema_id" || key == "x_schema_id") && child != nil {
				if schemaID, ok := child.(string); ok && schemaID != "" {
					ids[schemaID] = struct{}{}
				}
			}
			collectImportTargetSchemaIDs(child, ids)
		}
	case []any:
		for _, child := range typed {
			collectImportTargetSchemaIDs(child, ids)
		}
	}
}

func readDeclaredImportTargetInput[T any](root, path string) (declaredImportTargetInput[T], error) {
	var zero declaredImportTargetInput[T]
	if err := validateImportTargetDeclaredPath(path); err != nil {
		return zero, err
	}
	fullPath := filepath.Join(root, filepath.FromSlash(path))
	if !pathWithinRoot(root, fullPath) {
		return zero, fmt.Errorf("declared import-target input %s escapes the repository", path)
	}
	info, err := os.Lstat(fullPath)
	if err != nil {
		return zero, fmt.Errorf("stat declared import-target input %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return zero, fmt.Errorf("declared import-target input %s must be a regular file", path)
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return zero, fmt.Errorf("read declared import-target input %s: %w", path, err)
	}
	raw, err := decodeContract(data)
	if err != nil {
		return zero, fmt.Errorf("decode declared import-target input %s: %w", path, err)
	}
	canonical, err := canonicalizeDecoded(raw)
	if err != nil {
		return zero, fmt.Errorf("canonicalize declared import-target input %s: %w", path, err)
	}
	var value T
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return zero, fmt.Errorf("strictly decode declared import-target input %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return zero, fmt.Errorf("declared import-target input %s must contain one JSON document", path)
	}
	return declaredImportTargetInput[T]{
		Path:      path,
		Canonical: canonical,
		Raw:       raw,
		Value:     value,
	}, nil
}

func validateImportTargetDeclaredPath(path string) error {
	if path == "" ||
		filepath.ToSlash(filepath.Clean(path)) != path ||
		!strings.HasPrefix(path, "contracts/") ||
		!strings.HasSuffix(path, ".json") ||
		strings.Contains(path, "\\") ||
		strings.Contains(path, "/../") {
		return fmt.Errorf("declared path %q must be a normalized contracts/*.json path", path)
	}
	return nil
}

func requireUniqueImportTargetStrings(values []string, label string, sortedValues bool) error {
	if len(values) == 0 {
		return fmt.Errorf("%s must not be empty", label)
	}
	seen := map[string]struct{}{}
	for index, value := range values {
		if !importTargetIdentifierPattern.MatchString(value) {
			return fmt.Errorf("%s[%d] is invalid", label, index)
		}
		if _, duplicate := seen[value]; duplicate {
			return fmt.Errorf("%s contains duplicate %s", label, value)
		}
		if sortedValues && index > 0 && values[index-1] >= value {
			return fmt.Errorf("%s must be sorted and unique", label)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func stringSliceSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func importTargetRowDigest(row importTargetRegistryRow) (string, error) {
	encoded, err := json.Marshal(row)
	if err != nil {
		return "", fmt.Errorf("encode import-target row %s: %w", row.TargetID, err)
	}
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		return "", fmt.Errorf("decode import-target row %s: %w", row.TargetID, err)
	}
	delete(object, "row_sha256")
	canonical, err := canonicalizeDecoded(object)
	if err != nil {
		return "", fmt.Errorf("canonicalize import-target row %s: %w", row.TargetID, err)
	}
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:]), nil
}

func importTargetVerificationID(targetID string) string {
	var builder strings.Builder
	builder.WriteString("module.imports.target.")
	previousUnderscore := false
	for _, current := range targetID {
		isToken := (current >= 'a' && current <= 'z') || (current >= '0' && current <= '9')
		if isToken {
			builder.WriteRune(current)
			previousUnderscore = false
			continue
		}
		if !previousUnderscore {
			builder.WriteByte('_')
			previousUnderscore = true
		}
	}
	return strings.TrimSuffix(builder.String(), "_")
}

func importTargetArtifact(path string, value any) (artifact, error) {
	canonical, err := canonicalizeDecoded(value)
	if err != nil {
		return artifact{}, fmt.Errorf("canonicalize %s: %w", path, err)
	}
	sum := sha256.Sum256([]byte(canonical))
	return artifact{
		Path:   path,
		JSON:   canonical,
		SHA256: hex.EncodeToString(sum[:]),
	}, nil
}

func importTargetArtifactByPath(families []family, path string) (artifact, error) {
	for _, current := range families {
		if current.Dir != "imports" {
			continue
		}
		for _, currentArtifact := range current.Artifacts {
			if currentArtifact.Path == path {
				return currentArtifact, nil
			}
		}
	}
	return artifact{}, fmt.Errorf("missing generated import-target artifact %s", path)
}

func writeImportTargetRegistryGo(root string, families []family) error {
	registryArtifact, err := importTargetArtifactByPath(families, importTargetRegistryPath)
	if err != nil {
		return err
	}
	adapterArtifact, err := importTargetArtifactByPath(families, importTargetAdapterDescriptorPath)
	if err != nil {
		return err
	}
	verificationArtifact, err := importTargetArtifactByPath(families, importTargetVerificationPath)
	if err != nil {
		return err
	}
	var registry importTargetRegistry
	if err := json.Unmarshal([]byte(registryArtifact.JSON), &registry); err != nil {
		return fmt.Errorf("decode generated import-target registry for Go: %w", err)
	}
	var adapters importTargetAdapterDescriptorSet
	if err := json.Unmarshal([]byte(adapterArtifact.JSON), &adapters); err != nil {
		return fmt.Errorf("decode generated import-target adapters for Go: %w", err)
	}
	var verification importTargetVerificationSet
	if err := json.Unmarshal([]byte(verificationArtifact.JSON), &verification); err != nil {
		return fmt.Errorf("decode generated import-target verification projection for Go: %w", err)
	}

	var buffer bytes.Buffer
	buffer.WriteString("// Code generated by tools/contractgen from cartulary.import_target_registry.v1; DO NOT EDIT.\n\n")
	buffer.WriteString("package importtargetregistry\n\n")
	buffer.WriteString("type Target struct {\n")
	buffer.WriteString("\tRegistryOrder int\n\tTargetID string\n\tTargetKind string\n\tTargetViewSchemaID *string\n\tExtensionProfileID *string\n")
	buffer.WriteString("\tOwnerContractRef string\n\tSourceResourceFamily string\n\tFacadeKind string\n\tFacadeBindingID *string\n\tBindingSchemaID *string\n\tContractMajor *int\n\tFacadeID *string\n")
	buffer.WriteString("\tAvailabilityKind string\n\tActivationPolicy string\n\tMappingContractSchemaID string\n")
	buffer.WriteString("\tCreateRequestSchemaID *string\n\tCreateResultSchemaID *string\n\tPreviewRequestSchemaID *string\n\tPreviewResultSchemaID *string\n")
	buffer.WriteString("\tApplyRequestSchemaID *string\n\tApplyResultSchemaID *string\n\tErrorSchemaID string\n\tErrorTranslationID *string\n")
	buffer.WriteString("\tCommitProtocolID string\n\tDefaultUnknownColumnPolicy string\n\tEntityBearingDefault string\n")
	buffer.WriteString("\tPublicProjectionDisposition string\n\tRowSHA256 string\n}\n\n")
	buffer.WriteString("type AdapterDescriptor struct {\n\tTargetID string\n\tFacadeKind string\n\tFacadeBindingID *string\n\tBindingSchemaID *string\n\tContractMajor *int\n\tFacadeID *string\n")
	buffer.WriteString("\tOwnerContractRef string\n\tCreateRequestSchemaID *string\n\tCreateResultSchemaID *string\n")
	buffer.WriteString("\tPreviewRequestSchemaID *string\n\tPreviewResultSchemaID *string\n\tApplyRequestSchemaID *string\n\tApplyResultSchemaID *string\n")
	buffer.WriteString("\tErrorSchemaID string\n\tErrorTranslationID *string\n\tCommitProtocolID string\n}\n\n")
	buffer.WriteString("type VerificationTarget struct {\n\tRegistryOrder int\n\tTargetID string\n\tVerificationID string\n\tAvailabilityKind string\n}\n\n")
	fmt.Fprintf(&buffer, "const SourceSHA256 = %s\n", strconv.Quote(registry.SourceSHA256))
	fmt.Fprintf(&buffer, "const RegistrySHA256 = %s\n\n", strconv.Quote(registryArtifact.SHA256))
	buffer.WriteString("func stringPointer(value string) *string { return &value }\n")
	buffer.WriteString("func intPointer(value int) *int { return &value }\n\n")
	adaptersByTargetID := make(map[string]importTargetAdapterDescriptor, len(adapters.Adapters))
	for _, descriptor := range adapters.Adapters {
		adaptersByTargetID[descriptor.TargetID] = descriptor
	}
	buffer.WriteString("var Targets = []Target{\n")
	for _, row := range registry.Rows {
		descriptor, present := adaptersByTargetID[row.TargetID]
		if !present {
			return fmt.Errorf("generated import target %s has no adapter descriptor", row.TargetID)
		}
		buffer.WriteString("\t{\n")
		fmt.Fprintf(&buffer, "\t\tRegistryOrder: %d,\n", row.RegistryOrder)
		fmt.Fprintf(&buffer, "\t\tTargetID: %s,\n", strconv.Quote(row.TargetID))
		fmt.Fprintf(&buffer, "\t\tTargetKind: %s,\n", strconv.Quote(row.TargetKind))
		writeImportTargetGoPointer(&buffer, "TargetViewSchemaID", row.TargetViewSchemaID)
		writeImportTargetGoPointer(&buffer, "ExtensionProfileID", row.ExtensionProfileID)
		fmt.Fprintf(&buffer, "\t\tOwnerContractRef: %s,\n", strconv.Quote(row.OwnerContractRef))
		fmt.Fprintf(&buffer, "\t\tSourceResourceFamily: %s,\n", strconv.Quote(row.SourceResourceFamily))
		fmt.Fprintf(&buffer, "\t\tFacadeKind: %s,\n", strconv.Quote(row.FacadeKind))
		writeImportTargetGoPointer(&buffer, "FacadeBindingID", row.FacadeBindingID)
		writeImportTargetGoPointer(&buffer, "BindingSchemaID", descriptor.BindingSchemaID)
		writeImportTargetGoIntPointer(&buffer, "ContractMajor", descriptor.ContractMajor)
		writeImportTargetGoPointer(&buffer, "FacadeID", row.FacadeID)
		fmt.Fprintf(&buffer, "\t\tAvailabilityKind: %s,\n", strconv.Quote(row.AvailabilityKind))
		fmt.Fprintf(&buffer, "\t\tActivationPolicy: %s,\n", strconv.Quote(row.ActivationPolicy))
		fmt.Fprintf(&buffer, "\t\tMappingContractSchemaID: %s,\n", strconv.Quote(row.MappingContractSchemaID))
		writeImportTargetGoPointer(&buffer, "CreateRequestSchemaID", row.CreateRequestSchemaID)
		writeImportTargetGoPointer(&buffer, "CreateResultSchemaID", row.CreateResultSchemaID)
		writeImportTargetGoPointer(&buffer, "PreviewRequestSchemaID", row.PreviewRequestSchemaID)
		writeImportTargetGoPointer(&buffer, "PreviewResultSchemaID", row.PreviewResultSchemaID)
		writeImportTargetGoPointer(&buffer, "ApplyRequestSchemaID", row.ApplyRequestSchemaID)
		writeImportTargetGoPointer(&buffer, "ApplyResultSchemaID", row.ApplyResultSchemaID)
		fmt.Fprintf(&buffer, "\t\tErrorSchemaID: %s,\n", strconv.Quote(row.ErrorSchemaID))
		writeImportTargetGoPointer(&buffer, "ErrorTranslationID", row.ErrorTranslationID)
		fmt.Fprintf(&buffer, "\t\tCommitProtocolID: %s,\n", strconv.Quote(row.CommitProtocolID))
		fmt.Fprintf(&buffer, "\t\tDefaultUnknownColumnPolicy: %s,\n", strconv.Quote(row.DefaultUnknownColumnPolicy))
		fmt.Fprintf(&buffer, "\t\tEntityBearingDefault: %s,\n", strconv.Quote(row.EntityBearingDefault))
		fmt.Fprintf(&buffer, "\t\tPublicProjectionDisposition: %s,\n", strconv.Quote(row.PublicProjectionDisposition))
		fmt.Fprintf(&buffer, "\t\tRowSHA256: %s,\n", strconv.Quote(row.RowSHA256))
		buffer.WriteString("\t},\n")
	}
	buffer.WriteString("}\n\n")
	buffer.WriteString("var AdapterDescriptors = []AdapterDescriptor{\n")
	for _, descriptor := range adapters.Adapters {
		buffer.WriteString("\t{\n")
		fmt.Fprintf(&buffer, "\t\tTargetID: %s,\n", strconv.Quote(descriptor.TargetID))
		fmt.Fprintf(&buffer, "\t\tFacadeKind: %s,\n", strconv.Quote(descriptor.FacadeKind))
		writeImportTargetGoPointer(&buffer, "FacadeBindingID", descriptor.FacadeBindingID)
		writeImportTargetGoPointer(&buffer, "BindingSchemaID", descriptor.BindingSchemaID)
		writeImportTargetGoIntPointer(&buffer, "ContractMajor", descriptor.ContractMajor)
		writeImportTargetGoPointer(&buffer, "FacadeID", descriptor.FacadeID)
		fmt.Fprintf(&buffer, "\t\tOwnerContractRef: %s,\n", strconv.Quote(descriptor.OwnerContractRef))
		writeImportTargetGoPointer(&buffer, "CreateRequestSchemaID", descriptor.CreateRequestSchemaID)
		writeImportTargetGoPointer(&buffer, "CreateResultSchemaID", descriptor.CreateResultSchemaID)
		writeImportTargetGoPointer(&buffer, "PreviewRequestSchemaID", descriptor.PreviewRequestSchemaID)
		writeImportTargetGoPointer(&buffer, "PreviewResultSchemaID", descriptor.PreviewResultSchemaID)
		writeImportTargetGoPointer(&buffer, "ApplyRequestSchemaID", descriptor.ApplyRequestSchemaID)
		writeImportTargetGoPointer(&buffer, "ApplyResultSchemaID", descriptor.ApplyResultSchemaID)
		fmt.Fprintf(&buffer, "\t\tErrorSchemaID: %s,\n", strconv.Quote(descriptor.ErrorSchemaID))
		writeImportTargetGoPointer(&buffer, "ErrorTranslationID", descriptor.ErrorTranslationID)
		fmt.Fprintf(&buffer, "\t\tCommitProtocolID: %s,\n", strconv.Quote(descriptor.CommitProtocolID))
		buffer.WriteString("\t},\n")
	}
	buffer.WriteString("}\n\n")
	buffer.WriteString("var VerificationTargets = []VerificationTarget{\n")
	for _, target := range verification.Targets {
		fmt.Fprintf(
			&buffer,
			"\t{RegistryOrder: %d, TargetID: %s, VerificationID: %s, AvailabilityKind: %s},\n",
			target.RegistryOrder,
			strconv.Quote(target.TargetID),
			strconv.Quote(target.VerificationID),
			strconv.Quote(target.AvailabilityKind),
		)
	}
	buffer.WriteString("}\n")
	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("format generated import-target Go registry: %w", err)
	}
	return stageGeneratedFile(
		filepath.Join(root, "internal", "gen", "importtargetregistry", "registry_gen.go"),
		formatted,
		0o644,
	)
}

func writeImportTargetGoPointer(buffer *bytes.Buffer, field string, value *string) {
	if value == nil {
		return
	}
	fmt.Fprintf(buffer, "\t\t%s: stringPointer(%s),\n", field, strconv.Quote(*value))
}

func writeImportTargetGoIntPointer(buffer *bytes.Buffer, field string, value *int) {
	if value == nil {
		return
	}
	fmt.Fprintf(buffer, "\t\t%s: intPointer(%d),\n", field, *value)
}

func writeImportTargetRegistryTypeScript(root string, families []family) error {
	registryArtifact, err := importTargetArtifactByPath(families, importTargetRegistryPath)
	if err != nil {
		return err
	}
	var registry importTargetRegistry
	if err := json.Unmarshal([]byte(registryArtifact.JSON), &registry); err != nil {
		return fmt.Errorf("decode generated import-target registry for TypeScript: %w", err)
	}
	targets := make([]map[string]any, 0, len(registry.Rows))
	for _, row := range registry.Rows {
		targets = append(targets, map[string]any{
			"registry_order":                row.RegistryOrder,
			"target_id":                     row.TargetID,
			"target_kind":                   row.TargetKind,
			"target_view_schema_id":         row.TargetViewSchemaID,
			"extension_profile_id":          row.ExtensionProfileID,
			"owner_contract_ref":            row.OwnerContractRef,
			"availability_kind":             row.AvailabilityKind,
			"activation_policy":             row.ActivationPolicy,
			"mapping_contract_schema_id":    row.MappingContractSchemaID,
			"default_unknown_column_policy": row.DefaultUnknownColumnPolicy,
			"entity_bearing_default":        row.EntityBearingDefault,
			"public_projection_disposition": row.PublicProjectionDisposition,
			"row_sha256":                    row.RowSHA256,
		})
	}
	projection := map[string]any{
		"schema_id":       "cartulary.import_target_frontend_projection.v1",
		"source_sha256":   registry.SourceSHA256,
		"registry_sha256": registryArtifact.SHA256,
		"targets":         targets,
	}
	encoded, err := json.MarshalIndent(projection, "", "  ")
	if err != nil {
		return fmt.Errorf("encode generated import-target TypeScript projection: %w", err)
	}
	source := `// Code generated by tools/contractgen from cartulary.import_target_registry.v1; DO NOT EDIT.

export type ImportTargetFrontendDisposition =
  | "selectable"
  | "hidden_reserved"
  | "extension_claim_gated";

export type ImportTargetFrontendRow = {
  readonly registry_order: number;
  readonly target_id: string;
  readonly target_kind: "view_schema" | "network_flow_table";
  readonly target_view_schema_id: string | null;
  readonly extension_profile_id: string | null;
  readonly owner_contract_ref: string;
  readonly availability_kind: "enabled" | "reserved" | "claim_gated";
  readonly activation_policy:
    | "always"
    | "unavailable_reserved"
    | "extension_claim_required";
  readonly mapping_contract_schema_id: string;
  readonly default_unknown_column_policy: string;
  readonly entity_bearing_default: string;
  readonly public_projection_disposition: ImportTargetFrontendDisposition;
  readonly row_sha256: string;
};

export type ImportTargetFrontendProjection = {
  readonly schema_id: "cartulary.import_target_frontend_projection.v1";
  readonly source_sha256: string;
  readonly registry_sha256: string;
  readonly targets: readonly ImportTargetFrontendRow[];
};

export const importTargetRegistry = ` + string(encoded) + ` as const satisfies ImportTargetFrontendProjection;
`
	return stageGeneratedFile(
		filepath.Join(root, "packages", "protocol-ts", "src", "generated", "import-target-registry.ts"),
		[]byte(source),
		0o644,
	)
}
