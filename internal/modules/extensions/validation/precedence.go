// Package validation owns the result precedence of the Extensions protocols.
// Callers retain their typed callbacks and contract-specific schema checks.
package validation

const FindingLimit = 4096

type Disposition uint8

const (
	Success Disposition = iota
	InvocationFailed
	StructureInvalid
	Overflow
	SchemaInvalid
	Findings
)

// Classify deliberately delays all element and remaining-schema checks until
// after the findings count. Typed adapters establish structure; byte adapters
// must first strictly decode their actual byte boundary.
func Classify(invocationErr error, structureValid bool, count, schemaLimit int, schemaValid func() bool) Disposition {
	if invocationErr != nil {
		return InvocationFailed
	}
	if !structureValid {
		return StructureInvalid
	}
	if count > FindingLimit {
		return Overflow
	}
	if count < 0 || count > schemaLimit || schemaValid == nil || !schemaValid() {
		return SchemaInvalid
	}
	if count != 0 {
		return Findings
	}
	return Success
}
