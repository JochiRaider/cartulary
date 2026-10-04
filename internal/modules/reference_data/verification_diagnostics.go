package reference_data

import (
	"context"
	"errors"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

// Retained inventory has at most 67 objects, so its complete summary can be
// combined with the streamed container findings without losing discarded
// findings. Other summaries must never be re-emitted after truncation.
func emitRetainedFindings(err error, check string, emit packformat.FindingSink) error {
	if err == nil {
		return nil
	}
	var rejected *ContentRejection
	if !errors.As(err, &rejected) {
		return err
	}
	if rejected.CheckID != check {
		return errors.New("reference pack: unexpected retained inventory check")
	}
	if rejected.Summary == nil {
		return emit(packformat.Finding{Path: "$"})
	}
	s := rejected.Summary
	if s.Truncated || s.Total < 1 || s.Total != len(s.Issues) || s.Total > 67 {
		return errors.New("reference pack: incomplete retained inventory diagnostics")
	}
	for _, issue := range s.Issues {
		if issue.CheckID != check {
			return errors.New("reference pack: inconsistent retained inventory check")
		}
		if err := emit(packformat.Finding{Path: issue.Path, EntryID: issue.EntryID, Details: issue.Details}); err != nil {
			return err
		}
	}
	return nil
}

type verificationDiagnosticScratch struct {
	storage   VerificationStorage
	workspace VerificationWorkspace
}

func (s *verificationDiagnosticScratch) Write(ctx context.Context, name string, write func(io.Writer) error) error {
	if s.workspace == nil {
		var err error
		s.workspace, err = s.storage.NewWorkspace(ctx)
		if err != nil {
			return err
		}
	}
	return s.workspace.Write(ctx, name, write)
}
func (s *verificationDiagnosticScratch) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	if s.workspace == nil {
		return nil, errors.New("reference pack: diagnostic scratch unavailable")
	}
	return s.workspace.Open(ctx, name)
}
func (s *verificationDiagnosticScratch) Remove(ctx context.Context, name string) error {
	if s.workspace == nil {
		return errors.New("reference pack: diagnostic scratch unavailable")
	}
	return s.workspace.Remove(ctx, name)
}
func (s *verificationDiagnosticScratch) close() error {
	if s.workspace == nil {
		return nil
	}
	return s.workspace.Close()
}
