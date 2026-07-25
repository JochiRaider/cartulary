package server

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	telemetryconfiguration "github.com/JochiRaider/cartulary/internal/platform/telemetry/configuration"
)

func TestModuleSettingsProjection_Unit(t *testing.T) {
	cfg := configassembly.Deployment{
		Application: config.ApplicationConfig{PublicOrigin: "https://cartulary.example"},
		Telemetry: telemetryconfiguration.Config{
			Resource: telemetryconfiguration.ResourceConfig{ServiceVersion: "2026.7.25"},
			Shutdown: telemetryconfiguration.ShutdownConfig{FlushTimeoutMS: 71},
		},
		Timeouts: config.TimeoutConfig{Extensions: config.ExtensionTimeoutConfig{
			ReconciliationSeconds: 72,
			ShutdownDrainSeconds:  73,
		}},
		Intervals: config.IntervalConfig{Extensions: config.ExtensionIntervalConfig{StagedObjectSweepSeconds: 74}},
		Limits: config.LimitConfig{
			ObjectBlobs: config.ObjectBlobLimits{MaxDeclaredByteSize: 11},
			Imports: config.ImportLimits{
				MaxCSVSourceBytes: 21, MaxXLSXSourceBytes: 22, MaxRows: 23, MaxColumns: 24, MaxCells: 25,
			},
			Archives:        config.ArchiveLimits{DefaultMaxExtractedBytes: 31, MaxCompressionRatio: 32, MaxMembers: 33},
			ReferencePacks:  config.ReferencePackLimits{MaxExtractedBytes: 41},
			IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 51},
			Previews:        config.PreviewLimits{MaxPreviewablePayloadBytes: 61, MaxTextInlineBytes: 62},
		},
	}

	if got, want := collaborationSettings(cfg), (collaboration.Settings{
		PublicOrigin: "https://cartulary.example", ServiceVersion: "2026.7.25",
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("collaboration settings = %#v, want %#v", got, want)
	}
	if got, want := evidenceSettings(cfg), (evidence.Settings{
		MaxBlobBytes: 11, PreviewMax: 61, TextPreviewMax: 62,
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("evidence settings = %#v, want %#v", got, want)
	}
	importOwner, importArchive := importLimits(cfg)
	if want := (imports.Limits{
		MaxCSVSourceBytes: 21, MaxXLSXSourceBytes: 22, MaxRows: 23, MaxColumns: 24, MaxCells: 25,
	}); !reflect.DeepEqual(importOwner, want) {
		t.Fatalf("import limits = %#v, want %#v", importOwner, want)
	}
	if want := (imports.ArchiveLimits{
		DefaultMaxExtractedBytes: 31, MaxCompressionRatio: 32, MaxMembers: 33,
	}); !reflect.DeepEqual(importArchive, want) {
		t.Fatalf("import archive limits = %#v, want %#v", importArchive, want)
	}
	if got, want := incidentBundleLimits(cfg), (incidentbundles.Limits{
		Archives: incidentbundles.ArchiveLimits{
			DefaultMaxExtractedBytes: 31, MaxCompressionRatio: 32, MaxMembers: 33,
		},
		IncidentBundles: incidentbundles.IncidentBundleLimits{MaxExtractedBytes: 51},
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("incident-bundle limits = %#v, want %#v", got, want)
	}
	if got, want := referenceDataLimits(cfg), (reference_data.Limits{
		Archives: reference_data.ArchiveLimits{
			DefaultMaxExtractedBytes: 31, MaxCompressionRatio: 32, MaxMembers: 33,
		},
		ReferencePacks: reference_data.ReferenceLimits{MaxExtractedBytes: 41},
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("reference-data limits = %#v, want %#v", got, want)
	}
	if got, want := runtimeSettings(cfg), (RuntimeSettings{
		TelemetryFlushTimeoutMS:  71,
		ReconciliationSeconds:    72,
		ShutdownDrainSeconds:     73,
		StagedObjectSweepSeconds: 74,
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("runtime settings = %#v, want %#v", got, want)
	}
}
