import {
  networkAnalysisMappingColumnTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, useEffect, useRef } from "react";
import { terminalCommonJob } from "../services/commonJobContract";
import { importFailureMessage } from "../services/importClient";

import type { NetworkFlowImportPreviewResult } from "../services/networkFlowContractAdapter";
import {
  networkFlowMappingMetadata,
  networkFlowTimestampMetadata,
} from "../services/networkFlowContractAdapter";
import {
  NetworkFlowActionGroup,
  NetworkFlowButton,
  NetworkFlowField,
  NetworkFlowSelect,
  NetworkFlowTextInput,
} from "./NetworkFlowControls";
import type { NetworkFlowImportController } from "./NetworkFlowImportController";
import type { NetworkFlowImportDiscovery } from "./networkFlowImportModel";
import {
  ignoredColumnChoice,
  mappedRequiredFieldCount,
  type NetworkFlowMappingDraft,
  networkFlowMappingDraftReadyForPreview,
  networkFlowMappingFields,
  networkFlowMappingIssues,
  networkFlowRequiredFieldKeys,
  networkFlowSourceProfile,
  sourceColumnLabel,
  withNetworkFlowColumnChoice,
} from "./networkFlowImportModel";
import {
  type NetworkFlowImportState,
  networkFlowImportMappingState,
  unresolvedNetworkFlowWrite,
} from "./networkFlowImportState";
import { localizedNetworkFlowDiagnosticMessage } from "./networkFlowPresentation";
import { useNetworkFlowModalFocus } from "./useNetworkFlowModalFocus";

const unmappedColumnChoice = "__unmapped__";

