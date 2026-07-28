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

func collaborationSettings(cfg configassembly.Deployment, hub *collaboration.Hub) collaboration.Settings {
	return collaboration.Settings{
		AcceptSocket:       acceptCollaborationSocket(cfg.Application.PublicOrigin),
		CheckBrowserOrigin: checkCollaborationBrowserOrigin(cfg.Application.PublicOrigin),
		Hub:                hub,
		ServiceVersion:     cfg.Telemetry.Resource.ServiceVersion,
	}
}

func evidenceSettings(cfg configassembly.Deployment) evidence.Settings {
	return evidence.Settings{
		MaxBlobBytes:   cfg.Limits.ObjectBlobs.MaxDeclaredByteSize,
		PreviewMax:     cfg.Limits.Previews.MaxPreviewablePayloadBytes,
		TextPreviewMax: cfg.Limits.Previews.MaxTextInlineBytes,
	}
}

func importLimits(cfg configassembly.Deployment) (imports.Limits, imports.ArchiveLimits) {
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

func incidentBundleLimits(cfg configassembly.Deployment) incidentbundles.Limits {
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

func referenceDataLimits(cfg configassembly.Deployment) reference_data.Limits {
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

func testResetBootstrap(cfg configassembly.Deployment) func(context.Context, pgx.Tx) error {
	return func(ctx context.Context, tx pgx.Tx) error {
		return bootstrap.PreflightTx(ctx, configassembly.BootstrapSettings(cfg), tx)
	}
}

func runtimeSettings(cfg configassembly.Deployment) RuntimeSettings {
	return RuntimeSettings{
		TelemetryFlushTimeoutMS:  cfg.Telemetry.Shutdown.FlushTimeoutMS,
		ReconciliationSeconds:    cfg.Timeouts.Extensions.ReconciliationSeconds,
		StagedObjectSweepSeconds: cfg.Intervals.Extensions.StagedObjectSweepSeconds,
		ShutdownDrainSeconds:     cfg.Timeouts.Extensions.ShutdownDrainSeconds,
		ProcessModel:             cfg.Application.ProcessModel,
	}
}
