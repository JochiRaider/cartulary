package extensions

import (
	"context"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

func dependencyCoordinator(edges map[string][]string) *Coordinator {
	c := &Coordinator{registrySHA256: testDigest("a"), profiles: map[string]profileRecord{}, statePlans: map[string]StatePlan{}}
	for id, targets := range edges {
		d := Descriptor{ProfileID: id, Claimable: true, ContractMajor: 1}
		for _, target := range targets {
			d.Dependencies = append(d.Dependencies, Dependency{ProfileID: target, RequiredContractMajor: 1})
		}
		c.profiles[id] = profileRecord{descriptor: d, descriptorSHA256: testDigest("b"), bindingSHA256: testDigest("c"),
			bindingObject:         map[string]any{"preflight_algorithm_id": id + ".admit", "post_migration_algorithm_id": id + ".admit"},
			configurationContract: map[string]any{"keys": []any{}}, configurationContractSHA256: testDigest("d")}
		c.orderedProfileIDs = append(c.orderedProfileIDs, id)
	}
	sort.Strings(c.orderedProfileIDs)
	return c
}

func verifyDependencyFindingBound(t *testing.T) {
	for _, mode := range []string{"missing", "wrong phase", "wrong formatter"} {
		t.Run("overflow registration "+mode, func(t *testing.T) {
			source := newMutableGeneratedSource()
			found := false
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				rows, ok := object["conditions"].([]any)
				if !ok || object["schema_id"] != "cartulary.extension_validation_condition_registry.v2" {
					return
				}
				kept := []any{}
				for _, raw := range rows {
					condition := raw.(map[string]any)
					if condition["condition_id"] == registryfacts.DependencyOverflowCondition {
						found = true
						if mode == "missing" {
							continue
						}
						if mode == "wrong phase" {
							condition["phase"] = "profile_preflight"
						} else {
							condition["actual_formatter_id"] = "diagnostic_redacted_v1"
						}
					}
					kept = append(kept, condition)
				}
				object["conditions"] = kept
			})
			if !found {
				t.Fatal("current overflow registration fixture missing")
			}
			_, err := NewCoordinator(source)
			var failure *StartupValidationError
			if !errors.As(err, &failure) || len(failure.Findings) != 1 || failure.Findings[0].ReasonCode != "extension_registry_invalid" {
				t.Fatalf("invalid overflow registration admitted: %v", err)
			}
		})
	}
	registered := requireGeneratedCoordinator(t)
	for _, tc := range []struct {
		name     string
		claims   int
		oneExtra bool
	}{
		{"at capacity", 128, false}, {"first excess", 129, true}, {"many excess", 240, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			edges := map[string][]string{}
			requested := []string{}
			for i := 0; i < 16; i++ {
				edges[fmt.Sprintf("target_%02d", i)] = nil
			}
			for i := 0; i < tc.claims; i++ {
				id := fmt.Sprintf("claim_%03d", i)
				requested = append(requested, id)
				for j := 0; j < 16; j++ {
					edges[id] = append(edges[id], fmt.Sprintf("target_%02d", j))
				}
			}
			if tc.oneExtra {
				edges["claim_128"] = []string{"target_00"}
			}
			c := dependencyCoordinator(edges)
			c.validateFinding = registered.validateFinding
			c.validationConditions = registered.validationConditions
			for i := 0; i < 16; i++ {
				id := fmt.Sprintf("target_%02d", i)
				record := c.profiles[id]
				record.descriptor.ContractMajor = 2
				c.profiles[id] = record
			}
			if tc.oneExtra {
				record := c.profiles["claim_128"]
				record.descriptor.Dependencies[0].RequiredContractMajor = 2
				c.profiles["claim_128"] = record
			}
			resolution, err := c.ResolveClaims(requested)
			if len(resolution.AdmissionOrder()) != 0 {
				t.Fatal("dependency failure produced partial execution order")
			}
			var failure *StartupValidationError
			if !errors.As(err, &failure) {
				t.Fatalf("missing closed dependency diagnostics: %v", err)
			}
			if tc.claims == 128 {
				if len(failure.Findings) != 4096 {
					t.Fatalf("at-capacity list truncated: %d", len(failure.Findings))
				}
				counts := map[string]int{}
				for _, finding := range failure.Findings {
					counts[finding.ReasonCode]++
				}
				if counts["extension_dependency_not_claimed"] != 2048 || counts["extension_dependency_incompatible"] != 2048 {
					t.Fatalf("lost dependency classes: %#v", counts)
				}
				return
			}
			if len(failure.Findings) != 1 || failure.Findings[0].ReasonCode != "extension_diagnostic_overflow" {
				t.Fatalf("oversized dependency result: %d findings", len(failure.Findings))
			}
			finding := failure.Findings[0]
			if finding.Path != "$" || finding.Details["phase"] != "dependency_validation" || fmt.Sprint(finding.Details["limit"]) != "4096" || fmt.Sprint(finding.Details["actual"]) != "4097" {
				t.Fatalf("incorrect first-excess finding: %#v", finding)
			}
			var cause *ValidationError
			if !errors.As(err, &cause) || len(cause.findings) != 1 || cause.findings[0].Code != "extension_diagnostic_overflow" {
				t.Fatal("overflow retained the ordinary phase findings")
			}
		})
	}
}

