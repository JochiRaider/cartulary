package server

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/testutil/browserbuild"
)

func newRuntimeWithTestDependencies(t testing.TB,
	ctx context.Context,
	loaded configassembly.Loaded,
	options Options,
	dependencies runtimeDependencies,
) (*Runtime, error) {
	if options.BrowserBuild == nil {
		options.BrowserBuild = browserbuild.New(t)
	}
	return newRuntimeWithDependencies(ctx, loaded, options, dependencies)
}
