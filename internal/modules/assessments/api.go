package assessments

const (
	AssessmentsViewSchemaID = "cartulary.view.assessments.v1"
	maxSupportActions       = 64
)

type CreateValidationError struct {
	Field      string
	ReasonCode string
}

func (e *CreateValidationError) Error() string {
	return "assessments: invalid create request"
}
