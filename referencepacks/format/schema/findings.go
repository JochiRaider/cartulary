package schema

import "strconv"

// Finding is independent of application attestations and retained summaries.
// Callers may enumerate a check into their own bounded, private diagnostic sink.
type Finding struct {
	Path    string
	EntryID *string
	Details SafeDetails
}
type SafeDetails struct {
	LimitID            *string `json:"limit_id"`
	ExpectedToken      *string `json:"expected_token"`
	ActualToken        *string `json:"actual_token"`
	RelatedPackKey     *string `json:"related_pack_key"`
	RelatedPackVersion *string `json:"related_pack_version"`
}
type FindingSink func(Finding) error

func LimitFinding(path, id string, maximum, observed int64) Finding {
	return UnsignedLimitFinding(path, id, uint64(maximum), uint64(observed))
}
func UnsignedLimitFinding(path, id string, maximum, observed uint64) Finding {
	expected, actual := strconv.FormatUint(maximum, 10), strconv.FormatUint(observed, 10)
	return Finding{Path: path, Details: SafeDetails{LimitID: &id, ExpectedToken: &expected, ActualToken: &actual}}
}