func verifyDependencyGraphs(t *testing.T) {
	t.Helper()
	for _, tc := range []struct {
		name  string
		edges map[string][]string
		want  []string
	}{
		{"independent cycles exclude blocked dependent", map[string][]string{"a": {"b"}, "b": {"a"}, "c": {"d"}, "d": {"c"}, "e": {"a"}}, []string{"dependency_cycle:a->b,b->a", "dependency_cycle:c->d,d->c"}},
		{"self multiplicity and duplicate grouping", map[string][]string{"a": {"a", "a"}}, []string{"dependency_duplicate:a->a", "dependency_self:a->a", "dependency_self:a->a"}},
		{"duplicate pair", map[string][]string{"a": {"b", "b"}, "b": {}}, []string{"dependency_duplicate:a->b"}},
		{"missing edges and independent cycle", map[string][]string{"a": {"missing", "missing"}, "b": {"c"}, "c": {"b"}}, []string{"dependency_cycle:b->c,c->b", "dependency_duplicate:a->missing", "extension_dependency_not_claimed:a->missing", "extension_dependency_not_claimed:a->missing"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := dependencyCoordinator(tc.edges)
			if _, exists := tc.edges["a"]; exists {
				c.profiles["missing"] = profileRecord{descriptor: Descriptor{ProfileID: "missing", Claimable: true, ContractMajor: 1}, bindingSHA256: testDigest("a")}
			}
			var previous []Finding
			for attempt := 0; attempt < 4; attempt++ {
				resolution, err := c.ResolveClaims(c.orderedProfileIDs)
				var failure *ValidationError
				if !errors.As(err, &failure) || len(resolution.AdmissionOrder()) != 0 {
					t.Fatalf("invalid graph yielded resolution: %v / %v", resolution.AdmissionOrder(), err)
				}
				findings := failure.Findings()
				got := []string{}
				for _, f := range findings {
					if f.CollisionClass != "" {
						wantCount := 0
						if f.CollisionClass == "dependency_duplicate" {
							wantCount = 2
						}
						if len(f.ConflictingProfileIDs) == 0 || !sort.StringsAreSorted(f.ConflictingProfileIDs) || f.DuplicateCount != wantCount {
							t.Fatalf("incomplete collision: %#v", f)
						}
						if f.CollisionClass == "dependency_cycle" && (len(f.ConflictingProfileIDs) != 2 || slices.Contains(f.ConflictingProfileIDs, "e")) {
							t.Fatalf("wrong SCC members: %#v", f)
						}
					}
					if f.Code == "extension_dependency_not_claimed" {
						got = append(got, f.Code+":"+f.ProfileID+"->"+f.DependencyProfileID)
					} else {
						got = append(got, f.CollisionClass+":"+strings.Join(f.ConflictingTokens, ","))
					}
				}
				sort.Strings(got)
				if !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("findings %v, want %v", got, tc.want)
				}
				if attempt > 0 && !reflect.DeepEqual(previous, findings) {
					t.Fatal("nondeterministic findings")
				}
				previous = findings
			}
		})
	}
	c := dependencyCoordinator(map[string][]string{"a": {"z"}, "m": {}, "z": {}})
	r, err := c.ResolveClaims([]string{"z", "a", "m"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.AdmissionOrder(), []string{"m", "z", "a"}) || !reflect.DeepEqual(r.Claims().ProfileIDs(), []string{"a", "m", "z"}) {
		t.Fatal("identity and execution order conflated")
	}
	canonical, _ := NewResolvedClaimSet([]string{"a", "m", "z"})
	if r.Claims().SHA256() != canonical.SHA256() {
		t.Fatal("dependency ordering changed claim identity")
	}
	t.Run("maximal bounded SCC keeps every edge", func(t *testing.T) {
		edges := map[string][]string{}
		for i := 0; i < 256; i++ {
			id := fmt.Sprintf("p%03d", i)
			for offset := 1; offset <= 16; offset++ {
				edges[id] = append(edges[id], fmt.Sprintf("p%03d", (i+offset)%256))
			}
		}
		c := dependencyCoordinator(edges)
		registered := requireGeneratedCoordinator(t)
		c.validateFinding = registered.validateFinding
		c.validationConditions = registered.validationConditions
		_, err := c.ResolveClaims(c.orderedProfileIDs)
		var failure *ValidationError
		if !errors.As(err, &failure) {
			t.Fatal("dense SCC admitted")
		}
		findings := failure.Findings()
		if len(findings) != 1 || findings[0].CollisionClass != "dependency_cycle" || len(findings[0].ConflictingProfileIDs) != 256 || len(findings[0].ConflictingTokens) != 4096 {
			t.Fatal("SCC finding omitted members or edges")
		}
		var outward *StartupValidationError
		if !errors.As(err, &outward) || len(outward.Findings) != 1 || outward.Findings[0].ReasonCode != "extension_registry_conflict" || !c.validStartupFinding(outward.Findings[0]) {
			t.Fatalf("cycle projection = %v", err)
		}
		tokens, ok := outward.Findings[0].Details["conflicting_tokens"].([]string)
		if !ok || len(tokens) != 4096 {
			t.Fatal("outward cycle projection lost edges")
		}
		ordinary := outward.Findings[0]
		ordinary.Details = cloneObject(ordinary.Details)
		ordinary.Details["collision_class"] = "dependency_self"
		if c.validStartupFinding(ordinary) {
			t.Fatal("ordinary collision admitted cycle-only capacity")
		}
		oversized := outward.Findings[0]
		oversized.Details = cloneObject(oversized.Details)
		oversized.Details["conflicting_tokens"] = append(tokens, "z->a")
		if c.validStartupFinding(oversized) {
			t.Fatal("cycle exceeded graph capacity")
		}

	})
	order := r.AdmissionOrder()
	order[0] = "changed"
	if r.AdmissionOrder()[0] != "m" {
		t.Fatal("resolution leaked mutable order")
	}
	for _, change := range []string{"major", "binding", "claimable", "malformed"} {
		t.Run(change, func(t *testing.T) {
			c := dependencyCoordinator(map[string][]string{"a": {"z"}, "z": {}})
			want := "extension_dependency_incompatible"
			z := c.profiles["z"]
			switch change {
			case "major":
				z.descriptor.ContractMajor = 2
			case "binding":
				z.bindingSHA256 = ""
				want = "extension_implementation_unavailable"
			case "claimable":
				z.descriptor.Claimable = false
				want = "extension_profile_not_claimable"
			case "malformed":
				a := c.profiles["a"]
				a.descriptor.Dependencies[0].RequiredContractMajor = 0
				c.profiles["a"] = a
				want = "extension_registry_invalid"
			}
			c.profiles["z"] = z
			_, err := c.ResolveClaims([]string{"a", "z"})
			var failure *ValidationError
			if !errors.As(err, &failure) || failure.Findings()[0].Code != want {
				t.Fatalf("target admission: %v", err)
			}
		})
	}
	for _, kind := range []string{"self", "duplicate", "cycle"} {
		t.Run("packaged "+kind, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				target, major := "import", 1
				if kind != "self" {
					target, major = "network_flow_activity", 8
				}
				if _, hasFacts := object["facts"]; hasFacts && object["schema_id"] == "cartulary.extension_owner_fragment.v3" && object["owner_id"] == "core00" {
					facts := object["facts"].([]any)
					fact := map[string]any{"fact_kind": "runtime_dependency", "profile_id": "import", "dependency_profile_id": target, "required_contract_major": major}
					facts = append(facts, fact)
					if kind == "duplicate" {
						facts = append(facts, fact)
					}
					object["facts"] = facts
					normalized, err := registryfacts.NormalizeFragments([]map[string]any{object})
					if err != nil {
						t.Fatal(err)
					}
					object["facts"] = normalized[0]["facts"]
				}
				if object["schema_id"] != "cartulary.extension_profile_descriptor.v3" || object["profile_id"] != "import" {
					return
				}
				deps := []any{map[string]any{"profile_id": target, "required_contract_major": major}}
				if kind == "duplicate" {
					deps = append(deps, deps[0])
				}
				object["runtime_dependencies"] = deps
			})
			_, err := NewCoordinator(source)
			requireCollision(t, err, "dependency_"+kind)
			var outward *StartupValidationError
			if !errors.As(err, &outward) || len(outward.Findings) == 0 {
				t.Fatal("packaged collision lacks registered diagnostics")
			}
			for _, finding := range outward.Findings {
				if finding.ReasonCode != "extension_registry_conflict" {
					t.Fatalf("incorrect packaged classification: %#v", finding)
				}
			}
		})
	}
}

