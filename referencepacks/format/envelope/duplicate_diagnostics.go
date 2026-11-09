package envelope

import "context"

func duplicateManifestFindings(ctx context.Context, data []byte, emit FindingSink) error {
	return manifestShape.DuplicateFindings(ctx, data, emit)
}
