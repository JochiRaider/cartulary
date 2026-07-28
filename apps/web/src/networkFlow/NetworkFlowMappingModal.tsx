import {
  networkAnalysisMappingColumnTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import type { ExtensionImportDiscovery } from "../imports/importCoordinator";
import type { NetworkFlowImportPreviewResult } from "../services/networkFlowContractAdapter";
import { networkFlowMappingMetadata } from "../services/networkFlowContractAdapter";
import {
  ignoredColumnChoice,
  mappedRequiredFieldCount,
  type NetworkFlowMappingDraft,
  networkFlowMappingDraftReadyForPreview,
  networkFlowMappingFields,
  networkFlowRequiredFieldKeys,
  sourceColumnLabel,
  withNetworkFlowColumnChoice,
} from "./networkFlowImportModel";
import { localizedNetworkFlowDiagnosticMessage } from "./networkFlowPresentation";
import type { NetworkFlowImportStage } from "./useNetworkFlowImportController";
import { useNetworkFlowModalFocus } from "./useNetworkFlowModalFocus";

const unmappedColumnChoice = "__unmapped__";

export function NetworkFlowMappingModal({
  canApply,
  discovery,
  draft,
  onApply,
  onCancel,
  onDraftChange,
  onPreview,
  preview,
  stage,
}: {
  readonly canApply: boolean;
  readonly discovery: ExtensionImportDiscovery;
  readonly draft: NetworkFlowMappingDraft;
  readonly onApply: () => void;
  readonly onCancel: () => void;
  readonly onDraftChange: (draft: NetworkFlowMappingDraft) => void;
  readonly onPreview: () => void;
  readonly preview: NetworkFlowImportPreviewResult | null;
  readonly stage: NetworkFlowImportStage;
}) {
  const requiredCount = mappedRequiredFieldCount(draft);
  const sourceProfile = networkFlowMappingMetadata.source_profiles.find(
    (candidate) => candidate.source_profile_id === draft.sourceProfileId,
  );
  const working = stage === "previewing" || stage === "applying";
  const timestampReady = networkFlowMappingDraftReadyForPreview(draft);
  const modalFocus = useNetworkFlowModalFocus<HTMLElement>({
    dismissDisabled: working,
    initialFocusTestId: networkAnalysisTestId("mapping-profile"),
    onDismiss: onCancel,
  });

  return (
    <div style={backdropStyle}>
      <section
        ref={modalFocus.dialogRef}
        aria-describedby="network-flow-mapping-description"
        aria-labelledby="network-flow-mapping-title"
        aria-modal="true"
        data-testid={networkAnalysisTestId("mapping-dialog")}
        role="dialog"
        style={dialogStyle}
        onKeyDown={modalFocus.onKeyDown}
      >
        <header style={headerStyle}>
          <div>
            <h2 id="network-flow-mapping-title" style={titleStyle}>
              Review Network Flow mapping
            </h2>
            <p id="network-flow-mapping-description" style={mutedStyle}>
              Suggestions are not approval. Preview the current mapping, then
              explicitly apply the matching fingerprint.
            </p>
          </div>
          <button disabled={working} type="button" onClick={onCancel}>
            Cancel
          </button>
        </header>

        <div style={settingsGridStyle}>
          <label style={fieldStyle}>
            Source profile
            <select
              data-testid={networkAnalysisTestId("mapping-profile")}
              disabled={working}
              value={draft.sourceProfileId}
              onChange={(event) =>
                onDraftChange({
                  ...draft,
                  sourceProfileId: event.currentTarget.value,
                })
              }
            >
              {networkFlowMappingMetadata.source_profiles.map((profile) => (
                <option
                  key={profile.source_profile_id}
                  value={profile.source_profile_id}
                >
                  {profile.display_name}
                </option>
              ))}
            </select>
          </label>
          <label style={fieldStyle}>
            Table display name (optional)
            <input
              data-testid={networkAnalysisTestId("mapping-display-name")}
              disabled={working}
              maxLength={64}
              value={draft.displayNameOverride}
              onChange={(event) =>
                onDraftChange({
                  ...draft,
                  displayNameOverride: event.currentTarget.value,
                })
              }
            />
          </label>
          <label style={fieldStyle}>
            Timestamp interpretation
            <select
              data-testid={networkAnalysisTestId("mapping-timestamp-mode")}
              disabled={working}
              value={draft.timestampMode}
              onChange={(event) =>
                onDraftChange({
                  ...draft,
                  timestampMode: event.currentTarget
                    .value as typeof draft.timestampMode,
                })
              }
            >
              {sourceProfile?.supported_timestamp_modes.map((mode) => (
                <option key={mode} value={mode}>
                  {humanize(mode)}
                </option>
              ))}
            </select>
          </label>
          <label style={fieldStyle}>
            Unknown-column policy
            <select
              data-testid={networkAnalysisTestId("mapping-unknown-policy")}
              disabled={working}
              value={draft.unknownColumnPolicy}
              onChange={(event) =>
                onDraftChange({
                  ...draft,
                  unknownColumnPolicy: event.currentTarget
                    .value as typeof draft.unknownColumnPolicy,
                })
              }
            >
              {sourceProfile?.supported_unknown_column_policies.map(
                (policy) => (
                  <option key={policy} value={policy}>
                    {humanize(policy)}
                  </option>
                ),
              )}
            </select>
          </label>
          {draft.timestampMode === "rfc3339" ? (
            <label style={fieldStyle}>
              Source timezone (blank for offset-bearing timestamps)
              <input
                data-testid={networkAnalysisTestId("mapping-timezone")}
                disabled={working}
                placeholder="UTC"
                value={draft.timezone}
                onChange={(event) =>
                  onDraftChange({
                    ...draft,
                    timezone: event.currentTarget.value,
                  })
                }
              />
            </label>
          ) : null}
        </div>

        {draft.timestampMode === "netflow_sys_uptime_milliseconds" ? (
          <NetFlowUptimeSettings
            discovery={discovery}
            draft={draft}
            disabled={working}
            onDraftChange={onDraftChange}
          />
        ) : null}

        <div style={summaryStyle}>
          <strong>
            {requiredCount} of {networkFlowRequiredFieldKeys.length} required
            fields mapped
          </strong>
          <span>
            {discovery.preview.columns.length} discovered source columns;
            headers are qualified by ordinal.
          </span>
        </div>

        <section aria-label="Source-column mappings" style={mappingListStyle}>
          {draft.unresolvedAliasCollisionOrdinals.length === 0 ? null : (
            <div aria-label="Alias collision" role="alert" style={alertStyle}>
              Duplicate source aliases suggest the same target field. Explicitly
              map or ignore every highlighted column before previewing.
            </div>
          )}
          {discovery.preview.columns.map((column) => {
            const sourceColumn = preview?.source_columns.find(
              (candidate) =>
                candidate.source_column_ordinal ===
                column.source_column_ordinal,
            );
            const choice =
              draft.columnChoices[column.source_column_ordinal] ?? null;
            return (
              <div
                key={column.source_column_ordinal}
                data-alias-collision={
                  draft.unresolvedAliasCollisionOrdinals.includes(
                    column.source_column_ordinal,
                  )
                    ? "unresolved"
                    : undefined
                }
                style={mappingRowStyle}
              >
                <div>
                  <strong>{sourceColumnLabel(column)}</strong>
                  {sourceColumn === undefined ? null : (
                    <div style={sampleStyle}>
                      Safe samples: {safeSamples(sourceColumn.sample_values)} ·
                      empty values: {sourceColumn.detected_empty_count}
                    </div>
                  )}
                </div>
                <label style={fieldStyle}>
                  Target field
                  <select
                    aria-label={`Target for ${sourceColumnLabel(column)}`}
                    data-testid={networkAnalysisMappingColumnTestId(
                      column.source_column_ordinal,
                    )}
                    disabled={working}
                    value={choice ?? unmappedColumnChoice}
                    onChange={(event) =>
                      onDraftChange(
                        withNetworkFlowColumnChoice(
                          draft,
                          column.source_column_ordinal,
                          event.currentTarget.value === unmappedColumnChoice
                            ? null
                            : event.currentTarget.value,
                        ),
                      )
                    }
                  >
                    <option value={unmappedColumnChoice}>Unmapped</option>
                    <option value={ignoredColumnChoice}>
                      Explicitly ignore
                    </option>
                    {networkFlowMappingFields.map((field) => (
                      <option key={field.field_key} value={field.field_key}>
                        {humanize(field.field_key.replace("network_flow.", ""))}
                        {field.requirement === "required" ? " (required)" : ""}
                      </option>
                    ))}
                  </select>
                </label>
              </div>
            );
          })}
        </section>

        {preview === null ? null : <PreviewSummary preview={preview} />}

        {!timestampReady ? (
          <p role="alert" style={alertStyle}>
            Choose distinct export-time and exporter-uptime columns before
            previewing this timestamp mode.
          </p>
        ) : null}

        <footer style={footerStyle}>
          <button disabled={working} type="button" onClick={onCancel}>
            Cancel
          </button>
          <button
            data-testid={networkAnalysisTestId("mapping-preview")}
            disabled={working || !timestampReady}
            type="button"
            onClick={onPreview}
          >
            {stage === "previewing" ? "Previewing…" : "Preview mapping"}
          </button>
          <button
            data-testid={networkAnalysisTestId("mapping-apply")}
            disabled={!canApply || working}
            type="button"
            onClick={onApply}
          >
            {stage === "applying" ? "Applying…" : "Approve and apply"}
          </button>
        </footer>
      </section>
    </div>
  );
}

function NetFlowUptimeSettings({
  disabled,
  discovery,
  draft,
  onDraftChange,
}: {
  readonly disabled: boolean;
  readonly discovery: ExtensionImportDiscovery;
  readonly draft: NetworkFlowMappingDraft;
  readonly onDraftChange: (draft: NetworkFlowMappingDraft) => void;
}) {
  return (
    <div style={settingsGridStyle}>
      <ColumnOrdinalSelect
        disabled={disabled}
        discovery={discovery}
        label="Export-time source column"
        value={draft.netflowExportTimeColumnOrdinal}
        onChange={(ordinal) =>
          onDraftChange({ ...draft, netflowExportTimeColumnOrdinal: ordinal })
        }
      />
      <label style={fieldStyle}>
        Export-time interpretation
        <select
          disabled={disabled}
          value={draft.netflowExportTimeMode}
          onChange={(event) =>
            onDraftChange({
              ...draft,
              netflowExportTimeMode: event.currentTarget
                .value as typeof draft.netflowExportTimeMode,
            })
          }
        >
          <option value="rfc3339">RFC 3339</option>
          <option value="epoch_seconds">Epoch seconds</option>
          <option value="epoch_milliseconds">Epoch milliseconds</option>
        </select>
      </label>
      <ColumnOrdinalSelect
        disabled={disabled}
        discovery={discovery}
        label="Exporter uptime-at-export source column"
        value={draft.netflowExporterUptimeColumnOrdinal}
        onChange={(ordinal) =>
          onDraftChange({
            ...draft,
            netflowExporterUptimeColumnOrdinal: ordinal,
          })
        }
      />
    </div>
  );
}

function ColumnOrdinalSelect({
  disabled,
  discovery,
  label,
  onChange,
  value,
}: {
  readonly disabled: boolean;
  readonly discovery: ExtensionImportDiscovery;
  readonly label: string;
  readonly onChange: (ordinal: number | null) => void;
  readonly value: number | null;
}) {
  return (
    <label style={fieldStyle}>
      {label}
      <select
        disabled={disabled}
        value={value ?? ""}
        onChange={(event) =>
          onChange(
            event.currentTarget.value === ""
              ? null
              : Number(event.currentTarget.value),
          )
        }
      >
        <option value="">Choose a column</option>
        {discovery.preview.columns.map((column) => (
          <option
            key={column.source_column_ordinal}
            value={column.source_column_ordinal}
          >
            {sourceColumnLabel(column)}
          </option>
        ))}
      </select>
    </label>
  );
}

function PreviewSummary({
  preview,
}: {
  readonly preview: NetworkFlowImportPreviewResult;
}) {
  return (
    <section
      aria-live="polite"
      data-testid={networkAnalysisTestId("mapping-preview-summary")}
      style={previewStyle}
    >
      <h3 style={subtitleStyle}>Preview result</h3>
      <p>
        {preview.preview_accepted_count} accepted ·{" "}
        {preview.preview_rejected_count} rejected ·{" "}
        {preview.preview_record_count} examined
      </p>
      <p style={fingerprintStyle}>
        Mapping fingerprint: {preview.mapping_fingerprint}
      </p>
      {preview.diagnostics.length === 0 ? (
        <p>No preview diagnostics.</p>
      ) : (
        <ul>
          {preview.diagnostics.map((diagnostic) => (
            <li key={diagnostic.diagnostic_id}>
              Row {diagnostic.source_row_number}
              {diagnostic.source_column_ordinal === null
                ? ""
                : `, column ${diagnostic.source_column_ordinal}`}
              : {localizedNetworkFlowDiagnosticMessage(diagnostic)} (
              {diagnostic.reason_code})
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function safeSamples(
  samples: readonly { readonly safe_sample: string | null }[],
): string {
  const values = samples
    .map((sample) => sample.safe_sample)
    .filter((sample): sample is string => sample !== null);
  return values.length === 0 ? "none" : values.join(", ");
}

function humanize(value: string): string {
  return value.replaceAll("_", " ");
}

const backdropStyle: CSSProperties = {
  alignItems: "center",
  background: "rgba(15, 23, 42, 0.65)",
  display: "flex",
  inset: 0,
  justifyContent: "center",
  padding: 24,
  position: "fixed",
  zIndex: 40,
};
const dialogStyle: CSSProperties = {
  background: "var(--color-surface, #fff)",
  borderRadius: 10,
  boxShadow: "0 24px 64px rgba(15, 23, 42, 0.35)",
  color: "var(--color-text, #172033)",
  display: "grid",
  gap: 16,
  maxHeight: "calc(100vh - 48px)",
  maxWidth: 1080,
  overflow: "auto",
  padding: 20,
  width: "100%",
};
const headerStyle: CSSProperties = {
  alignItems: "start",
  display: "flex",
  gap: 16,
  justifyContent: "space-between",
};
const titleStyle: CSSProperties = { margin: 0 };
const subtitleStyle: CSSProperties = { margin: 0 };
const mutedStyle: CSSProperties = { margin: "6px 0 0", opacity: 0.75 };
const settingsGridStyle: CSSProperties = {
  display: "grid",
  gap: 12,
  gridTemplateColumns: "repeat(auto-fit, minmax(230px, 1fr))",
};
const fieldStyle: CSSProperties = {
  display: "grid",
  fontSize: 13,
  gap: 5,
};
const summaryStyle: CSSProperties = {
  alignItems: "center",
  display: "flex",
  flexWrap: "wrap",
  gap: 16,
  justifyContent: "space-between",
};
const mappingListStyle: CSSProperties = {
  border: "1px solid var(--color-border, #cbd5e1)",
  borderRadius: 6,
  display: "grid",
  maxHeight: 330,
  overflow: "auto",
};
const mappingRowStyle: CSSProperties = {
  alignItems: "center",
  borderBottom: "1px solid var(--color-border, #e2e8f0)",
  display: "grid",
  gap: 16,
  gridTemplateColumns: "minmax(260px, 1fr) minmax(240px, 0.8fr)",
  padding: "10px 12px",
};
const sampleStyle: CSSProperties = {
  fontSize: 12,
  marginTop: 4,
  opacity: 0.75,
};
const previewStyle: CSSProperties = {
  background: "var(--color-surface-subtle, #f1f5f9)",
  borderRadius: 6,
  padding: 12,
};
const fingerprintStyle: CSSProperties = {
  fontFamily: "ui-monospace, monospace",
  fontSize: 11,
  overflowWrap: "anywhere",
};
const alertStyle: CSSProperties = { color: "var(--color-danger, #b42318)" };
const footerStyle: CSSProperties = {
  display: "flex",
  gap: 8,
  justifyContent: "flex-end",
};
