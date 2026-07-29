package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

const (
	RestoreTargetMarkerSchemaID         = "cartulary.restore_target_marker.v2"
	RestoreTargetMarkerMaximumBytes     = int64(65536)
	RestoreTargetGenerationMaximumBytes = int64(64)
	RestoreTargetMarkerMaximumLifetime  = 24 * time.Hour
	RestoreTargetServingLeaseAcquireMax = time.Second

	RestoreTargetPurpose             = "restore_target"
	RestoreVerificationTargetPurpose = "restore_verification_target"
)

var ErrTargetServingLeaseLost = errors.New("restore target serving lease lost")

type TargetMarkerMaterial struct {
	MarkerBody     []byte
	GenerationBody []byte
}

type TargetMarkerReader func(bindingKind string, rootPath string) (TargetMarkerMaterial, error)

type TargetBindingDigests struct {
	DatabaseSHA256    string `json:"database_sha256"`
	ObjectStoreSHA256 string `json:"object_store_sha256"`
}

type RestoreTargetMarker struct {
	SchemaID           string               `json:"schema_id"`
	Purpose            string               `json:"purpose"`
	TargetGenerationID string               `json:"target_generation_id"`
	BindingDigests     TargetBindingDigests `json:"binding_digests"`
	IssuedAt           string               `json:"issued_at"`
	ExpiresAt          string               `json:"expires_at"`
}

type TargetServingAdmission interface {
	Context() context.Context
	AssertHeld() error
	Release(context.Context) error
}

type TargetServingAdmissionFactory func(context.Context, PostgresPool, time.Duration, time.Duration) (TargetServingAdmission, error)

func TargetBindingDigestsFor(deployment Deployment) TargetBindingDigests {
	return TargetBindingDigests{
		DatabaseSHA256:    bindingDigest(rootBindingBasis(deployment.DatabaseStorage)),
		ObjectStoreSHA256: bindingDigest(rootBindingBasis(deployment.ObjectStorage)),
	}
}

func ValidateRestoreTargetMarker(material TargetMarkerMaterial, purpose string, expected TargetBindingDigests, now time.Time) error {
	if purpose != RestoreTargetPurpose && purpose != RestoreVerificationTargetPurpose {
		return fmt.Errorf("restore target marker purpose %q is invalid", purpose)
	}
	if err := strictjson.ValidateObject(material.MarkerBody); err != nil {
		return fmt.Errorf("restore target marker is not strict JSON: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(material.MarkerBody))
	decoder.DisallowUnknownFields()
	var marker RestoreTargetMarker
	if err := decoder.Decode(&marker); err != nil {
		return fmt.Errorf("decode restore target marker: %w", err)
	}
	if marker.SchemaID != RestoreTargetMarkerSchemaID || marker.Purpose != purpose {
		return errors.New("restore target marker has the wrong schema or purpose")
	}
	generationText := strings.TrimSpace(string(material.GenerationBody))
	generationID, err := uuid.Parse(generationText)
	if err != nil || generationID == uuid.Nil || generationText != generationID.String() {
		return errors.New("restore target generation proof is invalid")
	}
	markerGenerationID, err := uuid.Parse(marker.TargetGenerationID)
	if err != nil || markerGenerationID == uuid.Nil ||
		marker.TargetGenerationID != markerGenerationID.String() ||
		markerGenerationID != generationID {
		return errors.New("restore target marker has the wrong target generation")
	}
	if marker.BindingDigests != expected ||
		!isLowerSHA256(marker.BindingDigests.DatabaseSHA256) ||
		!isLowerSHA256(marker.BindingDigests.ObjectStoreSHA256) {
		return errors.New("restore target marker has the wrong target binding")
	}
	issuedAt, err := parseCanonicalMarkerTime(marker.IssuedAt)
	if err != nil {
		return fmt.Errorf("restore target marker issued_at: %w", err)
	}
	expiresAt, err := parseCanonicalMarkerTime(marker.ExpiresAt)
	if err != nil {
		return fmt.Errorf("restore target marker expires_at: %w", err)
	}
	now = now.UTC()
	lifetime := expiresAt.Sub(issuedAt)
	if lifetime <= 0 || lifetime > RestoreTargetMarkerMaximumLifetime {
		return errors.New("restore target marker lifetime is invalid")
	}
	if issuedAt.After(now) || !expiresAt.After(now) {
		return errors.New("restore target marker is not currently valid")
	}
	return nil
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
