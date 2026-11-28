package suiteservices

import (
	"fmt"
	"strconv"
)

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

// PostgresAllocation is the PostgreSQL allocation a suite process works in:
// its isolation, allocation reference and the memory-backed data capacity its
// resident claim affords (zero when it names none). Environment always writes
// every name, so an empty allocation clears an inherited one.
type PostgresAllocation struct {
	Isolation string
	Ref       string
	DataBytes int64
}

// Environment is the only Go writer of the allocation names (TH-HARNESS
// Section 5.5, single writer); postgresAllocationEnvironment in
// tools/harness/services/suite-attachment.mjs is its Node form.
func (a PostgresAllocation) Environment() map[string]string {
	dataBytes := ""
	if a.DataBytes > 0 {
		dataBytes = strconv.FormatInt(a.DataBytes, 10)
	}
	return map[string]string{PGIsolationEnv: a.Isolation, PGAllocationRefEnv: a.Ref, PGDataBytesEnv: dataBytes}
}

// PostgresDataBytes is the allocation's memory-backed data capacity, or zero
// when the environment names none.
func PostgresDataBytes(env map[string]string) (int64, error) {
	raw := env[PGDataBytesEnv]
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive byte count", PGDataBytesEnv)
	}
	return value, nil
}
