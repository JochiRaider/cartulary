package server

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
)

func newRuntimeWithTestDependencies(
	ctx context.Context,
	deployment configassembly.Deployment,
	options Options,
	dependencies runtimeDependencies,
) (*Runtime, error) {
	loaded, err := configassembly.Admit(deployment)
	if err != nil {
		return nil, err
	}
	return newRuntimeWithDependencies(ctx, loaded, options, dependencies)
}
