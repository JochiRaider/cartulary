package harnessruntime

import (
	"cmp"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessplan"
)

// adoptPlan binds a validated Go plan to the runtime's execution model. The
// plan itself is the published run-root document; units receive their
// launch environment with runtime binaries bound.
func adoptPlan(source *harnessplan.Plan) (*Plan, error) {
	var manifest map[string]any
	data, err := json.Marshal(source.Manifest)
	if err != nil || json.Unmarshal(data, &manifest) != nil {
		return nil, errPlan
	}
	facts := planWire{
		BrowserLaneLimit: source.BrowserLaneLimit,
		InputEnvironment: inputEnvironment{Clear: slices.Clone(source.InputEnvironment.Clear), Forward: source.InputEnvironment.Forward},
		CacheInputs:      source.CacheInputs,
		Manifest:         manifest,
		Projections:      map[string]projection{},
		Capacities:       source.Capacities,
		AgingMS:          source.AgingQuantumMS,
		Bindings:         source.Bindings,
		Graph:            workGraph{Schema: harnessplan.GraphSchema, Digest: source.Graph.Digest},
	}
	for _, input := range source.SourceInputs {
		facts.SourceInputs = append(facts.SourceInputs, sourceInput{Path: input.Path, Kind: input.Kind, Mode: input.Mode, Digest: input.ByteDigest})
	}
	for name, p := range source.Projections {
		facts.Projections[name] = projection{CommandID: p.CommandID, Units: slices.Clone(p.UnitIDs)}
	}
	for _, fact := range source.Cache {
		facts.Cache = append(facts.Cache, cacheFact{UnitID: fact.UnitID, Eligible: fact.Eligible, Reason: fact.Reason, Profile: fact.ProfileID, Policy: fact.Policy,
			InputDigest: fact.InputDigest, Strategy: fact.DependencyStrategy, Vulnerability: fact.VulnerabilityDatabaseRevision, DependencyDigest: fact.DependencyDigest})
	}
	for _, u := range source.Graph.Sorted() {
		adopted := unit{
			ID: u.ID, Owner: u.OwnerID, Kind: u.Kind, PostgresIsolation: u.PostgresIsolation, Renderer: cmp.Or(u.RendererRequirement, "none"), Affinity: u.AffinityKey,
			Command: command{Executable: u.Command.Executable, Args: slices.Clone(u.Command.Args), Environment: source.LaunchEnvironment(u)},
			Needs:   slices.Clone(u.Needs), Claims: u.Claims, SharedLocks: slices.Clone(u.SharedLocks), ExclusiveLocks: slices.Clone(u.ExclusiveLocks),
			Fixture: u.FixtureLease, FixtureProfile: u.FixtureProfileID, SnapshotKey: u.SnapshotKey,
			Services: slices.Clone(u.ServiceDependencies), CachePolicy: u.CachePolicy, TimeoutMS: u.TimeoutMS,
			Evidence: slices.Clone(u.EvidenceOutputs), EstimateMS: u.EstimatedWorkMS, Digest: u.SemanticDigest,
			FailurePolicy: failurePolicy{BlockDescendants: u.FailurePolicy.BlockDescendants, ContinueIndependent: true, AggregateEffect: "required"},
			Residents:     make([]harnessfixtures.Resident, 0, len(u.Residents)),
		}
		for _, binding := range u.RowBindings {
			adopted.Rows = append(adopted.Rows, rowBinding{ID: binding.RowID, Runner: binding.Runner})
		}
		for _, output := range u.ArtifactOutputs {
			adopted.Artifacts = append(adopted.Artifacts, artifact{Type: output.ArtifactType, Path: output.RelativePath, Destination: output.DestinationClass, Mode: output.Mode, Producer: output.ProducerIdentity})
		}
		for _, resident := range u.Residents {
			scope := ""
			if resident.Scope == harnessplan.ScopeAttached {
				scope = harnessfixtures.ResidentAttached
			}
			adopted.Residents = append(adopted.Residents, harnessfixtures.Resident{Key: resident.Key, Scope: scope, Claims: resident.Claims})
		}
		facts.Graph.Units = append(facts.Graph.Units, adopted)
	}
	return &Plan{facts: facts, source: source}, nil
}

// planRequest reads the runtime's graph arguments as a planner request.
func planRequest(args []string, environment map[string]string) (harnessplan.Request, error) {
	request := harnessplan.Request{Environment: environment}
	for i := 0; i < len(args); i++ {
		value := func() string {
			if i+1 < len(args) {
				i++
				return args[i]
			}
			return ""
		}
		switch args[i] {
		case "--selection":
			request.Kind = value()
		case "--target":
			request.Target = value()
		case "--owner":
			request.Owner = value()
		case "--rows":
			request.Rows = strings.Split(value(), ",")
		case "--service-backed-only":
			request.ServiceBackedOnly = true
		default:
			return request, errors.New("unknown graph argument")
		}
	}
	return request, nil
}

// plannerFailure classifies a planner rejection as a closed preparation
// diagnostic (TH-HARNESS-REQ-854). The planner's messages are built from
// authored identifiers, so the message is the single subject.
func plannerFailure(err error) error {
	diagnostic := PreparationDiagnostic{FailureReason: "configuration_error", Cause: "unclassified", Subjects: []string{}}
	switch {
	case errors.Is(err, harnessplan.ErrSelection):
		diagnostic.FailureReason, diagnostic.Cause = "usage_error", "command_selection"
	case errors.Is(err, harnessplan.ErrInput):
		diagnostic.Cause = "catalog_shape"
	}
	subject := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, err.Error())
	if len(subject) > 256 {
		runes := []rune(subject)
		for len(string(runes)) > 253 {
			runes = runes[:len(runes)-1]
		}
		subject = string(runes) + "..."
	}
	if strings.TrimSpace(subject) != "" {
		diagnostic.Subjects = append(diagnostic.Subjects, subject)
	}
	return preparationFailure(diagnostic)
}
