package postgres

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

const (
	FilesystemRuntimeDSNFile   = "postgres.runtime.dsn"
	FilesystemMigrationDSNFile = "postgres.migration.dsn"
	FilesystemRecoveryDSNFile  = "postgres.recovery.dsn"

	filesystemRetiredDSNFile = "postgres.dsn"

	filesystemRootDSNMaximumBytes int64 = 65536
	managedServiceDSNPrefix             = "CARTULARY_POSTGRES_"
	managedServiceDSNSuffix             = "_DSN"
)

const (
	ReasonPurposeUnknown                     = "unsupported_postgres_purpose"
	ReasonBindingInvalid                     = "postgres_binding_invalid"
	ReasonRetiredCredentialPresent           = "retired_postgres_binding_present"
	ReasonUnselectedPurposeCredentialPresent = "cross_purpose_postgres_binding_present"
	ReasonSelectedCredentialMissing          = "postgres_binding_missing"
	ReasonSelectedCredentialInvalid          = "postgres_binding_invalid"
	ReasonEffectiveRoleMismatch              = "postgres_effective_role_mismatch"
)

type Purpose uint8

const (
	PurposeRuntime Purpose = iota + 1
	PurposeMigration
	PurposeRecovery
)

type ConfigurationError struct {
	reason string
}

func (err *ConfigurationError) Error() string {
	return err.reason
}

func (err *ConfigurationError) Reason() string {
	return err.reason
}

func configurationError(reason string) error {
	return &ConfigurationError{reason: reason}
}

type Settings struct {
	BindingKind  string
	RootPath     string
	DSN          string
	ServiceRef   string
	Purpose      Purpose
	ExpectedRole string
}

type Binding struct {
	BindingKind string
	RootPath    string
	ServiceRef  string
}

func ResolveSettings(binding Binding, purpose Purpose, env map[string]string) (Settings, error) {
	selected, ok := purposeContractFor(purpose)
	if !ok {
		return Settings{}, configurationError(ReasonPurposeUnknown)
	}

	switch binding.BindingKind {
	case "filesystem_root":
		if binding.RootPath == "" || binding.ServiceRef != "" {
			return Settings{}, configurationError(ReasonBindingInvalid)
		}
		root, err := rootedfs.Open(binding.RootPath)
		if err != nil {
			return Settings{}, configurationError(ReasonBindingInvalid)
		}
		defer root.Close()

		retiredPresent, err := root.Exists(rootedfs.MustParseReference(filesystemRetiredDSNFile))
		if err != nil {
			return Settings{}, configurationError(ReasonBindingInvalid)
		}
		if retiredPresent {
			return Settings{}, configurationError(ReasonRetiredCredentialPresent)
		}
		for _, candidate := range purposeContracts() {
			if candidate.purpose == purpose {
				continue
			}
			present, presenceErr := root.Exists(rootedfs.MustParseReference(candidate.file))
			if presenceErr != nil {
				return Settings{}, configurationError(ReasonBindingInvalid)
			}
			if present {
				return Settings{}, configurationError(ReasonUnselectedPurposeCredentialPresent)
			}
		}

		selectedReference := rootedfs.MustParseReference(selected.file)
		present, err := root.Exists(selectedReference)
		if err != nil {
			return Settings{}, configurationError(ReasonBindingInvalid)
		}
		if !present {
			return Settings{}, configurationError(ReasonSelectedCredentialMissing)
		}
		payload, _, err := root.ReadRegular(selectedReference, filesystemRootDSNMaximumBytes)
		if err != nil {
			return Settings{}, configurationError(ReasonSelectedCredentialInvalid)
		}
		dsn, ok := decodeDSNFile(payload)
		if !ok || !validDSN(dsn) {
			return Settings{}, configurationError(ReasonSelectedCredentialInvalid)
		}
		return Settings{
			BindingKind:  "filesystem_root",
			RootPath:     binding.RootPath,
			DSN:          dsn,
			Purpose:      purpose,
			ExpectedRole: selected.role,
		}, nil

	case "managed_service":
		if binding.RootPath != "" || normalizeServiceRef(binding.ServiceRef) == "" {
			return Settings{}, configurationError(ReasonBindingInvalid)
		}
		retiredKey, ok := retiredEnvKeyForServiceRef(binding.ServiceRef)
		if !ok {
			return Settings{}, configurationError(ReasonBindingInvalid)
		}
		if envPresent(env, retiredKey) {
			return Settings{}, configurationError(ReasonRetiredCredentialPresent)
		}
		for _, candidate := range purposeContracts() {
			if candidate.purpose == purpose {
				continue
			}
			key, keyErr := EnvKeyForServiceRef(binding.ServiceRef, candidate.purpose)
			if keyErr != nil {
				return Settings{}, configurationError(ReasonBindingInvalid)
			}
			if envPresent(env, key) {
				return Settings{}, configurationError(ReasonUnselectedPurposeCredentialPresent)
			}
		}
		key, err := EnvKeyForServiceRef(binding.ServiceRef, purpose)
		if err != nil {
			return Settings{}, configurationError(ReasonBindingInvalid)
		}
		dsn, present := lookupEnv(env, key)
		if !present {
			return Settings{}, configurationError(ReasonSelectedCredentialMissing)
		}
		if dsn == "" || !validDSN(dsn) {
			return Settings{}, configurationError(ReasonSelectedCredentialInvalid)
		}
		return Settings{
			BindingKind:  "managed_service",
			DSN:          dsn,
			ServiceRef:   binding.ServiceRef,
			Purpose:      purpose,
			ExpectedRole: selected.role,
		}, nil
	default:
		return Settings{}, configurationError(ReasonBindingInvalid)
	}
}

