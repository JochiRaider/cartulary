package harnesslocal

// ReadRejection exposes only closed diagnostic codes. The underlying error is
// retained for errors.Is (notably optional ENOENT), never rendered as evidence.
type ReadRejection struct {
	stage, reason string
	cause         error
}

func (e *ReadRejection) Error() string  { return "private harness read rejected" }
func (e *ReadRejection) Unwrap() error  { return e.cause }
func (e *ReadRejection) Stage() string  { return e.stage }
func (e *ReadRejection) Reason() string { return e.reason }
func rejectRead(stage, reason string, cause error) error {
	if cause == nil {
		cause = ErrUnsafe
	}
	return &ReadRejection{stage: stage, reason: reason, cause: cause}
}