func verifyDependencyExecution(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	c := dependencyCoordinator(map[string][]string{"a": {"z"}, "z": {}})
	r, err := c.ResolveClaims([]string{"a", "z"})
	if err != nil {
		t.Fatal(err)
	}
	views := map[string]ProfileConfigurationView{}
	algorithms := map[string]AdmissionAlgorithm{}
	var calls []string
	fail := false
	for _, id := range c.orderedProfileIDs {
		view, err := c.ConfigurationView(id, nil)
		if err != nil {
			t.Fatal(err)
		}
		views[id] = view
		algorithms[id+".admit"] = func(_ context.Context, input AdmissionValidationContext) (AdmissionValidationResult, error) {
			calls = append(calls, input.ProfileID)
			if fail {
				return AdmissionValidationResult{}, errors.New("test failure")
			}
			return ValidAdmissionResult(input, input.ProfileID+".admit"), nil
		}
	}
	stateFailureProfile := ""
	runtime := newStateRuntimeForTest(t, fixture.store, StateRuntimeOptions{FinalValidators: map[string]FinalStateValidator{
		"test_profile.final_state_v1": func(_ context.Context, input FinalStateValidationContext, _ StateReadCapability) (StateValidationResult, error) {
			calls = append(calls, input.ProfileID)
			if input.ProfileID == stateFailureProfile {
				return StateValidationResult{}, errors.New("state failure")
			}
			return ValidFinalStateValidationResult(), nil
		},
	}})
	for _, phase := range []string{"preflight", "post_migration"} {
		t.Run(phase, func(t *testing.T) {
			calls = nil
			if err := runtime.ValidateClaimAdmission(context.Background(), c, r, phase, views, algorithms); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(calls, []string{"z", "a"}) {
				t.Fatalf("phase order: %v", calls)
			}
			calls, fail = nil, true
			if err := runtime.ValidateClaimAdmission(context.Background(), c, r, phase, views, algorithms); err == nil {
				t.Fatal("admitted failing phase")
			}
			want := []string{"z", "a"}
			if phase == "post_migration" {
				want = []string{"z"}
			}
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("failure order %v, want %v", calls, want)
			}
			fail = false
		})
	}
	for _, id := range c.orderedProfileIDs {
		plan := cloneStatePlan(fixture.plan)
		plan.ProfileID = id
		c.statePlans[id] = plan
	}
	for _, failing := range []string{"z", "a"} {
		calls, stateFailureProfile = nil, failing
		if err := runtime.AdmitClaims(context.Background(), c, r); err == nil {
			t.Fatal("state failure admitted")
		}
		want := []string{"z"}
		if failing == "a" {
			want = append(want, "a")
		}
		if !reflect.DeepEqual(calls, want) {
			t.Fatalf("state failure execution: %v", calls)
		}
		var initialized []string
		rows, err := fixture.pool.Query(context.Background(), `SELECT profile_id FROM extension_state_metadata ORDER BY profile_id`)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				t.Fatal(err)
			}
			initialized = append(initialized, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if failing == "z" && len(initialized) != 0 || failing == "a" && !reflect.DeepEqual(initialized, []string{"z"}) {
			t.Fatalf("incorrect cross-profile commit/rollback: %v", initialized)
		}
	}
	stateFailureProfile = ""
	calls = nil
	if err := runtime.AdmitClaims(context.Background(), c, r); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, []string{"z", "a"}) {
		t.Fatalf("state order %v", calls)
	}
	for _, id := range c.orderedProfileIDs {
		if readStateMetadata(t, fixture.pool, id).StateVersion != 1 {
			t.Fatal("state not initialized")
		}
	}
	for _, kind := range []string{"zero", "foreign", "wrong order", "wrong membership", "wrong digest"} {
		t.Run("resolution "+kind, func(t *testing.T) {
			invalid := r
			switch kind {
			case "zero":
				invalid = ClaimResolution{}
			case "foreign":
				invalid.registrySHA256 = testDigest("f")
			case "wrong order":
				invalid.admissionOrder = []string{"a", "z"}
			case "wrong membership":
				invalid.claims, _ = NewResolvedClaimSet([]string{"z"})
			case "wrong digest":
				invalid.claims.sha256 = testDigest("e")
			}
			calls = nil
			if err := runtime.ValidateClaimAdmission(context.Background(), c, invalid, "preflight", views, algorithms); err == nil {
				t.Fatal("invalid resolution reached admission")
			}
			if err := runtime.AdmitClaims(context.Background(), c, invalid); err == nil {
				t.Fatal("invalid resolution reached state")
			}
			if len(calls) != 0 {
				t.Fatal("invalid resolution invoked owner")
			}
		})
	}
}
