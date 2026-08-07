package server

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
)

func newRuntimeWithTestDependencies(
	ctx context.Context,
	loaded configassembly.Loaded,
	options Options,
	dependencies runtimeDependencies,
) (*Runtime, error) {
	return newRuntimeWithDependencies(ctx, loaded, options, dependencies)
}
