package server

import (
	"context"
	"os"
	"strings"

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
	if options.Env == nil {
		options.Env = snapshotProcessEnvironment(os.Environ())
	}
	return (runtimeAssembly{
		loadedConfiguration: loaded,
		options:             options,
		dependencies:        dependencies,
	}).build(ctx)
}

func snapshotProcessEnvironment(entries []string) map[string]string {
	env := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || key == "" {
			continue
		}
		env[key] = value
	}
	return env
}
