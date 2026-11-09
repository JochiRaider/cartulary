package producer

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	windowsinputs "github.com/JochiRaider/cartulary/referencepacks/windows/inputs"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

const WindowsKey = "enrichment.windows_event_ids"
const WindowsProfile = "cartulary.windows_event_ids.microsoft_references.dataset.v1"
const WindowsReleaseRepository = "cartulary.windows_event_ids"
const WindowsFixtureRepository = DatasetFixtureRepository
const windowsArtifactLimit int64 = 64 << 20
const windowsTotalLimit int64 = 512 << 20
const windowsSourceLimit = 2048

// The reviewed producer catalog is authored once; runtime entries remain provider-neutral.
var windowsProviders = func() []string {
	b, err := windowsinputs.Files.ReadFile("dataset-profile.v1.json")
	if err != nil {
		panic(err)
	}
	var profile struct {
		Providers []string `json:"providers"`
	}
	if err = json.Unmarshal(b, &profile); err != nil {
		panic(err)
	}
	if len(profile.Providers) == 0 {
		panic("empty Windows provider catalog")
	}
	return profile.Providers
}()

// Producer errors are not runtime/public API error registry members.
type WindowsError struct {
	Code   string `json:"code"`
	Source string `json:"source"`
	Detail string `json:"detail"`
}

func (e *WindowsError) Error() string {
	return fmt.Sprintf("windows producer %s (%s): %s", e.Code, e.Source, e.Detail)
}
func windowsFailure(code, source, detail string) error { return &WindowsError{code, source, detail} }

type WindowsSource struct {
	ID             string  `json:"id"`
	URL            string  `json:"url"`
	Kind           string  `json:"kind"`
	Revision       *string `json:"revision"`
	ExpectedSHA256 *string `json:"expected_sha256"`
}
type WindowsDiscovery struct {
	Parent      string `json:"parent"`
	URL         string `json:"url"`
	Disposition string `json:"disposition"`
}
type WindowsInventory struct {
	SchemaID  string             `json:"schema_id"`
	Purpose   string             `json:"purpose"`
	Sources   []WindowsSource    `json:"sources"`
	Discovery []WindowsDiscovery `json:"discovery"`
}
type WindowsAcquired struct {
	Source        WindowsSource `json:"source"`
	ResolvedURL   string        `json:"resolved_url"`
	RetrievedAt   *string       `json:"retrieved_at"`
	ReceivedAt    string        `json:"received_at"`
	Method        string        `json:"method"`
	OriginReceipt *string       `json:"origin_receipt_sha256"`
	ContentType   string        `json:"content_type"`
	Evidence      EvidenceFile  `json:"evidence"`
}
type WindowsLock struct {
	SchemaID        string            `json:"schema_id"`
	InventorySHA256 string            `json:"inventory_sha256"`
	Purpose         string            `json:"purpose"`
	Sources         []WindowsAcquired `json:"sources"`
}
type WindowsAcquisitionReport struct {
	SchemaID        string            `json:"schema_id"`
	Status          string            `json:"status"`
	InventorySHA256 string            `json:"inventory_sha256"`
	LockSHA256      *string           `json:"source_lock_sha256"`
	Sources         []WindowsAcquired `json:"sources"`
	Issues          []WindowsError    `json:"issues"`
}
type WindowsRecord struct {
	SourceID       string                 `json:"source_id"`
	Locator        string                 `json:"locator"`
	FragmentSHA256 string                 `json:"fragment_sha256"`
	Provider       string                 `json:"provider_name"`
	GUID           *string                `json:"provider_guid"`
	EventID        int64                  `json:"event_id"`
	Version        *int64                 `json:"event_version"`
	Level          *int64                 `json:"level"`
	Task           *int64                 `json:"task"`
	Opcode         *int64                 `json:"opcode"`
	Kind           string                 `json:"evidence_kind"`
	Context        string                 `json:"context"`
	Message        *string                `json:"message"`
	Template       *string                `json:"template"`
	Documentation  []WindowsDocumentation `json:"documentation"`
}

// WindowsDocumentation is a closed tagged union in the runtime entry contract.
// Producer blocks carry exact source locators until reconciliation assigns the
// immutable archive reference. Source text is deliberately not trimmed.
type WindowsDocumentation struct {
	Kind    string         `json:"kind"`
	Scope   string         `json:"scope"`
	Locator string         `json:"locator"`
	Content map[string]any `json:"content"`
}
type WindowsDisposition struct {
	SourceID    string  `json:"source_id"`
	Locator     string  `json:"locator"`
	Disposition string  `json:"disposition"`
	Reason      string  `json:"reason"`
	EntryID     *string `json:"entry_id"`
}
type WindowsNote struct {
	ID             string `json:"id"`
	EntryID        string `json:"entry_id"`
	SourceID       string `json:"source_id"`
	Locator        string `json:"locator"`
	FragmentSHA256 string `json:"fragment_sha256"`
	Text           string `json:"text"`
}
type WindowsEditorial struct {
	SchemaID           string                     `json:"schema_id"`
	Notes              []WindowsNote              `json:"notes"`
	ProviderBindings   []WindowsProviderBinding   `json:"provider_bindings"`
	ProviderExclusions []WindowsProviderExclusion `json:"provider_exclusions,omitempty"`
}

// A user-reviewed exclusion applies only to one unchanged CSV row whose
// provider remains unresolved. Other defects and future rows still block.
type WindowsProviderExclusion struct {
	SourceID       string `json:"source_id"`
	Locator        string `json:"locator"`
	FragmentSHA256 string `json:"fragment_sha256"`
}

