package main

import (
	"fmt"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
)

func extensionLogicalBackupArtifacts(profileID string, bindings []map[string]any, indexed map[string]map[string]any) ([]map[string]any, []map[string]any, error) {
	definitions := map[string]map[string]any{}
	definitionFamilies := map[string]map[string]bool{}
	mappings := []map[string]any{}
	physicalTables := map[string]bool{}
	usedMappings := map[string]bool{}
	for _, binding := range bindings {
		if binding["storage_kind"] != "postgres" {
			return nil, nil, fmt.Errorf("profile %s has an unsupported logical codec storage kind", profileID)
		}
		codec, err := extensionUniqueArtifact(indexed, "cartulary.extension_backup_binding_codec.v4", "backup_codec_id", stringValue(binding["backup_codec_id"]))
		if err != nil {
			return nil, nil, err
		}
		ref := stringValue(codec["codec_contract_ref"])
		if ref != extensioncodec.NetworkFlowRowsSchemaID {
			return nil, nil, fmt.Errorf("profile %s has an unresolved logical codec contract", profileID)
		}
		definitionObject, err := extensionUniqueArtifact(indexed, ref, "schema_id", ref)
		if err != nil {
			return nil, nil, err
		}
		definitionJSON, err := canonicalizeExtension(definitionObject)
		if err != nil {
			return nil, nil, err
		}
		definition, err := extensioncodec.ParseDefinition([]byte(definitionJSON))
		if err != nil || definition.ProfileID != profileID {
			return nil, nil, fmt.Errorf("profile %s has an invalid logical row definition", profileID)
		}
		mappingObject, err := extensionUniqueArtifact(indexed, extensioncodec.PostgresMappingSchemaID, "binding_id", stringValue(binding["binding_id"]))
		if err != nil {
			return nil, nil, err
		}
		mappingJSON, err := canonicalizeExtension(mappingObject)
		if err != nil {
			return nil, nil, err
		}
		mapping, err := extensioncodec.ParseMapping([]byte(mappingJSON))
		if err != nil {
			return nil, nil, err
		}
		usedMappings[mapping.BindingID] = true
		maxItems, itemsErr := positiveJSONInt(codec["max_items"], "codec max_items")
		maxEntry, entryErr := positiveJSONInt(codec["max_entry_bytes"], "codec max_entry_bytes")
		maxBinding, bindingErr := positiveJSONInt(codec["max_binding_bytes"], "codec max_binding_bytes")
		if itemsErr != nil || entryErr != nil || bindingErr != nil {
			return nil, nil, fmt.Errorf("profile %s has invalid codec limits", profileID)
		}
		if _, err := extensioncodec.NewPostgres(definition, mapping, extensioncodec.Binding{
			BindingID: stringValue(binding["binding_id"]), LogicalFamilyID: stringValue(binding["logical_family_id"]), CodecContractRef: ref,
			LogicalIdentityAlgorithmID: stringValue(codec["logical_identity_algorithm_id"]), ContentEncodingAlgorithmID: stringValue(codec["content_encoding_algorithm_id"]),
			Limits: extensioncodec.Limits{MaxItems: maxItems, MaxEntryBytes: maxEntry, MaxBindingBytes: maxBinding},
		}); err != nil {
			return nil, nil, err
		}
		mappingDigest, err := extensionCanonicalDigest(mappingObject)
		if err != nil {
			return nil, nil, err
		}
		if binding["physical_ref"] != "sha256:"+mappingDigest {
			return nil, nil, fmt.Errorf("profile %s mapping selector digest is stale", profileID)
		}
		tableIdentity := mapping.SchemaName + "." + mapping.TableName
		if physicalTables[tableIdentity] {
			return nil, nil, fmt.Errorf("profile %s maps multiple bindings to one physical table", profileID)
		}
		physicalTables[tableIdentity] = true
		if definitions[ref] == nil {
			definitionDigest, err := extensionCanonicalDigest(definitionObject)
			if err != nil {
				return nil, nil, err
			}
			definitions[ref] = map[string]any{"codec_contract_ref": ref, "codec_contract_sha256": definitionDigest, "contract": cloneObject(definitionObject)}
			definitionFamilies[ref] = map[string]bool{}
			for _, family := range definition.Families {
				definitionFamilies[ref][family.LogicalFamilyID] = false
			}
		}
		familyID := stringValue(binding["logical_family_id"])
		if definitionFamilies[ref][familyID] {
			return nil, nil, fmt.Errorf("profile %s repeats a logical family", profileID)
		}
		definitionFamilies[ref][familyID] = true
		mappings = append(mappings, map[string]any{"binding_id": mapping.BindingID, "mapping_sha256": mappingDigest, "mapping": cloneObject(mappingObject)})
	}
	definitionRows := make([]map[string]any, 0, len(definitions))
	for _, object := range indexed {
		if object["schema_id"] == extensioncodec.PostgresMappingSchemaID && strings.HasPrefix(stringValue(object["binding_id"]), profileID+".") && !usedMappings[stringValue(object["binding_id"])] {
			return nil, nil, fmt.Errorf("profile %s has an unbound PostgreSQL mapping", profileID)
		}
	}
	for ref, definition := range definitions {
		for _, bound := range definitionFamilies[ref] {
			if !bound {
				return nil, nil, fmt.Errorf("profile %s omits a logical codec family binding", profileID)
			}
		}
		definitionRows = append(definitionRows, definition)
	}
	sortObjectRows(definitionRows, "codec_contract_ref")
	sortObjectRows(mappings, "binding_id")
	return definitionRows, mappings, nil
}

func extensionUniqueArtifact(indexed map[string]map[string]any, schemaID, key, identity string) (map[string]any, error) {
	var found map[string]any
	for _, object := range indexed {
		if object["schema_id"] != schemaID || object[key] != identity {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("extension logical codec artifact identity is ambiguous")
		}
		found = object
	}
	if found == nil {
		return nil, fmt.Errorf("extension logical codec artifact is unresolved")
	}
	return found, nil
}
