package reference_data

import (
	"context"
	"time"
)

// AdministrativeVersion projects current administrative state. Successful
// metadata is nullable until a successful envelope exists; failed attempts
// never manufacture digests, signatures or a verification timestamp.
type AdministrativeVersion struct {
	PackKey                  string                     `json:"pack_key"`
	PackKind                 string                     `json:"pack_kind"`
	PackVersion              string                     `json:"pack_version"`
	Condition                string                     `json:"pack_version_state"`
	Health                   string                     `json:"health"`
	AdministrativelyDisabled bool                       `json:"administratively_disabled"`
	DistributionKind         string                     `json:"distribution_kind"`
	Active                   bool                       `json:"active"`
	PendingWork              bool                       `json:"pending_work"`
	ReproducibilityPinned    bool                       `json:"reproducibility_pinned"`
	FallbackFromVersion      *string                    `json:"fallback_from_version"`
	Dependencies             []AdministrativeDependency `json:"dependencies"`
	Removed                  bool                       `json:"removed"`
	MissingReason            *string                    `json:"missing_reason"`
	LastFailureCode          *string                    `json:"last_failure_code"`
	SourceIdentifier         *string                    `json:"source_identifier"`
	ManifestSHA256           *string                    `json:"manifest_sha256"`
	PayloadSHA256            *string                    `json:"payload_sha256"`
	PackContractVersion      *string                    `json:"pack_contract_version"`
	PackReleaseSequence      *int64                     `json:"pack_release_sequence"`
	ContentProfileID         *string                    `json:"content_profile_id"`
	ContentProfileVersion    *string                    `json:"content_profile_version"`
	SourceProfileID          *string                    `json:"source_profile_id"`
	SourceProfileSHA256      *string                    `json:"source_profile_sha256"`
	SourceVersion            *string                    `json:"source_version"`
	SourceAsOf               *string                    `json:"source_as_of"`
	LicenseExpression        *string                    `json:"license_expression"`
	Redistribution           *string                    `json:"redistribution"`
	TrustRepositoryID        *string                    `json:"trust_repository_id"`
	VerificationMethod       *string                    `json:"verification_method"`
	LastVerifiedAt           *time.Time                 `json:"last_verified_at"`
	TrustValidUntil          *time.Time                 `json:"trust_valid_until"`
	VerifiedSignerKeyIDs     []string                   `json:"verified_signer_key_ids"`
	ImportedByUserID         *string                    `json:"imported_by_user_id"`
	ImportedAt               time.Time                  `json:"imported_at"`
	PreviousActiveVersion    *string                    `json:"previous_active_version"`
	ActivatedByUserID        *string                    `json:"activated_by_user_id"`
	ActivatedAt              *time.Time                 `json:"activated_at"`
}

type AdministrativeDependency struct {
	PackKey       string `json:"pack_key"`
	PackVersion   string `json:"pack_version"`
	PayloadSHA256 string `json:"payload_sha256"`
}

// AdministrativeApplication is the semantic surface used by transport adapters.
// Application assembly constructs it and registers its worker before binding routes.
type AdministrativeApplication interface {
	ImportAdmission
	ListVersions(context.Context) ([]AdministrativeVersion, error)
	GetVersion(context.Context, string, string) (AdministrativeVersion, error)
	GetValidationSummary(context.Context, string) (*ValidationSummary, error)
	VerifyRetained(context.Context, VerificationRequest) (JobAcceptedResult, error)
	Activate(context.Context, ActionParams) (ActionResult, error)
	Disable(context.Context, ActionParams) (ActionResult, error)
	Remove(context.Context, ActionParams) (ActionResult, error)
}