// A reviewed cross-source association selects exact bytes, not a provider to
// assume. The adapter independently verifies the definition and reads its name.
type WindowsProviderBinding struct {
	CSVSourceID    string `json:"csv_source_id"`
	CSVLocator     string `json:"csv_locator"`
	CSVSHA256      string `json:"csv_fragment_sha256"`
	SourceID       string `json:"source_id"`
	SourceSHA256   string `json:"source_sha256"`
	Locator        string `json:"locator"`
	FragmentSHA256 string `json:"fragment_sha256"`
}
type WindowsEvidence struct {
	SchemaID   string             `json:"schema_id"`
	Records    []WindowsRecord    `json:"records"`
	Notes      []WindowsNote      `json:"notes"`
	Assertions []WindowsAssertion `json:"assertions"`
}
type WindowsCoverage struct {
	Provider        string  `json:"provider_name"`
	Assertions      int     `json:"assertion_count"`
	Entries         int     `json:"entry_count"`
	NumericVersions []int64 `json:"numeric_versions"`
	Unspecified     int     `json:"unspecified_version_entries"`
}
type WindowsReconciliation struct {
	SchemaID                  string               `json:"schema_id"`
	Status                    string               `json:"status"`
	SourceLockSHA256          string               `json:"source_lock_sha256"`
	SourceProfileSHA256       string               `json:"source_profile_sha256"`
	EditorialSHA256           string               `json:"editorial_sha256"`
	Builder                   Builder              `json:"builder"`
	EntryCount                int                  `json:"entry_count"`
	PayloadFileSHA256         string               `json:"payload_file_sha256"`
	EvidenceSHA256            string               `json:"evidence_sha256"`
	MaximumDescriptionScalars int                  `json:"maximum_description_scalars"`
	MaximumLineBytes          int                  `json:"maximum_line_bytes"`
	Coverage                  []WindowsCoverage    `json:"coverage"`
	Dispositions              []WindowsDisposition `json:"dispositions"`
	Issues                    []WindowsError       `json:"issues"`
	Limitations               []string             `json:"limitations"`
}

var windowsSchemas = sync.OnceValues(func() (map[string]*jsonschema.Schema, error) {
	b, err := windowsinputs.Files.ReadFile("contracts.v3.schema.json")
	if err != nil {
		return nil, err
	}
	schemas := map[string]*jsonschema.Schema{}
	for _, name := range []string{"inventory", "lock", "acquisition", "acquisition_inputs", "evidence", "editorial", "reconciliation", "build", "license", "error", "csv_row", "summaries"} {
		s, e := compileProducerSchema(b, "#/$defs/"+name)
		if e != nil {
			return nil, e
		}
		schemas[name] = s
	}
	for _, name := range []string{"docx_sample_metadata_for_row_version_json", "docx_other_sample_metadata_json", "docx_section_contexts_json", "docx_field_definitions_json", "docx_reference_tables_json", "docx_links_json", "docx_image_references_json", "docx_subcategory_guidance_json", "xlsx_catalog_rows_json", "source_locators_json", "source_artifact_metadata_json", "quality_notes_json"} {
		s, e := compileProducerSchema(b, "#/$defs/"+name)
		if e != nil {
			return nil, e
		}
		schemas[name] = s
	}
	return schemas, nil
})

func windowsDecode(b []byte, name string, target any) error {
	schemas, err := windowsSchemas()
	if err != nil {
		return err
	}
	v, err := canonicaljson.DecodeStrict(b)
	if err != nil {
		return windowsFailure("input_invalid", name, "strict JSON required")
	}
	if err = schemas[name].Validate(v); err != nil {
		return windowsFailure("input_invalid", name, err.Error())
	}
	return DecodeClosed(b, target)
}
func windowsInventory(b []byte) (WindowsInventory, error) {
	var v WindowsInventory
	if err := windowsDecode(b, "inventory", &v); err != nil {
		return v, err
	}
	seenID, seenURL := map[string]bool{}, map[string]bool{}
	for _, s := range v.Sources {
		if !windowsSourceURL(s) || seenID[s.ID] || seenURL[s.URL] {
			return v, windowsFailure("input_invalid", s.ID, "unique approved source URL required")
		}
		seenID[s.ID], seenURL[s.URL] = true, true
		if s.Kind == "sysmon_manifest" || s.Kind == "sysmon_template" {
			name := "manifest.xml"
			if s.Kind == "sysmon_template" {
				name = "manifest.tt"
			}
			if s.Revision == nil || !commitPattern.MatchString(*s.Revision) || s.URL != "https://raw.githubusercontent.com/microsoft/SysmonCommon/"+*s.Revision+"/"+name {
				return v, windowsFailure("input_invalid", s.ID, "Sysmon source requires a matching pinned revision and resource")
			}
		}
	}
	slices.SortFunc(v.Sources, func(a, b WindowsSource) int { return strings.Compare(a.ID, b.ID) })
	return v, nil
}
func windowsProvider(s string) (string, bool) {
	for _, p := range windowsProviders {
		if strings.EqualFold(p, s) {
			return p, true
		}
	}
	return "", false
}
func windowsIssue(err error, source string) WindowsError {
	var w *WindowsError
	if errors.As(err, &w) {
		issue := *w
		if issue.Source == "" {
			issue.Source = source
		}
		return issue
	}
	return WindowsError{"acquisition_failed", source, "source operation failed"}
}

var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