var _ func(Binding, Purpose, map[string]string) (Settings, error) = ResolveSettings

func Setup(ctx context.Context, settings Settings) (*pgxpool.Pool, error) {
	if !validSettingsPurpose(settings) {
		return nil, configurationError(ReasonPurposeUnknown)
	}
	poolConfig, err := pgxpool.ParseConfig(settings.DSN)
	if err != nil {
		return nil, configurationError(ReasonSelectedCredentialInvalid)
	}
	poolConfig.AfterConnect = connectionInitializer(settings, poolConfig.ConnConfig.User)
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, configurationError(ReasonEffectiveRoleMismatch)
	}
	return pool, nil
}

func OpenSQL(settings Settings) (*sql.DB, error) {
	if !validSettingsPurpose(settings) {
		return nil, configurationError(ReasonPurposeUnknown)
	}
	config, err := pgx.ParseConfig(settings.DSN)
	if err != nil {
		return nil, configurationError(ReasonSelectedCredentialInvalid)
	}
	db := stdlib.OpenDB(*config, stdlib.OptionAfterConnect(connectionInitializer(settings, config.User)))
	return db, nil
}

func EnvKeyForServiceRef(serviceRef string, purpose Purpose) (string, error) {
	normalized := normalizeServiceRef(serviceRef)
	contract, ok := purposeContractFor(purpose)
	if normalized == "" || !ok {
		return "", configurationError(ReasonBindingInvalid)
	}
	return managedServiceDSNPrefix + normalized + "_" + contract.envSegment + managedServiceDSNSuffix, nil
}

type purposeContract struct {
	purpose    Purpose
	file       string
	envSegment string
	role       string
}

func purposeContracts() []purposeContract {
	return []purposeContract{
		{purpose: PurposeRuntime, file: FilesystemRuntimeDSNFile, envSegment: "RUNTIME", role: "cartulary_runtime"},
		{purpose: PurposeMigration, file: FilesystemMigrationDSNFile, envSegment: "MIGRATION", role: "cartulary_schema_owner"},
		{purpose: PurposeRecovery, file: FilesystemRecoveryDSNFile, envSegment: "RECOVERY", role: "cartulary_recovery"},
	}
}

func purposeContractFor(purpose Purpose) (purposeContract, bool) {
	for _, contract := range purposeContracts() {
		if contract.purpose == purpose {
			return contract, true
		}
	}
	return purposeContract{}, false
}

func validSettingsPurpose(settings Settings) bool {
	contract, ok := purposeContractFor(settings.Purpose)
	return ok && settings.ExpectedRole == contract.role
}

func connectionInitializer(settings Settings, expectedSessionUser string) func(context.Context, *pgx.Conn) error {
	return func(ctx context.Context, connection *pgx.Conn) error {
		contract, ok := purposeContractFor(settings.Purpose)
		if !ok || settings.ExpectedRole != contract.role || expectedSessionUser == "" {
			return configurationError(ReasonEffectiveRoleMismatch)
		}
		if _, err := connection.Exec(ctx, "SET ROLE "+contract.role); err != nil {
			return configurationError(ReasonEffectiveRoleMismatch)
		}
		var sessionUser string
		var currentUser string
		if err := connection.QueryRow(ctx, "SELECT session_user::text, current_user::text").Scan(&sessionUser, &currentUser); err != nil {
			return configurationError(ReasonEffectiveRoleMismatch)
		}
		if sessionUser != expectedSessionUser || currentUser != contract.role {
			return configurationError(ReasonEffectiveRoleMismatch)
		}
		return nil
	}
}

func decodeDSNFile(payload []byte) (string, bool) {
	if len(payload) == 0 || !utf8.Valid(payload) || strings.IndexByte(string(payload), 0) >= 0 {
		return "", false
	}
	if strings.HasSuffix(string(payload), "\r\n") {
		payload = payload[:len(payload)-2]
	} else if strings.HasSuffix(string(payload), "\n") {
		payload = payload[:len(payload)-1]
	}
	if len(payload) == 0 || strings.ContainsAny(string(payload), "\r\n") {
		return "", false
	}
	return string(payload), true
}

func validDSN(dsn string) bool {
	_, err := pgx.ParseConfig(dsn)
	return err == nil
}

func retiredEnvKeyForServiceRef(serviceRef string) (string, bool) {
	normalized := normalizeServiceRef(serviceRef)
	if normalized == "" {
		return "", false
	}
	return managedServiceDSNPrefix + normalized + managedServiceDSNSuffix, true
}

func envPresent(env map[string]string, key string) bool {
	if env != nil {
		_, ok := env[key]
		return ok
	}
	_, ok := os.LookupEnv(key)
	return ok
}

func lookupEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}
	return os.LookupEnv(key)
}

func normalizeServiceRef(value string) string {
	var builder strings.Builder
	previousUnderscore := false
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch {
		case c >= 'a' && c <= 'z':
			builder.WriteByte(c - ('a' - 'A'))
			previousUnderscore = false
		case c >= 'A' && c <= 'Z' || c >= '0' && c <= '9':
			builder.WriteByte(c)
			previousUnderscore = false
		case !previousUnderscore:
			builder.WriteByte('_')
			previousUnderscore = true
		}
	}

	return strings.Trim(builder.String(), "_")
}
