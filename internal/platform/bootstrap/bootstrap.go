package bootstrap

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/securefile"
)

const (
	bootstrapManifestSchemaID = "cartulary.bootstrap_admin.v1"
	bootstrapManifestRootPath = "bootstrap.first_admin_manifest"
	bootstrapManifestPathKey  = "bootstrap.first_admin_manifest_path"

	ManifestSchemaID = bootstrapManifestSchemaID

	ManifestMaximumBytes int64 = 1048576
)

type bootstrapManifest struct {
	BootstrapSchemaID   string
	BootstrapArtifactID uuid.UUID
	Email               string
	DisplayName         string
	InitialPassword     string
	MFARequired         bool
}

type bootstrapState struct {
	ActiveDeploymentAdmins int
	BootstrapCompleted     bool
}

type bootstrapCreateRequest struct {
	Manifest       bootstrapManifest
	ArtifactSHA256 []byte
	PasswordHash   string
}

type bootstrapStore interface {
	ReadBootstrapState(ctx context.Context) (bootstrapState, error)
	CreateBootstrapAdmin(ctx context.Context, request bootstrapCreateRequest) error
}

type postgresBootstrapStore struct {
	pool *pgxpool.Pool
}

type txBootstrapStore struct {
	tx pgx.Tx
}

type bootstrapManifestFS interface {
	Stat(name string) (fs.FileInfo, error)
	ReadFile(name string) ([]byte, error)
}

type secureBootstrapManifestFS struct {
	path     string
	document securefile.Document
	loaded   bool
}

func (manifestFS *secureBootstrapManifestFS) Stat(name string) (fs.FileInfo, error) {
	document, err := securefile.Read(name, ManifestMaximumBytes)
	if err != nil {
		var secureErr *securefile.Error
		if errors.As(err, &secureErr) && secureErr.Kind == securefile.FailureUnsafeObject {
			return secureManifestFileInfo{name: "bootstrap-manifest", mode: fs.ModeIrregular}, nil
		}
		return nil, err
	}
	manifestFS.path = name
	manifestFS.document = document
	manifestFS.loaded = true
	return secureManifestFileInfo{
		name:    "bootstrap-manifest",
		size:    document.Size(),
		mode:    document.Mode(),
		modTime: document.ModTime(),
	}, nil
}

func (manifestFS *secureBootstrapManifestFS) ReadFile(name string) ([]byte, error) {
	if !manifestFS.loaded || manifestFS.path != name {
		return nil, fs.ErrInvalid
	}
	return manifestFS.document.Bytes(), nil
}

type secureManifestFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

func (info secureManifestFileInfo) Name() string       { return info.name }
func (info secureManifestFileInfo) Size() int64        { return info.size }
func (info secureManifestFileInfo) Mode() fs.FileMode  { return info.mode }
func (info secureManifestFileInfo) ModTime() time.Time { return info.modTime }
func (info secureManifestFileInfo) IsDir() bool        { return info.mode.IsDir() }
func (info secureManifestFileInfo) Sys() any           { return nil }

func newSecureBootstrapManifestFS() bootstrapManifestFS {
	return &secureBootstrapManifestFS{}
}

func Preflight(ctx context.Context, settings Settings, pool *pgxpool.Pool) error {
	return bootstrapPreflight(ctx, settings, postgresBootstrapStore{pool: pool}, newSecureBootstrapManifestFS(), deriveBootstrapPasswordHash)
}

func PreflightTx(ctx context.Context, settings Settings, tx pgx.Tx) error {
	return bootstrapPreflight(ctx, settings, txBootstrapStore{tx: tx}, newSecureBootstrapManifestFS(), deriveBootstrapPasswordHash)
}

type Settings struct {
	ManifestPath string
}

func bootstrapPreflight(ctx context.Context, settings Settings, store bootstrapStore, manifestFS bootstrapManifestFS, hashPassword func(string) (string, error)) error {
	state, err := store.ReadBootstrapState(ctx)
	if err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_persist_failed", "query bootstrap state", err)
	}

	if state.ActiveDeploymentAdmins > 0 {
		return nil
	}
	if state.BootstrapCompleted {
		return configError(config.Diagnostic{
			Path:       bootstrapManifestPathKey,
			ReasonCode: "bootstrap_recovery_not_supported",
			Message:    "bootstrap completion state exists but no active deployment admin remains",
		})
	}

	manifestPath := settings.ManifestPath
	if manifestPath == "" {
		return configError(config.Diagnostic{
			Path:       bootstrapManifestPathKey,
			ReasonCode: "bootstrap_manifest_path_missing",
			Message:    "bootstrap manifest path is required for first deployment admin bootstrap",
		})
	}

	info, err := manifestFS.Stat(manifestPath)
	if err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_manifest_not_readable", "read bootstrap manifest", err)
	}
	if !info.Mode().IsRegular() {
		return configError(config.Diagnostic{
			Path:       bootstrapManifestPathKey,
			ReasonCode: "bootstrap_manifest_not_regular_file",
			Message:    "bootstrap manifest path must reference one regular file",
		})
	}

	raw, err := manifestFS.ReadFile(manifestPath)
	if err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_manifest_not_readable", "read bootstrap manifest", err)
	}

	manifest, err := parseBootstrapManifest(raw)
	if err != nil {
		return err
	}

	passwordHash, err := hashPassword(manifest.InitialPassword)
	if err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_persist_failed", "derive bootstrap password hash", err)
	}

	artifactSHA256, err := manifestSHA256(raw)
	if err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_persist_failed", "hash bootstrap manifest", err)
	}

	return store.CreateBootstrapAdmin(ctx, bootstrapCreateRequest{
		Manifest:       manifest,
		ArtifactSHA256: artifactSHA256,
		PasswordHash:   passwordHash,
	})
}

