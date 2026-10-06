package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

const (
	RestoreTargetMarkerSchemaID         = "cartulary.restore_target_marker.v5"
	RestoreTargetMarkerMaximumBytes     = int64(65536)
	RestoreTargetGenerationMaximumBytes = int64(64)
	RestoreTargetMarkerMaximumLifetime  = 24 * time.Hour
	RestoreTargetServingLeaseAcquireMax = time.Second

	RestoreTargetPurpose             = "restore_target"
	RestoreVerificationTargetPurpose = "restore_verification_target"
)

var ErrTargetServingLeaseLost = errors.New("restore target serving lease lost")
var errTargetMarkerExpired = errors.New("restore target marker has expired")

type TargetMarkerMaterial struct {
	MarkerBody     []byte
	GenerationBody []byte
}

type TargetMarkerReader func(bindingKind string, rootPath string) (TargetMarkerMaterial, error)

// TargetMarkerWriter publishes owner-issued material under the target serving lease.
type TargetMarkerWriter func(context.Context, string, TargetMarkerMaterial, TargetMarkerMaterial) error

type TargetBindingDigests struct {
	DatabaseSHA256             string `json:"database_sha256"`
	ObjectStoreSHA256          string `json:"object_store_sha256"`
	ReferencePackStorageSHA256 string `json:"reference_pack_storage_sha256"`
	ExportOutputsSHA256        string `json:"export_outputs_sha256"`
}

type RestoreTargetMarker struct {
	ApplicationCryptoFormat string               `json:"application_crypto_format"`
	SchemaID                string               `json:"schema_id"`
	Purpose                 string               `json:"purpose"`
	TargetGenerationID      string               `json:"target_generation_id"`
	BindingDigests          TargetBindingDigests `json:"binding_digests"`
	IssuedAt                string               `json:"issued_at"`
	ExpiresAt               string               `json:"expires_at"`
}

type TargetServingAdmission interface {
	Context() context.Context
	AssertHeld() error
	Release(context.Context) error
}

type TargetServingAdmissionFactory func(context.Context, PostgresPool, time.Duration, time.Duration) (TargetServingAdmission, error)

func TargetBindingDigestsFor(deployment Deployment) TargetBindingDigests {
	return TargetBindingDigests{
		DatabaseSHA256:             bindingDigest(databaseTargetIdentity(deployment)),
		ObjectStoreSHA256:          bindingDigest(objectTargetIdentity(deployment)),
		ReferencePackStorageSHA256: bindingDigest(rootBindingBasis(deployment.ReferencePackStorage)),
		ExportOutputsSHA256:        bindingDigest(rootBindingBasis(deployment.ExportOutputs)),
	}
}

func ValidateRestoreTargetMarker(material TargetMarkerMaterial, purpose string, expected TargetBindingDigests, now time.Time) error {
	_, err := AdmitRestoreTargetMarker(material, purpose, expected, now)
	return err
}

