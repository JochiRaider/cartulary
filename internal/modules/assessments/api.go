package assessments

const AssessmentsViewSchemaID = "cartulary.view.assessments.v1"

type CreateValidationError struct {
	Field      string
	ReasonCode string
}

func (e *CreateValidationError) Error() string {
	return "assessments: invalid create request"
}
