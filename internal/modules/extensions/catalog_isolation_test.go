package extensions

import (
	"bytes"
	"encoding/json"
	"testing"
)

// Snapshot the expected value as bytes before mutation, independently of every
// returned slice/map/pointer. Comparing two aliased Go values could pass falsely.
func assertReturnedIsolation[T any](t *testing.T, read func() T, mutate func(T)) {
	t.Helper()
	encode := func(value T) []byte {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	previous := read()
	want := encode(previous)
	changed := read()
	mutate(changed)
	if bytes.Equal(encode(changed), want) {
		t.Fatal("fixture mutation did not change the returned value")
	}
	if !bytes.Equal(encode(previous), want) {
		t.Fatal("mutation changed a previously returned independent value")
	}
	if !bytes.Equal(encode(read()), want) {
		t.Fatal("mutation changed the owner's subsequent accessor result")
	}
}

func TestCoordinatorCatalogCopyIsolation_Unit(t *testing.T) {
	// Populated local catalogs exercise supported nested fields even when the
	// packaged release has no executable migration or historical codec for them.
	state := StatePlan{
		ProfileID: "fixture", PhysicalStateBindingSHA256: "state-binding",
		DatabaseFamilyIDs: []string{"database"}, ObjectReferenceFamilyIDs: []string{"objects"},
		MigrationDefinitions:       []MigrationDefinition{{MigrationID: "forward", DefinitionSHA256: "forward-digest"}},
		MigrationLedgerDefinitions: []MigrationLedgerDefinition{{MigrationID: "ledger", DefinitionSHA256: "ledger-digest"}},
	}
	backup := BackupPlan{
		ProfileID: "fixture", PhysicalStateBindingSHA256: "backup-binding",
		Bindings: []BackupBinding{{BindingID: "binding"}},
		Codecs:   map[string]BackupCodec{"codec": {CodecID: "codec", SHA256: "codec-digest", HistoricalCodecs: []BackupCodecIdentity{{CodecID: "historical", SHA256: "historical-digest"}}}},
	}
	participant := ParticipantContract{
		ParticipantID: "participant", ContractSHA256: "participant-digest",
		AlgorithmIDs: []string{"algorithm"}, SerializationKeyKinds: []string{"key"}, OwnedStateFamilyIDs: []string{"family"},
		Operations: []ParticipantOperation{{OperationKind: "emit", StateFamilyIDs: []string{"operation-family"}}},
	}
	c := &Coordinator{
		orderedProfileIDs: []string{"fixture"}, registrySHA256: "registry-digest",
		statePlans: map[string]StatePlan{"fixture": state}, backupPlans: map[string]BackupPlan{"fixture": backup},
		participantContracts: map[string]ParticipantContract{"participant": participant},
	}
	registry := c.RegistrySHA256()
	for name, mutate := range map[string]func(StatePlan){
		"database_families": func(p StatePlan) { p.DatabaseFamilyIDs[0] = "changed" },
		"object_families":   func(p StatePlan) { p.ObjectReferenceFamilyIDs[0] = "changed" },
		"migrations":        func(p StatePlan) { p.MigrationDefinitions[0].DefinitionSHA256 = "changed" },
		"ledger":            func(p StatePlan) { p.MigrationLedgerDefinitions[0].DefinitionSHA256 = "changed" },
	} {
		t.Run(name, func(t *testing.T) {
			assertReturnedIsolation(t, func() StatePlan {
				p, ok := c.StatePlan("fixture")
				if !ok {
					t.Fatal("missing populated state fixture")
				}
				return p
			}, mutate)
		})
	}
	for name, mutate := range map[string]func([]BackupPlan){
		"bindings":      func(p []BackupPlan) { p[0].Bindings[0].BindingID = "changed" },
		"codec_replace": func(p []BackupPlan) { p[0].Codecs["codec"] = BackupCodec{CodecID: "changed"} },
		"codec_delete":  func(p []BackupPlan) { delete(p[0].Codecs, "codec") },
		"codec_insert":  func(p []BackupPlan) { p[0].Codecs["new"] = BackupCodec{CodecID: "new"} },
		"codec_history": func(p []BackupPlan) { p[0].Codecs["codec"].HistoricalCodecs[0].SHA256 = "changed" },
	} {
		t.Run(name, func(t *testing.T) { assertReturnedIsolation(t, c.BackupPlans, mutate) })
	}
	for name, mutate := range map[string]func([]ParticipantContract){
		"algorithms":         func(p []ParticipantContract) { p[0].AlgorithmIDs[0] = "changed" },
		"serialization_keys": func(p []ParticipantContract) { p[0].SerializationKeyKinds[0] = "changed" },
		"owned_families":     func(p []ParticipantContract) { p[0].OwnedStateFamilyIDs[0] = "changed" },
		"operations":         func(p []ParticipantContract) { p[0].Operations[0].OperationKind = "changed" },
		"operation_families": func(p []ParticipantContract) { p[0].Operations[0].StateFamilyIDs[0] = "changed" },
	} {
		t.Run(name, func(t *testing.T) { assertReturnedIsolation(t, c.ParticipantContracts, mutate) })
	}
	if c.RegistrySHA256() != registry {
		t.Fatal("catalog identity changed")
	}
}

func TestPublicationPlanCopyIsolation_Unit(t *testing.T) {
	for name, claims := range map[string][]string{
		"empty": {},
		"mixed": {"import", "network_flow_activity"},
		"full":  {"enterprise_authentication", "import", "incident_portability", "network_flow_activity", "reference_pack", "snapshot_reporting"},
	} {
		t.Run(name, func(t *testing.T) {
			c := requireGeneratedCoordinator(t)
			resolution, err := c.ResolveClaims(claims)
			if err != nil {
				t.Fatal(err)
			}
			plan, err := c.BuildPublicationPlan(resolution)
			if err != nil {
				t.Fatal(err)
			}
			identity := plan.Summary()
			t.Run("route_rows", func(t *testing.T) {
				assertReturnedIsolation(t, plan.Routes, func(v []RouteDispatch) { v[0].RouteFamily = "changed" })
			})
			for i, route := range plan.Routes() {
				if route.ContributionID == nil {
					if route.DispatchState == "claimed" {
						t.Fatal("claimed route has nil contribution")
					}
					continue
				}
				t.Run("route_pointer_"+route.ProfileID, func(t *testing.T) {
					assertReturnedIsolation(t, plan.Routes, func(v []RouteDispatch) { *v[i].ContributionID = "changed" })
				})
			}
			for i, profile := range plan.Discovery() {
				t.Run("discovery_"+profile.ProfileID, func(t *testing.T) {
					if profile.ContractMajor != nil {
						assertReturnedIsolation(t, plan.Discovery, func(v []DiscoveryProfile) { *v[i].ContractMajor = -1 })
					}
					if len(profile.RouteFamilies) > 0 {
						assertReturnedIsolation(t, plan.Discovery, func(v []DiscoveryProfile) { v[i].RouteFamilies[0] = "changed" })
					}
					if len(profile.WorkspaceKeys) > 0 {
						assertReturnedIsolation(t, plan.Discovery, func(v []DiscoveryProfile) { v[i].WorkspaceKeys[0] = "changed" })
					}
					if len(profile.Workspaces) > 0 {
						assertReturnedIsolation(t, plan.Discovery, func(v []DiscoveryProfile) { v[i].Workspaces[0].MinimumRole = "changed" })
					}
				})
			}
			if len(claims) == 0 {
				if len(plan.Workers()) != 0 || len(plan.JobKindContracts()) != 0 || len(plan.Workspaces()) != 0 {
					t.Fatal("empty claims published executable components")
				}
			} else {
				if len(plan.Workers()) == 0 || len(plan.JobKindContracts()) == 0 {
					t.Fatal("populated claim fixture has no workers or jobs")
				}
				t.Run("workers", func(t *testing.T) {
					assertReturnedIsolation(t, plan.Workers, func(v []WorkerPublication) { v[0].JobKinds[0] = "changed" })
				})
				t.Run("job_rows", func(t *testing.T) {
					assertReturnedIsolation(t, plan.JobKindContracts, func(v []JobKindContract) { v[0].JobKind = "changed" })
				})
				refs := 0
				for i, job := range plan.JobKindContracts() {
					if len(job.ResourceRefContracts) == 0 {
						continue
					}
					refs++
					t.Run("job_refs_"+job.JobKind, func(t *testing.T) {
						assertReturnedIsolation(t, plan.JobKindContracts, func(v []JobKindContract) { v[i].ResourceRefContracts[0].MaxRefs = -1 })
					})
				}
				if refs == 0 {
					t.Fatal("populated jobs have no resource-reference fixture")
				}
			}
			if plan.Summary() != identity {
				t.Fatal("publication identity changed")
			}
			rebuilt, err := c.BuildPublicationPlan(resolution)
			if err != nil || rebuilt.Summary() != identity {
				t.Fatalf("coordinator publication identity changed: %v", err)
			}
		})
	}
	// Capabilities are currently empty in the packaged profiles; retain a
	// populated accessor fixture and a legitimate absent contract major.
	t.Run("discovery_optional_fields", func(t *testing.T) {
		plan := PublicationPlan{discovery: []DiscoveryProfile{{ProfileID: "fixture", Capabilities: []string{"capability"}}}}
		assertReturnedIsolation(t, plan.Discovery, func(v []DiscoveryProfile) { v[0].Capabilities[0] = "changed" })
		if plan.Discovery()[0].ContractMajor != nil {
			t.Fatal("absent contract major became present")
		}
	})
}
