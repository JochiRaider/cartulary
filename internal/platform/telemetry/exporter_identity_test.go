package telemetry

import (
	"runtime/debug"
	"testing"
)

func TestExporterDependencyIdentity(t *testing.T) {
	const path = "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	for _, tc := range []struct {
		name  string
		info  *debug.BuildInfo
		valid bool
	}{
		{"actual identity", &debug.BuildInfo{Deps: []*debug.Module{{Path: path, Version: "v1.45.0"}}}, true},
		{"missing metadata", nil, false},
		{"missing exporter", &debug.BuildInfo{}, false},
		{"missing version", &debug.BuildInfo{Deps: []*debug.Module{{Path: path}}}, false},
		{"local version", &debug.BuildInfo{Deps: []*debug.Module{{Path: path, Version: "(devel)"}}}, false},
		{"replacement", &debug.BuildInfo{Deps: []*debug.Module{{Path: path, Version: "v1.45.0", Replace: &debug.Module{Path: "local"}}}}, false},
		{"duplicate", &debug.BuildInfo{Deps: []*debug.Module{{Path: path, Version: "v1.45.0"}, {Path: path, Version: "v1.45.0"}}}, false},
		{"contradictory", &debug.BuildInfo{Deps: []*debug.Module{{Path: path, Version: "v1.45.0"}, {Path: path, Version: "v1.43.0"}}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			version, err := exporterVersionFromBuild(tc.info, "otlp_http", "metrics")
			if (err == nil) != tc.valid {
				t.Fatalf("identity admission: version=%q error=%v", version, err)
			}
			if tc.valid && version != "v1.45.0" {
				t.Fatal("wrong admitted version")
			}
		})
	}
	for _, kind := range []string{"otlp_http", "otlp_grpc"} {
		for _, signal := range []string{"traces", "metrics", "logs"} {
			want := "v1.45.0"
			if signal == "logs" {
				want = "v0.21.0"
			}
			got, err := exporterDependencyVersion(kind, signal)
			if err != nil || got != want {
				t.Fatalf("actual %s/%s dependency=%q error=%v", kind, signal, got, err)
			}
		}
	}
	if _, err := exporterVersionFromBuild(&debug.BuildInfo{}, "legacy", "metrics"); err == nil {
		t.Fatal("unsupported protocol accepted")
	}
	if _, err := exporterVersionFromBuild(&debug.BuildInfo{}, "otlp_http", "unknown"); err == nil {
		t.Fatal("unsupported signal accepted")
	}
}
