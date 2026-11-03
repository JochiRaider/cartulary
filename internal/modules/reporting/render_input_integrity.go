package reporting

// Only readers that have distinguished storage/transport failure from rejected
// retained data may construct this classification. Its text carries no content.
type retainedRenderInputError struct{ missing bool }

func (e *retainedRenderInputError) Error() string {
	return "reporting: retained render input failed integrity validation"
}
func retainedRenderInputMismatch() error { return &retainedRenderInputError{} }
func retainedRenderInputMissing() error  { return &retainedRenderInputError{missing: true} }
