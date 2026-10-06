// cryptographicpolicy is a build/admission diagnostic, not an application facade.
package main

import (
	"crypto/fips140"
	"encoding/json"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
)

func main() {
	if err := cryptography.AdmitExecution(); err != nil {
		os.Exit(2)
	}
	info, _ := debug.ReadBuildInfo()
	selector := ""
	for _, s := range info.Settings {
		if s.Key == "GOFIPS140" {
			selector = s.Value
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Toolchain string `json:"toolchain"`
		Module    string `json:"module"`
		Selector  string `json:"selector"`
		Enabled   bool   `json:"enabled"`
	}{runtime.Version(), fips140.Version(), selector, fips140.Enabled()}); err != nil {
		os.Exit(1)
	}
}
