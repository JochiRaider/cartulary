package native

func policy(Manifest) hooks      { return hooks{} }
func httpsReference(string) bool { return false }
func entry(m map[string]any, c *semanticChecks, manifest Manifest) {
	id := m["entry_id"].(string)
	c.require(id == m["app_id"], ".app_id")
	c.require(nullableText(m["publisher"], 256, false), ".publisher")
	c.ordered(m["service_principal_names"], ".service_principal_names", asciiLower, func(v string) bool { return isASCII(v) && singleLine(v, 1024) })
}
