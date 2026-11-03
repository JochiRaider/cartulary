package extensions

// ApplicationLeaseAdmissionFailure reports the configured acquisition budget,
// without exporting a database session, lock key or driver error.
func (c *Coordinator) ApplicationLeaseAdmissionFailure(timeoutSeconds int64, cause error) error {
	return startupFailure(cause, c.registeredResultFinding("extension_application_lease_timeout", StartupFinding{Path: "$", ReasonCode: "extension_application_process_active", Message: "Another Cartulary application process is active for this deployment.", Details: map[string]any{"timeout_seconds": timeoutSeconds}}))
}

func (c *Coordinator) RecoveryLeaseAdmissionFailure(cause error) error {
	finding := c.registeredResultFinding("recovery_serving_lease_active", StartupFinding{Path: "$", ReasonCode: "recovery_serving_lease_active", Message: "Recovery is active for this deployment; application listener startup is closed.", Details: map[string]any{}})
	return &StartupValidationError{Findings: []StartupFinding{finding}, cause: cause, envelopeCode: "recovery_serving_lease_active"}
}

func (c *Coordinator) RecoveryTargetAdmissionFailure(cause error) error {
	finding := c.registeredResultFinding("recovery_target_incomplete", StartupFinding{Path: "$", ReasonCode: "recovery_target_incomplete", Message: "Recovery target completion is not established; application startup is closed.", Details: map[string]any{}})
	return &StartupValidationError{Findings: []StartupFinding{finding}, cause: cause, envelopeCode: "recovery_target_incomplete"}
}
