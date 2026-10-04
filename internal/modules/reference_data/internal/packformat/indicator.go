package packformat

import (
	"encoding/json"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type IndicatorPolicy struct {
	Type      string   `json:"indicator_type_id"`
	Kinds     []string `json:"allowed_value_kinds"`
	Normalize string   `json:"normalization_algorithm_id"`
	Validate  string   `json:"validation_algorithm_id"`
	Defang    string   `json:"defang_algorithm_id"`
	Dedupe    string   `json:"dedupe_algorithm_id"`
}
type Evaluation struct {
	Type                   string  `json:"indicator_type_id"`
	Kind                   string  `json:"value_kind"`
	Raw                    string  `json:"raw_value"`
	Valid                  bool    `json:"valid"`
	Code                   *string `json:"validation_code"`
	Display                *string `json:"display_value"`
	Normalized             *string `json:"normalized_value"`
	Defanged               *string `json:"defanged_value"`
	DedupeKey              *string `json:"dedupe_key"`
	NormalizationAlgorithm string  `json:"normalization_algorithm_id"`
	ValidationAlgorithm    string  `json:"validation_algorithm_id"`
	DefangAlgorithm        string  `json:"defang_algorithm_id"`
	DedupeAlgorithm        string  `json:"dedupe_algorithm_id"`
}

var indicatorPolicies = loadIndicatorPolicies()

func loadIndicatorPolicies() map[string]IndicatorPolicy {
	var catalog struct {
		Policies []IndicatorPolicy `json:"policies"`
	}
	if err := json.Unmarshal(projection("indicator_algorithms.v1.json"), &catalog); err != nil || len(catalog.Policies) != 9 {
		panic("invalid indicator algorithm projection")
	}
	result := map[string]IndicatorPolicy{}
	for _, p := range catalog.Policies {
		result[p.Type] = p
	}
	return result
}
func IndicatorPolicies() []IndicatorPolicy {
	result := []IndicatorPolicy{}
	for _, p := range indicatorPolicies {
		p.Kinds = slices.Clone(p.Kinds)
		result = append(result, p)
	}
	slices.SortFunc(result, func(a, b IndicatorPolicy) int { return strings.Compare(a.Type, b.Type) })
	return result
}

// Evaluate applies a resolved registry policy. Operation errors return no
// evaluation (and therefore no raw input echo); invalid admitted values return
// a complete evaluation with null derived values.
func Evaluate(policy IndicatorPolicy, kind, raw string) (Evaluation, error) {
	if !utf8.ValidString(raw) {
		return Evaluation{}, fail("invalid_pack_request")
	}
	if utf8.RuneCountInString(raw) > 8192 {
		return Evaluation{}, fail("indicator_input_too_long")
	}
	registered, ok := indicatorPolicies[policy.Type]
	if !ok {
		return Evaluation{}, fail("indicator_type_unsupported")
	}
	if !slices.Contains(policy.Kinds, kind) {
		return Evaluation{}, fail("indicator_value_kind_unsupported")
	}
	if policy.Normalize != registered.Normalize || policy.Validate != registered.Validate || policy.Defang != registered.Defang || policy.Dedupe != registered.Dedupe || !slices.Equal(policy.Kinds, registered.Kinds) {
		return Evaluation{}, fail("indicator_algorithm_unsupported")
	}
	result := Evaluation{Type: policy.Type, Kind: kind, Raw: raw, NormalizationAlgorithm: policy.Normalize, ValidationAlgorithm: policy.Validate, DefangAlgorithm: policy.Defang, DedupeAlgorithm: policy.Dedupe}
	family := strings.TrimSuffix(strings.TrimPrefix(policy.Normalize, "cartulary.indicator.normalize."), ".v1")
	value, code := normalizeIndicator(family, raw)
	if code != "" {
		result.Code = &code
		return result, nil
	}
	defanged := value
	switch policy.Defang {
	case "cartulary.indicator.defang.ipv4.v1", "cartulary.indicator.defang.domain.v1":
		defanged = strings.ReplaceAll(value, ".", "[.]")
	case "cartulary.indicator.defang.ipv6.v1":
		defanged = strings.ReplaceAll(value, ":", "[:]")
	case "cartulary.indicator.defang.email.v1":
		local, host, _ := strings.Cut(value, "@")
		defanged = local + "[@]" + strings.ReplaceAll(host, ".", "[.]")
	case "cartulary.indicator.defang.http_url.v1":
		scheme, rest, _ := strings.Cut(value, "://")
		end := strings.IndexByte(rest, '/')
		authority := rest[:end]
		suffix := rest[end:]
		if strings.HasPrefix(authority, "[") {
			close := strings.IndexByte(authority, ']')
			authority = "[" + strings.ReplaceAll(authority[1:close], ":", "[:]") + authority[close:]
		} else {
			authority = strings.ReplaceAll(authority, ".", "[.]")
		}
		defanged = strings.Replace(scheme, "http", "hxxp", 1) + "://" + authority + suffix
	}
	dedupe := policy.Type + ":" + value
	if policy.Dedupe == "cartulary.indicator.dedupe.sha256.v1" {
		dedupe = "sha256:" + value
	}
	if policy.Dedupe == "cartulary.indicator.dedupe.text_hash.v1" {
		dedupe = "text:" + Digest([]byte(value))
	}
	result.Valid = true
	result.Display = &value
	result.Normalized = &value
	result.Defanged = &defanged
	result.DedupeKey = &dedupe
	return result, nil
}
func TrimIndicatorInput(value string) string {
	return strings.TrimFunc(value, func(r rune) bool {
		return r >= 9 && r <= 13 || r == 32 || r == 0x85 || r == 0xa0 || r == 0x1680 || r >= 0x2000 && r <= 0x200a || r == 0x2028 || r == 0x2029 || r == 0x202f || r == 0x205f || r == 0x3000
	})
}
func asciiLower(value string) string {
	data := []byte(value)
	for i, c := range data {
		if c >= 'A' && c <= 'Z' {
			data[i] = c + 32
		}
	}
	return string(data)
}
func asciiUpper(value string) string {
	data := []byte(value)
	for i, c := range data {
		if c >= 'a' && c <= 'z' {
			data[i] = c - 32
		}
	}
	return string(data)
}
func isASCII(value string) bool {
	for _, c := range []byte(value) {
		if c > 127 {
			return false
		}
	}
	return true
}
func noControls(value string, multiline bool) bool {
	for _, r := range value {
		if multiline && (r == '\n' || r == '\t') {
			continue
		}
		if r < 32 || r >= 127 && r <= 159 {
			return false
		}
	}
	return true
}
func normalizeIndicator(family, raw string) (string, string) {
	v := TrimIndicatorInput(raw)
	switch family {
	case "ipv4":
		a, err := netip.ParseAddr(v)
		if err != nil || !a.Is4() {
			return "", "invalid_ipv4"
		}
		return a.String(), ""
	case "ipv6":
		a, err := netip.ParseAddr(v)
		if err != nil || !a.Is6() || a.Is4In6() || a.Zone() != "" || strings.Contains(v, ".") {
			return "", "invalid_ipv6"
		}
		b := a.As16()
		compatible := true
		for _, part := range b[:12] {
			compatible = compatible && part == 0
		}
		if compatible && !a.IsUnspecified() && !a.IsLoopback() {
			return "", "invalid_ipv6"
		}
		return a.String(), ""
	case "domain_ascii":
		if result, ok := normalizeDomain(v); ok {
			return result, ""
		}
		return "", "invalid_domain_name"
	case "http_url":
		if result, ok := NormalizeURL(v); ok {
			return result, ""
		}
		return "", "invalid_http_url"
	case "sha256":
		if _, ok := hexBytes(asciiLower(v), 32); ok {
			return asciiLower(v), ""
		}
		return "", "invalid_sha256"
	case "email_ascii":
		if !isASCII(v) || strings.Count(v, "@") != 1 {
			return "", "invalid_email_addr"
		}
		local, host, _ := strings.Cut(v, "@")
		domain, ok := normalizeDomain(host)
		if !ok || len(local) < 1 || len(local) > 64 || len(local)+1+len(domain) > 254 || strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") || strings.Contains(local, "..") {
			return "", "invalid_email_addr"
		}
		for _, c := range []byte(local) {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("!#$%&'*+-/=?^_`{|}~.", rune(c))) {
				return "", "invalid_email_addr"
			}
		}
		return local + "@" + domain, ""
	case "windows_registry_key":
		v = norm.NFC.String(v)
		v = strings.ReplaceAll(v, "/", `\`)
		for strings.Contains(v, `\\`) {
			v = strings.ReplaceAll(v, `\\`, `\`)
		}
		v = strings.TrimSuffix(v, `\`)
		root, rest, child := strings.Cut(v, `\`)
		root = asciiUpper(root)
		aliases := map[string]string{"HKLM": "HKEY_LOCAL_MACHINE", "HKCU": "HKEY_CURRENT_USER", "HKCR": "HKEY_CLASSES_ROOT", "HKU": "HKEY_USERS", "HKCC": "HKEY_CURRENT_CONFIG"}
		if full, ok := aliases[root]; ok {
			root = full
		} else {
			found := false
			for _, full := range aliases {
				found = found || root == full
			}
			if !found {
				return "", "invalid_registry_key"
			}
		}
		v = root
		if child {
			v += `\` + rest
		}
		v = asciiUpper(v)
		if !noControls(v, false) || utf8.RuneCountInString(v) > 1024 {
			return "", "invalid_registry_key"
		}
		return v, ""
	case "process_name":
		v = norm.NFC.String(v)
		if v == "" || utf8.RuneCountInString(v) > 260 || !noControls(v, false) || strings.ContainsAny(v, `/\`) {
			return "", "invalid_process_name"
		}
		return v, ""
	case "text":
		v = norm.NFC.String(v)
		v = strings.ReplaceAll(strings.ReplaceAll(v, "\r\n", "\n"), "\r", "\n")
		if v == "" || !noControls(v, true) || utf8.RuneCountInString(v) > 8192 {
			return "", "invalid_text"
		}
		return v, ""
	default:
		return "", "indicator_algorithm_unsupported"
	}
}
func normalizeDomain(value string) (string, bool) {
	if !isASCII(value) {
		return "", false
	}
	v := asciiLower(strings.TrimSuffix(value, "."))
	if len(v) < 1 || len(v) > 253 {
		return "", false
	}
	for _, label := range strings.Split(v, ".") {
		if len(label) < 1 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", false
		}
		for _, c := range []byte(label) {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return "", false
			}
		}
	}
	return v, true
}
func uriComponent(value string, query bool) (string, bool) {
	var out strings.Builder
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c == '%' {
			if i+2 >= len(value) {
				return "", false
			}
			v, err := strconv.ParseUint(value[i+1:i+3], 16, 8)
			if err != nil {
				return "", false
			}
			c = byte(v)
			if unreserved(c) {
				out.WriteByte(c)
			} else {
				out.WriteByte('%')
				out.WriteString(asciiUpper(value[i+1 : i+3]))
			}
			i += 2
			continue
		}
		if !(unreserved(c) || strings.ContainsRune("!$&'()*+,;=:@/", rune(c)) || query && c == '?') {
			return "", false
		}
		out.WriteByte(c)
	}
	return out.String(), true
}
func unreserved(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("-._~", rune(c))
}
func removeDotSegments(path string) string {
	output := ""
	for path != "" {
		switch {
		case strings.HasPrefix(path, "../"):
			path = path[3:]
		case strings.HasPrefix(path, "./"):
			path = path[2:]
		case strings.HasPrefix(path, "/./"):
			path = path[2:]
		case path == "/.":
			path = "/"
		case strings.HasPrefix(path, "/../") || path == "/..":
			if path == "/.." {
				path = "/"
			} else {
				path = path[3:]
			}
			if i := strings.LastIndexByte(output, '/'); i >= 0 {
				output = output[:i]
			} else {
				output = ""
			}
		case path == "." || path == "..":
			path = ""
		default:
			start := 0
			if path[0] == '/' {
				start = 1
			}
			end := strings.IndexByte(path[start:], '/')
			if end < 0 {
				output += path
				path = ""
			} else {
				end += start
				output += path[:end]
				path = path[end:]
			}
		}
	}
	return output
}

func NormalizeURL(raw string) (string, bool) {
	v := TrimIndicatorInput(raw)
	if !isASCII(v) {
		return "", false
	}
	scheme, rest, ok := strings.Cut(v, "://")
	scheme = asciiLower(scheme)
	if !ok || (scheme != "http" && scheme != "https") {
		return "", false
	}
	rest, fragment, hasFragment := strings.Cut(rest, "#")
	rest, query, hasQuery := strings.Cut(rest, "?")
	authority, path, hasPath := strings.Cut(rest, "/")
	if strings.Contains(authority, "@") || authority == "" {
		return "", false
	}
	if hasPath {
		path = "/" + path
	} else {
		path = "/"
	}
	host := authority
	port := ""
	if strings.HasPrefix(authority, "[") {
		end := strings.IndexByte(authority, ']')
		if end < 0 {
			return "", false
		}
		canonical, code := normalizeIndicator("ipv6", authority[1:end])
		if code != "" {
			return "", false
		}
		host = "[" + canonical + "]"
		if end+1 < len(authority) {
			if authority[end+1] != ':' {
				return "", false
			}
			port = authority[end+2:]
			if port == "" {
				return "", false
			}
		}
	} else {
		if strings.Contains(authority, ":") {
			host, port, _ = strings.Cut(authority, ":")
			if port == "" || strings.Contains(port, ":") {
				return "", false
			}
		}
		host = strings.TrimSuffix(host, ".")
		// A second trailing dot remains invalid; do not let domain normalization
		// remove another dot and make normalization change on a second pass.
		if strings.HasSuffix(host, ".") {
			return "", false
		}
		labels := strings.Split(host, ".")
		ipv4 := len(labels) == 4
		for _, label := range labels {
			if label == "" {
				ipv4 = false
			}
			for _, c := range []byte(label) {
				if c < '0' || c > '9' {
					ipv4 = false
				}
			}
		}
		if ipv4 {
			var code string
			host, code = normalizeIndicator("ipv4", host)
			if code != "" {
				return "", false
			}
		} else {
			var valid bool
			host, valid = normalizeDomain(host)
			if !valid {
				return "", false
			}
		}
	}
	if port != "" {
		for _, c := range []byte(port) {
			if c < '0' || c > '9' {
				return "", false
			}
		}
		n, err := strconv.ParseUint(port, 10, 16)
		if err != nil || n == 0 {
			return "", false
		}
		port = strconv.FormatUint(n, 10)
		if scheme == "http" && port == "80" || scheme == "https" && port == "443" {
			port = ""
		}
	}
	path, ok = uriComponent(path, false)
	if !ok {
		return "", false
	}
	path = removeDotSegments(path)
	if path == "" {
		path = "/"
	}
	query, ok = uriComponent(query, true)
	if !ok {
		return "", false
	}
	fragment, ok = uriComponent(fragment, true)
	if !ok {
		return "", false
	}
	result := scheme + "://" + host
	if port != "" {
		result += ":" + port
	}
	result += path
	if hasQuery {
		result += "?" + query
	}
	if hasFragment {
		result += "#" + fragment
	}
	return result, len(result) <= 8192
}
