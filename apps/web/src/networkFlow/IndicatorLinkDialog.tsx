import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import { useEffect, useRef, useSyncExternalStore } from "react";
import {
  useWorkbookRecoveryPresentation,
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../shared/WorkbookRecoveryBoundary";
import {
  NetworkFlowActionGroup,
  NetworkFlowButton,
  NetworkFlowChoice,
  NetworkFlowChromeStyles,
  NetworkFlowField,
  NetworkFlowSelect,
  NetworkFlowTextInput,
  networkFlowChromeRootClassName,
} from "./NetworkFlowControls";
import type {
  IndicatorLinkSnapshot,
  NetworkFlowIndicatorLinkController,
} from "./NetworkFlowIndicatorLinkController";
import { networkFlowIndicatorRecoveryItems } from "./networkFlowIndicatorRecoveryItems";

export function NetworkFlowIndicatorLinkSurface({
  controller,
}: {
  readonly controller: NetworkFlowIndicatorLinkController;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const selected = useWorkbookRecoverySource(
    "network-indicator",
    networkFlowIndicatorRecoveryItems(state),
    {
      activate: (id) => {
        if (String(state.attempt?.workId) === id) controller.reopen();
        else if (String(state.draft?.workId) === id) controller.showDraft();
        else return false;
        return true;
      },
      detach: controller.dismiss,
    },
  );
  const presented =
    state.hidden || state.presentation === null
      ? null
      : state.presentation === "draft"
        ? state.draft?.workId
        : state.attempt?.workId;
  useWorkbookRecoveryPresentation(
    "network-indicator",
    presented == null ? null : String(presented),
    selected,
  );
  return (
    <WorkbookRecoveryDetail source="network-indicator" item={selected}>
      <div className={networkFlowChromeRootClassName}>
        <NetworkFlowChromeStyles />
        {!state.hidden && state.presentation !== null ? (
          <IndicatorLinkDialog
            key={selected}
            controller={controller}
            state={state}
          />
        ) : null}
      </div>
    </WorkbookRecoveryDetail>
  );
}

function IndicatorLinkDialog({
  controller,
  state,
}: {
  readonly controller: NetworkFlowIndicatorLinkController;
  readonly state: IndicatorLinkSnapshot;
}) {
  const draft = state.draft;
  const attempt = state.attempt;
  const editing = state.presentation === "draft" && draft !== null;
  const candidate = editing ? draft.candidate : attempt?.candidate;
  const discovery = useSyncExternalStore(
    controller.targets.subscribe,
    controller.targets.getSnapshot,
  );
  const heading = useRef<HTMLHeadingElement | null>(null);
  const panelRef = useRef<HTMLElement>(null);
  const settlement = state.settlement;
  const pending =
    settlement?.kind === "pending" || settlement?.kind === "queued";
  const unresolved = pending || settlement?.kind === "uncertain";
  const confirmed =
    settlement?.kind === "confirmed" || settlement?.kind === "reused";
  useEffect(() => {
    if (
      editing &&
      draft.targetMode === "existing_indicator" &&
      discovery.phase === "idle" &&
      draft.applicable
    )
      controller.targets.load();
  }, [controller, editing, draft, discovery.phase]);
  useEffect(() => {
    const field = editing ? draft.feedback?.field : null;
    if (field === null || field === undefined) return;
    panelRef.current
      ?.querySelector<HTMLElement>(
        field === "target"
          ? "#network-flow-existing-indicator-id"
          : "#network-flow-indicator-confirmation",
      )
      ?.focus();
  }, [editing, draft?.feedback]);
  if (candidate === undefined) return null;
  const targetError =
    editing && draft.feedback?.field === "target"
      ? draft.feedback.message
      : null;
  const confirmationError =
    editing && draft.feedback?.field === "confirmation"
      ? draft.feedback.message
      : null;
  return (
    <div>
      <section
        ref={panelRef}
        aria-labelledby="network-flow-indicator-link-title"
        data-testid={networkAnalysisTestId("indicator-link-dialog")}
        className="network-flow-recovery-detail"
      >
        <h3 id="network-flow-indicator-link-title" ref={heading} tabIndex={-1}>
          Link Core Indicator
        </h3>
        <p>{candidate.label}</p>
        <code className="network-flow-link-value">
          {candidate.candidateValue}
        </code>
        <p>
          Link this canonical IP endpoint to an indicator in this incident.
          Linking creates or reuses a binding; flow rows and indicator
          observations stay unchanged.
        </p>
        {state.limitError !== null ? (
          <div>
            <p role="alert">{state.limitError}</p>
            <NetworkFlowButton onClick={controller.loadLimit}>
              Retry link limits
            </NetworkFlowButton>
          </div>
        ) : null}
        {editing ? (
          <form
            className="network-flow-dialog-form"
            noValidate
            onSubmit={(event) => {
              event.preventDefault();
              controller.submit();
            }}
          >
            {draft.feedback !== null && draft.feedback.field === null ? (
              <p role="alert">{draft.feedback.message}</p>
            ) : null}
            {unresolved ? (
              <div role="status">
                <p>
                  An earlier link request is unresolved. Recover it before
                  submitting another link.
                </p>
                <NetworkFlowButton onClick={controller.reopen}>
                  Review earlier request
                </NetworkFlowButton>
              </div>
            ) : null}
            <fieldset className="network-flow-link-targets">
              <legend>Indicator target</legend>
              <label htmlFor="network-flow-indicator-target-create">
                <NetworkFlowChoice
                  id="network-flow-indicator-target-create"
                  name="network-flow-indicator-target"
                  type="radio"
                  checked={draft.targetMode === "create_indicator"}
                  disabled={!state.writable}
                  onChange={() =>
                    controller.editDraft({ targetMode: "create_indicator" })
                  }
                />{" "}
                Create or reuse indicator
              </label>
              <label htmlFor="network-flow-indicator-target-existing">
                <NetworkFlowChoice
                  id="network-flow-indicator-target-existing"
                  name="network-flow-indicator-target"
                  type="radio"
                  checked={draft.targetMode === "existing_indicator"}
                  disabled={!state.writable}
                  onChange={() =>
                    controller.editDraft({ targetMode: "existing_indicator" })
                  }
                />{" "}
                Existing indicator
              </label>
            </fieldset>
            {draft.targetMode === "existing_indicator" ? (
              <section
                aria-label="Existing indicator selection"
                className="network-flow-dialog-form"
              >
                <NetworkFlowField
                  htmlFor="network-flow-indicator-target-list"
                  label="Compatible indicators on this page"
                  help="One page of up to 100 visible atomic IP indicators is checked for this exact value."
                  helpId="network-flow-indicator-discovery-help"
                >
                  <NetworkFlowSelect
                    id="network-flow-indicator-target-list"
                    aria-describedby={`network-flow-indicator-discovery-help${targetError ? " network-flow-indicator-target-error" : ""}`}
                    aria-invalid={targetError ? true : undefined}
                    disabled={!state.writable || discovery.phase === "loading"}
                    value={
                      discovery.items.some(
                        (item) => item.id === draft.existingId,
                      )
                        ? draft.existingId
                        : ""
                    }
                    onChange={(event) =>
                      controller.editDraft({
                        existingId: event.currentTarget.value,
                      })
                    }
                  >
                    <option value="">Select an indicator</option>
                    {discovery.items.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.value} · {item.id}
                      </option>
                    ))}
                  </NetworkFlowSelect>
                </NetworkFlowField>
                <div aria-live="polite" aria-atomic="true">
                  {discovery.phase === "loading"
                    ? "Loading compatible indicators…"
                    : discovery.phase === "ready" &&
                        discovery.items.length === 0
                      ? "No compatible indicators were found on this page."
                      : null}
                </div>
                {discovery.error ? <p role="alert">{discovery.error}</p> : null}
                <NetworkFlowActionGroup>
                  {discovery.phase === "failed" ? (
                    <NetworkFlowButton onClick={controller.targets.retry}>
                      Retry indicator page
                    </NetworkFlowButton>
                  ) : null}
                  {discovery.nextCursor !== null ? (
                    <NetworkFlowButton onClick={controller.targets.next}>
                      Check next indicator page
                    </NetworkFlowButton>
                  ) : null}
                  {discovery.phase === "ready" ? (
                    <NetworkFlowButton
                      onClick={() => controller.targets.load()}
                    >
                      Check first indicator page
                    </NetworkFlowButton>
                  ) : null}
                </NetworkFlowActionGroup>
                <NetworkFlowField
                  htmlFor="network-flow-existing-indicator-id"
                  label="Known indicator ID (alternative)"
                  error={targetError ?? undefined}
                  errorId="network-flow-indicator-target-error"
                  help="The server checks current visibility and exact atomic IP compatibility."
                  helpId="network-flow-indicator-target-help"
                >
                  <NetworkFlowTextInput
                    id="network-flow-existing-indicator-id"
                    data-testid={networkAnalysisTestId(
                      "indicator-link-existing-id",
                    )}
                    value={draft.existingId}
                    readOnly={!state.writable}
                    aria-invalid={targetError ? true : undefined}
                    aria-describedby={`network-flow-indicator-target-help${targetError ? " network-flow-indicator-target-error" : ""}`}
                    onChange={(event) =>
                      controller.editDraft({
                        existingId: event.currentTarget.value,
                      })
                    }
                    autoComplete="off"
                    spellCheck={false}
                  />
                </NetworkFlowField>
              </section>
            ) : targetError ? (
              <p role="alert">{targetError}</p>
            ) : null}
            <NetworkFlowField
              htmlFor="network-flow-indicator-confirmation"
              label="Confirm exact canonical value"
              error={confirmationError ?? undefined}
              errorId="network-flow-indicator-confirmation-error"
              help="Match the value byte for byte. Changing the source or target requires a new confirmation."
              helpId="network-flow-indicator-confirmation-help"
            >
              <NetworkFlowTextInput
                id="network-flow-indicator-confirmation"
                data-testid={networkAnalysisTestId(
                  "indicator-link-confirmation",
                )}
                value={draft.confirmation}
                readOnly={!state.writable || !draft.applicable}
                aria-invalid={confirmationError ? true : undefined}
                aria-describedby={`network-flow-indicator-confirmation-help${confirmationError ? " network-flow-indicator-confirmation-error" : ""}`}
                onChange={(event) =>
                  controller.editDraft({
                    confirmation: event.currentTarget.value,
                  })
                }
                autoComplete="off"
                spellCheck={false}
              />
            </NetworkFlowField>
            {candidate.selector.kind === "graph_edge" ||
            candidate.selector.kind === "graph_vertex" ? (
              <p>
                Graph links retain at most {state.sourceLimit} source
                references. A contributor page may show only part of the
                matching source set.
              </p>
            ) : (
              <p>
                {candidate.sourceRefs.length} accepted source row
                {candidate.sourceRefs.length === 1 ? "" : "s"} selected.
              </p>
            )}
            <NetworkFlowActionGroup>
              <NetworkFlowButton
                data-testid={networkAnalysisTestId("indicator-link-cancel")}
                onClick={controller.dismiss}
              >
                Close
              </NetworkFlowButton>
              <NetworkFlowButton
                data-testid={networkAnalysisTestId("indicator-link-submit")}
                type="submit"
                variant="primary"
                disabled={!state.writable || !draft.applicable || unresolved}
              >
                Link Indicator
              </NetworkFlowButton>
            </NetworkFlowActionGroup>
          </form>
        ) : (
          <div className="network-flow-dialog-form">
            <p>
              Captured target:{" "}
              {attempt?.request.target.mode === "existing_indicator" ? (
                <code>{attempt.request.target.indicator_id}</code>
              ) : (
                "Create or reuse a compatible indicator"
              )}
              .
            </p>
            {pending ? (
              <p role="status">
                {settlement?.kind === "queued"
                  ? "Waiting to dispatch this captured request…"
                  : "Linking… Waiting for the binding receipt."}
              </p>
            ) : null}
            {settlement !== null && "feedback" in settlement ? (
              <p role="alert">{settlement.feedback.message}</p>
            ) : null}
            {confirmed ? (
              <section
                aria-label="Indicator binding receipt"
                className="network-flow-dialog-form"
              >
                <strong role="status">
                  {settlement.kind === "reused"
                    ? "Existing indicator binding reused."
                    : "Indicator binding created."}
                </strong>
                {settlement.replay ? (
                  <p>Recovered by replaying the exact original request.</p>
                ) : null}
                <p>
                  Binding{" "}
                  <code>
                    {
                      settlement.receipt.binding
                        .network_flow_indicator_binding_id
                    }
                  </code>
                </p>
                <p>
                  Indicator{" "}
                  <code>
                    {
                      settlement.receipt.binding.target_indicator_ref
                        .indicator_id
                    }
                  </code>
                </p>
                <p>
                  {settlement.receipt.binding.source_row_refs.length} source
                  {settlement.receipt.binding.source_row_refs.length === 1
                    ? " reference"
                    : " references"}{" "}
                  retained
                  {` from ${settlement.receipt.binding.source_row_refs_total_count} matching row${settlement.receipt.binding.source_row_refs_total_count === 1 ? "" : "s"}${settlement.receipt.binding.source_row_refs_truncated ? " (truncated)" : ""}`}
                  .
                </p>
                {settlement.kind === "reused" ? (
                  <p>
                    This receipt retains the binding's original creation
                    metadata.
                  </p>
                ) : null}
              </section>
            ) : null}
            {!confirmed ? (
              <p>
                Closing this dialog only dismisses it. Server work may continue.
                Use indicator link recovery to review this captured request.
              </p>
            ) : null}
            <NetworkFlowActionGroup>
              {confirmed ? (
                <NetworkFlowButton variant="primary" onClick={controller.done}>
                  Done
                </NetworkFlowButton>
              ) : (
                <NetworkFlowButton
                  data-testid={networkAnalysisTestId("indicator-link-cancel")}
                  onClick={controller.dismiss}
                >
                  Close
                </NetworkFlowButton>
              )}
              {!pending && !confirmed ? (
                <NetworkFlowButton
                  disabled={!state.writable}
                  variant="primary"
                  onClick={controller.replay}
                >
                  Replay exact request
                </NetworkFlowButton>
              ) : null}
              {!confirmed && draft !== null ? (
                <NetworkFlowButton onClick={controller.showDraft}>
                  Edit draft
                </NetworkFlowButton>
              ) : null}
            </NetworkFlowActionGroup>
            {!confirmed && !pending ? (
              <details>
                <summary>Forget local recovery</summary>
                <p>
                  This removes the recovery copy from this workbook session.
                  Indicators or bindings may already exist on the server. It
                  does not undo linking.
                </p>
                <NetworkFlowButton
                  variant="danger"
                  onClick={controller.abandonRecovery}
                >
                  Forget this request locally
                </NetworkFlowButton>
              </details>
            ) : null}
          </div>
        )}
      </section>
    </div>
  );
}
