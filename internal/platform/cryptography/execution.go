// Package cryptography owns cryptographic mechanics and executable admission.
// Payload schemas, authorization, key rotation and storage lifecycles belong to callers.
package cryptography

import (
	"crypto/fips140"
	"errors"
	"runtime"
	"runtime/debug"
)

const (
	Toolchain           = "go1.27.1"
	ModuleVersion       = "v1.0.0"
	ModuleSelector      = "v1.0.0-c2097c7c"
	ModuleArchiveSHA256 = "daf3614e0406f67ae6323c902db3f953a1effb199142362a039e7526dfb9368b"
)

// ErrExecutionPolicy is deliberately independent of environment values and secrets.
var ErrExecutionPolicy = errors.New("cryptographic execution policy rejected")

// AdmitExecution checks the running executable, never caller-supplied metadata.
func AdmitExecution() error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ErrExecutionPolicy
	}
	return admitExecution(info, runtime.Version(), fips140.Version(), fips140.Enabled())
}

func admitExecution(info *debug.BuildInfo, toolchain, module string, enabled bool) error {
	if info == nil || !enabled || toolchain != Toolchain || info.GoVersion != Toolchain || module != ModuleVersion {
		return ErrExecutionPolicy
	}
	count := 0
	for _, setting := range info.Settings {
		if setting.Key == "GOFIPS140" {
			count++
			if setting.Value != ModuleSelector {
				return ErrExecutionPolicy
			}
		}
	}
	if count != 1 {
		return ErrExecutionPolicy
	}
	return nil
}
