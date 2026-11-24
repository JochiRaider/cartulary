package harnessplan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// Graph is a compiled work graph: units keyed by ID, every need resolved
// inside the graph.
type Graph struct {
	Units map[string]*Unit
}

// Sorted returns the units in ID order.
func (g *Graph) Sorted() []*Unit {
	ids := make([]string, 0, len(g.Units))
	for id := range g.Units {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	units := make([]*Unit, len(ids))
	for i, id := range ids {
		units[i] = g.Units[id]
	}
	return units
}

// validate checks the graph's closure: every need names a unit in the graph,
// the needs are acyclic, and no row is bound to two units.
func (g *Graph) validate() error {
	rows := map[string]string{}
	for id, unit := range g.Units {
		for _, need := range unit.Needs {
			if _, ok := g.Units[need]; !ok {
				return fmt.Errorf("%w: %s needs missing unit %s", ErrSelection, id, need)
			}
		}
		for _, binding := range unit.RowBindings {
			if other, ok := rows[binding.RowID]; ok && other != id {
				return fmt.Errorf("%w: row %s is bound to %s and %s", ErrSelection, binding.RowID, other, id)
			}
			rows[binding.RowID] = id
		}
	}
	state := map[string]int{}
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("%w: dependency cycle through %s", ErrSelection, id)
		case 2:
			return nil
		}
		state[id] = 1
		for _, need := range g.Units[id].Needs {
			if err := visit(need); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for id := range g.Units {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

// Unit is one schedulable piece of work (graph v9 unit).
type Unit struct {
	ID                  string
	OwnerID             string
	Kind                string
	Command             Command
	Needs               []string
	Claims              map[string]int64
	Residents           []Resident
	SharedLocks         []string
	ExclusiveLocks      []string
	FixtureLease        string
	ServiceDependencies []string
	PostgresIsolation   string
	CachePolicy         string
	TimeoutMS           int64
	EvidenceOutputs     []string
	RowBindings         []RowBinding
	ArtifactOutputs     []ArtifactOutput
	RuntimeBinaryIDs    []string
	EstimatedWorkMS     int64
}

// Command is what the runtime launches for a unit. Runtime binary paths are
// not part of it: the unit lists the binaries it consumes and the runtime
// binds their paths at launch.
type Command struct {
	Executable  string
	Args        []string
	Environment map[string]string
}

// Resident is a physical allocation that outlives a single process: a
// PostgreSQL generation or the shared object store. Its key identifies the
// allocation wherever it appears; its scope fixes its lifetime.
type Resident struct {
	Key    string
	Scope  string
	Claims map[string]int64
}

// Resident scopes.
const (
	// ScopeUnit: created for one unit and retired with it.
	ScopeUnit = "unit"
	// ScopeGraph: shared by consumers in the graph and retired after the last.
	ScopeGraph = "graph"
)

// RowBinding binds a catalog row to the unit that produces its evidence.
type RowBinding struct {
	RowID  string
	Runner string
}

// digest is the SHA-256 of value's canonical JSON, as hex.
func digest(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		panic(err)
	}
	canonical, err := canonicaljson.Marshal(document)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:])
}
