package reference_data

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type minimumDisconnectedPack struct {
	PackKey string
	Version string
}

var minimumDisconnectedPacks = []minimumDisconnectedPack{
	{PackKey: "type_registry.host", Version: "1"},
	{PackKey: "type_registry.evidence", Version: "1"},
	{PackKey: "type_registry.indicator", Version: "1"},
}

func EnsureMinimumDisconnectedBundle(ctx context.Context, cfg config.Config, pool *pgxpool.Pool, now time.Time) error {
	if cfg.DeploymentProfile != "disconnected" {
		return nil
	}
	var existing int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_packs`).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	store := &Service{deps: httpapi.DependencySet{Config: cfg}}
	type seedRecord struct {
		Verification VerificationResult
		BundlePath   string
	}
	seeds := make([]seedRecord, 0, len(minimumDisconnectedPacks))
	for _, pack := range minimumDisconnectedPacks {
		bundle, err := buildMinimumDisconnectedBundle(pack)
		if err != nil {
			return err
		}
		verification, err := VerifyBundle(VerificationInput{
			Bundle:          bundle,
			ContentType:     MediaTypeZip,
			ArchiveLimits:   cfg.Limits.Archives,
			ReferenceLimits: cfg.Limits.ReferencePacks,
		})
		if err != nil {
			return err
		}
		path, err := store.persistBundle(verification.BundleSHA256, bundle)
		if err != nil {
			return err
		}
		seeds = append(seeds, seedRecord{Verification: verification, BundlePath: path})
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, seed := range seeds {
		if err := insertMinimumDisconnectedPackTx(ctx, tx, seed.Verification, seed.BundlePath, now.UTC()); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func buildMinimumDisconnectedBundle(pack minimumDisconnectedPack) ([]byte, error) {
	payloadPath := "payloads/" + strings.ReplaceAll(pack.PackKey, ".", "_") + ".json"
	payload, err := json.Marshal(map[string]any{
		"pack_key": pack.PackKey,
		"version":  pack.Version,
		"entries":  []any{},
	})
	if err != nil {
		return nil, err
	}
	payloadSHA := sha256.Sum256(payload)
	manifest, err := json.Marshal(bundleManifest{
		PackKey:             pack.PackKey,
		PackKind:            "type_registry",
		PackVersion:         pack.Version,
		PackContractVersion: PackContractVersionV1,
		VerificationMethod:  "manifest_sha256_v1",
		Payloads: []manifestPayload{
			{Path: payloadPath, SHA256: hex.EncodeToString(payloadSHA[:])},
		},
	})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := writeDeterministicZipMember(zw, "manifest.json", manifest); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeDeterministicZipMember(zw, payloadPath, payload); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeDeterministicZipMember(zw *zip.Writer, name string, data []byte) error {
	header := &zip.FileHeader{Name: name, Method: zip.Store}
	header.Modified = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

func insertMinimumDisconnectedPackTx(ctx context.Context, tx pgx.Tx, verification VerificationResult, bundlePath string, now time.Time) error {
	metadata := map[string]any{
		"payload_count":                verification.Metadata["payload_count"],
		"minimum_disconnected_builtin": true,
	}
	_, err := tx.Exec(ctx, `
INSERT INTO reference_packs (
    pack_key, version, pack_kind, source_identifier, manifest_sha256, payload_sha256,
    pack_contract_version, verification_method, signer_key_id, status, imported_at,
    imported_by_user_id, verification_result, bundle_sha256, bundle_storage_path, metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'available', $10, NULL, 'passed', $11, $12, $13)
`, verification.PackKey, verification.PackVersion, verification.PackKind, verification.SourceIdentifier, verification.ManifestSHA256, verification.PayloadSHA256, verification.PackContractVersion, verification.VerificationMethod, verification.SignerKeyID, now, verification.BundleSHA256, bundlePath, mustJSON(metadata))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO reference_pack_activation_state (
    pack_key, active_version, previous_active_version, activated_at, activated_by_user_id, operator_note
)
VALUES ($1, $2, NULL, $3, NULL, $4)
`, verification.PackKey, verification.PackVersion, now, "minimum disconnected bundle")
	if err != nil {
		return err
	}
	importAttestation := attestationParams{
		PackKey: verification.PackKey, PackVersion: verification.PackVersion, PackKind: verification.PackKind,
		EventKind: "import", ManifestSHA256: verification.ManifestSHA256, PayloadSHA256: verification.PayloadSHA256,
		SourceIdentifier: verification.SourceIdentifier, VerificationMethod: verification.VerificationMethod,
		SignerKeyID: verification.SignerKeyID, VerificationResult: VerificationPassed, OccurredAt: now, Metadata: metadata,
	}
	if err := insertAttestationTx(ctx, tx, importAttestation); err != nil {
		return err
	}
	activateAttestation := importAttestation
	activateAttestation.EventKind = "activate"
	activateAttestation.OperatorNote = stringPtr("minimum disconnected bundle")
	return insertAttestationTx(ctx, tx, activateAttestation)
}

func (s *Service) persistBundle(bundleSHA string, data []byte) (string, error) {
	root := s.deps.Config.Roots.ReferencePackStorage.Path
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(os.TempDir(), "cartulary-reference-packs")
	}
	bundleDir := filepath.Join(root, "bundles")
	if err := os.MkdirAll(bundleDir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(bundleDir, bundleSHA+".bundle")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func stringPtr(value string) *string {
	return &value
}
