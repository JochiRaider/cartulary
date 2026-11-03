package extensions

// PublicationFailure reports only the admitted phase and effective configured
// budget. Component identities, private preparation errors and plans stay local.
func (c *Coordinator) PublicationFailure(timeoutSeconds int64, cause error) error {
	return startupFailure(cause, c.registeredResultFinding("extension_publication_failed", StartupFinding{
		Path: "$", ReasonCode: "extension_publication_failed", Message: "Extension publication did not reach the serving state.",
		Details: map[string]any{"phase": "publication", "timeout_seconds": timeoutSeconds},
	}))
}
