package server

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
)

type applicationSettingsProjection struct {
	deployment configassembly.Deployment
}

func newApplicationSettingsProjection(cfg configassembly.Deployment) applicationSettingsProjection {
	return applicationSettingsProjection{deployment: cfg}
}

func (projection applicationSettingsProjection) Collaboration(
	hub *collaboration.Hub,
	transport collaborationSocketTransport,
) collaboration.Settings {
	cfg := projection.deployment
	return collaboration.Settings{
		AcceptSocket:       transport.Accept,
		CheckBrowserOrigin: transport.CheckBrowserOrigin,
		Hub:                hub,
		ServiceVersion:     cfg.Telemetry.Resource.ServiceVersion,
	}
}

func (projection applicationSettingsProjection) Evidence() evidence.Settings {
	cfg := projection.deployment
	return evidence.Settings{
		MaxBlobBytes:   cfg.Limits.ObjectBlobs.MaxDeclaredByteSize,
		PreviewMax:     cfg.Limits.Previews.MaxPreviewablePayloadBytes,
		TextPreviewMax: cfg.Limits.Previews.MaxTextInlineBytes,
	}
}

func (projection applicationSettingsProjection) Imports() (imports.Limits, imports.ArchiveLimits) {
	cfg := projection.deployment
	return imports.Limits{
			MaxCSVSourceBytes:  cfg.Limits.Imports.MaxCSVSourceBytes,
			MaxXLSXSourceBytes: cfg.Limits.Imports.MaxXLSXSourceBytes,
			MaxRows:            cfg.Limits.Imports.MaxRows,
			MaxColumns:         cfg.Limits.Imports.MaxColumns,
			MaxCells:           cfg.Limits.Imports.MaxCells,
		}, imports.ArchiveLimits{
			DefaultMaxExtractedBytes: cfg.Limits.Archives.DefaultMaxExtractedBytes,
			MaxCompressionRatio:      cfg.Limits.Archives.MaxCompressionRatio,
			MaxMembers:               cfg.Limits.Archives.MaxMembers,
		}
}

func (projection applicationSettingsProjection) IncidentBundles() incidentbundles.Limits {
	cfg := projection.deployment
	return incidentbundles.Limits{
		Archives: incidentbundles.ArchiveLimits{
			DefaultMaxExtractedBytes: cfg.Limits.Archives.DefaultMaxExtractedBytes,
			MaxCompressionRatio:      cfg.Limits.Archives.MaxCompressionRatio,
			MaxMembers:               cfg.Limits.Archives.MaxMembers,
		},
		IncidentBundles: incidentbundles.IncidentBundleLimits{
			MaxExtractedBytes: cfg.Limits.IncidentBundles.MaxExtractedBytes,
		},
	}
}

func (projection applicationSettingsProjection) ReferenceData() reference_data.Limits {
	cfg := projection.deployment
	return reference_data.Limits{
		Archives: reference_data.ArchiveLimits{
			DefaultMaxExtractedBytes: cfg.Limits.Archives.DefaultMaxExtractedBytes,
			MaxCompressionRatio:      cfg.Limits.Archives.MaxCompressionRatio,
			MaxMembers:               cfg.Limits.Archives.MaxMembers,
		},
		ReferencePacks: reference_data.ReferenceLimits{
			MaxExtractedBytes: cfg.Limits.ReferencePacks.MaxExtractedBytes,
		},
	}
}

func (projection applicationSettingsProjection) TestResetBootstrap() func(context.Context, pgx.Tx) error {
	cfg := projection.deployment
	return func(ctx context.Context, tx pgx.Tx) error {
		return bootstrap.PreflightTx(ctx, configassembly.BootstrapSettings(cfg), tx)
	}
}

func (projection applicationSettingsProjection) Runtime() runtimeSettings {
	cfg := projection.deployment
	return runtimeSettings{
		TelemetryFlushTimeoutMS:  cfg.Telemetry.Shutdown.FlushTimeoutMS,
		ReconciliationSeconds:    cfg.Timeouts.Extensions.ReconciliationSeconds,
		StagedObjectSweepSeconds: cfg.Intervals.Extensions.StagedObjectSweepSeconds,
		ShutdownDrainSeconds:     cfg.Timeouts.Extensions.ShutdownDrainSeconds,
	}
}
