package producer

import (
	"fmt"
	inputs "github.com/JochiRaider/cartulary/referencepacks/attack/inputs"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"sync"
)

var auditReportSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	raw, err := inputs.Files.ReadFile("audit-report.schema.json")
	if err != nil {
		return nil, err
	}
	return compileProducerSchema(raw, "")
})

// Producer evidence is closed before it can be delivered as successful output.
func validateAuditReport(report any) error {
	schema, err := auditReportSchema()
	if err != nil {
		return err
	}
	raw, err := canonicaljson.Marshal(report)
	if err != nil {
		return err
	}
	value, err := canonicaljson.DecodeStrict(raw)
	if err != nil {
		return err
	}
	if err := schema.Validate(value); err != nil {
		return fmt.Errorf("audit report contract: %w", err)
	}
	return nil
}
