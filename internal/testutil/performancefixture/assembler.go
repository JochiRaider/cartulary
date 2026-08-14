package performancefixture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

const SemanticReceiptSchemaID = "cartulary.performance_fixture_semantic_receipt.v1"

type Descriptor struct {
	ContributionID string         `json:"contribution_id"`
	Version        string         `json:"version"`
	OwnerID        string         `json:"owner_id"`
	Dependencies   []string       `json:"dependencies"`
	ExpectedCounts map[string]int `json:"expected_counts"`
}

type Receipt struct {
	ContributionID string         `json:"contribution_id"`
	Version        string         `json:"version"`
	OwnerID        string         `json:"owner_id"`
	Counts         map[string]int `json:"counts"`
}

type BuildState struct {
	SnapshotKey       string
	Seed              int
	RuntimeBundle     RuntimeBundle
	BackgroundUserIDs []string
	IncidentID        string
}

type Contribution interface {
	Descriptor() Descriptor
	Apply(context.Context, *BuildState) (Receipt, error)
}

type SemanticValidation struct {
	Counts                   map[string]int `json:"counts"`
	RelationshipDistribution bool           `json:"relationship_distribution"`
	DefaultView              bool           `json:"default_view"`
	Authorization            bool           `json:"authorization"`
	ProjectionReadiness      bool           `json:"projection_readiness"`
	NoActiveSessions         bool           `json:"no_active_sessions"`
}

type Validator interface {
	Validate(context.Context, *BuildState) (SemanticValidation, error)
}

type Result struct {
	Receipts                 []Receipt
	Validation               SemanticValidation
	SemanticValidationDigest string
}

type Assembler struct {
	contributions    []Contribution
	expected         []Descriptor
	expectedSemantic map[string]int
	validator        Validator
}

func NewAssembler(expected []Descriptor, expectedSemantic map[string]int, validator Validator, contributions ...Contribution) (*Assembler, error) {
	if validator == nil {
		return nil, errors.New("performance fixture semantic validator is required")
	}
	if err := validateDescriptorClosure(expected); err != nil {
		return nil, fmt.Errorf("validate expected performance fixture contribution closure: %w", err)
	}
	if len(contributions) != len(expected) {
		return nil, fmt.Errorf("performance fixture contribution count mismatch: got %d want %d", len(contributions), len(expected))
	}
	for index, contribution := range contributions {
		if contribution == nil {
			return nil, fmt.Errorf("performance fixture contribution %d is nil", index)
		}
		got := normalizeDescriptor(contribution.Descriptor())
		want := normalizeDescriptor(expected[index])
		if !reflect.DeepEqual(got, want) {
			return nil, fmt.Errorf("performance fixture contribution %d descriptor mismatch: got %#v want %#v", index, got, want)
		}
	}
	if len(expectedSemantic) == 0 {
		return nil, errors.New("performance fixture semantic expectations are required")
	}
	return &Assembler{
		contributions:    slices.Clone(contributions),
		expected:         cloneDescriptors(expected),
		expectedSemantic: maps.Clone(expectedSemantic),
		validator:        validator,
	}, nil
}

func (a *Assembler) Assemble(ctx context.Context, state *BuildState) (Result, error) {
	if a == nil {
		return Result{}, errors.New("performance fixture assembler is nil")
	}
	if state == nil {
		return Result{}, errors.New("performance fixture build state is required")
	}
	if err := validateSnapshotKey(state.SnapshotKey); err != nil {
		return Result{}, err
	}
	if state.Seed <= 0 {
		return Result{}, errors.New("performance fixture seed must be positive")
	}
	receipts := make([]Receipt, 0, len(a.contributions))
	for index, contribution := range a.contributions {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		receipt, err := contribution.Apply(ctx, state)
		if err != nil {
			return Result{}, fmt.Errorf("apply performance fixture contribution %s: %w", a.expected[index].ContributionID, err)
		}
		if err := validateReceipt(receipt, a.expected[index]); err != nil {
			return Result{}, err
		}
		receipts = append(receipts, cloneReceipt(receipt))
	}
	validation, err := a.validator.Validate(ctx, state)
	if err != nil {
		return Result{}, fmt.Errorf("validate performance fixture semantics: %w", err)
	}
	if err := validateSemanticResult(validation, a.expectedSemantic); err != nil {
		return Result{}, err
	}
	digest, err := semanticDigest(receipts, validation)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Receipts:                 receipts,
		Validation:               cloneSemanticValidation(validation),
		SemanticValidationDigest: digest,
	}, nil
}

