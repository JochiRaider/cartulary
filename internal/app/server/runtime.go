package server

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
)

// NewRuntime is the repository-internal construction facade. The private
// runtimeAssembly owns ordered composition and cleanup; callers receive the
// existing Runtime surface until the dedicated caller-migration slice.
func NewRuntime(ctx context.Context, deployment configassembly.Deployment, options Options) (*Runtime, error) {
	loaded, err := configassembly.Admit(deployment)
	if err != nil {
		return nil, err
	}
	return newRuntime(ctx, loaded, options)
}

func newRuntime(ctx context.Context, loaded configassembly.Loaded, options Options) (*Runtime, error) {
	return (runtimeAssembly{loadedConfiguration: loaded, options: options}).build(ctx)
}
