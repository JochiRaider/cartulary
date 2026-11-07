package testsupport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	entrainputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/entra-app-ids"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// checkEntraBaseline compares public bytes with authored factual expectations.
// It shares no extraction/projection code with the source producer or JS witness.
func checkEntraBaseline(manifest packformat.Manifest, payload []byte) error {
	read := func(name string, value any) error {
		b, err := entrainputs.Files.ReadFile(name)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, value)
	}
	var baseline struct {
		Entries []struct {
			ID       string   `json:"app_id"`
			Label    string   `json:"display_label"`
			Aliases  []string `json:"aliases"`
			Locators []string `json:"source_locators"`
			Status   string   `json:"source_status"`
		} `json:"entries"`
	}
	var editorial struct {
		Historical string `json:"historical_note"`
		Summaries  []struct {
			ID   string `json:"app_id"`
			Text string `json:"text"`
		} `json:"summaries"`
	}
	var inventory struct {
		Sources []struct {
			ID     string `json:"source_id"`
			URL    string `json:"url"`
			Commit string `json:"commit"`
			SHA    string `json:"sha256"`
		} `json:"sources"`
	}
	for _, input := range []struct {
		name  string
		value any
	}{
		{"baseline.v2.json", &baseline}, {"editorial.v2.json", &editorial}, {"inventory.v2.json", &inventory},
	} {
		if err := read(input.name, input.value); err != nil {
			return err
		}
	}
	indices := map[string]int{}
	for _, source := range inventory.Sources {
		for i, artifact := range manifest.Artifacts {
			if artifact.Ref != source.URL {
				continue
			}
			if _, exists := indices[source.ID]; exists || artifact.Version != source.Commit || artifact.SHA256 != source.SHA {
				return fmt.Errorf("entra source identity mismatch: %s", source.ID)
			}
			indices[source.ID] = i
		}
		if _, exists := indices[source.ID]; !exists {
			return fmt.Errorf("entra source absent: %s", source.ID)
		}
	}
	summaries := map[string]string{}
	for _, summary := range editorial.Summaries {
		if _, exists := summaries[summary.ID]; exists {
			return fmt.Errorf("duplicate editorial identity: %s", summary.ID)
		}
		summaries[summary.ID] = summary.Text
	}
	if !bytes.HasSuffix(payload, []byte{'\n'}) || bytes.ContainsRune(payload, '\r') {
		return fmt.Errorf("entra corpus framing")
	}
	lines := bytes.Split(payload[:len(payload)-1], []byte{'\n'})
	if len(lines) != 139 || len(baseline.Entries) != 139 || len(summaries) != 16 {
		return fmt.Errorf("entra corpus cardinality")
	}
	historical, original, nulls, aliases, aliased := 0, 0, 0, 0, 0
	type sourceRef struct {
		Index   int    `json:"artifact_index"`
		Locator string `json:"locator"`
	}
	for i, want := range baseline.Entries {
		var description any
		switch want.Status {
		case "historical_documented":
			description = editorial.Historical
			historical++
		case "current_documented":
			if text, ok := summaries[want.ID]; ok {
				description = text
				original++
			} else {
				nulls++
			}
		default:
			return fmt.Errorf("unknown Entra baseline status: %s", want.Status)
		}
		refs := []sourceRef{}
		for _, locator := range want.Locators {
			source, line, ok := strings.Cut(locator, ":")
			index, exists := indices[source]
			if !ok || !exists {
				return fmt.Errorf("unknown Entra baseline locator: %s", locator)
			}
			refs = append(refs, sourceRef{index, line})
		}
		slices.SortFunc(refs, func(a, b sourceRef) int {
			if a.Index < b.Index {
				return -1
			}
			if a.Index > b.Index {
				return 1
			}
			return strings.Compare(a.Locator, b.Locator)
		})
		expected, err := canonicaljson.Marshal(map[string]any{
			"schema_id": "cartulary.reference_pack.enrichment.entra_app_ids.entry.v2", "entry_id": want.ID,
			"app_id": want.ID, "display_label": want.Label, "aliases": want.Aliases, "description": description,
			"source_refs": refs, "publisher": "Microsoft", "service_principal_names": []string{},
			"deprecated": false, "replacement_entry_id": nil, "extensions": map[string]any{},
		})
		if err != nil {
			return err
		}
		if !bytes.Equal(expected, lines[i]) {
			return fmt.Errorf("entra factual corpus mismatch at entry %d (%s): identity, text, defaults or exact provenance", i, want.ID)
		}
		aliases += len(want.Aliases)
		if len(want.Aliases) != 0 {
			aliased++
		}
	}
	if historical != 119 || original != 16 || nulls != 4 || aliases != 139 || aliased != 135 {
		return fmt.Errorf("entra classification accounting")
	}
	return nil
}
