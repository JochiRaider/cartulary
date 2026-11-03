package extensions

import (
	"errors"
	"fmt"
	"sort"
)

// InactiveConfigurationPolicy is the narrow, immutable configuration-policy
// projection supplied by app composition to the platform configuration owner.
type InactiveConfigurationPolicy struct {
	ProfileID string
	Namespace string
	ClaimKey  string
	Key       string
	Kind      string
	Schema    map[string]any
}

func cloneInactiveConfigurationPolicy(policy InactiveConfigurationPolicy) InactiveConfigurationPolicy {
	policy.Schema = cloneObject(policy.Schema)
	return policy
}

func (c *Coordinator) InactiveConfigurationPolicies() []InactiveConfigurationPolicy {
	if c == nil {
		return nil
	}
	result := make([]InactiveConfigurationPolicy, len(c.inactiveConfigurationPolicies))
	for index, policy := range c.inactiveConfigurationPolicies {
		result[index] = cloneInactiveConfigurationPolicy(policy)
	}
	return result
}

func admitInactiveConfigurationPolicies(source ArtifactSource, records map[string]profileRecord, orderedProfileIDs []string) ([]InactiveConfigurationPolicy, error) {
	schemaSet, _, err := readArtifactObject(source, "contracts/extensions/specification/inactive-value-schemas.json")
	if err != nil {
		return nil, err
	}
	if schemaSet["schema_id"] != "cartulary.extension_inactive_value_schema_set.v1" {
		return nil, invalidArtifact("inactive_value_schema_set", errors.New("schema id mismatch"))
	}
	schemas := map[string]map[string]any{}
	schemaRows, ok := objectSlice(schemaSet["schemas"])
	if !ok {
		return nil, invalidArtifact("inactive_value_schema_set", errors.New("schemas must be an array"))
	}
	for _, row := range schemaRows {
		schemaID := stringValue(row["inactive_value_schema_id"])
		schema, ok := row["schema"].(map[string]any)
		if schemaID == "" || !ok {
			return nil, invalidArtifact("inactive_value_schema_set", errors.New("schema row is incomplete"))
		}
		if _, duplicate := schemas[schemaID]; duplicate {
			return nil, invalidArtifact("inactive_value_schema_set", fmt.Errorf("duplicate schema %s", schemaID))
		}
		schemas[schemaID] = cloneObject(schema)
	}

	policies := make([]InactiveConfigurationPolicy, 0)
	seenKeys := map[string]struct{}{}
	for _, profileID := range orderedProfileIDs {
		contract, contractArtifact, readErr := readArtifactObject(source, "contracts/extensions/profiles/"+profileID+"/configuration.json")
		if readErr != nil {
			return nil, readErr
		}
		if contract["schema_id"] != "cartulary.extension_profile_configuration_contract.v4" || stringValue(contract["profile_id"]) != profileID {
			return nil, invalidArtifact("configuration_contract", fmt.Errorf("profile mismatch for %s", profileID))
		}
		record := records[profileID]
		record.configurationContract = cloneObject(contract)
		record.configurationContractSHA256 = contractArtifact.SHA256
		records[profileID] = record
		keys, ok := objectSlice(contract["keys"])
		if !ok {
			return nil, invalidArtifact("configuration_contract", fmt.Errorf("keys are not an array for %s", profileID))
		}
		prestage := make([]string, 0)
		for _, row := range keys {
			key := stringValue(row["key"])
			kind := stringValue(row["inactive_policy"])
			if key == "" || (kind != "forbidden" && kind != "syntax_only") {
				return nil, invalidArtifact("configuration_contract", fmt.Errorf("inactive policy is invalid for %s", profileID))
			}
			if _, duplicate := seenKeys[key]; duplicate {
				return nil, registryInvalid("duplicate_configuration_key_projection")
			}
			seenKeys[key] = struct{}{}
			policy := InactiveConfigurationPolicy{
				ProfileID: profileID,
				Namespace: stringValue(contract["configuration_namespace"]),
				ClaimKey:  records[profileID].descriptor.ClaimConfigKey,
				Key:       key,
				Kind:      kind,
			}
			if kind == "syntax_only" {
				schemaID := stringValue(row["inactive_value_schema_id"])
				schema, exists := schemas[schemaID]
				if !exists {
					return nil, invalidArtifact("configuration_contract", fmt.Errorf("inactive schema %s is unavailable", schemaID))
				}
				policy.Schema = cloneObject(schema)
				prestage = append(prestage, key)
			} else if row["inactive_value_schema_id"] != nil {
				return nil, invalidArtifact("configuration_contract", fmt.Errorf("forbidden policy %s has an inactive schema", key))
			}
			policies = append(policies, policy)
		}
		sort.Strings(prestage)
		descriptorPrestage := catalogStrings(records[profileID].descriptorObject["prestage_config_keys"])
		if fmt.Sprint(prestage) != fmt.Sprint(descriptorPrestage) {
			return nil, unavailableBinding(profileID, records[profileID].descriptor.ContractMajor)
		}
	}
	sort.Slice(policies, func(i, j int) bool { return policies[i].Key < policies[j].Key })
	return policies, nil
}

// ConfigurationSchemaContract is the immutable metadata projection that
// application composition matches to actual typed configuration owners.
type ConfigurationSchemaContract struct {
	ProfileID string
	Namespace string
	Values    []ConfigurationValueSchema
}
type ConfigurationValueSchema struct {
	Key      string
	SchemaID string
}

func (c *Coordinator) ConfigurationSchemaContracts() []ConfigurationSchemaContract {
	if c == nil {
		return nil
	}
	contracts := make([]ConfigurationSchemaContract, 0, len(c.orderedProfileIDs))
	for _, profile := range c.orderedProfileIDs {
		object := c.profiles[profile].configurationContract
		contract := ConfigurationSchemaContract{ProfileID: profile, Namespace: stringValue(object["configuration_namespace"]), Values: []ConfigurationValueSchema{}}
		rows, _ := objectSlice(object["keys"])
		for _, row := range rows {
			contract.Values = append(contract.Values, ConfigurationValueSchema{Key: stringValue(row["key"]), SchemaID: stringValue(row["value_schema_id"])})
		}
		contracts = append(contracts, contract)
	}
	return contracts
}
