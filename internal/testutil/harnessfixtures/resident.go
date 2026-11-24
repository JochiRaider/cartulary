package harnessfixtures

// Resident is one physical allocation a unit holds by stable key: a
// PostgreSQL generation, the shared object store or a browser stack.
type Resident struct {
	Key    string
	Scope  string
	Claims map[string]int64
}

// ResidentAttached marks an allocation held by a persistent service session
// the run borrows; admission does not charge it.
const ResidentAttached = "attached"
