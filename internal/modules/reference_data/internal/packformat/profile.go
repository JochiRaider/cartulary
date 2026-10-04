package packformat

import (
	"fmt"
	"golang.org/x/text/unicode/norm"
	"net/netip"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

var registryToken = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
var stableToken = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,191}$`)
var cveToken = regexp.MustCompile(`^CVE-[0-9]{4}-[0-9]{4,}$`)
var serviceToken = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,127}$`)
var uuidToken = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func ValidateEntryIdentity(key, id string) error {
	valid := false
	switch key {
	case "type_registry.host", "type_registry.evidence", "type_registry.indicator":
		valid = registryToken.MatchString(id)
	case "enrichment.windows_event_ids":
		_, _, _, valid = ParseEventIdentity(id)
	case "enrichment.windows_sids":
		valid = ValidSID(id)
	case "enrichment.cisa_kev":
		valid = len(id) <= 192 && cveToken.MatchString(id)
	case "enrichment.entra_app_ids":
		valid = uuidToken.MatchString(id)
	case "enrichment.tor":
		prefix, err := netip.ParsePrefix(id)
		valid = err == nil && prefix == prefix.Masked() && prefix.String() == id
	case "enrichment.ms_portals":
		valid = serviceToken.MatchString(id)
	default:
		_, registered := profiles[key]
		valid = registered && stableToken.MatchString(id)
	}
	if !valid {
		return fail("invalid_pack_request")
	}
	return nil
}

// Registry compatibility is intentionally later than ordinary row semantics.
// Algorithms and vocabulary are accepted only after all content has passed its
// structural and semantic checks.
func validateRegistryEntry(profile string, m map[string]any, builtin bool, c *semanticChecks) error {
	id := m["entry_id"].(string)
	text := func(name string) string { return m[name].(string) }
	values := func(name string) []string {
		result := []string{}
		for _, v := range m[name].([]any) {
			result = append(result, v.(string))
		}
		return result
	}
	switch profile {
	case "type_registry.host", "type_registry.evidence":
		if builtin && id == "unknown" {
			c.require(!m["deprecated"].(bool), ".deprecated")
			c.require(m["replacement_entry_id"] == nil, ".replacement_entry_id")
		}
	case "type_registry.indicator":
		p, ok := indicatorPolicies[id]
		c.require(ok, ".entry_id")
		c.require(!m["deprecated"].(bool), ".deprecated")
		c.require(m["replacement_entry_id"] == nil, ".replacement_entry_id")
		if !ok {
			return c.finish("type_registry_incompatible")
		}
		c.require(slices.Equal(values("allowed_value_kinds"), p.Kinds), ".allowed_value_kinds")
		c.require(text("normalization_algorithm_id") == p.Normalize, ".normalization_algorithm_id")
		c.require(text("validation_algorithm_id") == p.Validate, ".validation_algorithm_id")
		c.require(text("defang_algorithm_id") == p.Defang, ".defang_algorithm_id")
		c.require(text("dedupe_algorithm_id") == p.Dedupe, ".dedupe_algorithm_id")
		mapping := map[string][2]string{"ipv4_addr": {"ipv4-addr", "value"}, "ipv6_addr": {"ipv6-addr", "value"}, "domain_name": {"domain-name", "value"}, "url": {"url", "value"}, "sha256": {"file", "hashes.'SHA-256'"}, "email_addr": {"email-addr", "value"}, "registry_key": {"windows-registry-key", "key"}, "process_name": {"process", "name"}}
		if id == "text" {
			c.require(m["stix_mapping"] == nil, ".stix_mapping")
		} else if actual, ok := object(m["stix_mapping"]); ok {
			c.require(actual["stix_object_type"] == mapping[id][0], ".stix_mapping.stix_object_type")
			c.require(actual["pattern_property"] == mapping[id][1], ".stix_mapping.pattern_property")
		} else {
			c.require(false, ".stix_mapping")
		}
	}
	return c.finish("type_registry_incompatible")
}

