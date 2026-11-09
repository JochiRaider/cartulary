package native

func policy(m Manifest) hooks {
	a := newVERISState(m)
	return hooks{a.object, a.relationship, a.finishObjects, a.finish}
}
func entry(map[string]any, *semanticChecks, Manifest) {}
