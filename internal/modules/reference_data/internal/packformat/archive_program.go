package packformat

import (
	"context"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"io"
)

func Extract(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits, destination Destination) (Inventory, error) {
	return extract(ctx, source, size, limits, destination, nil, false)
}
func extract(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits, destination Destination, scratch DiagnosticScratch, diagnostics bool) (Inventory, error) {
	program, extract, err := ArchiveChecks(ctx, source, size, limits)
	if err != nil {
		return nil, err
	}
	if err := runCheckRange(ctx, "container_format", "archive_member_type", scratch, program, diagnostics); err != nil {
		return nil, err
	}
	return extract(destination)
}
func ArchiveChecks(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits) (map[string]CheckFunc, func(Destination) (Inventory, error), error) {
	checks, extract, err := envelope.ArchiveChecks(ctx, source, size, limits)
	if err != nil {
		return nil, nil, formatError(err)
	}
	return formatChecks(checks), func(destination Destination) (Inventory, error) {
		result, err := extract(destination)
		return result, formatError(err)
	}, nil
}
