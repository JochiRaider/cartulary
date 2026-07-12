//go:build !cartulary_harness

package app

import (
	"context"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
)

type productionServerProfile struct{}

func newServerProfile() serverProfile {
	return productionServerProfile{}
}

func (productionServerProfile) validateEnvironment(lookup func(string) (string, bool)) error {
	for _, key := range harnessOnlyServerEnv {
		if _, configured := lookup(key); configured {
			return config.NewDiagnosticsError(config.Diagnostic{
				Path:       key,
				ReasonCode: "harness_profile_required",
				Message:    "requires the harness server build profile",
			})
		}
	}
	return nil
}

func (productionServerProfile) runtimeOptions(func(string) (string, bool)) Options {
	return Options{}
}

func (productionServerProfile) inheritedListenerFD(func(string) (string, bool)) string {
	return ""
}

func (productionServerProfile) serve(ctx context.Context, handler http.Handler, options httpruntime.Options) error {
	return httpruntime.Serve(ctx, handler, options)
}
