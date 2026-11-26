package suiteservices

// Attachment is a process's relation to a test-services suite: the suite it
// belongs to and how it calls and borrows it.
type Attachment struct {
	SuiteID            string
	CallMode           string
	LifecycleMode      string
	PersistentBorrower bool
}

// Environment is the only Go writer of the suite attachment names (TH-HARNESS
// Section 5.5, single writer); tools/harness/services/suite-attachment.mjs is
// its Node form. Unset modes are omitted rather than written empty.
func (a Attachment) Environment() map[string]string {
	environment := map[string]string{SuiteIDEnv: a.SuiteID}
	if a.CallMode != "" {
		environment[CallModeEnv] = a.CallMode
	}
	if a.LifecycleMode != "" {
		environment[LifecycleModeEnv] = a.LifecycleMode
	}
	if a.PersistentBorrower {
		environment[PersistentBorrowerEnv] = "1"
	}
	return environment
}
