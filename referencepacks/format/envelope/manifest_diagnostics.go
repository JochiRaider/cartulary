package envelope

import (
	"errors"
	"slices"
	"strconv"
	"strings"
)

// Application profile registration, capability support, required registries and
// active-set dependency policy are deliberately outside this library.
func manifestCompatibility(m Manifest, c *semanticChecks) {
	c.require(m.Contract == ContractID, ".pack_contract_version")
	if m.ProfileID == "cartulary.reference_pack.dataset.v1" {
		c.require(m.ProfileVersion == "1", ".content_profile_version")
		c.require(m.Summary.Kind == "objects_relationships", ".content_summary.kind")
		c.require(m.Kind == "framework" || m.Kind == "enrichment", ".pack_kind")
	}
}

func ManifestCompatibilityFindings(m Manifest, emit FindingSink) error {
	if emit == nil {
		return errors.New("reference pack: missing compatibility findings sink")
	}
	c := &semanticChecks{path: "$", emit: emit}
	manifestCompatibility(m, c)
	return c.err
}

func manifestFieldFindings(m Manifest, operator bool, emit FindingSink) error {
	report := func(path string) error { return emit(Finding{Path: path}) }
	if operator != (m.Repository != nil) {
		if err := report("$.trust_repository_id"); err != nil {
			return err
		}
	}
	if operator && m.SourceAsOf == nil {
		if err := report("$.source_as_of"); err != nil {
			return err
		}
	}
	built, validBuilt := packTime(m.BuiltAt)
	if !validBuilt {
		if err := report("$.built_at"); err != nil {
			return err
		}
	}
	if m.SourceAsOf != nil {
		asOf, ok := packTime(*m.SourceAsOf)
		if !ok || validBuilt && asOf.After(built) {
			if err := report("$.source_as_of"); err != nil {
				return err
			}
		}
	}
	for _, field := range []struct {
		path, value       string
		characters, bytes int
	}{
		{"$.source_identifier", m.SourceIdentifier, 512, 0}, {"$.source_version", m.SourceVersion, 256, 0},
		{"$.builder.builder_id", m.Builder.ID, 192, 192}, {"$.builder.builder_version", m.Builder.Version, 128, 128},
	} {
		if !singleLine(field.value, field.characters) || field.bytes != 0 && len(field.value) > field.bytes {
			if err := report(field.path); err != nil {
				return err
			}
		}
	}
	previous := ""
	for i, artifact := range m.Artifacts {
		path := "$.source_artifacts[" + strconv.Itoa(i) + "]"
		if !singleLine(artifact.Ref, 1024) {
			if err := report(path + ".source_ref"); err != nil {
				return err
			}
		}
		if !singleLine(artifact.Version, 256) {
			if err := report(path + ".source_version"); err != nil {
				return err
			}
		}
		identity := artifact.Ref + "\x00" + artifact.Version + "\x00" + artifact.SHA256
		if identity <= previous {
			if err := report(path); err != nil {
				return err
			}
		}
		previous = identity
	}
	previous = ""
	for i, dependency := range m.Dependencies {
		identity := dependency.Key + "\x00" + dependency.Version + "\x00" + dependency.SHA256
		if identity <= previous {
			if err := report("$.dependencies[" + strconv.Itoa(i) + "]"); err != nil {
				return err
			}
		}
		previous = identity
	}
	previous = ""
	for i, conflict := range m.Conflicts {
		identity := conflict.Key + "\x00"
		if conflict.Version != nil {
			identity += *conflict.Version
		}
		if identity <= previous {
			if err := report("$.conflicts[" + strconv.Itoa(i) + "]"); err != nil {
				return err
			}
		}
		previous = identity
	}
	previous = ""
	payloads, notices := []string{}, []string{}
	for i, file := range m.Files {
		path := "$.files[" + strconv.Itoa(i) + "]"
		if ValidatePath(file.Path, false) != nil || file.Path <= previous {
			if err := report(path + ".path"); err != nil {
				return err
			}
		}
		previous = file.Path
		if file.Role == "payload" {
			expectedMedia := "application/x-ndjson"
			if m.ProfileID == "cartulary.reference_pack.dataset.v1" && file.Path == "payload/descriptor.json" {
				expectedMedia = "application/json"
			}
			if file.MediaType != expectedMedia {
				if err := report(path + ".media_type"); err != nil {
					return err
				}
			}
			payloads = append(payloads, file.Path)
		} else {
			if file.MediaType != "text/plain" {
				if err := report(path + ".media_type"); err != nil {
					return err
				}
			}
			if !strings.HasPrefix(file.Path, "notices/") {
				if err := report(path + ".path"); err != nil {
					return err
				}
			}
			if file.Size == 0 {
				if err := report(path + ".size_bytes"); err != nil {
					return err
				}
			}
			notices = append(notices, file.Path)
		}
	}
	expected := []string{"payload/entries.ndjson"}
	if m.Summary.Kind == "objects_relationships" {
		expected = []string{"payload/objects.ndjson", "payload/relationships.ndjson"}
	}
	if m.ProfileID == "cartulary.reference_pack.dataset.v1" {
		expected = []string{"payload/descriptor.json", "payload/objects.ndjson", "payload/relationships.ndjson"}
	}
	if !slices.Equal(payloads, expected) {
		if err := report("$.files"); err != nil {
			return err
		}
	}
	if !slices.Equal(notices, m.License.Notices) {
		if err := report("$.license.notice_paths"); err != nil {
			return err
		}
	}
	return nil
}
