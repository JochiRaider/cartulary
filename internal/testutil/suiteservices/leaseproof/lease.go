// Package leaseproof defines suite-service lifecycle proof shared by acquisition,
// cleanup and read-only observation. Retained evidence uses opaque identities.
package leaseproof

const (
	LabelManaged          = "cartulary.test-services.managed"
	LabelSuiteID          = "cartulary.test-services.suite-id"
	LabelRunID            = "cartulary.test-services.run-id"
	LabelTarget           = "cartulary.test-services.target"
	LabelService          = "cartulary.test-services.service"
	LabelSessionID        = "cartulary.test-services.session-id"
	LabelSessionExpiresAt = "cartulary.test-services.session-expires-at"
	ManagedValue          = "true"
)

type PostgresProvisioning struct {
	PostgresIsolation string `json:"postgres_isolation"`
	AllocationRef     string `json:"allocation_ref"`
	SchemaID          string `json:"schema_id"`
	Generation        string `json:"generation"`
	PolicyDigest      string `json:"policy_digest"`
	TemplateDigest    string `json:"template_digest"`
}

type Lease struct {
	PostgresProvisioning *PostgresProvisioning `json:"postgres_provisioning"`
	SchemaID             string                `json:"schema_id"`
	LeaseID              string                `json:"lease_id"`
	SuiteID              string                `json:"suite_id"`
	RunID                string                `json:"run_id"`
	ResultRoot           string                `json:"result_root"`
	RunRoot              string                `json:"run_root"`
	Target               string                `json:"target"`
	Mode                 string                `json:"mode"`
	OwnershipMode        string                `json:"ownership_mode"`
	OwnerPID             int                   `json:"owner_pid"`
	CreatedAt            string                `json:"created_at"`
	ExpiresAt            string                `json:"expires_at,omitempty"`
	Resources            []LeaseResource       `json:"resources"`
	ProofLabels          map[string]string     `json:"proof_labels"`
	ProofPrefixes        map[string]string     `json:"proof_prefixes"`
	CleanupState         string                `json:"cleanup_state"`
}

type LeaseResource struct {
	Kind        string            `json:"kind"`
	Service     string            `json:"service"`
	ContainerID string            `json:"container_id"`
	Name        string            `json:"name,omitempty"`
	Image       string            `json:"image"`
	Labels      map[string]string `json:"labels"`
}

func RequiredLeaseLabels(lease Lease, service LeaseResource) map[string]string {
	labels := map[string]string{
		LabelManaged: ManagedValue,
		LabelSuiteID: lease.SuiteID,
		LabelRunID:   lease.RunID,
		LabelService: service.Service,
	}
	for key, value := range service.Labels {
		if _, required := labels[key]; required {
			labels[key] = value
		}
	}
	return labels
}
