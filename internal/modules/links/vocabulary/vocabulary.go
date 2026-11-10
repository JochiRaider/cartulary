package vocabulary

import "fmt"

type LinkType uint8

const (
	LinkTypeInvalid LinkType = iota
	LinkTypeObservedOnHost
	LinkTypeObservedAsIdentity
	LinkTypeReferencesIndicator
	LinkTypeAttachedEvidence
	LinkTypeReferencesArtifact
	LinkTypeDerivedFrom
	LinkTypeMergedInto
	LinkTypeSupportedBy
	LinkTypeReferencesRecord
	LinkTypeSupersedes
)

var linkTypeTokens = [...]string{
	LinkTypeInvalid:             "",
	LinkTypeObservedOnHost:      "observed_on_host",
	LinkTypeObservedAsIdentity:  "observed_as_identity",
	LinkTypeReferencesIndicator: "references_indicator",
	LinkTypeAttachedEvidence:    "attached_evidence",
	LinkTypeReferencesArtifact:  "references_artifact",
	LinkTypeDerivedFrom:         "derived_from",
	LinkTypeMergedInto:          "merged_into",
	LinkTypeSupportedBy:         "supported_by",
	LinkTypeReferencesRecord:    "references_record",
	LinkTypeSupersedes:          "supersedes",
}

func ParseLinkType(value string) (LinkType, error) {
	for candidate := LinkTypeObservedOnHost; candidate <= LinkTypeSupersedes; candidate++ {
		if linkTypeTokens[candidate] == value {
			return candidate, nil
		}
	}
	return LinkTypeInvalid, fmt.Errorf("links: invalid link type %q", value)
}

func (value LinkType) String() string {
	if value <= LinkTypeInvalid || int(value) >= len(linkTypeTokens) {
		return ""
	}
	return linkTypeTokens[value]
}

func (value LinkType) MarshalText() ([]byte, error) {
	return []byte(value.String()), nil
}

type LinkProvenance uint8

const (
	LinkProvenanceInvalid LinkProvenance = iota
	LinkProvenanceManual
	LinkProvenanceAutoMatch
	LinkProvenanceImport
	LinkProvenanceRollback
	LinkProvenanceSystem
)

var linkProvenanceTokens = [...]string{
	LinkProvenanceInvalid:   "",
	LinkProvenanceManual:    "manual",
	LinkProvenanceAutoMatch: "auto_match",
	LinkProvenanceImport:    "import",
	LinkProvenanceRollback:  "rollback",
	LinkProvenanceSystem:    "system",
}

func ParseLinkProvenance(value string) (LinkProvenance, error) {
	for candidate := LinkProvenanceManual; candidate <= LinkProvenanceSystem; candidate++ {
		if linkProvenanceTokens[candidate] == value {
			return candidate, nil
		}
	}
	return LinkProvenanceInvalid, fmt.Errorf("links: invalid link provenance %q", value)
}

func (value LinkProvenance) String() string {
	if value <= LinkProvenanceInvalid || int(value) >= len(linkProvenanceTokens) {
		return ""
	}
	return linkProvenanceTokens[value]
}

func (value LinkProvenance) MarshalText() ([]byte, error) {
	return []byte(value.String()), nil
}
