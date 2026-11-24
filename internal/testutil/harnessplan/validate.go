package harnessplan

import (
	"fmt"
	"slices"
)

// Fixed planning vocabularies.
var (
	postgresIsolations = []string{"none", "shared", "exclusive"}
	processIsolations  = []string{"", "exclusive"}
	cachePolicies      = []string{"none", "same_run", "content_addressed"}
)

// requiredCapacityPolicies are the capacity rules capture resolves.
var requiredCapacityPolicies = []string{"cpu_tokens", "io_tokens", "memory_bytes", "object_store_lanes", "port_lanes", "postgres_heavy_lanes", "postgres_lanes", "process_slots", "writable_volume"}

// validate checks every invariant planning depends on, so the planner can
// index the model without defensive lookups.
func (in *Inputs) validate() error {
	fail := func(source, format string, args ...any) error {
		return fmt.Errorf("%w: %s: %s", ErrInput, source, fmt.Sprintf(format, args...))
	}
	for _, name := range requiredCapacityPolicies {
		if _, ok := in.Registry.CapacityPolicies[name]; !ok {
			return fail(registryPath, "missing capacity policy %s", name)
		}
	}
	resources := map[string]bool{}
	for _, profile := range in.Topology.ResourceProfiles {
		if resources[profile.ID] {
			return fail(topologyPath, "duplicate resource profile %s", profile.ID)
		}
		resources[profile.ID] = true
	}
	runtimes := map[string]bool{}
	for _, profile := range in.Topology.RuntimeProfiles {
		runtimes[profile.ID] = true
	}
	binaries := map[string]RuntimeBinary{}
	for _, binary := range in.Topology.RuntimeBinaries {
		binaries[binary.ID] = binary
	}
	for id, binary := range binaries {
		if _, ok := in.Owner.Operations[binary.ProducerOperation]; !ok {
			return fail(topologyPath, "runtime binary %s names unknown producer %s", id, binary.ProducerOperation)
		}
	}
	services := in.Topology.ServiceResourceMinimums
	checkServices := func(source, subject string, deps []string, isolation string) error {
		if !slices.Contains(postgresIsolations, isolation) {
			return fail(source, "%s: unknown postgres_isolation %q", subject, isolation)
		}
		for _, service := range deps {
			if _, ok := services[service]; !ok {
				return fail(source, "%s: unknown service %s", subject, service)
			}
		}
		if slices.Contains(deps, "postgres") != (isolation != "none") {
			return fail(source, "%s: postgres_isolation %q disagrees with its service dependencies", subject, isolation)
		}
		return nil
	}

	owner := in.Owner
	if owner.DefaultTimeoutMS <= 0 || owner.AgingQuantumMS <= 0 || owner.GoUnitTargetMS <= 0 || owner.GoHeavyUnitTargetMS <= 0 {
		return fail(ownerPath, "default timeout, aging quantum and Go unit targets must be positive")
	}
	if _, ok := owner.InvocationPolicies[RunnerGo]; !ok {
		return fail(ownerPath, "missing the go invocation policy")
	}
	if owner.InvocationPolicies[RunnerVitest].UnitTargetMS <= 0 {
		return fail(ownerPath, "the vitest invocation policy needs a positive unit target")
	}
	for runner, policy := range owner.InvocationPolicies {
		for _, id := range policy.PreparationOperations {
			if _, ok := owner.Operations[id]; !ok {
				return fail(ownerPath, "the %s invocation policy prepares unknown operation %s", runner, id)
			}
		}
	}
	for id, unit := range owner.PolicyUnits {
		subject := "policy unit " + id
		if _, ok := owner.Operations[id]; !ok {
			return fail(ownerPath, "%s has no operation", subject)
		}
		if !resources[unit.ResourceProfileID] {
			return fail(ownerPath, "%s: unknown resource profile %s", subject, unit.ResourceProfileID)
		}
		if !slices.Contains(cachePolicies, unit.CachePolicy) {
			return fail(ownerPath, "%s: unknown cache policy %q", subject, unit.CachePolicy)
		}
		if unit.EstimatedWorkMS <= 0 {
			return fail(ownerPath, "%s: estimated_work_ms must be positive", subject)
		}
		if err := checkServices(ownerPath, subject, unit.ServiceDependencies, unit.PostgresIsolation); err != nil {
			return err
		}
		for _, need := range unit.Needs {
			_, operation := owner.Operations[need]
			_, union := owner.TargetMembers[need]
			if !operation && !union {
				return fail(ownerPath, "%s needs %s, which is neither an operation nor a target", subject, need)
			}
		}
	}
	tiers := map[string]bool{}
	for _, tier := range owner.TierOrder {
		tiers[tier] = true
	}
	for target, tier := range owner.AggregateTiers {
		if !tiers[tier] {
			return fail(ownerPath, "aggregate %s uses unknown tier %s", target, tier)
		}
		for _, root := range owner.AggregatePolicyRoots[target] {
			_, scheduled := owner.PolicyUnits[root]
			_, union := owner.TargetMembers[root]
			if !scheduled && !union {
				return fail(ownerPath, "aggregate %s root %s is neither a policy unit nor a target", target, root)
			}
		}
		for _, parent := range owner.AggregatePolicyInherits[target] {
			if _, ok := owner.AggregateTiers[parent]; !ok {
				return fail(ownerPath, "aggregate %s inherits unknown aggregate %s", target, parent)
			}
		}
	}

	rows := map[string]bool{}
	owners := map[string]bool{}
	for _, registered := range in.Catalog.Owners {
		owners[registered.OwnerID] = true
	}
	for _, row := range in.Catalog.Rows {
		source := "catalog row " + row.RowID
		if rows[row.RowID] {
			return fail(catalogRegistryPath, "duplicate row %s", row.RowID)
		}
		rows[row.RowID] = true
		if !owners[row.OwnerID] {
			return fail(catalogRegistryPath, "%s: unregistered owner %s", source, row.OwnerID)
		}
		if row.Status != "active" {
			return fail(catalogRegistryPath, "%s: status %q is not active", source, row.Status)
		}
		if err := row.Selector.check(row.Runner, owner.Operations); err != nil {
			return fail(catalogRegistryPath, "%s: %v", source, err)
		}
		if !resources[row.ResourceProfileID] {
			return fail(catalogRegistryPath, "%s: unknown resource profile %s", source, row.ResourceProfileID)
		}
		if !runtimes[row.RuntimeProfileID] {
			return fail(catalogRegistryPath, "%s: unknown runtime profile %s", source, row.RuntimeProfileID)
		}
		if !slices.Contains(processIsolations, row.ProcessIsolation) {
			return fail(catalogRegistryPath, "%s: unknown process_isolation %q", source, row.ProcessIsolation)
		}
		if row.AggregatePlacement != "explicit_only" && !tiers[row.AggregatePlacement] {
			return fail(catalogRegistryPath, "%s: unknown aggregate placement %q", source, row.AggregatePlacement)
		}
		if _, ok := owner.EvidenceEstimatesMS[row.EvidenceClass]; !ok && row.EstimatedWorkMS == 0 {
			return fail(catalogRegistryPath, "%s: evidence class %s has no estimate", source, row.EvidenceClass)
		}
		if err := checkServices(catalogRegistryPath, source, row.ServiceDependencies, row.PostgresIsolation); err != nil {
			return err
		}
		for _, id := range row.RuntimeBinaryIDs {
			if _, ok := binaries[id]; !ok {
				return fail(catalogRegistryPath, "%s: unknown runtime binary %s", source, id)
			}
		}
		for _, id := range row.PreparationOperations {
			if _, ok := owner.Operations[id]; !ok {
				return fail(catalogRegistryPath, "%s: unknown preparation operation %s", source, id)
			}
		}
	}
	for id, unit := range owner.PolicyUnits {
		for _, required := range unit.RequiredRows {
			if !rows[required.RowID] {
				return fail(ownerPath, "policy unit %s requires unknown row %s", id, required.RowID)
			}
		}
	}
	for id := range owner.RowExecutionContexts {
		if !rows[id] {
			return fail(ownerPath, "row execution context names unknown row %s", id)
		}
	}
	for _, name := range owner.ExplicitSelections {
		if _, ok := owner.TargetMembers[name]; !ok {
			return fail(ownerPath, "explicit selection %s is not a target", name)
		}
	}
	for name, members := range owner.TargetMembers {
		for _, member := range members {
			switch {
			case member.Name != "":
				_, operation := owner.Operations[member.Name]
				_, union := owner.TargetMembers[member.Name]
				if !operation && !union {
					return fail(ownerPath, "target %s member %s is neither an operation nor a target", name, member.Name)
				}
			case member.RowID != "":
				if !rows[member.RowID] {
					return fail(ownerPath, "target %s member names unknown row %s", name, member.RowID)
				}
			case member.OperationID != "":
				if _, ok := owner.Operations[member.OperationID]; !ok {
					return fail(ownerPath, "target %s member names unknown operation %s", name, member.OperationID)
				}
			case member.Catalog != nil:
				if !slices.Contains([]string{RunnerGo, RunnerVitest, RunnerPlaywright, RunnerShell}, member.Catalog.Runner) {
					return fail(ownerPath, "target %s member names unknown runner %s", name, member.Catalog.Runner)
				}
			}
		}
	}
	bound := map[string]string{}
	for _, profile := range in.Cache.Profiles {
		source := "cache profile " + profile.ProfileID
		if _, ok := bound["profile:"+profile.ProfileID]; ok {
			return fail(cacheRegistryPath, "duplicate %s", source)
		}
		bound["profile:"+profile.ProfileID] = profile.ProfileID
		if !slices.Contains([]string{"same_run", "content_addressed"}, profile.Policy) || (profile.DependencyStrategy == "unit_aware") != (profile.UnitInputs != nil) {
			return fail(cacheRegistryPath, "%s: unknown policy or unit inputs that disagree with its strategy", source)
		}
		for _, id := range profile.OperationIDs {
			if _, ok := owner.Operations[id]; !ok {
				return fail(cacheRegistryPath, "%s binds unknown operation %s", source, id)
			}
		}
		for _, key := range append(slices.Clone(profile.OperationIDs), profile.RunnerTypes...) {
			if other, ok := bound[key]; ok {
				return fail(cacheRegistryPath, "%s and %s both bind %s", source, other, key)
			}
			bound[key] = profile.ProfileID
		}
	}
	policies := map[string]string{}
	for _, profile := range in.Cache.Profiles {
		policies[profile.ProfileID] = profile.Policy
	}
	for id, unit := range owner.PolicyUnits {
		if profile, ok := bound[id]; unit.CachePolicy != "none" && (!ok || policies[profile] != unit.CachePolicy) {
			return fail(ownerPath, "policy unit %s has cache policy %s but no cache profile with that policy binds it", id, unit.CachePolicy)
		}
	}
	for _, builder := range in.Fixtures.Builders {
		source := "fixture builder " + builder.FixtureProfileID
		if !resources[builder.ResourceProfileID] || !runtimes[builder.RuntimeProfileID] {
			return fail(fixtureBuilderPath, "%s: unknown resource or runtime profile", source)
		}
		if err := checkServices(fixtureBuilderPath, source, builder.ServiceDependencies, builder.PostgresIsolation); err != nil {
			return err
		}
	}
	return in.resolveFixtures()
}

