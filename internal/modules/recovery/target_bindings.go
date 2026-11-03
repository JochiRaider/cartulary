package recovery

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type RootBinding struct {
	BindingKind string
	Path        string
	ServiceRef  string
}

func (binding RootBinding) Identity() string {
	switch binding.BindingKind {
	case "filesystem_root":
		return "filesystem_root:" + filepath.Clean(binding.Path)
	case "managed_service":
		return "managed_service:" + strings.TrimSpace(binding.ServiceRef)
	default:
		return strings.TrimSpace(binding.BindingKind)
	}
}

type TargetBindingDigests struct {
	DatabaseSHA256             string `json:"database_sha256"`
	ObjectStoreSHA256          string `json:"object_store_sha256"`
	ReferencePackStorageSHA256 string `json:"reference_pack_storage_sha256"`
	ExportOutputsSHA256        string `json:"export_outputs_sha256"`
}

func (bindings TargetBindingDigests) valid() bool {
	return validSHA256Hex(bindings.DatabaseSHA256) && validSHA256Hex(bindings.ObjectStoreSHA256) &&
		validSHA256Hex(bindings.ReferencePackStorageSHA256) && validSHA256Hex(bindings.ExportOutputsSHA256)
}

// TargetBindingInputs contains storage coordinates, never operational factories.
// Credentials are excluded from the resulting installation identity.
type TargetBindingInputs struct {
	DatabaseStorage      RootBinding
	ObjectStorage        RootBinding
	ReferencePackStorage RootBinding
	ExportOutputs        RootBinding
	PostgresSettings     postgres.Settings
	ObjectSettings       objectstore.Settings
}

func TargetBindingDigestsFor(input TargetBindingInputs) TargetBindingDigests {
	database := input.DatabaseStorage.Identity()
	if strings.TrimSpace(input.PostgresSettings.DSN) != "" {
		config, err := pgx.ParseConfig(input.PostgresSettings.DSN)
		if err != nil {
			database += ":invalid_database_binding"
		} else {
			physical, _ := json.Marshal([]any{database, strings.ToLower(config.Host), config.Port, config.Database})
			database = string(physical)
		}
	}
	settings := input.ObjectSettings
	object, _ := json.Marshal([]any{input.ObjectStorage.Identity(), strings.ToLower(settings.Endpoint), settings.Secure, settings.Bucket, filepath.Clean(settings.RootPath)})
	return TargetBindingDigests{
		DatabaseSHA256:             SHA256String(database),
		ObjectStoreSHA256:          SHA256String(string(object)),
		ReferencePackStorageSHA256: SHA256String(input.ReferencePackStorage.Identity()),
		ExportOutputsSHA256:        SHA256String(input.ExportOutputs.Identity()),
	}
}
