package workbookroutetest

import (
	"context"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/workbook"
)

// CreateWorkbookRow passes the caller's bytes to the registered production decoder.
// The catalog, command and payload are explicit; admission and outcomes remain owner-owned.
func CreateWorkbookRow(catalog *workbook.WorkbookContributionCatalog, ctx context.Context, command workbook.CreateCommand, payload io.Reader) (workbook.MutationOutcome, error) {
	if catalog == nil {
		return workbook.MutationOutcome{}, fmt.Errorf("workbook contribution catalog is required")
	}
	provider, ok := catalog.CreateFor(command.ViewSchemaID)
	if !ok {
		return workbook.MutationOutcome{}, fmt.Errorf("create provider missing for %q", command.ViewSchemaID)
	}
	operation, failure, err := provider.DecodeCreate(payload)
	if err != nil {
		return workbook.MutationOutcome{}, err
	}
	if failure != nil {
		return workbook.RejectedMutation(failure), nil
	}
	return operation.Execute(ctx, command)
}

// PatchWorkbookRow uses the authoritative record type supplied by the caller.
func PatchWorkbookRow(catalog *workbook.WorkbookContributionCatalog, ctx context.Context, command workbook.PatchCommand, payload io.Reader) (workbook.MutationOutcome, error) {
	if catalog == nil {
		return workbook.MutationOutcome{}, fmt.Errorf("workbook contribution catalog is required")
	}
	provider, ok := catalog.PatchFor(command.AuthoritativeRecordType)
	if !ok {
		return workbook.MutationOutcome{}, fmt.Errorf("patch provider missing for %q", command.AuthoritativeRecordType)
	}
	operation, failure, err := provider.DecodePatch(payload)
	if err != nil {
		return workbook.MutationOutcome{}, err
	}
	if failure != nil {
		return workbook.RejectedMutation(failure), nil
	}
	return operation.Execute(ctx, command)
}