func parseBootstrapManifest(raw []byte) (bootstrapManifest, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return bootstrapManifest{}, configError(config.Diagnostic{
			Path:       bootstrapManifestPathKey,
			ReasonCode: "bootstrap_manifest_parse_error",
			Message:    fmt.Sprintf("parse bootstrap manifest JSON: %v", err),
		})
	}

	allowedKeys := map[string]struct{}{
		"bootstrap_schema_id":   {},
		"bootstrap_artifact_id": {},
		"email":                 {},
		"display_name":          {},
		"initial_password":      {},
		"mfa_required":          {},
	}

	diagnostics := make([]config.Diagnostic, 0)
	for key := range root {
		if _, ok := allowedKeys[key]; ok {
			continue
		}
		diagnostics = append(diagnostics, config.Diagnostic{
			Path:       bootstrapManifestRootPath + "." + key,
			ReasonCode: "bootstrap_manifest_schema_invalid",
			Message:    fmt.Sprintf("bootstrap manifest field %q is not allowed", key),
		})
	}

	manifest := bootstrapManifest{
		MFARequired: true,
	}

	if value, ok := root["bootstrap_schema_id"]; !ok {
		diagnostics = append(diagnostics, requiredBootstrapFieldDiagnostic("bootstrap_schema_id"))
	} else if parsed, ok := decodeJSONString(value, "bootstrap_schema_id", &diagnostics); ok {
		manifest.BootstrapSchemaID = parsed
		if manifest.BootstrapSchemaID != bootstrapManifestSchemaID {
			diagnostics = append(diagnostics, config.Diagnostic{
				Path:       bootstrapManifestRootPath + ".bootstrap_schema_id",
				ReasonCode: "bootstrap_manifest_schema_invalid",
				Message:    fmt.Sprintf("bootstrap_schema_id must equal %q", bootstrapManifestSchemaID),
			})
		}
	}

	if value, ok := root["bootstrap_artifact_id"]; !ok {
		diagnostics = append(diagnostics, requiredBootstrapFieldDiagnostic("bootstrap_artifact_id"))
	} else if parsed, ok := decodeJSONString(value, "bootstrap_artifact_id", &diagnostics); ok {
		artifactID, err := uuid.Parse(parsed)
		if err != nil {
			diagnostics = append(diagnostics, config.Diagnostic{
				Path:       bootstrapManifestRootPath + ".bootstrap_artifact_id",
				ReasonCode: "bootstrap_manifest_schema_invalid",
				Message:    fmt.Sprintf("bootstrap_artifact_id must be a UUID: %v", err),
			})
		} else {
			manifest.BootstrapArtifactID = artifactID
		}
	}

	if value, ok := root["email"]; !ok {
		diagnostics = append(diagnostics, requiredBootstrapFieldDiagnostic("email"))
	} else if parsed, ok := decodeJSONString(value, "email", &diagnostics); ok {
		normalized, err := normalizeBootstrapEmail(parsed)
		if err != nil {
			diagnostics = append(diagnostics, bootstrapManifestDiagnostic("email", err.Error()))
		} else {
			manifest.Email = normalized
		}
	}

	if value, ok := root["display_name"]; !ok {
		diagnostics = append(diagnostics, requiredBootstrapFieldDiagnostic("display_name"))
	} else if parsed, ok := decodeJSONString(value, "display_name", &diagnostics); ok {
		normalized, err := normalizeBootstrapDisplayName(parsed)
		if err != nil {
			diagnostics = append(diagnostics, bootstrapManifestDiagnostic("display_name", err.Error()))
		} else {
			manifest.DisplayName = normalized
		}
	}

	if value, ok := root["initial_password"]; !ok {
		diagnostics = append(diagnostics, requiredBootstrapFieldDiagnostic("initial_password"))
	} else if parsed, ok := decodeJSONString(value, "initial_password", &diagnostics); ok {
		if err := validateBootstrapPassword(parsed); err != nil {
			diagnostics = append(diagnostics, bootstrapManifestDiagnostic("initial_password", err.Error()))
		} else {
			manifest.InitialPassword = parsed
		}
	}

	if value, ok := root["mfa_required"]; ok {
		var explicit bool
		if err := json.Unmarshal(value, &explicit); err != nil {
			diagnostics = append(diagnostics, config.Diagnostic{
				Path:       bootstrapManifestRootPath + ".mfa_required",
				ReasonCode: "bootstrap_manifest_schema_invalid",
				Message:    fmt.Sprintf("mfa_required must be a boolean: %v", err),
			})
		} else if !explicit {
			diagnostics = append(diagnostics, config.Diagnostic{
				Path:       bootstrapManifestRootPath + ".mfa_required",
				ReasonCode: "bootstrap_manifest_schema_invalid",
				Message:    "mfa_required must be true when supplied",
			})
		}
	}

	if len(diagnostics) > 0 {
		return bootstrapManifest{}, configError(diagnostics...)
	}

	return manifest, nil
}

