package registryfacts

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const CollisionCondition = "extension_registry_collision"

// RegistryInput contains admitted declarations, never their repeated descriptor,
// binding, or publication projections. Configuration and implementation rows are
// supplied by the caller's explicit catalog. References are not declarations.
type RegistryInput struct {
	Fragments              []map[string]any
	ConfigurationContracts []map[string]any
	ImplementationBindings []map[string]any
	BaseReservations       []map[string]any
}

type collisionDeclaration struct {
	profiles  []string
	facts     []string
	signature string
}

type collisionGroup struct {
	class, token string
	declarations []collisionDeclaration
}

type registryRoute struct {
	profile, root string
	facts         []string
}

type registryFacets struct {
	profile, token string
	definition     map[string]any
	members        map[string][]string
}

// RegistryFindings collects the complete safe owner-input/collision stage before
// descriptor derivation or execution. Logical facets are paired, not added: a
// workspace declaration and its contribution, or an executable migration and
// its matching ledger definition, describe one logical declaration.
func RegistryFindings(input RegistryInput) ([]OwnerFinding, error) {
	fragments, err := NormalizeFragments(input.Fragments)
	if err != nil {
		return nil, err
	}
	recognized := map[string]bool{}
	primaryOwners := map[string]map[string]bool{}
	for _, fragment := range fragments {
		if fragment["owner_id"] != "core00" {
			continue
		}
		for _, raw := range fragment["facts"].([]any) {
			fact := raw.(map[string]any)
			if fact["fact_kind"] == "recognized_profile" {
				profile := fact["profile_id"].(string)
				recognized[profile] = true
				if primaryOwners[profile] == nil {
					primaryOwners[profile] = map[string]bool{}
				}
				primaryOwners[profile][fact["primary_owner_id"].(string)] = true
			}
		}
	}
	// An unallocated owner cannot authorize diagnostic identifiers. Preserve all
	// declared primary owners here so duplicate recognition still reaches the
	// collision stage without arbitrarily selecting one declaration.
	for _, fragment := range fragments {
		for _, raw := range fragment["facts"].([]any) {
			fact := raw.(map[string]any)
			profile, kind := fact["profile_id"].(string), fact["fact_kind"].(string)
			if !recognized[profile] {
				continue
			}
			allocated := false
			for primary := range primaryOwners[profile] {
				allocated = allocated || fragment["owner_id"] == FactOwner(kind, primary)
			}
			if !allocated {
				return nil, fmt.Errorf("owner allocation invalid")
			}
		}
	}
	for _, contract := range input.ConfigurationContracts {
		profile := contract["profile_id"].(string)
		if !recognized[profile] {
			return nil, fmt.Errorf("configuration namespace identity invalid")
		}
		namespace := contract["configuration_namespace"].(string)
		for _, raw := range contract["keys"].([]any) {
			key := raw.(map[string]any)["key"].(string)
			if !strings.HasPrefix(key, namespace+".") || key == profile+".claimed" {
				return nil, fmt.Errorf("configuration namespace membership invalid")
			}
		}
	}
	findings, err := ownerInputFindings(fragments, recognized)
	if err != nil || len(findings) == 1 && findings[0].ConditionID == OwnerOverflowCondition {
		return findings, err
	}
	overflow := func() ([]OwnerFinding, error) {
		return []OwnerFinding{{ConditionID: OwnerOverflowCondition, Path: "$", Details: map[string]any{"phase": "registry_generation", "limit": DiagnosticLimit, "actual": DiagnosticLimit + 1}}}, nil
	}
	emit := func(class string, profiles, tokens []string, count int) bool {
		var duplicate any
		if count != 0 {
			duplicate = count
		}
		findings = append(findings, OwnerFinding{ConditionID: CollisionCondition, Path: "$", Details: map[string]any{"collision_class": class, "conflicting_profile_ids": uniqueStrings(profiles), "conflicting_tokens": uniqueStrings(tokens), "duplicate_count": duplicate}})
		return len(findings) <= DiagnosticLimit
	}
	groups := map[string]*collisionGroup{}
	add := func(class, token string, declaration collisionDeclaration) {
		key := class + "\x00" + token
		if groups[key] == nil {
			groups[key] = &collisionGroup{class: class, token: token}
		}
		groups[key].declarations = append(groups[key].declarations, declaration)
	}
	fallback := map[string]*collisionGroup{}
	covered := map[string]bool{}
	routeFacets, workspaceFacets, migrationFacets := map[string]*registryFacets{}, map[string]*registryFacets{}, map[string]*registryFacets{}
	pair := func(catalog map[string]*registryFacets, key, profile, token, facet, factKey string, definition map[string]any) {
		if catalog[key] == nil {
			catalog[key] = &registryFacets{profile: profile, token: token, definition: definition, members: map[string][]string{}}
		}
		catalog[key].members[facet] = append(catalog[key].members[facet], factKey)
	}
	graph := map[string][]string{}
	selfEdges := []collisionDeclaration{}
	for fi, fragment := range fragments {
		fragmentProfiles := []string{}
		for factIndex, raw := range fragment["facts"].([]any) {
			fact := raw.(map[string]any)
			identity, err := Identify(fact)
			if err != nil {
				return nil, err
			}
			profile, kind, key := identity.profile, identity.kind, identity.Key()
			occurrence := fmt.Sprintf("%d/%d", fi, factIndex)
			if !recognized[profile] {
				continue
			}
			fragmentProfiles = append(fragmentProfiles, profile)
			declaration := collisionDeclaration{profiles: []string{profile}, facts: []string{occurrence}}
			if fallback[key] == nil {
				fallback[key] = &collisionGroup{class: "owner_fact_identity", token: identity.Token()}
			}
			fallback[key].declarations = append(fallback[key].declarations, declaration)
			switch kind {
			case "recognized_profile":
				if fragment["owner_id"] == "core00" {
					add("profile_id", profile, declaration)
				}
			case "claim_configuration":
				add("claim_key", fact["claim_config_key"].(string), declaration)
			case "public_schema":
				add("public_schema_id", fact["public_schema_id"].(string), declaration)
			case "worker_kind":
				add("worker_kind", fact["worker_kind"].(string), declaration)
			case "job_kind":
				add("job_kind", fact["job_kind_contract"].(map[string]any)["job_kind"].(string), declaration)
			case "workspace":
				token := profile + "/" + fact["workspace_key"].(string)
				pair(workspaceFacets, token, profile, token, "workspace", occurrence, nil)
			case "route_family":
				root := fact["route_family"].(string)
				pair(routeFacets, profile+"\x00"+root, profile, root, "route_family", occurrence, nil)
			case "runtime_dependency":
				target := fact["dependency_profile_id"].(string)
				// Unknown targets are never exposed by collision diagnostics. Their
				// invalid reference is handled by the dependency owner admission gate.
				if !recognized[target] {
					delete(fallback, key)
					continue
				}
				add("dependency_duplicate", profile+"->"+target, declaration)
				graph[profile] = append(graph[profile], target)
				if target == profile {
					selfEdges = append(selfEdges, declaration)
				}
			case "state_ownership":
				state := fact["state_ownership"].(map[string]any)
				if state["kind"] == "extension_versioned" {
					body, err := canonicaljson.MarshalExtension(state)
					if err != nil {
						return nil, err
					}
					declaration.signature = string(body)
					add("migration_lineage_id", state["migration_lineage_id"].(string), declaration)
				}
			case "migration_definition", "migration_ledger_definition":
				definition := fact[kind].(map[string]any)
				digest := ""
				if kind == "migration_definition" {
					body, err := canonicaljson.MarshalExtension(definition)
					if err != nil {
						return nil, err
					}
					sum := sha256.Sum256(body)
					digest = hex.EncodeToString(sum[:])
				} else {
					digest = definition["migration_definition_sha256"].(string)
				}
				logical := map[string]any{"profile_id": profile, "migration_id": definition["migration_id"], "migration_lineage_id": definition["migration_lineage_id"], "from_state_version": definition["from_state_version"], "to_state_version": definition["to_state_version"], "migration_definition_sha256": digest}
				body, err := canonicaljson.MarshalExtension(logical)
				if err != nil {
					return nil, err
				}
				pair(migrationFacets, string(body), profile, definition["migration_id"].(string), kind, occurrence, definition)
			case "contribution":
				contribution := fact["contribution"].(map[string]any)
				add("contribution_id", contribution["contribution_id"].(string), declaration)
				switch contribution["kind"] {
				case "http_route_family":
					root := contribution["route_family"].(string)
					pair(routeFacets, profile+"\x00"+root, profile, root, "contribution", occurrence, nil)
				case "incident_workspace":
					token := profile + "/" + contribution["workspace_key"].(string)
					pair(workspaceFacets, token, profile, token, "contribution", occurrence, nil)
				case "extension_resource_kind":
					add("extension_resource_kind", contribution["resource_kind"].(string), declaration)
				case "import_target":
					add("import_target", contribution["target_kind"].(string), declaration)
				case "job_resource_ref_kind":
					add("job_resource_ref_kind", contribution["resource_ref_kind"].(string), declaration)
				case "cross_owner_transaction_participant", "snapshot_reporting_participant", "incident_portability_participant":
					add("transaction_participant", contribution["participant_id"].(string), declaration)
				case "deployment_admin_panel":
					add("deployment_admin_panel", contribution["panel_key"].(string), declaration)
				case "authentication_entry":
					add("authentication_entry", contribution["entry_key"].(string), declaration)
				}
			}
		}
		add("owner_fragment_id", fragment["owner_fragment_id"].(string), collisionDeclaration{profiles: uniqueStrings(fragmentProfiles)})
	}
	for _, contract := range input.ConfigurationContracts {
		profile := contract["profile_id"].(string)
		if recognized[profile] {
			add("configuration_namespace", contract["configuration_namespace"].(string), collisionDeclaration{profiles: []string{profile}})
		}
	}
	for _, binding := range input.ImplementationBindings {
		profile := binding["profile_id"].(string)
		if recognized[profile] {
			add("implementation_binding_duplicate", profile, collisionDeclaration{profiles: []string{profile}})
		}
	}
	routes := []registryRoute{}
	for _, family := range []struct {
		kind    string
		catalog map[string]*registryFacets
	}{{"route", routeFacets}, {"workspace", workspaceFacets}, {"migration", migrationFacets}} {
		catalog := family.catalog
		for _, key := range sortedKeys(catalog) {
			facet := catalog[key]
			maximum := 0
			for _, members := range facet.members {
				maximum = max(maximum, len(members))
			}
			for i := 0; i < maximum; i++ {
				facts := []string{}
				for _, kind := range sortedKeys(facet.members) {
					if i < len(facet.members[kind]) {
						facts = append(facts, facet.members[kind][i])
					}
				}
				declaration := collisionDeclaration{profiles: []string{facet.profile}, facts: facts}
				switch family.kind {
				case "route":
					routes = append(routes, registryRoute{profile: facet.profile, root: facet.token, facts: facts})
				case "workspace":
					add("workspace_identity", facet.token, declaration)
				default:
					add("migration_id", facet.token, declaration)
					add("migration_edge", fmt.Sprintf("%s:%v->%v", facet.profile, facet.definition["from_state_version"], facet.definition["to_state_version"]), declaration)
				}
			}
		}
	}
	for _, key := range sortedKeys(groups) {
		group := groups[key]
		if len(group.declarations) < 2 {
			continue
		}
		if group.class == "migration_lineage_id" {
			first := group.declarations[0]
			compatible := true
			for _, declaration := range group.declarations[1:] {
				if declaration.profiles[0] != first.profiles[0] || declaration.signature != first.signature {
					compatible = false
					break
				}
			}
			if compatible {
				continue
			}
		}
		profiles := []string{}
		for _, declaration := range group.declarations {
			profiles = append(profiles, declaration.profiles...)
			for _, fact := range declaration.facts {
				covered[fact] = true
			}
		}
		if !emit(group.class, profiles, []string{group.token}, len(group.declarations)) {
			return overflow()
		}
	}
	for _, edge := range selfEdges {
		profile := edge.profiles[0]
		for _, fact := range edge.facts {
			covered[fact] = true
		}
		if !emit("dependency_self", edge.profiles, []string{profile + "->" + profile}, 0) {
			return overflow()
		}
	}
	vertices := sortedKeys(recognized)
	for profile, targets := range graph {
		graph[profile] = uniqueStrings(targets)
	}
	for _, component := range DependencyComponents(vertices, graph) {
		if len(component) < 2 {
			continue
		}
		members := map[string]bool{}
		for _, profile := range component {
			members[profile] = true
		}
		edges := []string{}
		for _, profile := range component {
			for _, target := range graph[profile] {
				if members[target] {
					edges = append(edges, profile+"->"+target)
				}
			}
		}
		if !emit("dependency_cycle", component, edges, 0) {
			return overflow()
		}
	}
	for i, left := range routes {
		for _, right := range routes[i+1:] {
			if RoutesOverlap(left.root, right.root) {
				for _, fact := range append(append([]string{}, left.facts...), right.facts...) {
					covered[fact] = true
				}
				if !emit("route_family_overlap", []string{left.profile, right.profile}, []string{left.root, right.root}, 0) {
					return overflow()
				}
			}
		}
		for _, base := range input.BaseReservations {
			root := base["path_template"].(string)
			if BaseReservationOverlaps(left.root, root, base["match_scope"].(string)) {
				if !emit("base_route_capture", []string{left.profile}, []string{left.root, root}, 0) {
					return overflow()
				}
			}
		}
	}
	for _, key := range sortedKeys(fallback) {
		group := fallback[key]
		allCovered := true
		for _, declaration := range group.declarations {
			for _, occurrence := range declaration.facts {
				if !covered[occurrence] {
					allCovered = false
				}
			}
		}
		if len(group.declarations) > 1 && !allCovered {
			if !emit(group.class, group.declarations[0].profiles, []string{group.token}, len(group.declarations)) {
				return overflow()
			}
		}
	}
	return findings, nil
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func uniqueStrings(values []string) []string {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	return sortedKeys(set)
}
