package extensions

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
)

func resolvePostgresCodecs(plan *BackupPlan, row map[string]any) error {
	definitions, ok := objectSlice(row["logical_contracts"])
	if !ok || len(definitions) == 0 || len(definitions) > len(plan.Codecs) {
		return errors.New("logical codec definitions are missing or extra")
	}
	logical := map[string]extensioncodec.Definition{}
	familyUses := map[string]map[string]bool{}
	previous := ""
	for _, encoded := range definitions {
		if err := requireExactKeys(encoded, "codec_contract_ref", "codec_contract_sha256", "contract"); err != nil {
			return err
		}
		ref := stringValue(encoded["codec_contract_ref"])
		if ref != extensioncodec.NetworkFlowRowsSchemaID || ref <= previous {
			return errors.New("logical codec contract identity is unsupported or unordered")
		}
		previous = ref
		body, err := admittedCodecObject(encoded, "contract", "codec_contract_sha256")
		if err != nil {
			return err
		}
		definition, err := extensioncodec.ParseDefinition(body)
		if err != nil || definition.SchemaID != ref || definition.ProfileID != plan.ProfileID {
			return errors.New("logical codec contract identity mismatch")
		}
		logical[ref] = definition
		familyUses[ref] = map[string]bool{}
		for _, family := range definition.Families {
			familyUses[ref][family.LogicalFamilyID] = false
		}
	}
	encodedMappings, ok := objectSlice(row["postgres_mappings"])
	if !ok || len(encodedMappings) != len(plan.Bindings) || len(plan.Codecs) != len(plan.Bindings) {
		return errors.New("logical codec mapping set is incomplete or extra")
	}
	mappings := map[string]extensioncodec.Mapping{}
	mappingDigests := map[string]string{}
	tables := map[string]bool{}
	previous = ""
	for _, encoded := range encodedMappings {
		if err := requireExactKeys(encoded, "binding_id", "mapping_sha256", "mapping"); err != nil {
			return err
		}
		bindingID := stringValue(encoded["binding_id"])
		if bindingID <= previous {
			return errors.New("logical codec mappings are duplicated or unordered")
		}
		previous = bindingID
		body, err := admittedCodecObject(encoded, "mapping", "mapping_sha256")
		if err != nil {
			return err
		}
		mapping, err := extensioncodec.ParseMapping(body)
		if err != nil || mapping.BindingID != bindingID {
			return errors.New("logical codec mapping identity mismatch")
		}
		table := mapping.SchemaName + "." + mapping.TableName
		if tables[table] {
			return errors.New("physical table has multiple logical codec mappings")
		}
		tables[table] = true
		mappings[bindingID] = mapping
		mappingDigests[bindingID] = stringValue(encoded["mapping_sha256"])
	}
	for _, binding := range plan.Bindings {
		codec := plan.Codecs[binding.BackupCodecID]
		definition, resolved := logical[codec.CodecContractRef]
		mapping, mapped := mappings[binding.BindingID]
		if !resolved || !mapped || binding.StorageKind != "postgres" || binding.PhysicalRef != "sha256:"+mappingDigests[binding.BindingID] {
			return errors.New("backup binding logical codec or mapping is unresolved")
		}
		compiled, err := extensioncodec.NewPostgres(definition, mapping, extensioncodec.Binding{
			BindingID: binding.BindingID, LogicalFamilyID: binding.LogicalFamilyID, CodecContractRef: codec.CodecContractRef,
			LogicalIdentityAlgorithmID: codec.LogicalIdentityAlgorithmID, ContentEncodingAlgorithmID: codec.ContentEncodingAlgorithmID,
			Limits: extensioncodec.Limits{MaxItems: int64(codec.MaxItems), MaxEntryBytes: codec.MaxEntryBytes, MaxBindingBytes: codec.MaxBindingBytes},
		})
		if err != nil {
			return err
		}
		if familyUses[codec.CodecContractRef][binding.LogicalFamilyID] {
			return errors.New("logical codec family is bound more than once")
		}
		familyUses[codec.CodecContractRef][binding.LogicalFamilyID] = true
		codec.Postgres = compiled
		plan.Codecs[binding.BackupCodecID] = codec
	}
	for _, families := range familyUses {
		for _, used := range families {
			if !used {
				return errors.New("logical codec family has no binding")
			}
		}
	}
	return nil
}

func admittedCodecObject(encoded map[string]any, objectKey, digestKey string) ([]byte, error) {
	body, ok := encoded[objectKey].(map[string]any)
	if !ok {
		return nil, errors.New("logical codec object is absent")
	}
	canonical, err := canonicalJSON(body, true)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(canonical)
	if stringValue(encoded[digestKey]) != hex.EncodeToString(digest[:]) {
		return nil, errors.New("logical codec artifact digest mismatch")
	}
	return canonical, nil
}
