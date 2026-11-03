package jobcontract

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// These are the imported public scalar contracts. Validation is structural:
// retained references require neither a live resource nor an active profile.
var packKey = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}(\.[a-z][a-z0-9_]{0,31}){1,7}$`)
var packVersion = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)

func resourceCollection(kind string) string {
	switch kind {
	case "incident":
		return "incidents"
	case "import_session":
		return "import-sessions"
	case "snapshot":
		return "snapshots"
	case "release":
		return "releases"
	case "incident_bundle":
		return "incident-bundles"
	}
	return ""
}

func knownResourceKind(kind string) bool {
	return resourceCollection(kind) != "" || kind == "reference_pack_version" || kind == "network_flow_table" || kind == "network_flow_graph_view"
}

func prefixedHexID(id, prefix string) bool {
	if len(id) != len(prefix)+32 || !strings.HasPrefix(id, prefix) {
		return false
	}
	for _, c := range id[len(prefix):] {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func validResourceReference(ref ResourceRef, requireRoute bool) bool {
	if requireRoute && ref.Route == "" {
		return false
	}
	if collection := resourceCollection(ref.Kind); collection != "" {
		return canonicalIdentityUUID(&ref.ID) && (ref.Route == "" || ref.Route == "/api/v1/"+collection+"/"+ref.ID)
	}
	switch ref.Kind {
	case "reference_pack_version":
		parts := strings.SplitN(ref.Route, "/", 7)
		return len(parts) == 6 && strings.Join(parts[:4], "/") == "/api/v1/reference-packs" &&
			len(parts[4]) <= 128 && len(parts[5]) <= 128 && packKey.MatchString(parts[4]) && packVersion.MatchString(parts[5]) && ref.ID == ref.Route
	case "network_flow_table", "network_flow_graph_view":
		prefix, collection := "nft_", "tables"
		if ref.Kind == "network_flow_graph_view" {
			prefix, collection = "nfgv_", "graph-views"
		}
		if !prefixedHexID(ref.ID, prefix) {
			return false
		}
		if ref.Route == "" {
			return true
		}
		parts := strings.SplitN(ref.Route, "/", 9)
		return len(parts) == 8 && canonicalIdentityUUID(&parts[4]) &&
			ref.Route == "/api/v1/incidents/"+parts[4]+"/network-flow/"+collection+"/"+ref.ID
	}
	return false
}

func resourceLess(a, b ResourceRef) bool {
	if a.Kind != b.Kind {
		return a.Kind < b.Kind
	}
	return a.ID < b.ID
}

func validateSuccess(summary Success, requireRoutes bool) error {
	if summary.Code == "" || summary.Message == "" || !utf8.ValidString(summary.Code) || !utf8.ValidString(summary.Message) || len(summary.ResourceRefs) > 1024 {
		return errors.New("invalid common-job success")
	}
	seen := make(map[ResourceRef]bool, len(summary.ResourceRefs))
	for _, ref := range summary.ResourceRefs {
		if !validResourceReference(ref, requireRoutes) {
			return errors.New("invalid common-job resource reference")
		}
		identity := ResourceRef{Kind: ref.Kind, ID: ref.ID}
		if seen[identity] {
			return errors.New("duplicate common-job resource reference")
		}
		seen[identity] = true
	}
	return nil
}

// ValidateCurrentSuccess validates the already committed public representation.
// Current Job families require read routes; the abstract schema remains wider.
func ValidateCurrentSuccess(summary Success) error {
	return validateSuccess(summary, true)
}

// CanonicalCurrentSuccess copies and orders references before a terminal write.
// It never rewrites retained bytes or resolves references against current state.
func CanonicalCurrentSuccess(summary Success) (Success, error) {
	if err := ValidateCurrentSuccess(summary); err != nil {
		return Success{}, err
	}
	result := summary
	result.ResourceRefs = append([]ResourceRef{}, summary.ResourceRefs...)
	sort.Slice(result.ResourceRefs, func(i, j int) bool { return resourceLess(result.ResourceRefs[i], result.ResourceRefs[j]) })
	return result, nil
}