func manifestSHA256(raw []byte) ([]byte, error) {
	sum := sha256.Sum256(raw)
	data := make([]byte, len(sum))
	copy(data, sum[:])
	return data, nil
}

func deriveBootstrapPasswordHash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := authn.DerivePasswordHash([]byte(password), salt)
	return fmt.Sprintf(
		"argon2id$v=19$m=65536,t=1,p=4$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (s postgresBootstrapStore) ReadBootstrapState(ctx context.Context) (bootstrapState, error) {
	var state bootstrapState
	if err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true),
			EXISTS(SELECT 1 FROM deployment_bootstrap_state)
	`).Scan(&state.ActiveDeploymentAdmins, &state.BootstrapCompleted); err != nil {
		return bootstrapState{}, err
	}
	return state, nil
}

func (s postgresBootstrapStore) CreateBootstrapAdmin(ctx context.Context, request bootstrapCreateRequest) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_persist_failed", "begin bootstrap transaction", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := createBootstrapAdminInTx(ctx, tx, request); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_persist_failed", "commit bootstrap transaction", err)
	}

	return nil
}

func (s txBootstrapStore) ReadBootstrapState(ctx context.Context) (bootstrapState, error) {
	var state bootstrapState
	if err := s.tx.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true),
			EXISTS(SELECT 1 FROM deployment_bootstrap_state)
	`).Scan(&state.ActiveDeploymentAdmins, &state.BootstrapCompleted); err != nil {
		return bootstrapState{}, err
	}
	return state, nil
}

func (s txBootstrapStore) CreateBootstrapAdmin(ctx context.Context, request bootstrapCreateRequest) error {
	return createBootstrapAdminInTx(ctx, s.tx, request)
}

func createBootstrapAdminInTx(ctx context.Context, tx pgx.Tx, request bootstrapCreateRequest) error {
	var emailExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, request.Manifest.Email).Scan(&emailExists); err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_persist_failed", "check bootstrap email uniqueness", err)
	}
	if emailExists {
		return configError(config.Diagnostic{
			Path:       bootstrapManifestRootPath + ".email",
			ReasonCode: "bootstrap_email_conflict",
			Message:    "bootstrap manifest email conflicts with an existing local user",
		})
	}

	var userID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (email, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin)
		VALUES ($1, $2, $3, now(), true, true, true)
		RETURNING id
	`, request.Manifest.Email, request.Manifest.DisplayName, request.PasswordHash).Scan(&userID); err != nil {
		return bootstrapPersistOrConflict(err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO deployment_bootstrap_state (slot, bootstrap_schema_id, bootstrap_artifact_id, artifact_sha256, created_user_id)
		VALUES ('first_deployment_admin', $1, $2, $3, $4)
	`, request.Manifest.BootstrapSchemaID, request.Manifest.BootstrapArtifactID.String(), request.ArtifactSHA256, userID); err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_persist_failed", "persist bootstrap completion marker", err)
	}

	after := map[string]any{
		"email":               request.Manifest.Email,
		"display_name":        request.Manifest.DisplayName,
		"mfa_required":        true,
		"is_active":           true,
		"is_deployment_admin": true,
	}
	occurredAt := time.Now().UTC()
	targetID := userID.String()
	if _, err := administrativeaudit.AppendTx(ctx, tx, administrativeaudit.RawEvent{
		TargetUserID: &userID,
		EventSource:  "bootstrap_manifest",
		EventKind:    "bootstrap_admin_created",
		After:        after,
		OccurredAt:   occurredAt,
	}, administrativeaudit.Event{
		ScopeKind:  administrativeaudit.ScopeDeployment,
		OccurredAt: occurredAt,
		ActorKind:  administrativeaudit.ActorSystem,
		Source:     administrativeaudit.SourceStartup,
		ActionCode: administrativeaudit.ActionBootstrapAdminCreated,
		TargetKind: administrativeaudit.TargetUser,
		TargetID:   &targetID,
		Changes: []administrativeaudit.Change{
			administrativeaudit.Visible("display_name", nil, request.Manifest.DisplayName),
			administrativeaudit.Visible("email", nil, request.Manifest.Email),
			administrativeaudit.Visible("is_active", nil, true),
			administrativeaudit.Visible("is_deployment_admin", nil, true),
			administrativeaudit.Visible("mfa_required", nil, true),
		},
	}); err != nil {
		return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_persist_failed", "persist bootstrap audit event", err)
	}

	return nil
}

