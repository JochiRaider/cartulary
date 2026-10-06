package testsupport

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type TargetFixture struct {
	Env         map[string]string
	Postgres    *pgxpool.Pool
	ObjectStore objectstore.Store
	cleanupOnce *sync.Once
}

func NewTargetFixture(env map[string]string, postgres *pgxpool.Pool, store objectstore.Store) TargetFixture {
	return TargetFixture{
		Env:         maps.Clone(env),
		Postgres:    postgres,
		ObjectStore: store,
		cleanupOnce: &sync.Once{},
	}
}

func (fixture TargetFixture) Cleanup() {
	if fixture.cleanupOnce == nil {
		return
	}
	fixture.cleanupOnce.Do(func() {
		clear(fixture.Env)
	})
}

type EvidenceLocation struct {
	ResultsRoot string
	RunID       string
	Target      string
	Group       string
}

func (location EvidenceLocation) Dir() (string, error) {
	parts := []struct {
		name  string
		value string
	}{
		{name: "results root", value: location.ResultsRoot},
		{name: "run ID", value: location.RunID},
		{name: "target", value: location.Target},
		{name: "group", value: location.Group},
	}
	for _, part := range parts {
		if strings.TrimSpace(part.value) == "" {
			return "", fmt.Errorf("recovery evidence %s is required", part.name)
		}
	}
	for _, part := range parts[1:] {
		if filepath.Base(filepath.Clean(part.value)) != part.value || part.value == "." || part.value == ".." {
			return "", fmt.Errorf("recovery evidence %s must be one normalized path segment", part.name)
		}
	}
	return filepath.Join(filepath.Clean(location.ResultsRoot), location.RunID, location.Target, location.Group), nil
}

func WriteEvidenceArtifact(t testing.TB, location EvidenceLocation, name string, body []byte) string {
	t.Helper()
	if filepath.Base(filepath.Clean(name)) != name || name == "." || name == ".." {
		t.Fatalf("Recovery evidence name must be one normalized path segment: %q", name)
	}
	dir, err := location.Dir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create Recovery evidence dir: %v", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatalf("secure Recovery evidence dir: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write Recovery evidence artifact: %v", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatalf("secure Recovery evidence artifact: %v", err)
	}
	return path
}

func RequireStoredArtifactProof(t testing.TB, storage recovery.BackupStorage, proof recovery.BackupArtifactProof) []byte {
	t.Helper()
	body, err := recovery.VerifyArtifactProof(context.Background(), storage, proof)
	if err != nil {
		t.Fatalf("verify stored artifact proof for %s: %v", proof.Key, err)
	}
	return body
}

type VerificationExpectation struct {
	BackupSetID        uuid.UUID
	ConsistencyPointAt time.Time
	IncidentID         string
	ObjectCount        int64
	RegistrationID     string
	ViewSchemaID       string
}

func RequireVerificationArtifact(t testing.TB, storage recovery.BackupStorage, result recovery.RestoreVerificationResult, location EvidenceLocation, expected VerificationExpectation) string {
	t.Helper()
	body := RequireStoredArtifactProof(t, storage, result.ArtifactProof)
	artifact, err := recovery.DecodeRestoreVerificationArtifact(body)
	if err != nil {
		t.Fatalf("decode restore verification artifact: %v", err)
	}
	if artifact.BackupSetID != expected.BackupSetID.String() ||
		!artifact.ConsistencyPointAt.Equal(expected.ConsistencyPointAt) ||
		artifact.SelectedIncidentID == nil || *artifact.SelectedIncidentID != expected.IncidentID ||
		artifact.WorkbookProbe.Status != "executed" ||
		artifact.WorkbookProbe.RegistrationID != expected.RegistrationID ||
		artifact.WorkbookProbe.ViewSchemaID != expected.ViewSchemaID ||
		artifact.RestoredObjectCount != expected.ObjectCount || artifact.Result != "pass" {
		t.Fatalf("restore verification artifact got %#v want %#v", artifact, expected)
	}
	return WriteEvidenceArtifact(t, location, "restore-verification.json", body)
}