func validateDescriptorClosure(descriptors []Descriptor) error {
	if len(descriptors) == 0 {
		return errors.New("closed contribution set is empty")
	}
	indexes := make(map[string]int, len(descriptors))
	for index, descriptor := range descriptors {
		descriptor = normalizeDescriptor(descriptor)
		if descriptor.ContributionID == "" || descriptor.Version == "" || descriptor.OwnerID == "" {
			return fmt.Errorf("contribution %d has an incomplete identity", index)
		}
		if descriptor.ContributionID != descriptor.Version {
			return fmt.Errorf("contribution %s version %s must equal its adopted contribution identity", descriptor.ContributionID, descriptor.Version)
		}
		if _, exists := indexes[descriptor.ContributionID]; exists {
			return fmt.Errorf("duplicate contribution %s", descriptor.ContributionID)
		}
		if len(descriptor.ExpectedCounts) == 0 {
			return fmt.Errorf("contribution %s has no expected receipt counts", descriptor.ContributionID)
		}
		for key, value := range descriptor.ExpectedCounts {
			if strings.TrimSpace(key) == "" || value < 0 {
				return fmt.Errorf("contribution %s has invalid expected count %q=%d", descriptor.ContributionID, key, value)
			}
		}
		indexes[descriptor.ContributionID] = index
	}
	for index, descriptor := range descriptors {
		seen := map[string]struct{}{}
		for _, dependency := range descriptor.Dependencies {
			if _, duplicate := seen[dependency]; duplicate {
				return fmt.Errorf("contribution %s duplicates dependency %s", descriptor.ContributionID, dependency)
			}
			seen[dependency] = struct{}{}
			dependencyIndex, exists := indexes[dependency]
			if !exists {
				return fmt.Errorf("contribution %s depends on unknown contribution %s", descriptor.ContributionID, dependency)
			}
			if dependencyIndex >= index {
				return fmt.Errorf("contribution %s dependency %s is cyclic or out of order", descriptor.ContributionID, dependency)
			}
		}
	}
	return nil
}

func validateReceipt(receipt Receipt, descriptor Descriptor) error {
	if receipt.ContributionID != descriptor.ContributionID || receipt.Version != descriptor.Version || receipt.OwnerID != descriptor.OwnerID {
		return fmt.Errorf("performance fixture receipt identity mismatch for %s", descriptor.ContributionID)
	}
	if !reflect.DeepEqual(receipt.Counts, descriptor.ExpectedCounts) {
		return fmt.Errorf("performance fixture receipt counts mismatch for %s: got %#v want %#v", descriptor.ContributionID, receipt.Counts, descriptor.ExpectedCounts)
	}
	return nil
}

func validateSemanticResult(validation SemanticValidation, expected map[string]int) error {
	if !reflect.DeepEqual(validation.Counts, expected) {
		return fmt.Errorf("performance fixture semantic counts mismatch: got %#v want %#v", validation.Counts, expected)
	}
	if !validation.RelationshipDistribution || !validation.DefaultView || !validation.Authorization || !validation.ProjectionReadiness || !validation.NoActiveSessions {
		return fmt.Errorf("performance fixture semantic predicates failed: relationships=%t default_view=%t authorization=%t projection_readiness=%t no_active_sessions=%t",
			validation.RelationshipDistribution,
			validation.DefaultView,
			validation.Authorization,
			validation.ProjectionReadiness,
			validation.NoActiveSessions,
		)
	}
	return nil
}

func semanticDigest(receipts []Receipt, validation SemanticValidation) (string, error) {
	payload, err := json.Marshal(struct {
		SchemaID   string             `json:"schema_id"`
		Receipts   []Receipt          `json:"contribution_receipts"`
		Validation SemanticValidation `json:"validation"`
	}{
		SchemaID:   SemanticReceiptSchemaID,
		Receipts:   receipts,
		Validation: validation,
	})
	if err != nil {
		return "", fmt.Errorf("encode performance fixture semantic receipt: %w", err)
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func validateSnapshotKey(value string) error {
	if len(value) != 64 || strings.ToLower(value) != value {
		return errors.New("performance fixture snapshot key must be 64 lowercase hexadecimal characters")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return errors.New("performance fixture snapshot key must be 64 lowercase hexadecimal characters")
	}
	return nil
}

func normalizeDescriptor(descriptor Descriptor) Descriptor {
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return descriptor
}

func cloneDescriptors(descriptors []Descriptor) []Descriptor {
	result := make([]Descriptor, len(descriptors))
	for index, descriptor := range descriptors {
		result[index] = normalizeDescriptor(descriptor)
	}
	return result
}

func cloneReceipt(receipt Receipt) Receipt {
	receipt.Counts = maps.Clone(receipt.Counts)
	return receipt
}

func cloneSemanticValidation(validation SemanticValidation) SemanticValidation {
	validation.Counts = maps.Clone(validation.Counts)
	return validation
}
