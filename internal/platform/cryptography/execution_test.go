package cryptography

import (
	"crypto/fips140"
	"errors"
	"runtime"
	"runtime/debug"
	"testing"
)

func TestExecutionIdentityMatrix(t *testing.T) {
	valid := func() *debug.BuildInfo {
		return &debug.BuildInfo{GoVersion: Toolchain, Settings: []debug.BuildSetting{{Key: "GOFIPS140", Value: ModuleSelector}}}
	}
	for _, tc := range []struct {
		name             string
		mutate           func(*debug.BuildInfo)
		compiler, module string
		enabled          bool
		valid            bool
	}{
		{"current", nil, Toolchain, ModuleVersion, true, true},
		{"disabled", nil, Toolchain, ModuleVersion, false, false},
		{"wrong compiler", nil, "go1.27.0", ModuleVersion, true, false},
		{"wrong module", nil, Toolchain, "inprocess", true, false},
		{"selector is not service version", nil, Toolchain, ModuleSelector, true, false},
		{"missing selector", func(i *debug.BuildInfo) { i.Settings = nil }, Toolchain, ModuleVersion, true, false},
		{"floating selector", func(i *debug.BuildInfo) { i.Settings[0].Value = "certified" }, Toolchain, ModuleVersion, true, false},
		{"wrong selector", func(i *debug.BuildInfo) { i.Settings[0].Value = "v1.0.0" }, Toolchain, ModuleVersion, true, false},
		{"contradictory compiler", func(i *debug.BuildInfo) { i.GoVersion = "go1.27.0" }, Toolchain, ModuleVersion, true, false},
		{"duplicate metadata", func(i *debug.BuildInfo) { i.Settings = append(i.Settings, i.Settings[0]) }, Toolchain, ModuleVersion, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := valid()
			if tc.mutate != nil {
				tc.mutate(i)
			}
			err := admitExecution(i, tc.compiler, tc.module, tc.enabled)
			if (err == nil) != tc.valid {
				t.Fatalf("admission = %v", err)
			}
			if err != nil && !errors.Is(err, ErrExecutionPolicy) {
				t.Fatal(err)
			}
		})
	}
	if admitExecution(nil, Toolchain, ModuleVersion, true) == nil {
		t.Fatal("missing build information admitted")
	}
}

func TestExecutionActualIdentity(t *testing.T) {
	info, ok := debug.ReadBuildInfo()
	expected := ok && runtime.Version() == Toolchain && info.GoVersion == Toolchain && fips140.Enabled() && fips140.Version() == ModuleVersion
	count := 0
	for _, s := range info.Settings {
		if s.Key == "GOFIPS140" {
			count++
			expected = expected && s.Value == ModuleSelector
		}
	}
	expected = expected && count == 1
	if err := AdmitExecution(); (err == nil) != expected {
		t.Fatalf("actual executable admission = %v, expected admitted=%v", err, expected)
	}
}