func AdmitRestoreTargetMarker(material TargetMarkerMaterial, purpose string, expected TargetBindingDigests, now time.Time) (uuid.UUID, error) {
	if purpose != RestoreTargetPurpose && purpose != RestoreVerificationTargetPurpose {
		return uuid.Nil, fmt.Errorf("restore target marker purpose %q is invalid", purpose)
	}
	if err := strictjson.ValidateObject(material.MarkerBody); err != nil {
		return uuid.Nil, fmt.Errorf("restore target marker is not strict JSON: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(material.MarkerBody))
	decoder.DisallowUnknownFields()
	var marker RestoreTargetMarker
	if err := decoder.Decode(&marker); err != nil {
		return uuid.Nil, fmt.Errorf("decode restore target marker: %w", err)
	}
	if marker.ApplicationCryptoFormat != recovery.ApplicationCryptoFormatID || marker.SchemaID != RestoreTargetMarkerSchemaID || marker.Purpose != purpose {
		return uuid.Nil, errors.New("restore target marker has the wrong schema or purpose")
	}
	generationText := strings.TrimSpace(string(material.GenerationBody))
	generationID, err := uuid.Parse(generationText)
	if err != nil || generationID == uuid.Nil || generationText != generationID.String() {
		return uuid.Nil, errors.New("restore target generation proof is invalid")
	}
	markerGenerationID, err := uuid.Parse(marker.TargetGenerationID)
	if err != nil || markerGenerationID == uuid.Nil ||
		marker.TargetGenerationID != markerGenerationID.String() ||
		markerGenerationID != generationID {
		return uuid.Nil, errors.New("restore target marker has the wrong target generation")
	}
	if marker.BindingDigests != expected ||
		!isLowerSHA256(marker.BindingDigests.DatabaseSHA256) ||
		!isLowerSHA256(marker.BindingDigests.ObjectStoreSHA256) || !isLowerSHA256(marker.BindingDigests.ReferencePackStorageSHA256) || !isLowerSHA256(marker.BindingDigests.ExportOutputsSHA256) {
		return uuid.Nil, errors.New("restore target marker has the wrong target binding")
	}
	issuedAt, err := parseCanonicalMarkerTime(marker.IssuedAt)
	if err != nil {
		return uuid.Nil, fmt.Errorf("restore target marker issued_at: %w", err)
	}
	expiresAt, err := parseCanonicalMarkerTime(marker.ExpiresAt)
	if err != nil {
		return uuid.Nil, fmt.Errorf("restore target marker expires_at: %w", err)
	}
	now = now.UTC()
	lifetime := expiresAt.Sub(issuedAt)
	if lifetime <= 0 || lifetime > RestoreTargetMarkerMaximumLifetime {
		return uuid.Nil, errors.New("restore target marker lifetime is invalid")
	}
	if issuedAt.After(now) {
		return uuid.Nil, errors.New("restore target marker is not currently valid")
	}
	if !expiresAt.After(now) {
		return markerGenerationID, errTargetMarkerExpired
	}
	return markerGenerationID, nil
}

func bindingDigest(identity string) string {
	digest := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(digest[:])
}

func isLowerSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && hex.EncodeToString(decoded) == value
}

func parseCanonicalMarkerTime(value string) (time.Time, error) {
	if !strings.HasSuffix(value, "Z") {
		return time.Time{}, errors.New("timestamp must use UTC Z form")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	if parsed.UTC().Format(time.RFC3339Nano) != value {
		return time.Time{}, errors.New("timestamp is not canonical")
	}
	return parsed.UTC(), nil
}

// prepareTargetMarker preserves admitted generations and only issues or renews a
// proof after the owner has inspected pristine state under exclusive admission.
func prepareTargetMarker(material TargetMarkerMaterial, readErr error, purpose string, expected TargetBindingDigests, now time.Time, pristine func() error, publish func(TargetMarkerMaterial, TargetMarkerMaterial) error) (uuid.UUID, error) {
	var generation uuid.UUID
	if readErr == nil {
		var err error
		generation, err = AdmitRestoreTargetMarker(material, purpose, expected, now)
		if err == nil {
			return generation, nil
		}
		if !errors.Is(err, errTargetMarkerExpired) {
			return uuid.Nil, err
		}
	} else {
		if !errors.Is(readErr, os.ErrNotExist) || len(material.MarkerBody) != 0 {
			return uuid.Nil, readErr
		}
		if len(material.GenerationBody) != 0 {
			value := strings.TrimSpace(string(material.GenerationBody))
			var err error
			generation, err = uuid.Parse(value)
			if err != nil || generation == uuid.Nil || generation.String() != value {
				return uuid.Nil, errors.New("invalid incomplete target generation")
			}
		}
	}
	if pristine == nil || publish == nil {
		return uuid.Nil, errors.New("target proof preparation unavailable")
	}
	if err := pristine(); err != nil {
		return uuid.Nil, err
	}
	if generation == uuid.Nil {
		generation = uuid.New()
	}
	marker := RestoreTargetMarker{ApplicationCryptoFormat: recovery.ApplicationCryptoFormatID, SchemaID: RestoreTargetMarkerSchemaID, Purpose: purpose, TargetGenerationID: generation.String(), BindingDigests: expected, IssuedAt: now.UTC().Format(time.RFC3339Nano), ExpiresAt: now.UTC().Add(RestoreTargetMarkerMaximumLifetime).Format(time.RFC3339Nano)}
	body, err := json.Marshal(marker)
	if err != nil {
		return uuid.Nil, err
	}
	prepared := TargetMarkerMaterial{MarkerBody: body, GenerationBody: []byte(generation.String() + "\n")}
	if _, err := AdmitRestoreTargetMarker(prepared, purpose, expected, now); err != nil {
		return uuid.Nil, err
	}
	if err := publish(material, prepared); err != nil {
		return uuid.Nil, err
	}
	return generation, nil
}

// Physical namespace identity supplements the logical owner binding, excluding
// credentials so rotation neither invalidates proofs nor discloses secrets.
func databaseTargetIdentity(deployment Deployment) string {
	basis := rootBindingBasis(deployment.DatabaseStorage)
	if strings.TrimSpace(deployment.PostgresSettings.DSN) == "" {
		return basis
	}
	config, err := pgx.ParseConfig(deployment.PostgresSettings.DSN)
	if err != nil {
		return basis + ":invalid_database_binding"
	}
	physical, _ := json.Marshal([]any{basis, strings.ToLower(config.Host), config.Port, config.Database})
	return string(physical)
}

func objectTargetIdentity(deployment Deployment) string {
	settings := deployment.ObjectSettings
	physical, _ := json.Marshal([]any{rootBindingBasis(deployment.ObjectStorage), strings.ToLower(settings.Endpoint), settings.Secure, settings.Bucket, filepath.Clean(settings.RootPath)})
	return string(physical)
}
