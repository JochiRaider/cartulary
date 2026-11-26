// Package browserbinding is the single Go writer of a browser binding
// (TH-HARNESS Section 5.5, single writer): the stage, runtime and resource
// profiles, service requirement, session, preparation and stack a browser
// process is bound to, and the performance-fixture snapshot clone it uses. The planner decides a cohort's binding, and the browser
// fixture passes it, with the session it allocates, to the stack it starts.
// Its Node form is tools/harness/browser/browser-binding.mjs. It has no
// dependencies, so the planner can use it.
package browserbinding

import "maps"

// Binding is a browser binding. Environment writes only the fields that are
// set; every reader treats an empty binding value as unset.
type Binding struct {
	Stage                string
	RuntimeProfile       string
	ResourceProfile      string
	ServiceRequirement   string
	SessionContract      string
	SessionGroup         string
	StackKey             string
	PreparationLabel     string
	PreparationTimeoutMS string
	ResetDrainTimeoutMS  string
	CohortLabel          string
	AcquisitionFile      string
	SessionLeaseFile     string
	PrivateSessionRoot   string
	// A performance-fixture cohort's snapshot: the planner decides the profile,
	// snapshot, builder, row and predicate; the fixture adds the clone lease and
	// ordinal it allocates.
	FixtureProfile      string
	SnapshotKey         string
	SnapshotBuilderUnit string
	FixtureRow          string
	FixturePredicate    string
	CloneLease          string
	CloneOrdinal        string
}

// Environment returns b as environment entries.
func (b Binding) Environment() map[string]string {
	environment := map[string]string{}
	for name, value := range map[string]string{
		"CARTULARY_BROWSER_STAGE":                    b.Stage,
		"CARTULARY_BROWSER_RUNTIME_PROFILE_ID":       b.RuntimeProfile,
		"CARTULARY_BROWSER_RESOURCE_PROFILE_ID":      b.ResourceProfile,
		"CARTULARY_BROWSER_SERVICE_REQUIREMENT":      b.ServiceRequirement,
		"CARTULARY_BROWSER_SESSION_CONTRACT":         b.SessionContract,
		"CARTULARY_BROWSER_SESSION_GROUP":            b.SessionGroup,
		"CARTULARY_BROWSER_STACK_KEY":                b.StackKey,
		"CARTULARY_BROWSER_PREPARATION_LABEL":        b.PreparationLabel,
		"CARTULARY_BROWSER_PREPARATION_TIMEOUT_MS":   b.PreparationTimeoutMS,
		"CARTULARY_BROWSER_RESET_DRAIN_TIMEOUT_MS":   b.ResetDrainTimeoutMS,
		"CARTULARY_BROWSER_COHORT_LABEL":             b.CohortLabel,
		"CARTULARY_BROWSER_ACQUISITION_FILE":         b.AcquisitionFile,
		"CARTULARY_WEB_E2E_SESSION_LEASE_FILE":       b.SessionLeaseFile,
		"CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT":     b.PrivateSessionRoot,
		"CARTULARY_FIXTURE_PROFILE_ID":               b.FixtureProfile,
		"CARTULARY_FIXTURE_SNAPSHOT_KEY":             b.SnapshotKey,
		"CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID": b.SnapshotBuilderUnit,
		"CARTULARY_FIXTURE_ROW_ID":                   b.FixtureRow,
		"CARTULARY_FIXTURE_PREDICATE_ID":             b.FixturePredicate,
		"CARTULARY_FIXTURE_CLONE_LEASE_ID":           b.CloneLease,
		"CARTULARY_FIXTURE_CLONE_ORDINAL":            b.CloneOrdinal,
	} {
		if value != "" {
			environment[name] = value
		}
	}
	return environment
}

// Apply writes b into environment and returns it.
func (b Binding) Apply(environment map[string]string) map[string]string {
	maps.Copy(environment, b.Environment())
	return environment
}
