import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import { Pencil, Trash2 } from "lucide-react";
import { useEffect, useRef, useSyncExternalStore } from "react";
import {
  NetworkFlowActionGroup,
  NetworkFlowButton,
  NetworkFlowChromeStyles,
  NetworkFlowField,
  NetworkFlowTextInput,
  networkFlowChromeRootClassName,
} from "./NetworkFlowControls";
import type {
  NetworkFlowTableController,
  TableSnapshot,
} from "./NetworkFlowTableController";
import { normalizeTableDisplayName } from "./networkFlowTableOperation";
import { useNetworkFlowModalFocus } from "./useNetworkFlowModalFocus";

type Props = { readonly controller: NetworkFlowTableController };
export function TableLifecycleControls({ controller }: Props) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  if (state.activeTableId === null || state.hidden) return null;
  const id = state.activeTableId;
  return (
    <>
      {state.canRename ? (
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("rename-trigger")}
          title="Rename active table"
          variant="secondary"
          onClick={() => controller.openAction("rename", id)}
        >
          <Pencil aria-hidden="true" size={16} />
          Rename
        </NetworkFlowButton>
      ) : null}
      {state.canDelete ? (
        <NetworkFlowButton
          data-testid={networkAnalysisTestId("delete-trigger")}
          variant="danger"
          onClick={() => controller.openAction("delete", id)}
        >
          <Trash2 aria-hidden="true" size={16} />
          Delete
        </NetworkFlowButton>
      ) : null}
    </>
  );
}
export function NetworkFlowTableRecovery({ controller }: Props) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  if (
    state.hidden ||
    (state.operation === null && state.draft === null) ||
    state.presentation !== null
  )
    return null;
  return (
    <span className={networkFlowChromeRootClassName}>
      {state.operation ? (
        <NetworkFlowButton
          onClick={controller.reopenOperation}
          aria-label="Review retained table change"
        >
          {state.operation.status === "uncertain" ||
          state.operation.status === "pending"
            ? "Table recovery"
            : "Review table change"}
        </NetworkFlowButton>
      ) : null}
      {state.draft ? (
        <NetworkFlowButton
          onClick={controller.reopenDraft}
          aria-label={
            state.operation
              ? "Review table draft"
              : "Review retained table change"
          }
        >
          Review table draft
        </NetworkFlowButton>
      ) : null}
    </span>
  );
}
export function NetworkFlowTableSurface({ controller }: Props) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  if (
    state.hidden ||
    state.presentation === null ||
    (state.presentation === "draft" && state.draft === null) ||
    (state.presentation === "operation" && state.operation === null)
  )
    return null;
  return (
    <div
      className={networkFlowChromeRootClassName}
      style={{ display: "contents" }}
    >
      <NetworkFlowChromeStyles />
      <TableDialog
        key={
          state.presentation === "draft"
            ? `draft-${state.draft?.action}-${state.draft?.target.network_flow_table_id}`
            : `operation-${state.operation?.attempt.transactionId}`
        }
        controller={controller}
        state={state}
      />
    </div>
  );
}
function TableDialog({
  controller,
  state,
}: Props & { readonly state: TableSnapshot }) {
  const draft = state.presentation === "draft" ? state.draft : null;
  const operation = state.operation;
  const target = draft?.target ?? operation?.attempt.target;
  const action = draft?.action ?? operation?.attempt.action ?? "rename";
  const rename = action === "rename";
  const name = draft?.name ?? operation?.attempt.normalizedName ?? "";
  const normalized = normalizeTableDisplayName(name);
  const current = state.tables.find(
    (table) => table.network_flow_table_id === target?.network_flow_table_id,
  );
  const pending = operation?.status === "pending";
  const uncertain = operation?.status === "uncertain";
  const unresolved = pending || uncertain;
  const sameAttempt = draft === null || operation?.dialogId === draft.id;
  const permitted = rename ? state.canRename : state.canDelete;
  const feedback =
    (sameAttempt && unresolved ? operation?.failure?.message : null) ??
    draft?.error ??
    (sameAttempt ? operation?.failure?.message : null);
  const inputId = rename ? "rename-input" : "delete-confirmation";
  const reviewRef = useRef<HTMLDivElement | null>(null);
  const modalFocus = useNetworkFlowModalFocus<HTMLFormElement>({
    initialFocusTestId: networkAnalysisTestId(inputId),
    onDismiss: controller.closeDialog,
    restoreFallbackFocus: () => {
      const tab = document.querySelector<HTMLElement>(
        '[role="tab"][aria-selected="true"][data-network-flow-table-id]',
      );
      if (!tab) return false;
      tab.focus();
      return true;
    },
  });
  useEffect(() => {
    if (draft?.reviewRequired || feedback) reviewRef.current?.focus();
  }, [draft?.reviewRequired, feedback]);
  return (
    <div className="network-flow-dialog-backdrop">
      {/* biome-ignore lint/a11y/useAriaPropsSupportedByRole: Both literal branches are modal dialog roles. */}
      <form
        ref={modalFocus.dialogRef}
        className="network-flow-dialog"
        role={rename ? "dialog" : "alertdialog"}
        aria-modal="true"
        aria-labelledby="network-flow-table-title"
        aria-describedby="network-flow-table-description"
        data-testid={networkAnalysisTestId(
          rename ? "rename-dialog" : "delete-dialog",
        )}
        onKeyDown={modalFocus.onKeyDown}
        onSubmit={(event) => {
          event.preventDefault();
          if (draft) void controller.submit();
        }}
      >
        <h3 id="network-flow-table-title">
          {rename ? "Rename" : "Delete"} Network Flow table
        </h3>
        <p id="network-flow-table-description">
          {rename ? (
            <>
              Captured table: <strong>{target?.display_name}</strong>, version{" "}
              {target?.table_version}.
            </>
          ) : (
            <>
              This soft-deletes <strong>{target?.display_name}</strong> and
              makes its rows, diagnostics, affected graph results, and cursors
              unavailable. Type the exact table name to confirm.
            </>
          )}
        </p>
        {draft ? (
          <NetworkFlowField
            htmlFor={`network-flow-${inputId}`}
            label={rename ? "Display name" : "Confirm table name"}
            help={
              rename
                ? normalized.ok
                  ? `${normalized.scalarCount} / 64 Unicode characters after normalization.`
                  : "Use 1–64 Unicode characters after normalization, without control characters."
                : `Type ${target?.display_name} exactly.`
            }
            helpId="network-flow-table-field-help"
          >
            <NetworkFlowTextInput
              id={`network-flow-${inputId}`}
              data-testid={networkAnalysisTestId(inputId)}
              value={rename ? name : draft.confirmation}
              readOnly={!permitted || unresolved}
              aria-invalid={feedback !== null || undefined}
              aria-describedby="network-flow-table-field-help network-flow-table-feedback"
              onChange={(event) =>
                rename
                  ? controller.setName(event.currentTarget.value)
                  : controller.setConfirmation(event.currentTarget.value)
              }
            />
          </NetworkFlowField>
        ) : null}
        <div
          ref={reviewRef}
          tabIndex={-1}
          id="network-flow-table-feedback"
          role={feedback ? "alert" : "status"}
        >
          {feedback ? <p>{feedback}</p> : null}
          {draft?.reviewRequired ? (
            <p>
              Current table:{" "}
              {current
                ? `${current.display_name}, version ${current.table_version}.`
                : "Metadata is unavailable."}{" "}
              Review this information before submitting.
            </p>
          ) : null}
          {!permitted && draft ? (
            <p>
              Your current access does not permit this action. The draft remains
              available to copy.
            </p>
          ) : null}
          {unresolved ? (
            <p>
              {pending
                ? "Waiting for acknowledgement. Closing stops observation; the server may still complete the request."
                : "The outcome is unknown. Replay sends the same transaction and request; it does not create a new change."}
            </p>
          ) : null}
          {!sameAttempt && unresolved ? (
            <p>
              Resolve the earlier table change before submitting this draft.
            </p>
          ) : null}
          {draft === null && operation?.status === "acknowledged" ? (
            <p>
              The{" "}
              {operation.attempt.action === "delete" ? "deletion" : "rename"}{" "}
              was acknowledged at table version{" "}
              {operation.receipt?.table_version}. This receipt records that
              operation; current table metadata is loaded separately.
              {state.loadState === "error"
                ? " Completed; refresh unavailable."
                : ""}
            </p>
          ) : null}
        </div>
        <NetworkFlowActionGroup>
          <NetworkFlowButton
            variant="secondary"
            data-testid={networkAnalysisTestId(
              rename ? "rename-cancel" : "delete-cancel",
            )}
            onClick={controller.closeDialog}
          >
            {unresolved || draft === null ? "Close" : "Cancel"}
          </NetworkFlowButton>
          {draft?.reviewRequired ? (
            <>
              <NetworkFlowButton
                variant="secondary"
                disabled={
                  state.loadState === "loading" ||
                  state.loadState === "refreshing"
                }
                onClick={() => void controller.loadTables()}
              >
                Refresh table metadata
              </NetworkFlowButton>
              <NetworkFlowButton
                disabled={
                  !permitted ||
                  unresolved ||
                  state.loadState !== "ready" ||
                  !current
                }
                onClick={controller.reviewCurrent}
              >
                Review current table
              </NetworkFlowButton>
            </>
          ) : null}
          {uncertain ? (
            <NetworkFlowButton
              disabled={
                operation?.attempt.action === "rename"
                  ? !state.canRename
                  : !state.canDelete
              }
              onClick={() => void controller.replay()}
            >
              Replay exact request
            </NetworkFlowButton>
          ) : null}
          {draft ? (
            <NetworkFlowButton
              type="submit"
              variant={rename ? "secondary" : "danger"}
              data-testid={networkAnalysisTestId(
                rename ? "rename-submit" : "delete-confirm",
              )}
              pending={pending && sameAttempt}
              disabled={
                !permitted ||
                unresolved ||
                draft.reviewRequired ||
                (!rename && draft.confirmation !== draft.target.display_name)
              }
            >
              {pending && sameAttempt
                ? rename
                  ? "Renaming…"
                  : "Deleting…"
                : rename
                  ? "Rename"
                  : "Delete table"}
            </NetworkFlowButton>
          ) : (
            <NetworkFlowButton
              variant="secondary"
              onClick={() => void controller.loadTables()}
            >
              Refresh table metadata
            </NetworkFlowButton>
          )}
        </NetworkFlowActionGroup>
      </form>
    </div>
  );
}
