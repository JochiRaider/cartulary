package packformat

import (
	"errors"
	"slices"
)

func manifestCompatibility(m Manifest, runtime bool, c *semanticChecks) {
	if runtime {
		c.require(slices.Equal(m.Compatibility.Majors, []int{1}), ".compatibility.cartulary_contract_majors")
		c.require(len(m.Compatibility.Capabilities) == 0, ".compatibility.required_capabilities")
		return
	}
	p, ok := ProfileForKey(m.Key)
	c.require(ok, ".pack_key")
	c.require(m.Contract == ContractID, ".pack_contract_version")
	if ok {
		c.require(m.ProfileID == p.ID, ".content_profile_id")
		c.require(m.ProfileVersion == p.Version, ".content_profile_version")
		if _, standard := profiles[m.Key]; standard {
			c.require(m.Kind == p.Kind, ".pack_kind")
		} else {
			c.require(m.Kind == "framework" || m.Kind == "enrichment", ".pack_kind")
		}
		shape := p.Shape
		if shape == "dataset" {
			shape = "objects_relationships"
		}
		c.require(m.Summary.Kind == shape, ".content_summary.kind")
	}
}

func ManifestCompatibilityFindings(m Manifest, runtime bool, emit FindingSink) error {
	if emit == nil {
		return errors.New("reference pack: missing compatibility findings sink")
	}
	c := &semanticChecks{path: "$", emit: emit}
	manifestCompatibility(m, runtime, c)
	return c.err
}
