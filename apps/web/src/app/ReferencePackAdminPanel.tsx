import {
  referencePackAdminPanelTestId,
  referencePackCancelButtonTestId,
  referencePackErrorTestId,
  referencePackFileInputTestId,
  referencePackImportButtonTestId,
  referencePackJobStatusTestId,
  referencePackListStatusTestId,
  referencePackRefreshAllButtonTestId,
  referencePackRefreshSelectedButtonTestId,
  referencePackReloadButtonTestId,
  referencePackRowTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import {
  type ReferencePackCommand,
  referencePackJobProblem,
  referencePackResultTarget,
  referencePackVersionRoute,
} from "../services/referencePacks";
import type { ReferencePackAdminController } from "./referencePackAdminController";
import {
  emptyReferencePackQuery,
  type ReferencePackKnownJob,
  referencePackBusy,
  referencePackCommandLabel,
  referencePackEligible,
  referencePackIdentity,
  referencePackListStatus,
  referencePackProblemText,
  terminalReferencePackJobStates,
} from "./referencePackAdminModel";
import { useReferencePackAdminPresentation } from "./useReferencePackAdmin";

const referencePackStyles = `
[data-reference-pack-admin] :is(button,input,select):focus-visible { outline: var(--ct-border-focus); outline-offset: var(--ct-component-focus-ring-offset); }
[data-reference-pack-admin] button:is(:disabled,[aria-disabled="true"]) { color: var(--ct-colors-ink-subtle) !important; background: var(--ct-colors-surface-3) !important; cursor: not-allowed; }
[data-reference-pack-admin] progress { appearance: none; box-sizing: border-box; height: var(--ct-spacing-sm); border: var(--ct-border-hairline); border-radius: var(--ct-rounded-pill); background: var(--ct-colors-surface-3); }
[data-reference-pack-admin] progress:indeterminate { background: repeating-linear-gradient(135deg, var(--ct-colors-accent) 0, var(--ct-colors-accent) var(--ct-spacing-xs), var(--ct-colors-surface-3) var(--ct-spacing-xs), var(--ct-colors-surface-3) var(--ct-spacing-sm)); }
[data-reference-pack-admin] progress::-webkit-progress-bar { background: inherit; border-radius: inherit; }
[data-reference-pack-admin] progress::-webkit-progress-value { background: var(--ct-colors-accent); border-radius: inherit; }
[data-reference-pack-admin] progress::-moz-progress-bar { background: var(--ct-colors-accent); border-radius: inherit; }
[data-reference-pack-admin] .rp-cell-label { display: none; }
@container (max-width: 42rem) {
  [data-reference-pack-admin] table, [data-reference-pack-admin] tbody, [data-reference-pack-admin] caption { display: block; }
  [data-reference-pack-admin] colgroup { display: none; }
  [data-reference-pack-admin] thead { position: absolute; inline-size: 1px; block-size: 1px; overflow: hidden; clip-path: inset(50%); }
  [data-reference-pack-admin] tbody tr { display: flex; flex-wrap: wrap; border-block-end: var(--ct-border-hairline); padding-block: var(--ct-spacing-sm); }
  [data-reference-pack-admin] td { display: block; box-sizing: border-box; min-inline-size: 0; flex: 1 1 12em; border-block-end: 0 !important; }
  [data-reference-pack-admin] td:first-child, [data-reference-pack-admin] td:last-child { flex-basis: 100%; }
  [data-reference-pack-admin] .rp-cell-label { display: block; font-weight: bold; margin-block-end: var(--ct-spacing-xs); }
}
`;

export function ReferencePackAdminPanel({
  controller,
  active,
}: {
  readonly controller: ReferencePackAdminController;
  readonly active: boolean;
}) {
  const binding = useReferencePackAdminPresentation(controller, active);
  const { state } = binding;
  const busy = referencePackBusy(state);
  const available =
    state.authority !== null &&
    state.access === "ready" &&
    !state.reconciling &&
    active;
  const canStart = available && !busy;
  const loadedKeys = new Set(state.catalog.rows.map((pack) => pack.pack_key));
  const hiddenKeys = state.selectedKeys.filter((key) => !loadedKeys.has(key));
  const operation = state.operation;
  const run = (element: HTMLElement, command: ReferencePackCommand) =>
    binding.run(element, () => controller.run(command));
  return (
    <section
      ref={binding.rootRef}
      data-reference-pack-admin=""
      data-testid={referencePackAdminPanelTestId()}
      style={panelStyle}
    >
      <style>{referencePackStyles}</style>
      <header style={rowStyle}>
        <h2 style={titleStyle}>Reference packs</h2>
        <button
          type="button"
          style={buttonStyle}
          data-testid={referencePackReloadButtonTestId()}
          disabled={state.access !== "ready"}
          onClick={() => void controller.reload()}
        >
          Reload catalog
        </button>
      </header>
      {state.authority === null ? (
        <p>
          Deployment admin access is required for reference-pack import,
          refresh, activation, and verification actions.
        </p>
      ) : (
        <>
          {state.access !== "ready" || state.reconciling ? (
            <p>
              {state.access === "unavailable"
                ? "Access could not be confirmed. Retained information may be stale."
                : "Checking current access and reconciling reference packs."}
            </p>
          ) : null}
          {state.access === "unavailable" ? (
            <button
              type="button"
              data-rp-recovery
              style={buttonStyle}
              onClick={(event) =>
                binding.run(event.currentTarget, controller.retryAccess)
              }
            >
              Retry access
            </button>
          ) : null}
          <form
            aria-label="Reference pack catalog query"
            style={rowStyle}
            onSubmit={(event) => {
              event.preventDefault();
              void controller.reload();
            }}
          >
            <label style={searchFieldStyle}>
              Search reference packs
              <input
                style={inputStyle}
                value={state.input.search}
                onChange={(event) =>
                  controller.setQuery({
                    ...state.input,
                    search: event.currentTarget.value,
                  })
                }
              />
            </label>
            <label style={fieldStyle}>
              State
              <select
                aria-label="Reference pack state"
                style={inputStyle}
                value={state.input.packVersionState}
                onChange={(event) =>
                  controller.setQuery({
                    ...state.input,
                    packVersionState: event.currentTarget.value,
                  })
                }
              >
                <option value="">Any state</option>
                {Object.entries(stateLabels).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </label>
            <label style={fieldStyle}>
              Verification
              <select
                aria-label="Reference pack verification result"
                style={inputStyle}
                value={state.input.verificationResult}
                onChange={(event) =>
                  controller.setQuery({
                    ...state.input,
                    verificationResult: event.currentTarget.value,
                  })
                }
              >
                <option value="">Any verification</option>
                <option value="pending">Pending</option>
                <option value="passed">Passed</option>
                <option value="failed">Failed</option>
              </select>
            </label>
            <label style={fieldStyle}>
              Active
              <select
                aria-label="Reference pack active state"
                style={inputStyle}
                value={state.input.active}
                onChange={(event) =>
                  controller.setQuery({
                    ...state.input,
                    active: event.currentTarget.value,
                  })
                }
              >
                <option value="">Any active state</option>
                <option value="true">Active</option>
                <option value="false">Inactive</option>
              </select>
            </label>
            <button type="submit" style={buttonStyle}>
              Search
            </button>
            <button
              type="button"
              style={buttonStyle}
              onClick={() =>
                controller.setQuery({ ...emptyReferencePackQuery })
              }
            >
              Clear filters
            </button>
          </form>
          <p role="status" data-testid={referencePackListStatusTestId()}>
            {referencePackListStatus(state)}
          </p>
          <div data-testid={referencePackErrorTestId()}>
            {state.catalog.problem ? (
              <p style={errorStyle}>
                {referencePackProblemText(state.catalog.problem)}{" "}
                <button
                  type="button"
                  style={buttonStyle}
                  onClick={() => void controller.reload()}
                >
                  Retry catalog
                </button>
              </p>
            ) : null}
          </div>
          <form
            aria-label="Import reference pack"
            style={sectionStyle}
            onSubmit={(event) => {
              event.preventDefault();
              if (canStart && state.file)
                run(
                  event.nativeEvent instanceof SubmitEvent &&
                    event.nativeEvent.submitter instanceof HTMLElement
                    ? event.nativeEvent.submitter
                    : event.currentTarget,
                  { kind: "import", filename: state.file.name },
                );
            }}
          >
            <h3 style={subtitleStyle}>Import a bundle</h3>
            <p style={mutedStyle}>
              Import verifies a candidate version. Activation requires a
              separate action.
            </p>
            <div style={rowStyle}>
              <label style={searchFieldStyle}>
                Reference pack bundle
                <input
                  type="file"
                  required={state.file === null}
                  ref={binding.fileInputRef}
                  data-testid={referencePackFileInputTestId()}
                  style={fileStyle}
                  disabled={busy || !available}
                  onChange={(event) =>
                    controller.setFile(event.currentTarget.files?.[0] ?? null)
                  }
                />
              </label>
              <button
                type="submit"
                style={primaryButtonStyle}
                data-testid={referencePackImportButtonTestId()}
                aria-disabled={!canStart || !state.file}
              >
                Import
              </button>
              {state.file ? (
                <button
                  type="button"
                  style={buttonStyle}
                  disabled={busy}
                  onClick={() => controller.setFile(null)}
                >
                  Clear file
                </button>
              ) : null}
            </div>
            {state.file ? (
              <p style={mutedStyle}>Selected file: {state.file.name}</p>
            ) : null}
          </form>
          <section aria-label="Refresh scope" style={sectionStyle}>
            <div style={rowStyle}>
              <button
                type="button"
                style={buttonStyle}
                data-testid={referencePackRefreshAllButtonTestId()}
                aria-disabled={!canStart}
                onClick={(event) => {
                  if (canStart)
                    run(event.currentTarget, { kind: "refresh_all" });
                }}
              >
                Refresh all
              </button>
              <button
                type="button"
                style={buttonStyle}
                data-testid={referencePackRefreshSelectedButtonTestId()}
                aria-disabled={!canStart || !state.selectedKeys.length}
                onClick={(event) => {
                  if (canStart && state.selectedKeys.length)
                    run(event.currentTarget, {
                      kind: "refresh_selected",
                      packKeys: state.selectedKeys,
                    });
                }}
              >
                Refresh selected
              </button>
              <span>
                {state.selectedKeys.length} pack keys selected ·{" "}
                {loadedKeys.size} pack keys loaded
              </span>
              <button
                type="button"
                style={buttonStyle}
                disabled={!state.selectedKeys.length}
                onClick={() => controller.clearSelection()}
              >
                Clear selection
              </button>
            </div>
            <p style={mutedStyle}>
              Refresh verifies imported packs. Refresh all covers the server's
              imported catalog at admission. Selection persists across filters
              and pages.
            </p>
            {hiddenKeys.length ? (
              <div>
                <p>
                  {hiddenKeys.length} selected keys are outside the loaded rows:
                </p>
                <ul style={listStyle}>
                  {hiddenKeys.map((key) => (
                    <li key={key} style={rowStyle}>
                      <span style={identityStyle}>{key}</span>
                      <button
                        type="button"
                        style={buttonStyle}
                        aria-label={`Deselect ${key}`}
                        onClick={() => controller.setSelected(key, false)}
                      >
                        Deselect
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
            ) : null}
          </section>
          <section
            ref={binding.feedbackRef}
            tabIndex={-1}
            aria-label="Current pack action"
            style={
              operation ? sectionStyle : { ...sectionStyle, display: "none" }
            }
          >
            {operation ? (
              <>
                <h3 style={subtitleStyle}>
                  {referencePackCommandLabel(operation.command)}
                </h3>
                <p>{phaseLabels[operation.phase]}</p>
                {operation.problem ? (
                  <p style={errorStyle}>
                    {referencePackProblemText(operation.problem)}
                  </p>
                ) : null}
                {operation.phase === "uncertain" ||
                operation.phase === "replaying" ? (
                  <button
                    type="button"
                    style={buttonStyle}
                    data-rp-recovery
                    onClick={(event) =>
                      binding.run(event.currentTarget, controller.retryAttempt)
                    }
                    aria-disabled={
                      !available || operation.phase === "replaying"
                    }
                  >
                    Retry exact request
                  </button>
                ) : null}
                {operation.phase === "rejected" ? (
                  <button
                    type="button"
                    style={buttonStyle}
                    data-rp-recovery
                    onClick={(event) =>
                      binding.run(event.currentTarget, controller.reload)
                    }
                  >
                    Review current catalog
                  </button>
                ) : null}
                {busy ? (
                  <p style={mutedStyle}>
                    Resolve this operation before starting another pack
                    operation. Catalog search and selection remain available.
                  </p>
                ) : null}
              </>
            ) : (
              <p style={mutedStyle}>No operation submitted in this session.</p>
            )}
          </section>
          {Object.keys(state.jobs).length ? (
            <section
              aria-label="Operations in this session"
              data-testid={referencePackJobStatusTestId()}
              style={sectionStyle}
            >
              <h3 style={subtitleStyle}>Operations in this session</h3>
              <p style={mutedStyle}>
                Only operations learned in this application session appear here.
                Leaving this panel pauses observation; it does not cancel server
                work.
              </p>
              <ol style={listStyle}>
                {Object.entries(state.jobs).map(([id, job]) => (
                  <li key={id} style={sectionStyle}>
                    <section
                      aria-label={referencePackCommandLabel(job.command)}
                    >
                      <h4 style={subtitleStyle}>
                        {referencePackCommandLabel(job.command)}
                      </h4>
                      <p>
                        {jobStatusLabels[job.snapshot.status]}
                        {job.observation === "reading"
                          ? " · Checking status"
                          : job.observation === "paused"
                            ? " · Observation paused"
                            : ""}
                      </p>
                      <progress
                        aria-label={`${referencePackCommandLabel(job.command)} progress`}
                        {...(job.snapshot.progress.total === null
                          ? {}
                          : {
                              value: job.snapshot.progress.completed,
                              max: job.snapshot.progress.total,
                            })}
                      />
                      <span>
                        {" "}
                        {job.snapshot.progress.completed}
                        {job.snapshot.progress.total === null
                          ? " processed; total unknown"
                          : ` of ${job.snapshot.progress.total}`}
                      </span>
                      {terminalReferencePackJobStates.has(
                        job.snapshot.status,
                      ) ? (
                        <p>{jobResultText(job)}</p>
                      ) : null}
                      {job.problem ? (
                        <p style={errorStyle}>
                          {referencePackProblemText(job.problem)} The last
                          observed state is retained.
                        </p>
                      ) : null}
                      {job.problem !== null ||
                      job.observation === "failed" ||
                      job.observation === "unavailable" ||
                      job.observation === "paused" ? (
                        <button
                          type="button"
                          style={buttonStyle}
                          data-rp-recovery
                          aria-disabled={
                            !available || job.observation === "reading"
                          }
                          onClick={(event) =>
                            binding.run(event.currentTarget, () =>
                              controller.retryObservation(id),
                            )
                          }
                        >
                          Retry observation
                        </button>
                      ) : null}
                      {job.snapshot.cancelable &&
                      !terminalReferencePackJobStates.has(
                        job.snapshot.status,
                      ) ? (
                        <button
                          type="button"
                          style={buttonStyle}
                          data-testid={referencePackCancelButtonTestId()}
                          aria-disabled={
                            !available ||
                            ["checking", "submitting", "uncertain"].includes(
                              job.cancellation?.phase ?? "",
                            )
                          }
                          onClick={(event) => {
                            if (available)
                              binding.run(event.currentTarget, () =>
                                controller.cancelJob(id),
                              );
                          }}
                        >
                          Cancel operation
                        </button>
                      ) : null}
                      {job.cancellation ? (
                        <p>
                          {cancelLabels[job.cancellation.phase]}{" "}
                          {job.cancellation.problem
                            ? referencePackProblemText(job.cancellation.problem)
                            : ""}
                        </p>
                      ) : null}
                      {job.cancellation?.phase === "uncertain" ? (
                        <button
                          type="button"
                          style={buttonStyle}
                          data-rp-recovery
                          disabled={!available}
                          onClick={(event) =>
                            binding.run(event.currentTarget, () =>
                              controller.cancelJob(id, true),
                            )
                          }
                        >
                          Retry cancellation request
                        </button>
                      ) : null}
                      {terminalReferencePackJobStates.has(
                        job.snapshot.status,
                      ) ? (
                        <button
                          type="button"
                          style={buttonStyle}
                          onClick={(event) =>
                            binding.run(event.currentTarget, () =>
                              controller.dismissJob(id),
                            )
                          }
                        >
                          Dismiss operation
                        </button>
                      ) : null}
                    </section>
                  </li>
                ))}
              </ol>
            </section>
          ) : null}
          <div style={catalogStyle}>
            <table style={tableStyle}>
              <caption style={captionStyle}>Imported pack versions</caption>
              <colgroup>
                {["40%", "15%", "20%", "7%", "18%"].map((width) => (
                  <col key={width} style={{ width }} />
                ))}
              </colgroup>
              <thead>
                <tr>
                  {[
                    "Pack selection",
                    "Version state",
                    "Verification",
                    "Active",
                    "Version actions",
                  ].map((label) => (
                    <th key={label} scope="col" style={cellStyle}>
                      {label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {state.catalog.rows.map((pack) => (
                  <tr
                    key={referencePackIdentity(pack)}
                    data-testid={referencePackRowTestId(
                      pack.pack_key,
                      pack.pack_version,
                    )}
                  >
                    <td style={cellStyle}>
                      <label style={rowStyle}>
                        <input
                          type="checkbox"
                          aria-label={`Select pack key ${pack.pack_key}, version ${pack.pack_version}`}
                          checked={state.selectedKeys.includes(pack.pack_key)}
                          onChange={(event) =>
                            controller.setSelected(
                              pack.pack_key,
                              event.currentTarget.checked,
                            )
                          }
                        />
                        <span style={identityStyle}>
                          <strong>{pack.pack_key}</strong>
                          <span>@{pack.pack_version}</span>
                          <small style={mutedStyle}> · {pack.pack_kind}</small>
                        </span>
                      </label>
                    </td>
                    <td style={cellStyle}>
                      <span className="rp-cell-label" aria-hidden="true">
                        Version state
                      </span>
                      {stateLabels[pack.pack_version_state]}
                    </td>
                    <td style={cellStyle}>
                      <span className="rp-cell-label" aria-hidden="true">
                        Verification
                      </span>
                      {verificationLabels[pack.verification_result]}
                      <small style={mutedStyle}>
                        {" "}
                        · {pack.verification_method}
                      </small>
                    </td>
                    <td style={cellStyle}>
                      <span className="rp-cell-label" aria-hidden="true">
                        Active
                      </span>
                      {pack.active ? "Yes" : "No"}
                    </td>
                    <td style={cellStyle}>
                      <span className="rp-cell-label" aria-hidden="true">
                        Version actions
                      </span>
                      <span style={rowStyle}>
                        {(["activate", "disable", "reverify"] as const).map(
                          (action) => (
                            <button
                              key={action}
                              type="button"
                              style={buttonStyle}
                              disabled={!referencePackEligible(pack, action)}
                              aria-disabled={!canStart}
                              onClick={(event) => {
                                if (canStart)
                                  run(event.currentTarget, {
                                    kind: action,
                                    target: {
                                      pack_key: pack.pack_key,
                                      pack_version: pack.pack_version,
                                    },
                                  });
                              }}
                            >
                              {actionLabels[action]}
                            </button>
                          ),
                        )}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {state.catalog.paging.has_more ? (
            <button
              type="button"
              style={buttonStyle}
              disabled={state.catalog.pending !== null || state.catalog.dirty}
              onClick={() => void controller.loadMore()}
            >
              Load more
            </button>
          ) : null}
        </>
      )}
      <span
        role="status"
        aria-live="polite"
        aria-atomic="true"
        style={visuallyHiddenStyle}
      >
        {active ? state.announcement.text : ""}
      </span>
    </section>
  );
}
const actionLabels = {
  activate: "Activate",
  disable: "Disable",
  reverify: "Reverify",
};
const stateLabels = {
  staged: "Staged",
  verified_available: "Verified, available",
  disabled: "Disabled",
  failed: "Failed",
  missing: "Missing",
};
const verificationLabels = {
  pending: "Pending",
  passed: "Passed",
  failed: "Failed",
};
const phaseLabels = {
  checking: "Checking current state before submission.",
  submitting: "Submitted; awaiting server acknowledgment.",
  replaying: "Exact request resubmitted; awaiting server acknowledgment.",
  accepted: "Accepted for processing. Completion has not yet been confirmed.",
  committed: "Committed. Catalog data may still need reloading.",
  rejected: "Request rejected.",
  uncertain: "Outcome uncertain. Recovery resends the exact captured request.",
  failed: "Operation failed.",
  canceled: "Operation canceled.",
};
const cancelLabels = {
  checking: "Checking whether cancellation is currently allowed.",
  submitting: "Cancellation submitted; awaiting acknowledgment.",
  uncertain: "Cancellation outcome uncertain.",
  rejected: "Cancellation was not confirmed.",
  acknowledged:
    "Cancellation acknowledged. This does not establish cancellation or rollback.",
};
const jobStatusLabels = {
  queued: "Queued",
  running: "Running",
  cancel_requested: "Cancellation requested",
  succeeded: "Succeeded",
  failed: "Failed",
  canceled: "Canceled",
};
function jobResultText(job: ReferencePackKnownJob) {
  if (job.snapshot.status === "failed") {
    const problem = referencePackJobProblem(job.snapshot);
    return problem.kind === "rejected"
      ? "The operation failed. Review current catalog state before starting another operation."
      : referencePackProblemText(problem);
  }
  if (job.snapshot.status === "canceled")
    return "The server confirmed cancellation. This does not imply rollback of earlier effects.";
  const expected = {
    import: "reference_pack_imported",
    activate: "reference_pack_activated",
    disable: "reference_pack_disabled",
    reverify: "reference_pack_reverified",
    refresh_all: "reference_packs_refreshed",
    refresh_selected: "reference_packs_refreshed",
  }[job.command.kind];
  if (job.snapshot.result_summary?.code !== expected)
    return "The operation succeeded with a result this client cannot interpret. Reload the catalog to review current state.";
  const refs = (job.snapshot.result_summary.resource_refs ?? []).filter(
    (ref) => ref.kind === "reference_pack_version",
  );
  if (
    job.command.kind === "refresh_all" ||
    job.command.kind === "refresh_selected"
  )
    return "Refresh succeeded. Returned references may omit changed versions; reload the catalog to review current state.";
  if (
    refs.length !== 1 ||
    ("target" in job.command &&
      refs[0]?.id !== referencePackVersionRoute(job.command.target))
  )
    return "The operation succeeded, but its version reference could not be confirmed. Reload the catalog to review current state.";
  const target = refs[0] ? referencePackResultTarget(refs[0]) : null;
  return `${job.command.kind === "import" ? "Import verified a candidate; activate it separately." : "The operation succeeded for the exact version."} Version: ${target?.pack_key}@${target?.pack_version}`;
}
const panelStyle: CSSProperties = {
  containerType: "inline-size",
  boxSizing: "border-box",
  minInlineSize: 0,
  maxInlineSize: "100%",
  padding: "var(--ct-spacing-lg)",
  color: "var(--ct-colors-ink)",
  background: "var(--ct-colors-surface-1)",
  overflow: "auto",
  overflowWrap: "anywhere",
};
const rowStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
  minInlineSize: 0,
};
const titleStyle: CSSProperties = {
  fontSize: "var(--ct-typography-surface-title-fontSize)",
  margin: 0,
  marginInlineEnd: "auto",
};
const subtitleStyle: CSSProperties = { fontSize: "inherit", margin: 0 };
const fieldStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  minInlineSize: 0,
  flex: "1 1 12em",
};
const searchFieldStyle: CSSProperties = { ...fieldStyle, flex: "2 1 16em" };
const inputStyle: CSSProperties = {
  boxSizing: "border-box",
  minInlineSize: 0,
  inlineSize: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
  padding: "var(--ct-component-text-input-padding)",
};
const buttonStyle: CSSProperties = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  border: "var(--ct-component-button-secondary-border)",
  padding: "var(--ct-component-button-secondary-padding)",
  font: "inherit",
  maxInlineSize: "100%",
  whiteSpace: "normal",
  overflowWrap: "anywhere",
};
const primaryButtonStyle: CSSProperties = {
  ...buttonStyle,
  background: "var(--ct-component-button-primary-backgroundColor)",
  color: "var(--ct-component-button-primary-textColor)",
  border: "var(--ct-border-hairline)",
};
const sectionStyle: CSSProperties = {
  borderBlockStart: "var(--ct-border-hairline)",
  paddingBlock: "var(--ct-spacing-md)",
  marginBlockStart: "var(--ct-spacing-sm)",
  minInlineSize: 0,
};
const mutedStyle: CSSProperties = {
  color: "var(--ct-colors-ink-muted)",
  overflowWrap: "anywhere",
};
const identityStyle: CSSProperties = {
  minInlineSize: 0,
  overflowWrap: "anywhere",
  flex: "1 1 0",
};
const fileStyle: CSSProperties = { maxInlineSize: "100%", minInlineSize: 0 };
const errorStyle: CSSProperties = {
  color: "var(--ct-colors-semantic-conflict)",
  overflowWrap: "anywhere",
};
const catalogStyle: CSSProperties = {
  maxInlineSize: "100%",
  overflowX: "auto",
};
const tableStyle: CSSProperties = {
  borderCollapse: "collapse",
  inlineSize: "100%",
  tableLayout: "fixed",
};
const cellStyle: CSSProperties = {
  textAlign: "start",
  verticalAlign: "top",
  padding: "var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
  overflowWrap: "anywhere",
};
const captionStyle: CSSProperties = {
  textAlign: "start",
  paddingBlock: "var(--ct-spacing-sm)",
  fontWeight: "bold",
};
const listStyle: CSSProperties = { listStyle: "none", padding: 0, margin: 0 };
const visuallyHiddenStyle: CSSProperties = {
  position: "absolute",
  inlineSize: 1,
  blockSize: 1,
  overflow: "hidden",
  clipPath: "inset(50%)",
  whiteSpace: "nowrap",
};
