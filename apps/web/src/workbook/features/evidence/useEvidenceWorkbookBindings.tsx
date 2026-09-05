import type { GridDensity } from "@cartulary/grid-adapter";
import {
  type EvidenceAccessContext,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  workbookGridDensityMetrics,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import {
  buildEvidenceAccessPresentation,
  type EvidenceOperationKind,
  type EvidenceOperationState,
  evidenceOperationFeedback,
} from "../../evidence/evidenceAccessPresentation";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import { workbookSurfaceOverlayPanelStyle } from "../../layout/WorkbookSurfaceLayout";
import { buildEvidenceLifecycleViewModel } from "../../models/evidenceLifecycleViewModel";
import type {
  EvidenceCapabilityPort,
  EvidenceHandleOutcome,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookOwnerBinding } from "../../policies/workbookSurfacePolicy";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import {
  EvidenceAccessActions,
  evidenceButtonStyle,
  evidenceMessageStyle,
} from "./EvidenceAccessActions";

type Target = { readonly recordId: string; readonly rowVersion: number };
type Ticket = Target & { readonly lifetime: number; readonly sequence: number };
type Preview = Ticket & {
  readonly title: string;
  readonly href: string | null;
  readonly invoker: HTMLElement | null;
};
type RecordOperation = {
  readonly rowVersion: number;
  readonly state: EvidenceOperationState;
};
type MutationPorts = Pick<GenericSurfaceMutationController, "beginMutation">;

function titleFor(row: WorkbookQueryRow) {
  return (
    stringifyGridValue(row.cells["evidence.title"]?.value).trim() || "Evidence"
  );
}
function rowLifecycle(row: WorkbookQueryRow) {
  return buildEvidenceLifecycleViewModel({
    uploadStateSource: "evidence_projection",
    evidenceLifecycleState: row.cells["evidence.lifecycle_state"]?.value,
    objectBlobUploadState: row.cells["evidence.upload_state"]?.value,
  });
}
const unknownFailure: WorkbookOperationFailure = {
  kind: "terminal",
  message: "Evidence request failed.",
};

export function useEvidenceWorkbookBindings(input: {
  readonly mutationCommands: EvidenceCapabilityPort;
  readonly mutation: MutationPorts;
  readonly onRefresh: () => Promise<void> | void;
  readonly ownerBindings: readonly WorkbookOwnerBinding[];
  readonly resetKey: string;
  readonly rows: readonly WorkbookQueryRow[];
  readonly subjectRecordId: string | null;
  readonly canRead: boolean;
  readonly attachDisabledReason: string | null;
  readonly density: GridDensity;
  readonly onInspect: (recordId: string) => void;
  readonly onRestoreFocus: (recordId: string) => void;
}) {
  const active = input.ownerBindings.includes("evidence_lifecycle");
  const inputRef = useRef(input);
  inputRef.current = input;
  const lifetime = useRef(0);
  const sequence = useRef(0);
  const latestFeedback = useRef(new Map<string, number>());
  const latestDownload = useRef(new Map<string, number>());
  const pendingPreview = useRef<Preview | null>(null);
  const attachment = useRef<number | null>(null);
  const [operations, setOperations] = useState<Record<string, RecordOperation>>(
    {},
  );
  const [preview, setPreview] = useState<Preview | null>(null);
  const [attaching, setAttaching] = useState(false);
  const denied = useRef(false);
  const [accessLost, setAccessLost] = useState(false);
  const [announcement, setAnnouncement] = useState<{
    text: string;
    sequence: number;
    priority: "polite" | "assertive";
  } | null>(null);

  const discardFeedback = useCallback(
    (recordId: string, requestSequence: number) => {
      if (latestFeedback.current.get(recordId) !== requestSequence) return;
      latestFeedback.current.delete(recordId);
      setOperations((current) => {
        const next = { ...current };
        delete next[recordId];
        return next;
      });
      setAnnouncement((current) =>
        current?.sequence === requestSequence ? null : current,
      );
    },
    [],
  );
  const clearPreview = useCallback(
    (discardOperation = false) => {
      const previous = pendingPreview.current;
      pendingPreview.current = null;
      setPreview(null);
      if (discardOperation && previous !== null)
        discardFeedback(previous.recordId, previous.sequence);
    },
    [discardFeedback],
  );
  const resetLocalState = useCallback(() => {
    lifetime.current += 1;
    latestFeedback.current.clear();
    latestDownload.current.clear();
    clearPreview();
    setOperations({});
    setAnnouncement(null);
  }, [clearPreview]);
  const observed = useRef({
    resetKey: input.resetKey,
    subjectRecordId: input.subjectRecordId,
    canRead: input.canRead,
  });
  useLayoutEffect(() => {
    const previous = observed.current;
    observed.current = {
      resetKey: input.resetKey,
      subjectRecordId: input.subjectRecordId,
      canRead: input.canRead,
    };
    if (
      previous.resetKey !== input.resetKey ||
      previous.canRead !== input.canRead ||
      !active
    ) {
      resetLocalState();
      denied.current = false;
      setAccessLost(false);
      return;
    }
    for (const [recordId, requestSequence] of latestDownload.current) {
      const row = input.rows.find(
        (candidate) => candidate.record_id === recordId,
      );
      const retargeted =
        previous.subjectRecordId !== input.subjectRecordId &&
        recordId !== input.subjectRecordId;
      if (
        retargeted ||
        row === undefined ||
        !rowLifecycle(row).accessEligible
      ) {
        latestDownload.current.delete(recordId);
        discardFeedback(recordId, requestSequence);
      }
    }
    const target = pendingPreview.current;
    if (target !== null) {
      const row = input.rows.find(
        (candidate) => candidate.record_id === target.recordId,
      );
      const retargeted =
        previous.subjectRecordId !== input.subjectRecordId &&
        input.subjectRecordId !== target.recordId;
      if (
        retargeted ||
        row === undefined ||
        row.row_version !== target.rowVersion ||
        !rowLifecycle(row).accessEligible
      )
        clearPreview(true);
    }
    setOperations((current) => {
      const entries = Object.entries(current).filter(([id, value]) =>
        input.rows.some(
          (row) => row.record_id === id && row.row_version === value.rowVersion,
        ),
      );
      return entries.length === Object.keys(current).length
        ? current
        : Object.fromEntries(entries);
    });
  }, [
    active,
    clearPreview,
    discardFeedback,
    input.canRead,
    input.resetKey,
    input.rows,
    input.subjectRecordId,
    resetLocalState,
  ]);
  useLayoutEffect(
    () => () => {
      lifetime.current += 1;
      pendingPreview.current = null;
    },
    [],
  );

  const targetIsCurrent = useCallback((ticket: Ticket) => {
    const current = inputRef.current;
    return (
      ticket.lifetime === lifetime.current &&
      current.canRead &&
      !denied.current &&
      current.rows.some(
        (row) =>
          row.record_id === ticket.recordId &&
          row.row_version === ticket.rowVersion,
      )
    );
  }, []);
  const publish = useCallback(
    (ticket: Ticket, state: EvidenceOperationState) => {
      if (
        !targetIsCurrent(ticket) ||
        latestFeedback.current.get(ticket.recordId) !== ticket.sequence
      )
        return;
      const row = inputRef.current.rows.find(
        (candidate) => candidate.record_id === ticket.recordId,
      );
      if (row === undefined) return;
      setOperations((current) => ({
        ...current,
        [ticket.recordId]: { rowVersion: ticket.rowVersion, state },
      }));
      const feedback = evidenceOperationFeedback(state);
      if (feedback.announcement !== "none")
        setAnnouncement({
          sequence: ticket.sequence,
          text: `${titleFor(row)}: ${feedback.message}`,
          priority: feedback.announcement,
        });
    },
    [targetIsCurrent],
  );
  const begin = useCallback(
    (row: WorkbookQueryRow, operation: EvidenceOperationKind): Ticket => {
      const ticket = {
        recordId: row.record_id,
        rowVersion: row.row_version,
        lifetime: lifetime.current,
        sequence: ++sequence.current,
      };
      latestFeedback.current.set(row.record_id, ticket.sequence);
      publish(ticket, { kind: "pending", operation });
      return ticket;
    },
    [publish],
  );
  const invalidateAccess = useCallback(
    (failure: WorkbookOperationFailure) => {
      if (
        failure.kind !== "authentication_required" &&
        failure.kind !== "authorization_lost" &&
        failure.presentation?.family !== "permission_or_incident_access_loss"
      )
        return;
      clearPreview();
      denied.current = true;
      lifetime.current += 1;
      latestDownload.current.clear();
      latestFeedback.current.clear();
      setAccessLost(true);
      // The existing query/root error path owns removal of protected materialization and navigation.
      void (async () => {
        try {
          await inputRef.current.onRefresh();
        } catch {
          /* Existing query error handling owns refresh failures. */
        }
      })();
    },
    [clearPreview],
  );

  const issueHandle = useCallback(
    async (
      row: WorkbookQueryRow,
      kind: "preview" | "download",
      invoker: HTMLElement | null,
    ) => {
      if (
        !inputRef.current.canRead ||
        denied.current ||
        !rowLifecycle(row).accessEligible
      )
        return;
      if (kind === "preview") clearPreview(true);
      const ticket = begin(row, kind);
      if (kind === "preview") {
        const next = { ...ticket, title: titleFor(row), href: null, invoker };
        pendingPreview.current = next;
        setPreview(next);
      } else latestDownload.current.set(row.record_id, ticket.sequence);
      let outcome: EvidenceHandleOutcome;
      try {
        outcome = await inputRef.current.mutationCommands.issueHandle({
          evidenceRecordId: row.record_id,
          kind,
        });
      } catch {
        outcome = { kind: "rejected" as const, failure: unknownFailure };
      }
      const currentPreview = pendingPreview.current;
      if (
        !targetIsCurrent(ticket) ||
        (kind === "preview"
          ? currentPreview?.sequence !== ticket.sequence
          : latestDownload.current.get(row.record_id) !== ticket.sequence)
      )
        return;
      if (kind === "download") latestDownload.current.delete(row.record_id);
      if (outcome.kind === "rejected") {
        if (kind === "preview") clearPreview();
        publish(ticket, {
          kind: "rejected",
          operation: kind,
          failure: outcome.failure,
        });
        invalidateAccess(outcome.failure);
        return;
      }
      if (kind === "preview" && currentPreview !== null) {
        const next = { ...currentPreview, href: outcome.value.href };
        pendingPreview.current = next;
        setPreview(next);
      } else {
        const anchor = document.createElement("a");
        anchor.href = outcome.value.href;
        anchor.download = outcome.value.filename || "evidence";
        anchor.rel = "noopener";
        document.body.append(anchor);
        anchor.click();
        anchor.remove();
      }
      publish(ticket, { kind: "accepted", operation: kind });
    },
    [begin, clearPreview, invalidateAccess, publish, targetIsCurrent],
  );

  const attachFile = useCallback(
    async (row: WorkbookQueryRow, file: File) => {
      const current = inputRef.current;
      if (
        current.attachDisabledReason !== null ||
        !current.canRead ||
        denied.current ||
        attachment.current !== null
      )
        return;
      const ticket = begin(row, "attach");
      if (file.size <= 0) {
        publish(ticket, {
          kind: "rejected",
          operation: "attach",
          failure: unknownFailure,
        });
        return;
      }
      attachment.current = ticket.sequence;
      setAttaching(true);
      const finish = current.mutation.beginMutation();
      try {
        const outcome = await current.mutationCommands.attach({
          baseRowVersion: row.row_version,
          evidenceRecordId: row.record_id,
          file,
        });
        if (outcome.kind === "rejected") {
          publish(ticket, {
            kind: "rejected",
            operation: "attach",
            failure: outcome.failure,
          });
          if (ticket.lifetime === lifetime.current)
            invalidateAccess(outcome.failure);
          return;
        }
        publish(ticket, { kind: "accepted", operation: "attach" });
        if (ticket.lifetime === lifetime.current) await current.onRefresh();
      } catch {
        publish(ticket, {
          kind: "rejected",
          operation: "attach",
          failure: unknownFailure,
        });
      } finally {
        finish();
        attachment.current = null;
        setAttaching(false);
      }
    },
    [begin, invalidateAccess, publish],
  );

  const closePreview = useCallback(() => {
    const previous = pendingPreview.current;
    if (previous === null) return;
    clearPreview(true);
    if (
      previous.invoker?.isConnected &&
      previous.invoker.getClientRects().length > 0 &&
      !previous.invoker.closest("[inert]")
    )
      previous.invoker.focus({ preventScroll: true });
    else inputRef.current.onRestoreFocus(previous.recordId);
  }, [clearPreview]);

  const renderActions = (
    row: WorkbookQueryRow,
    context: EvidenceAccessContext,
  ) =>
    active ? (
      <EvidenceAccessActions
        access={buildEvidenceAccessPresentation(
          rowLifecycle(row),
          operations[row.record_id]?.state ?? null,
        )}
        attachDisabledReason={input.attachDisabledReason}
        attaching={attaching}
        canRead={input.canRead && !accessLost}
        context={context}
        recordId={row.record_id}
        title={titleFor(row)}
        onAttach={(file) => void attachFile(row, file)}
        onInspect={() => input.onInspect(row.record_id)}
        onIssue={(kind, invoker) => void issueHandle(row, kind, invoker)}
      />
    ) : null;
  const metrics = workbookGridDensityMetrics(input.density);
  const visiblePreview =
    active &&
    input.canRead &&
    !accessLost &&
    preview !== null &&
    preview.lifetime === lifetime.current &&
    input.rows.some(
      (row) =>
        row.record_id === preview.recordId &&
        row.row_version === preview.rowVersion,
    );
  const announcements = (
    <>
      {active ? (
        <div style={announcementStyle}>
          <div role="status" aria-live="polite" aria-atomic="true">
            {announcement?.priority === "polite" ? announcement.text : ""}
          </div>
          <div role="alert" aria-live="assertive" aria-atomic="true">
            {announcement?.priority === "assertive" ? announcement.text : ""}
          </div>
        </div>
      ) : null}
    </>
  );
  const overlay = (
    <>
      {visiblePreview ? (
        <section
          data-testid={evidencePreviewPanelTestId()}
          aria-label={`Evidence preview: ${preview.title}`}
          style={previewPanelStyle}
        >
          <div style={previewHeaderStyle}>
            <h2 style={previewTitleStyle}>{preview.title}</h2>
            <button
              type="button"
              style={evidenceButtonStyle}
              onClick={closePreview}
            >
              Close
            </button>
          </div>
          {preview.href === null ? (
            <p style={evidenceMessageStyle}>Opening preview…</p>
          ) : (
            <iframe
              key={preview.sequence}
              data-testid={evidencePreviewFrameTestId(preview.recordId)}
              src={preview.href}
              title={`Evidence preview ${preview.title}`}
              style={previewFrameStyle}
            />
          )}
        </section>
      ) : null}
    </>
  );
  return {
    actionsWidth: active
      ? Math.ceil(
          metrics.fontSizeCssPx * 32 + metrics.cellPaddingInlineCssPx * 2,
        )
      : 76,
    hasRecordActions: active,
    renderRowActions: (row: WorkbookQueryRow) => renderActions(row, "row"),
    renderInspector: (row: WorkbookQueryRow) => renderActions(row, "inspector"),
    overlay,
    announcements,
    closePreview: visiblePreview ? closePreview : undefined,
    resetLocalState,
  };
}

const previewPanelStyle = {
  ...workbookSurfaceOverlayPanelStyle,
  zIndex: 9,
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-panel-padding)",
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
} satisfies CSSProperties;
const previewHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "start",
  gap: "var(--ct-spacing-sm)",
  minInlineSize: 0,
} satisfies CSSProperties;
const previewTitleStyle = {
  margin: 0,
  fontSize: "var(--ct-typography-section-heading-fontSize)",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const previewFrameStyle = {
  inlineSize: "100%",
  blockSize: "min(28rem, 34vh)",
  minBlockSize: 0,
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
} satisfies CSSProperties;
const announcementStyle = {
  position: "absolute",
  inlineSize: 1,
  blockSize: 1,
  overflow: "hidden",
  clipPath: "inset(50%)",
  whiteSpace: "nowrap",
} satisfies CSSProperties;