func bootstrapPersistOrConflict(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return configError(config.Diagnostic{
			Path:       bootstrapManifestRootPath + ".email",
			ReasonCode: "bootstrap_email_conflict",
			Message:    "bootstrap manifest email conflicts with an existing local user",
		})
	}
	return bootstrapDiagnostic(bootstrapManifestPathKey, "bootstrap_persist_failed", "persist bootstrap-created user", err)
}

func bootstrapDiagnostic(path string, reasonCode string, action string, err error) error {
	message := err.Error()
	if errors.Is(err, fs.ErrPermission) {
		message = fs.ErrPermission.Error()
	}

	return configError(config.Diagnostic{
		Path:       path,
		ReasonCode: reasonCode,
		Message:    fmt.Sprintf("%s: %s", action, message),
	})
}

func configError(diagnostics ...config.Diagnostic) error {
	return config.NewDiagnosticsError(diagnostics...)
}

func bootstrapManifestDiagnostic(field string, message string) config.Diagnostic {
	return config.Diagnostic{
		Path:       bootstrapManifestRootPath + "." + field,
		ReasonCode: "bootstrap_manifest_schema_invalid",
		Message:    message,
	}
}

func requiredBootstrapFieldDiagnostic(field string) config.Diagnostic {
	return config.Diagnostic{
		Path:       bootstrapManifestRootPath + "." + field,
		ReasonCode: "bootstrap_manifest_schema_invalid",
		Message:    fmt.Sprintf("%s is required", field),
	}
}

func decodeJSONString(raw json.RawMessage, field string, diagnostics *[]config.Diagnostic) (string, bool) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		*diagnostics = append(*diagnostics, config.Diagnostic{
			Path:       bootstrapManifestRootPath + "." + field,
			ReasonCode: "bootstrap_manifest_schema_invalid",
			Message:    fmt.Sprintf("%s must be a JSON string: %v", field, err),
		})
		return "", false
	}
	return value, true
}

func normalizeBootstrapEmail(value string) (string, error) {
	normalized := norm.NFC.String(strings.TrimSpace(value))
	if normalized == "" {
		return "", errors.New("email must not be empty")
	}
	if hasControlRunes(normalized) {
		return "", errors.New("email must not contain control characters")
	}
	for _, r := range normalized {
		if unicode.IsSpace(r) {
			return "", errors.New("email must not contain whitespace")
		}
	}
	if utf8.RuneCountInString(normalized) > 320 {
		return "", errors.New("email must not exceed 320 Unicode scalar values")
	}
	if strings.Count(normalized, "@") != 1 {
		return "", errors.New("email must contain exactly one @")
	}
	parts := strings.SplitN(normalized, "@", 2)
	if parts[0] == "" || parts[1] == "" {
		return "", errors.New("email must contain non-empty local and domain parts")
	}
	return normalized, nil
}

func normalizeBootstrapDisplayName(value string) (string, error) {
	normalized := norm.NFC.String(strings.TrimSpace(value))
	if normalized == "" {
		return "", errors.New("display_name must not be empty")
	}
	if hasControlRunes(normalized) {
		return "", errors.New("display_name must not contain control characters")
	}
	if utf8.RuneCountInString(normalized) > 256 {
		return "", errors.New("display_name must not exceed 256 Unicode scalar values")
	}
	return normalized, nil
}

func validateBootstrapPassword(value string) error {
	if hasControlRunes(value) {
		return errors.New("initial_password must not contain control characters")
	}
	if utf8.RuneCountInString(value) < 12 {
		return errors.New("initial_password must be at least 12 Unicode scalar values")
	}
	if utf8.RuneCountInString(value) > 1024 {
		return errors.New("initial_password must not exceed 1024 Unicode scalar values")
	}
	allWhitespace := true
	for _, r := range value {
		if !unicode.IsSpace(r) {
			allWhitespace = false
			break
		}
	}
	if allWhitespace {
		return errors.New("initial_password must not be all whitespace")
	}
	return nil
}

func hasControlRunes(value string) bool {
	for _, r := range value {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			return true
		}
	}
	return false
}
