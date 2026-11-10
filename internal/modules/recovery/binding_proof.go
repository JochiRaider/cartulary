package recovery

import (
	"context"
	"fmt"
	"os"

	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
	"github.com/JochiRaider/cartulary/internal/platform/maintenancequery"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

const PostgresBindingProofV1SchemaID = "cartulary.extension_postgres_binding_proof.v1"

// PostgresBindingProof authenticates the logical entries and the exact current
// physical/implementation/codec identities, independently of table transport.
type PostgresBindingProof struct {
	SchemaID                    string `json:"schema_id"`
	ProfileID                   string `json:"profile_id"`
	BindingID                   string `json:"binding_id"`
	PhysicalStateBindingSHA256  string `json:"physical_state_binding_sha256"`
	ImplementationBindingSHA256 string `json:"implementation_binding_sha256"`
	BackupCodecID               string `json:"backup_codec_id"`
	BackupCodecSHA256           string `json:"backup_codec_sha256"`
	StorageKind                 string `json:"storage_kind"`
	ItemCount                   int64  `json:"item_count"`
	ContentBytes                int64  `json:"content_bytes"`
	BindingDigest               string `json:"binding_digest"`
}

type bindingQuery func(context.Context, sqlshape.Fragment, ...any) (VNextRows, error)

// logicalBindingEntries confines a reader to the admitted mapping. Only the
// selected logical fields cross this boundary; no transaction or SQL port does.
func logicalBindingEntries(ctx context.Context, codec *extensioncodec.Postgres, query bindingQuery) extensioncodec.Entries {
	return func(yield func(extensioncodec.Entry) error) error {
		if codec == nil || query == nil {
			return ErrExtensionBindingInvalid
		}
		mapping, family := codec.Mapping(), codec.Family()
		columns := make([]sqlshape.Fragment, len(mapping.Columns))
		for i, column := range mapping.Columns {
			columns[i] = sqlshape.Identifier(column.ColumnName)
			// Preserve original JSON numbers. Driver decoding into interface values
			// would otherwise convert JSON numbers through floating point.
			if family.Fields[i].ValueKind == "json" {
				columns[i] = sqlshape.Concat(columns[i], sqlshape.Literal("::text"))
			}
		}
		if mapping.SchemaName != "public" {
			return ErrExtensionBindingInvalid
		}
		from, err := maintenancequery.From(mapping.TableName)
		if err != nil {
			return err
		}
		rows, err := query(ctx, sqlshape.Format("SELECT %s %s", sqlshape.Join(columns, sqlshape.Literal(",")), from))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			if err := ctx.Err(); err != nil {
				return err
			}
			values := make([]any, len(columns))
			destinations := make([]any, len(columns))
			for i := range values {
				destinations[i] = &values[i]
			}
			if err := rows.Scan(destinations...); err != nil {
				return err
			}
			entry, err := codec.EncodeRow(values)
			if err != nil {
				return err
			}
			if err := yield(entry); err != nil {
				return err
			}
		}
		return rows.Err()
	}
}

func computePostgresBindingProof(ctx context.Context, binding ExtensionBackupBinding, scratch *os.Root, query bindingQuery) (PostgresBindingProof, error) {
	codec := binding.CurrentCodec.Postgres
	if codec == nil {
		return PostgresBindingProof{}, ErrExtensionBindingInvalid
	}
	proof, err := extensioncodec.ComputeProof(ctx, binding.BindingID, binding.StorageKind, codec.Binding().Limits, scratch, logicalBindingEntries(ctx, codec, query))
	if err != nil {
		return PostgresBindingProof{}, err
	}
	return PostgresBindingProof{
		SchemaID: PostgresBindingProofV1SchemaID, ProfileID: binding.ProfileID, BindingID: binding.BindingID,
		PhysicalStateBindingSHA256: binding.PhysicalStateBindingSHA256, ImplementationBindingSHA256: binding.ImplementationBindingSHA256,
		BackupCodecID: binding.CurrentCodec.CodecID, BackupCodecSHA256: binding.CurrentCodec.CodecSHA256,
		StorageKind: binding.StorageKind, ItemCount: proof.ItemCount, ContentBytes: proof.ContentBytes, BindingDigest: proof.BindingDigest,
	}, nil
}

func admitPostgresBindingProof(proof PostgresBindingProof, binding ExtensionBackupBinding) error {
	if proof.BackupCodecID != binding.CurrentCodec.CodecID || proof.BackupCodecSHA256 != binding.CurrentCodec.CodecSHA256 {
		return ErrExtensionCodecUnsupported
	}
	if proof.SchemaID != PostgresBindingProofV1SchemaID || proof.ProfileID != binding.ProfileID || proof.BindingID != binding.BindingID ||
		proof.PhysicalStateBindingSHA256 != binding.PhysicalStateBindingSHA256 || proof.ImplementationBindingSHA256 != binding.ImplementationBindingSHA256 ||
		proof.StorageKind != "postgres" || proof.ItemCount < 0 || proof.ItemCount > int64(binding.CurrentCodec.MaxItems) ||
		proof.ContentBytes < 0 || proof.ContentBytes > binding.CurrentCodec.MaxTotalBytes || !validSHA256Hex(proof.BindingDigest) {
		return fmt.Errorf("%w: binding proof identity or bounds mismatch", ErrExtensionBindingInvalid)
	}
	return nil
}

// Resolve each declared column against the stopped target without reading row
// contents or invoking profile semantics. Logical value kinds remain the codec's
// responsibility; physical type aliases do not define binding identity.
func inspectPostgresBindingMappings(ctx context.Context, db restoreTargetReader, catalog *ExtensionBackupCatalog) error {
	if db == nil || catalog == nil {
		return ErrExtensionBindingInvalid
	}
	for _, binding := range catalog.Bindings() {
		mapping := binding.CurrentCodec.Postgres.Mapping()
		columns := make([]sqlshape.Fragment, len(mapping.Columns))
		for i, column := range mapping.Columns {
			columns[i] = sqlshape.Identifier(column.ColumnName)
		}
		if mapping.SchemaName != "public" {
			return ErrExtensionBindingInvalid
		}
		from, err := maintenancequery.From(mapping.TableName)
		if err != nil {
			return err
		}
		rows, err := db.Query(ctx, sqlshape.Text(sqlshape.Format("SELECT %s %s LIMIT 0", sqlshape.Join(columns, sqlshape.Literal(",")), from)))
		if err != nil {
			return fmt.Errorf("%w: binding mapping unavailable", ErrExtensionBindingInvalid)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return ErrExtensionBindingInvalid
		}
	}
	return nil
}