export function NetworkFlowMappingModal({
  controller,
  state,
}: {
  readonly controller: NetworkFlowImportController;
  readonly state: NetworkFlowImportState;
}) {
  const closeButton = useRef<HTMLButtonElement>(null);
  const hasDraft = state.draft !== null;
  const focus = useNetworkFlowModalFocus<HTMLElement>({
    initialFocusTestId: networkAnalysisTestId("mapping-profile"),
    onDismiss: () => controller.setPresented(false),
  });
  useEffect(() => {
    if (hasDraft && document.activeElement === closeButton.current) {
      focus.dialogRef.current
        ?.querySelector<HTMLSelectElement>(
          "#network-flow-mapping-profile:not(:disabled)",
        )
        ?.focus({ preventScroll: true });
    }
  }, [hasDraft, focus.dialogRef]);
  const job = state.applyJob ?? state.discoveryJob;
  const write = state.write;
  const pending = write?.disposition === "pending";
  const frozen =
    !state.workspaceActive ||
    !state.canWrite ||
    unresolvedNetworkFlowWrite(state) ||
    state.applyJob !== null;
  const issues = state.draft ? networkFlowMappingIssues(state.draft) : [];
  const failure =
    state.sourceFailure ??
    state.previewFailure ??
    write?.failure ??
    job?.failure ??
    state.handoff?.failure ??
    state.cancellation?.failure;
  const continuation = state.approval
    ? state.selection
      ? "Apply selected mapping"
      : "Continue selection"
    : "Approve and apply";
  return (
    <div className="network-flow-dialog-backdrop">
      <section
        ref={focus.dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="network-flow-mapping-title"
        aria-describedby="network-flow-mapping-description"
        data-testid={networkAnalysisTestId("mapping-dialog")}
        className="network-flow-dialog"
        style={mappingDialogStyle}
        onKeyDown={focus.onKeyDown}
      >
        <header style={headerStyle}>
          <div>
            <h2 id="network-flow-mapping-title" style={titleStyle}>
              Review Network Flow mapping
            </h2>
            <p id="network-flow-mapping-description" style={mutedStyle}>
              Suggestions are not approval. Preview the current mapping, then
              explicitly approve and apply it. Closing this dialog does not
              cancel server work.
            </p>
          </div>
          <NetworkFlowButton
            ref={closeButton}
            variant="secondary"
            onClick={() => controller.setPresented(false)}
          >
            Close
          </NetworkFlowButton>
        </header>
        <section
          data-testid={networkAnalysisTestId("import-progress")}
          aria-label="Import progress"
          className="network-flow-status"
          data-tone={
            failure ? "error" : state.previewStale ? "stale" : undefined
          }
        >
          <p role="status" aria-live="polite" aria-atomic="true">
            {state.message}
          </p>
          <p>
            Stage: {humanize(state.stage)}
            {state.draft
              ? ` · Mapping: ${humanize(networkFlowImportMappingState(state) ?? "")}`
              : ""}
          </p>
          {failure ? (
            <p role="alert">
              {importFailureMessage(failure)}{" "}
              <span className="network-flow-mono">
                ({failure.code}
                {failure.reason ? `: ${failure.reason}` : ""})
              </span>
            </p>
          ) : null}
          {write?.disposition === "uncertain" ? (
            <p>
              The {write.request.kind} request may have reached the server. Its
              submitted intent is retained. Retry the exact request to determine
              its disposition.
            </p>
          ) : null}
          {job ? (
            <p>
              Job {job.resource.job_id}: {humanize(job.resource.status)} ·{" "}
              {job.observing
                ? "observing"
                : job.current
                  ? "current observation"
                  : "observation paused"}
              . {job.resource.progress.completed} of{" "}
              {job.resource.progress.total ?? "unknown"} job units.
            </p>
          ) : null}
          {job?.resource.error_summary ? (
            <p>Job outcome: {job.resource.error_summary.code}</p>
          ) : null}
          {state.cancellation ? (
            <p>
              Cancellation:{" "}
              {state.cancellation.disposition === "accepted"
                ? job && terminalCommonJob(job.resource)
                  ? "acknowledged; terminal job outcome shown above"
                  : "acknowledged; awaiting terminal job outcome"
                : state.cancellation.disposition}
              .
            </p>
          ) : null}
          {!state.canWrite && state.draft ? (
            <p>
              {state.closed
                ? "Closed, read-only."
                : "Mapping is read-only under current access."}{" "}
              The retained draft is available for copying.
            </p>
          ) : null}
          {!state.workspaceActive ? (
            <p>
              Return to Network Analysis to resume observation or recover table
              handoff.
            </p>
          ) : null}
        </section>
        <NetworkFlowActionGroup>
          {!state.draft &&
          controller.canDiscardDraft() &&
          state.stage !== "finished" ? (
            <NetworkFlowButton onClick={() => controller.discardDraft()}>
              Discard unsubmitted draft
            </NetworkFlowButton>
          ) : null}
          {state.stage === "finished" ? (
            <NetworkFlowButton
              disabled={
                !state.canWrite ||
                !state.workspaceActive ||
                !controller.canStartNew()
              }
              onClick={() => controller.startNew()}
            >
              Start another import
            </NetworkFlowButton>
          ) : null}
          {write && ["uncertain", "rejected"].includes(write.disposition) ? (
            <NetworkFlowButton
              data-testid={networkAnalysisTestId("import-replay")}
              disabled={!state.canWrite || !state.workspaceActive || pending}
              onClick={() => {
                void controller.retryWrite();
              }}
            >
              {write.disposition === "uncertain"
                ? "Retry exact request"
                : `Retry ${write.request.kind} request`}
            </NetworkFlowButton>
          ) : null}
          {job &&
          !job.observing &&
          (!job.current || !terminalCommonJob(job.resource)) ? (
            <NetworkFlowButton
              data-testid={networkAnalysisTestId("import-resume")}
              disabled={state.closed || !state.workspaceActive}
              onClick={() => {
                void controller.resumeObservation();
              }}
            >
              Resume observation
            </NetworkFlowButton>
          ) : null}
          {state.sourceFailure &&
          state.discoveryJob?.resource.status === "succeeded" &&
          !state.applyJob ? (
            <NetworkFlowButton
              data-testid={networkAnalysisTestId("import-source-reload")}
              disabled={state.closed}
              onClick={() => {
                void controller.refreshSource();
              }}
            >
              Reload discovered source
            </NetworkFlowButton>
          ) : null}
          {state.handoff &&
          !["selected", "loading"].includes(state.handoff.status) ? (
            <NetworkFlowButton
              data-testid={networkAnalysisTestId("import-handoff")}
              disabled={
                state.closed || !state.workspaceActive || !state.handoff.tableId
              }
              onClick={() => {
                void controller.recoverHandoff();
              }}
            >
              Recover table handoff
            </NetworkFlowButton>
          ) : null}
          {controller.canCancel() ||
          state.cancellation?.disposition === "uncertain" ? (
            <NetworkFlowButton
              data-testid={networkAnalysisTestId("import-cancel")}
              disabled={state.closed}
              pending={
                state.cancellationChecking ||
                state.cancellation?.disposition === "pending"
              }
              onClick={() => {
                void controller.cancel();
              }}
            >
              {state.cancellation?.disposition === "uncertain"
                ? "Retry exact cancellation request"
                : "Request cancellation"}
            </NetworkFlowButton>
          ) : null}
          {job?.observing ? (
            <NetworkFlowButton onClick={() => controller.pauseObservation()}>
              Pause observation
            </NetworkFlowButton>
          ) : null}
        </NetworkFlowActionGroup>
        {state.discovery && state.draft ? (
          <>
            <p className="network-flow-mono">
              Source: {state.discovery.session.original_filename} ·{" "}
              {state.discovery.session.source_file_kind} · discovery parser{" "}
              {state.discovery.session.parser_profile_id} v
              {state.discovery.session.parser_version} · analytical parser{" "}
              {networkFlowMappingMetadata.source_profiles.find(
                (p) => p.source_profile_id === state.draft?.sourceProfileId,
              )?.parser_profile_id ?? "unavailable"}
            </p>
            {issues.length ? (
              <section
                aria-label="Mapping requirements"
                id="network-flow-mapping-issues"
              >
                <strong>Mapping requires attention</strong>
                <ul>
                  {issues.map((issue) => (
                    <li key={`${issue.field}:${issue.message}`}>
                      {issue.message}
                    </li>
                  ))}
                </ul>
              </section>
            ) : null}
            <NetworkFlowMappingEditor
              discovery={state.discovery}
              draft={state.draft}
              preview={state.preview?.value ?? null}
              working={frozen || pending}
              onDraftChange={(draft) => controller.updateDraft(draft)}
            />
            {!state.canWrite || frozen ? (
              <NetworkFlowField
                htmlFor="network-flow-copy-draft"
                label="Retained mapping draft (copy only)"
              >
                <textarea
                  id="network-flow-copy-draft"
                  readOnly
                  value={JSON.stringify(state.draft, null, 2)}
                  className="network-flow-control network-flow-input"
                  rows={5}
                />
              </NetworkFlowField>
            ) : null}
            <footer>
              <NetworkFlowActionGroup>
                {!state.approval && !state.applyJob ? (
                  <NetworkFlowButton
                    disabled={!controller.canDiscardDraft()}
                    onClick={() => controller.discardDraft()}
                  >
                    Discard unsubmitted draft
                  </NetworkFlowButton>
                ) : null}
                <NetworkFlowButton
                  data-testid={networkAnalysisTestId("mapping-preview")}
                  disabled={!controller.canPreview()}
                  pending={state.stage === "previewing"}
                  onClick={() => {
                    void controller.requestPreview();
                  }}
                >
                  Preview mapping
                </NetworkFlowButton>
                <NetworkFlowButton
                  data-testid={networkAnalysisTestId("mapping-apply")}
                  disabled={!controller.canContinue()}
                  pending={pending && write?.request.kind !== "upload"}
                  variant="primary"
                  onClick={() => {
                    void controller.continueApply();
                  }}
                >
                  {continuation}
                </NetworkFlowButton>
              </NetworkFlowActionGroup>
            </footer>
          </>
        ) : null}
        {state.stage === "finished" ? (
          <NetworkFlowButton
            disabled={!state.canWrite}
            onClick={() => controller.startNew()}
          >
            Finish review
          </NetworkFlowButton>
        ) : null}
      </section>
    </div>
  );
}

function NetworkFlowMappingEditor({
  discovery,
  draft,
  preview,
  working,
  onDraftChange,
}: {
  readonly discovery: NetworkFlowImportDiscovery;
  readonly draft: NetworkFlowMappingDraft;
  readonly preview: NetworkFlowImportPreviewResult | null;
  readonly working: boolean;
  readonly onDraftChange: (draft: NetworkFlowMappingDraft) => void;
}) {
  const requiredCount = mappedRequiredFieldCount(draft);
  const sourceProfile = networkFlowMappingMetadata.source_profiles.find(
    (candidate) => candidate.source_profile_id === draft.sourceProfileId,
  );
  const timestampReady = networkFlowMappingDraftReadyForPreview(draft);
  return (
    <>
      <div style={settingsGridStyle}>
        <NetworkFlowField
          htmlFor="network-flow-mapping-profile"
          label="Source profile"
        >
          <NetworkFlowSelect
            data-testid={networkAnalysisTestId("mapping-profile")}
            disabled={working}
            id="network-flow-mapping-profile"
            aria-describedby={
              networkFlowMappingIssues(draft).length
                ? "network-flow-mapping-issues"
                : undefined
            }
            value={draft.sourceProfileId}
            onChange={(event) =>
              onDraftChange({
                ...draft,
                sourceProfileId: event.currentTarget.value,
              })
            }
          >
            {networkFlowMappingMetadata.source_profiles
              .filter((profile) =>
                networkFlowSourceProfile(profile.source_profile_id),
              )
              .map((profile) => (
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
            readOnly={working}
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
            {sourceProfile?.supported_unknown_column_policies.map((policy) => (
              <option key={policy} value={policy}>
                {humanize(policy)}
              </option>
            ))}
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
          {requiredCount} of{" "}
          {networkFlowRequiredFieldKeys(draft.sourceProfileId).length} required
          fields mapped
        </strong>
        <span>
          {discovery.preview.columns.length} discovered source columns; headers
          are qualified by ordinal.
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
              candidate.source_column_ordinal === column.source_column_ordinal,
          );
          const columnIssues = networkFlowMappingIssues(draft).filter(
            (issue) => issue.field === `column:${column.source_column_ordinal}`,
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
                <strong style={{ whiteSpace: "pre-wrap" }}>
                  {sourceColumnLabel(column)}
                </strong>
                {sourceColumn === undefined ? (
                  <div style={sampleStyle}>
                    Safe samples available after preview.
                  </div>
                ) : (
                  <div style={sampleStyle}>
                    Safe samples: {safeSamples(sourceColumn.sample_values)} ·
                    empty values: {sourceColumn.detected_empty_count}
                  </div>
                )}
              </div>
              <NetworkFlowField
                htmlFor={`network-flow-mapping-target-${column.source_column_ordinal}`}
                label="Target field"
                error={columnIssues.map((issue) => issue.message).join(" ")}
                errorId={`network-flow-mapping-error-${column.source_column_ordinal}`}
              >
                <NetworkFlowSelect
                  aria-label={`Target for ${sourceColumnLabel(column)}`}
                  aria-invalid={columnIssues.length > 0}
                  aria-describedby={
                    columnIssues.length
                      ? `network-flow-mapping-error-${column.source_column_ordinal}`
                      : undefined
                  }
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
                  <option value={ignoredColumnChoice}>Explicitly ignore</option>
                  {networkFlowMappingFields(draft.sourceProfileId).map(
                    (field) => (
                      <option key={field.field_key} value={field.field_key}>
                        {humanize(field.field_key.replace("network_flow.", ""))}
                        {` (${field.requirement})`}
                      </option>
                    ),
                  )}
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
    </>
  );
}

function NetFlowUptimeSettings({
  disabled,
  discovery,
  draft,
  onDraftChange,
}: {
  readonly disabled: boolean;
  readonly discovery: NetworkFlowImportDiscovery;
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
          {networkFlowTimestampMetadata.exportTimeModes.map((mode) => (
            <option key={mode} value={mode}>
              {humanize(mode)}
            </option>
          ))}
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
  readonly discovery: NetworkFlowImportDiscovery;
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
      <h3 style={subtitleStyle}>Preview slice result</h3>
      <p>
        These counts describe only the preview slice. Full-file validation can
        differ; rejected preview rows do not by themselves prohibit import.
      </p>
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
  gridTemplateColumns: "repeat(auto-fit, minmax(min(14rem, 100%), 1fr))",
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
