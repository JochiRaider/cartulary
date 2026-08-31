import {
  networkAnalysisMappingColumnTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import type { ExtensionImportDiscovery } from "../imports/importCoordinator";
import type { NetworkFlowImportPreviewResult } from "../services/networkFlowContractAdapter";
import { networkFlowMappingMetadata } from "../services/networkFlowContractAdapter";
import {
  NetworkFlowActionGroup,
  NetworkFlowButton,
  NetworkFlowField,
  NetworkFlowSelect,
  NetworkFlowTextInput,
} from "./NetworkFlowControls";
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
    <div className="network-flow-dialog-backdrop">
      <section
        ref={modalFocus.dialogRef}
        aria-describedby="network-flow-mapping-description"
        aria-labelledby="network-flow-mapping-title"
        aria-modal="true"
        data-testid={networkAnalysisTestId("mapping-dialog")}
        role="dialog"
        className="network-flow-dialog"
        style={mappingDialogStyle}
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
          <NetworkFlowButton
            disabled={working}
            variant="secondary"
            onClick={onCancel}
          >
            Cancel
          </NetworkFlowButton>
        </header>

        <div style={settingsGridStyle}>
          <NetworkFlowField
            htmlFor="network-flow-mapping-profile"
            label="Source profile"
          >
            <NetworkFlowSelect
              data-testid={networkAnalysisTestId("mapping-profile")}
              disabled={working}
              id="network-flow-mapping-profile"
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
            </NetworkFlowSelect>
          </NetworkFlowField>
          <NetworkFlowField
            htmlFor="network-flow-mapping-display-name"
            label="Table display name (optional)"
          >
            <NetworkFlowTextInput
              data-testid={networkAnalysisTestId("mapping-display-name")}
              disabled={working}
              id="network-flow-mapping-display-name"
              maxLength={64}
              value={draft.displayNameOverride}
              onChange={(event) =>
                onDraftChange({
                  ...draft,
                  displayNameOverride: event.currentTarget.value,
                })
              }
            />
          </NetworkFlowField>
          <NetworkFlowField
            htmlFor="network-flow-mapping-timestamp-mode"
            label="Timestamp interpretation"
          >
            <NetworkFlowSelect
              data-testid={networkAnalysisTestId("mapping-timestamp-mode")}
              disabled={working}
              id="network-flow-mapping-timestamp-mode"
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
            </NetworkFlowSelect>
          </NetworkFlowField>
          <NetworkFlowField
            htmlFor="network-flow-mapping-unknown-policy"
            label="Unknown-column policy"
          >
            <NetworkFlowSelect
              data-testid={networkAnalysisTestId("mapping-unknown-policy")}
              disabled={working}
              id="network-flow-mapping-unknown-policy"
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
            </NetworkFlowSelect>
          </NetworkFlowField>
          {draft.timestampMode === "rfc3339" ? (
            <NetworkFlowField
              htmlFor="network-flow-mapping-timezone"
              label="Source timezone (blank for offset-bearing timestamps)"
            >
              <NetworkFlowTextInput
                data-testid={networkAnalysisTestId("mapping-timezone")}
                disabled={working}
                id="network-flow-mapping-timezone"
                placeholder="UTC"
                value={draft.timezone}
                onChange={(event) =>
                  onDraftChange({
                    ...draft,
                    timezone: event.currentTarget.value,
                  })
                }
              />
            </NetworkFlowField>
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
            <div
              aria-label="Alias collision"
              className="network-flow-status"
              data-tone="error"
              role="alert"
            >
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
                className="network-flow-mapping-row"
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
                <NetworkFlowField
                  htmlFor={`network-flow-mapping-target-${column.source_column_ordinal}`}
                  label="Target field"
                >
                  <NetworkFlowSelect
                    aria-label={`Target for ${sourceColumnLabel(column)}`}
                    data-testid={networkAnalysisMappingColumnTestId(
                      column.source_column_ordinal,
                    )}
                    disabled={working}
                    id={`network-flow-mapping-target-${column.source_column_ordinal}`}
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
                  </NetworkFlowSelect>
                </NetworkFlowField>
              </div>
            );
          })}
        </section>

        {preview === null ? null : <PreviewSummary preview={preview} />}

        {!timestampReady ? (
          <p className="network-flow-status" data-tone="error" role="alert">
            Choose distinct export-time and exporter-uptime columns before
            previewing this timestamp mode.
          </p>
        ) : null}

        <footer>
          <NetworkFlowActionGroup>
            <NetworkFlowButton
              disabled={working}
              variant="secondary"
              onClick={onCancel}
            >
              Cancel
            </NetworkFlowButton>
            <NetworkFlowButton
              data-testid={networkAnalysisTestId("mapping-preview")}
              disabled={!timestampReady || stage === "applying"}
              pending={stage === "previewing"}
              variant="secondary"
              onClick={onPreview}
            >
              {stage === "previewing" ? "Previewing…" : "Preview mapping"}
            </NetworkFlowButton>
            <NetworkFlowButton
              data-testid={networkAnalysisTestId("mapping-apply")}
              disabled={!canApply || stage === "previewing"}
              pending={stage === "applying"}
              variant="primary"
              onClick={onApply}
            >
              {stage === "applying" ? "Applying…" : "Approve and apply"}
            </NetworkFlowButton>
          </NetworkFlowActionGroup>
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
      <NetworkFlowField
        htmlFor="network-flow-export-time-mode"
        label="Export-time interpretation"
      >
        <NetworkFlowSelect
          disabled={disabled}
          id="network-flow-export-time-mode"
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
        </NetworkFlowSelect>
      </NetworkFlowField>
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
  const controlId = `network-flow-column-${label
    .toLowerCase()
    .replaceAll(/[^a-z0-9]+/gu, "-")}`;
  return (
    <NetworkFlowField htmlFor={controlId} label={label}>
      <NetworkFlowSelect
        disabled={disabled}
        id={controlId}
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
      </NetworkFlowSelect>
    </NetworkFlowField>
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

const mappingDialogStyle: CSSProperties = {
  inlineSize: "min(67.5rem, 100%)",
};
const headerStyle: CSSProperties = {
  alignItems: "start",
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-lg)",
  justifyContent: "space-between",
};
const titleStyle: CSSProperties = { margin: 0 };
const subtitleStyle: CSSProperties = { margin: 0 };
const mutedStyle: CSSProperties = {
  color: "var(--ct-colors-ink-muted)",
  margin: "var(--ct-spacing-sm) 0 0",
};
const settingsGridStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-md)",
  gridTemplateColumns: "repeat(auto-fit, minmax(14rem, 1fr))",
};
const summaryStyle: CSSProperties = {
  alignItems: "center",
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-lg)",
  justifyContent: "space-between",
};
const mappingListStyle: CSSProperties = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  display: "grid",
  maxHeight: 330,
  overflow: "auto",
};
const mappingRowStyle: CSSProperties = {
  alignItems: "center",
  borderBottom: "var(--ct-border-hairline)",
  display: "grid",
  gap: "var(--ct-spacing-lg)",
  gridTemplateColumns: "minmax(16rem, 1fr) minmax(15rem, 0.8fr)",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
};
const sampleStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "var(--ct-typography-compact-metadata-fontSize)",
  marginTop: "var(--ct-spacing-xs)",
  overflowWrap: "anywhere",
};
const previewStyle: CSSProperties = {
  background: "var(--ct-colors-surface-2)",
  borderRadius: "var(--ct-rounded-sm)",
  padding: "var(--ct-spacing-md)",
};
const fingerprintStyle: CSSProperties = {
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "var(--ct-typography-mono-fontSize)",
  overflowWrap: "anywhere",
};
