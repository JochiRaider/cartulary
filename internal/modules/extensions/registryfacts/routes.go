package registryfacts

import "strings"

func RoutesOverlap(left, right string) bool {
	a, aOK := RouteSegments(left)
	b, bOK := RouteSegments(right)
	return aOK && bOK && compatibleRoutePrefix(a, b)
}

func BaseReservationOverlaps(extension, base, scope string) bool {
	a, aOK := RouteSegments(extension)
	b, bOK := RouteSegments(base)
	if !aOK || !bOK || scope != "exact" && scope != "descendants" {
		return false
	}
	// An extension root owns descendants. A deeper exact Base route is captured;
	// a shallower exact Base route is outside the extension's path set.
	if scope == "exact" && len(a) > len(b) {
		return false
	}
	return compatibleRoutePrefix(a, b)
}

func compatibleRoutePrefix(a, b []string) bool {
	for i := 0; i < min(len(a), len(b)); i++ {
		if a[i] != b[i] && !strings.HasPrefix(a[i], "{") && !strings.HasPrefix(b[i], "{") {
			return false
		}
	}
	return true
}
