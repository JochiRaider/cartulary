package native

import (
	"net/netip"
	"strconv"
	"strings"
)

func asciiUpper(value string) string {
	data := []byte(value)
	for i, c := range data {
		if c >= 'a' && c <= 'z' {
			data[i] = c - 32
		}
	}
	return string(data)
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

	default:
		return "", "unsupported_address"
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
	// Validate the enclosing URI before component helpers can trim a host.
	// Percent escapes remain encoded here and are checked by each component.
	for _, c := range []byte(v) {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("-._~:/?#[]@!$&'()*+,;=%", rune(c))) {
			return "", false
		}
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

func httpsReference(v string) bool {
	out, ok := NormalizeURL(v)
	return ok && out == v && strings.HasPrefix(v, "https://") && !strings.Contains(v, "#")
}