func validateProfileEntry(profile string, m map[string]any, c *semanticChecks) ([]LookupKey, error) {
	id := m["entry_id"].(string)
	keys := []LookupKey{}
	add := func(kind, value, order string) { keys = append(keys, LookupKey{kind, value, order}) }
	text := func(name string) string { return m[name].(string) }
	line := func(name string, max int) { c.require(singleLine(text(name), max), "."+name) }
	nullable := func(name string, max int, multi bool) { c.require(nullableText(m[name], max, multi), "."+name) }
	array := func(name string, normalize func(string) string, valid func(string) bool) {
		c.ordered(m[name], "."+name, normalize, valid)
	}
	values := func(name string) []string {
		result := []string{}
		for _, v := range m[name].([]any) {
			result = append(result, v.(string))
		}
		return result
	}
	switch profile {
	case "type_registry.host", "type_registry.evidence", "type_registry.indicator":
		add("entry_id", id, id)
	case "enrichment.tor":
		p, err := netip.ParsePrefix(text("network"))
		c.require(err == nil && p.Masked() == p && p.String() == text("network"), ".network")
		c.require(text("network") == id, ".entry_id")
		if err == nil {
			family := "ipv6"
			if p.Addr().Is4() {
				family = "ipv4"
			}
			c.require(text("address_family") == family, ".address_family")
			_, code := normalizeIndicator(family, p.Addr().String())
			c.require(code == "", ".network")
			add("network", id, id)
			add("ip_literal", id, fmt.Sprintf("%03d\x00%s", 128-p.Bits(), id))
		}
		first, ok := dateValue(m["first_observed_at"])
		last, ok2 := dateValue(m["last_observed_at"])
		c.require(ok, ".first_observed_at")
		c.require(ok2 && (!ok || first == "" || last == "" || first <= last), ".last_observed_at")
	case "enrichment.cisa_kev":
		c.require(id == text("cve_id"), ".entry_id")
		line("vendor_project", 256)
		line("product", 256)
		line("vulnerability_name", 512)
		c.require(multiline(text("short_description")), ".short_description")
		c.require(multiline(text("required_action")), ".required_action")
		nullable("notes", 8192, true)
		added, ok := dateValue(m["date_added"])
		due, ok2 := dateValue(m["due_date"])
		c.require(ok, ".date_added")
		c.require(ok2 && (!ok || due >= added), ".due_date")
		add("cve_id", id, id)
	case "enrichment.ms_portals":
		c.require(id == text("service_id"), ".entry_id")
		array("portal_urls", unchanged, httpsReference)
		array("audiences", unchanged, commonLine)
		array("clouds", unchanged, commonLine)
		add("service_id", id, id)
		for _, v := range values("portal_urls") {
			if u, err := url.Parse(v); err == nil {
				add("url_host", u.Hostname(), id)
			}
		}
	case "enrichment.windows_event_ids":
		line("provider_name", 256)
		nullable("level", 128, false)
		nullable("task", 256, false)
		nullable("opcode", 128, false)
		event, _ := integer(m["event_id"], 0, 65535)
		var version *int64
		rank := int64(0)
		if m["event_version"] != nil {
			n, _ := integer(m["event_version"], 0, 255)
			version, rank = &n, n+1
		}
		c.require(id == EventIdentity(text("provider_name"), event, version), ".entry_id")
		provider := asciiLower(text("provider_name"))
		order := fmt.Sprintf("%03d\x00%s", rank, id)
		add("provider_event_id", provider+"\x00"+strconv.FormatInt(event, 10), order)
		if version != nil {
			add("provider_event_version", provider+"\x00"+strconv.FormatInt(event, 10)+"\x00"+strconv.FormatInt(*version, 10), order)
		}
		add("event_id", strconv.FormatInt(event, 10), provider+"\x00"+text("provider_name")+"\x00"+order)
	case "enrichment.entra_app_ids":
		c.require(id == text("app_id"), ".entry_id")
		nullable("publisher", 256, false)
		array("service_principal_names", asciiLower, func(v string) bool { return isASCII(v) && singleLine(v, 1024) })
		add("app_id", id, id)
		for _, v := range values("service_principal_names") {
			add("service_principal_name", asciiLower(v), id)
		}
	case "enrichment.lolbas", "enrichment.lolesxi":
		line("name", 256)
		array("platforms", asciiLower, commonLine)
		array("categories", asciiLower, commonLine)
		c.usage(m["usage_examples"], ".usage_examples")
		array("references", unchanged, httpsReference)
		add("entry_id", id, id)
		add("name", asciiLower(text("name")), id)
	case "enrichment.loldrivers":
		c.require(id == text("driver_id"), ".entry_id")
		array("filenames", asciiLower, basename)
		array("signer_names", asciiLower, commonLine)
		array("categories", asciiLower, commonLine)
		array("references", unchanged, httpsReference)
		previous := ""
		for i, value := range m["hashes"].([]any) {
			hash := value.(map[string]any)
			algorithm, digest := hash["algorithm"].(string), hash["value"].(string)
			size := map[string]int{"md5": 16, "sha1": 20, "sha256": 32}[algorithm]
			_, ok := hexBytes(digest, size)
			rank := map[string]string{"md5": "0", "sha1": "1", "sha256": "2"}[algorithm]
			order := rank + digest
			c.require(ok && order > previous, ".hashes["+strconv.Itoa(i)+"]")
			previous = order
			add("hash", algorithm+"\x00"+digest, id)
		}
		add("driver_id", id, id)
		for _, v := range values("filenames") {
			add("filename", asciiLower(v), id)
		}
	case "enrichment.hijacklibs":
		c.require(basename(text("library_name")), ".library_name")
		array("candidate_paths", WindowsPathLookup, func(v string) bool { return singleLine(v, 1024) })
		array("categories", asciiLower, commonLine)
		c.usage(m["usage_examples"], ".usage_examples")
		array("references", unchanged, httpsReference)
		add("library_name", asciiLower(text("library_name")), id)
		for _, v := range values("candidate_paths") {
			add("candidate_path", WindowsPathLookup(v), id)
		}
	case "enrichment.windows_sids":
		c.require(ValidSID(text("sid")), ".sid")
		c.require(id == text("sid"), ".entry_id")
		line("name", 256)
		add("sid", id, id)
		add("name", asciiLower(text("name")), id)
	default:
		return nil, fail("contract_incompatible")
	}
	if err := c.finish("content_semantic_invalid"); err != nil {
		return nil, err
	}
	// Distinct source values can produce the same lookup key for one entry.
	slices.SortFunc(keys, func(a, b LookupKey) int {
		if c := strings.Compare(a.Kind, b.Kind); c != 0 {
			return c
		}
		if c := strings.Compare(a.Value, b.Value); c != 0 {
			return c
		}
		return strings.Compare(a.Order, b.Order)
	})
	return slices.Compact(keys), nil
}
func basename(value string) bool {
	return isASCII(value) && singleLine(value, 260) && !strings.ContainsAny(value, `/\`)
}

// LookupInput returns a canonical index query without echoing malformed input.
// It does not choose a pack set or consult current deployment activation.
type LookupInput struct {
	Kind, Value    string
	NetworkAddress *netip.Addr
}

func NormalizeLookup(profile, kind string, input any) (LookupInput, error) {
	bad := func() (LookupInput, error) { return LookupInput{}, fail("invalid_pack_request") }
	unsupported := func() (LookupInput, error) { return LookupInput{}, fail("lookup_kind_unsupported") }
	result := LookupInput{Kind: kind}
	if _, ok := profiles[profile]; !ok {
		return unsupported()
	}
	if profile == "enrichment.windows_event_ids" {
		switch kind {
		case "event_id":
			n, ok := integer(input, 0, 65535)
			if !ok {
				return bad()
			}
			result.Value = strconv.FormatInt(n, 10)
			return result, nil
		case "provider_event_id":
			m, ok := object(input)
			if !ok || (len(m) != 2 && len(m) != 3) {
				return bad()
			}
			provider, ok := m["provider_name"].(string)
			event, ok2 := integer(m["event_id"], 0, 65535)
			if !ok || !ok2 || !singleLine(provider, 256) || !isASCII(provider) {
				return bad()
			}
			result.Value = asciiLower(provider) + "\x00" + strconv.FormatInt(event, 10)
			if v, present := m["event_version"]; present {
				n, ok := integer(v, 0, 255)
				if !ok {
					return bad()
				}
				result.Kind = "provider_event_version"
				result.Value += "\x00" + strconv.FormatInt(n, 10)
			} else if len(m) != 2 {
				return bad()
			}
			return result, nil
		default:
			return unsupported()
		}
	}
	if profile == "enrichment.loldrivers" && kind == "hash" {
		m, ok := exact(input, "algorithm", "value")
		if !ok {
			return bad()
		}
		algorithm, ok := m["algorithm"].(string)
		value, ok2 := m["value"].(string)
		size := map[string]int{"md5": 16, "sha1": 20, "sha256": 32}[algorithm]
		if !ok || !ok2 || size == 0 {
			return bad()
		}
		value = asciiLower(value)
		if _, ok := hexBytes(value, size); !ok {
			return bad()
		}
		result.Value = algorithm + "\x00" + value
		return result, nil
	}
	value, ok := input.(string)
	if !ok || !utf8.ValidString(value) || utf8.RuneCountInString(value) > 8192 || value == "" {
		return bad()
	}
	result.Value = value
	if kind == "alias" {
		if !profileAllowsAliases(profile) {
			return unsupported()
		}
		if profile == "type_registry.indicator" {
			value = TrimIndicatorInput(value)
		}
		value = norm.NFC.String(value)
		if !singleLine(value, 256) {
			return bad()
		}
		if profile == "type_registry.host" || profile == "type_registry.evidence" {
			if !isASCII(value) || value != asciiLower(value) {
				return bad()
			}
		}
		result.Value = aliasNormalizer(profile)(value)
		return result, nil
	}
	if kind == "candidate_path" {
		value = norm.NFC.String(TrimIndicatorInput(value))
		result.Value = value
	}
	if !singleLine(value, 8192) {
		return bad()
	}
	switch profile {
	case "type_registry.host", "type_registry.evidence", "type_registry.indicator":
		if kind != "entry_id" {
			return unsupported()
		}
		if !registryToken.MatchString(value) {
			return bad()
		}
	case "framework.attack", "framework.d3fend", "framework.veris":
		if kind != "object_id" && kind != "external_id" {
			return unsupported()
		}
		if kind == "object_id" && !stableToken.MatchString(value) {
			return bad()
		}
	case "enrichment.tor":
		if kind == "network" {
			p, err := netip.ParsePrefix(value)
			if err != nil || p.Masked() != p || p.String() != value {
				return bad()
			}
		}
		if kind == "ip_literal" {
			family := "ipv4"
			if strings.Contains(value, ":") {
				family = "ipv6"
			}
			v, code := normalizeIndicator(family, value)
			if code != "" {
				return bad()
			}
			a, _ := netip.ParseAddr(v)
			result.Value = v
			result.NetworkAddress = &a
		} else if kind != "network" {
			return unsupported()
		}
	case "enrichment.cisa_kev":
		if kind != "cve_id" {
			return unsupported()
		}
		result.Value = asciiUpper(value)
		if !cveToken.MatchString(result.Value) || len(value) > 192 {
			return bad()
		}
	case "enrichment.ms_portals":
		if kind == "url_host" {
			v, ok := normalizeLookupHost(value)
			if !ok {
				return bad()
			}
			result.Value = v
		} else if kind == "service_id" {
			if !serviceToken.MatchString(value) {
				return bad()
			}
		} else {
			return unsupported()
		}
	case "enrichment.entra_app_ids":
		if kind == "app_id" {
			result.Value = asciiLower(value)
			if !uuidToken.MatchString(result.Value) {
				return bad()
			}
		} else if kind == "service_principal_name" {
			if !isASCII(value) || len(value) > 1024 {
				return bad()
			}
			result.Value = asciiLower(value)
		} else {
			return unsupported()
		}
	case "enrichment.lolbas", "enrichment.lolesxi":
		if kind == "name" {
			if !singleLine(value, 256) {
				return bad()
			}
			result.Value = asciiLower(value)
		} else if kind == "entry_id" {
			if !stableToken.MatchString(value) {
				return bad()
			}
		} else {
			return unsupported()
		}
	case "enrichment.loldrivers":
		if kind == "filename" {
			if !basename(value) {
				return bad()
			}
			result.Value = asciiLower(value)
		} else if kind == "driver_id" {
			if !stableToken.MatchString(value) {
				return bad()
			}
		} else {
			return unsupported()
		}
	case "enrichment.hijacklibs":
		if kind == "library_name" {
			if !basename(value) {
				return bad()
			}
			result.Value = asciiLower(value)
		} else if kind == "candidate_path" {
			if !singleLine(value, 1024) {
				return bad()
			}
			result.Value = WindowsPathLookup(value)
		} else {
			return unsupported()
		}
	case "enrichment.windows_sids":
		if kind == "sid" {
			if !ValidSID(value) {
				return bad()
			}
		} else if kind == "name" {
			if !singleLine(value, 256) {
				return bad()
			}
			result.Value = asciiLower(value)
		} else {
			return unsupported()
		}
	}
	return result, nil
}
func normalizeLookupHost(value string) (string, bool) {
	if strings.Contains(value, ":") {
		v, code := normalizeIndicator("ipv6", value)
		return v, code == ""
	}
	result, ok := NormalizeURL("https://" + value + "/")
	if !ok {
		return "", false
	}
	u, err := url.Parse(result)
	if err != nil {
		return "", false
	}
	return u.Hostname(), u.Port() == "" && u.Path == "/" && u.RawQuery == "" && u.Fragment == ""
}