// check enforces the selector fields each runner requires and forbids the
// others, so a row selects exactly one kind of work.
func (s Selector) check(runner string, operations map[string]Operation) error {
	goFields := s.Package != "" || len(s.Tests) != 0
	fileFields := s.File != "" || s.ProjectID != "" || len(s.Titles) != 0
	browserFields := s.Stage != "" || len(s.ScenarioIDs) != 0
	switch runner {
	case RunnerGo:
		if s.Package == "" || len(s.Tests) == 0 || fileFields || browserFields || s.OperationID != "" {
			return fmt.Errorf("go selector needs exactly a package and tests")
		}
	case RunnerVitest:
		if s.File == "" || s.ProjectID == "" || len(s.Titles) == 0 || goFields || browserFields || s.OperationID != "" {
			return fmt.Errorf("vitest selector needs exactly a file, project and titles")
		}
	case RunnerPlaywright:
		if s.File == "" || s.ProjectID == "" || len(s.Titles) == 0 || s.Stage == "" || len(s.ScenarioIDs) == 0 || goFields || s.OperationID != "" {
			return fmt.Errorf("playwright selector needs exactly a file, project, stage, scenarios and titles")
		}
	case RunnerShell:
		if _, ok := operations[s.OperationID]; !ok || goFields || fileFields || browserFields {
			return fmt.Errorf("shell selector needs exactly a known operation")
		}
	default:
		return fmt.Errorf("unknown runner %q", runner)
	}
	return nil
}
