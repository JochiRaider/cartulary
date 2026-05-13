package incidents

type StartupCandidate struct {
	SheetRef map[string]string
	Valid    bool
}

func SelectStartupSheet(explicit StartupCandidate, home StartupCandidate, defaults StartupCandidate) (map[string]string, []string) {
	cleared := []string{}
	for _, candidate := range []struct {
		name string
		ref  StartupCandidate
	}{
		{name: "explicit", ref: explicit},
		{name: "home", ref: home},
		{name: "default", ref: defaults},
	} {
		if candidate.ref.SheetRef == nil {
			continue
		}
		if candidate.ref.Valid {
			return cloneSheetRef(candidate.ref.SheetRef), cleared
		}
		cleared = append(cleared, candidate.name)
	}
	return map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v1"}, cleared
}

func cloneSheetRef(value map[string]string) map[string]string {
	if value == nil {
		return nil
	}
	cloned := make(map[string]string, len(value))
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}
