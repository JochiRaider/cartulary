package telemetry

import (
	"errors"
	"runtime/debug"
)

var errExporterIdentity = errors.New("telemetry exporter dependency identity rejected")

func exporterDependencyVersion(kind, signal string) (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", errExporterIdentity
	}
	return exporterVersionFromBuild(info, kind, signal)
}

func exporterVersionFromBuild(info *debug.BuildInfo, kind, signal string) (string, error) {
	if info == nil {
		return "", errExporterIdentity
	}
	protocol := ""
	switch kind {
	case "otlp_http":
		protocol = "http"
	case "otlp_grpc":
		protocol = "grpc"
	default:
		return "", errExporterIdentity
	}
	family := ""
	switch signal {
	case "traces":
		family = "otlptrace"
	case "metrics":
		family = "otlpmetric"
	case "logs":
		family = "otlplog"
	default:
		return "", errExporterIdentity
	}
	path := "go.opentelemetry.io/otel/exporters/otlp/" + family + "/" + family + protocol
	version := ""
	for _, dependency := range info.Deps {
		if dependency.Path != path {
			continue
		}
		if version != "" || dependency.Replace != nil || !safeExporterVersion(dependency.Version) {
			return "", errExporterIdentity
		}
		version = dependency.Version
	}
	if version == "" {
		return "", errExporterIdentity
	}
	return version, nil
}
