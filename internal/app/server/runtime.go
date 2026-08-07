package server

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
)

// NewRuntime is the repository-internal construction facade. The private
// runtimeAssembly owns ordered composition and cleanup; Runtime exposes only
// the lifecycle capabilities needed by its repository callers.
func NewRuntime(ctx context.Context, loaded configassembly.Loaded, options Options) (*Runtime, error) {
	return newRuntime(ctx, loaded, options)
}

func newRuntime(ctx context.Context, loaded configassembly.Loaded, options Options) (*Runtime, error) {
	return newRuntimeWithDependencies(ctx, loaded, options, productionRuntimeDependencies())
}

func newRuntimeWithDependencies(
	ctx context.Context,
	loaded configassembly.Loaded,
	options Options,
	dependencies runtimeDependencies,
) (*Runtime, error) {
	return (runtimeAssembly{
		loadedConfiguration: loaded,
		options:             options,
		dependencies:        dependencies,
	}).build(ctx)
}
