import type { GridDensity } from "@cartulary/grid-adapter";
import {
  type EvidenceAccessContext,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  workbookGridDensityMetrics,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  useContext,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import { buildEvidenceAccessPresentation } from "../../evidence/evidenceAccessPresentation";
import {
  savedInspectorRegion,
  type WorkbookInspectorRegion,
} from "../../inspector/presentation/WorkbookInspectorPanelContent";
import { buildEvidenceLifecycleViewModel } from "../../models/evidenceLifecycleViewModel";
import type { EvidenceCapabilityPort } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOwnerBinding } from "../../policies/workbookSurfacePolicy";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import {
  EvidenceAccessActions,
  evidenceMessageStyle,
} from "./EvidenceAccessActions";

import { EvidenceAttachmentContext } from "./EvidenceAttachmentContext";
import { EvidenceAttachmentEntry } from "./EvidenceAttachmentEntry";
import { EvidenceFileRecovery } from "./EvidenceFileRecovery";
import { useEvidenceHandleAccess } from "./useEvidenceHandleAccess";
import type { EvidenceAttachmentSnapshot } from "./WorkbookEvidenceAttachmentOwner";

const emptyAttachments: readonly EvidenceAttachmentSnapshot[] = [];
const noAttachments = () => emptyAttachments;
const noSubscription = () => () => {};

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
export function useEvidenceWorkbookBindings(input: {
  readonly mutationCommands: EvidenceCapabilityPort;
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
  const owner = useContext(EvidenceAttachmentContext);
  const retained = useSyncExternalStore(
    owner?.subscribe ?? noSubscription,
    owner?.getSnapshot ?? noAttachments,
  );
  const presentationToken = useRef(Symbol());
  const access = useEvidenceHandleAccess({
    port: input.mutationCommands,
    canRead: active && input.canRead,
    scopeKey: input.resetKey,
    isCurrent: (target) =>
      input.rows.some(
        (row) =>
          row.record_id === target.recordId &&
          target.identity ===
            JSON.stringify([row.row_version, input.subjectRecordId]) &&
          rowLifecycle(row).accessEligible,
      ),
    onAccessFailure: input.onRefresh,
    onRestoreFocus: (target) => input.onRestoreFocus(target.recordId),
  });
  const targetFor = (row: WorkbookQueryRow) => ({
    recordId: row.record_id,
    identity: JSON.stringify([row.row_version, input.subjectRecordId]),
    title: titleFor(row),
  });
  useLayoutEffect(() => {
    if (!owner || !active || !input.canRead) return;
    for (const row of input.rows) owner.observe(row.record_id, row.row_version);
    if (input.subjectRecordId)
      owner.attach(presentationToken.current, input.subjectRecordId);
    const token = presentationToken.current;
    return () => owner.detach(token);
  }, [owner, active, input.canRead, input.rows, input.subjectRecordId]);
  const attachFiles = (row: WorkbookQueryRow, files: readonly File[]) => {
    if (
      !owner ||
      input.attachDisabledReason !== null ||
      !input.canRead ||
      access.accessLost
    )
      return;
    const message = owner.begin(row, files);
    if (message) access.announce(message, "polite");
  };

  const renderActions = (
    row: WorkbookQueryRow,
    context: EvidenceAccessContext,
  ) =>
    active ? (
      <EvidenceAccessActions
        access={buildEvidenceAccessPresentation(
          rowLifecycle(row),
          access.operations[row.record_id]?.state ?? null,
        )}
        canRead={input.canRead && !access.accessLost}
        context={context}
        recordId={row.record_id}
        title={titleFor(row)}
        onInspect={() => input.onInspect(row.record_id)}
        onIssue={(kind, invoker) =>
          void access.issue(targetFor(row), kind, invoker)
        }
      />
    ) : null;
  const renderAttachment = (
    row: WorkbookQueryRow,
    context: EvidenceAccessContext,
  ) => (
    <EvidenceAttachmentEntry
      compact={context === "row"}
      title={titleFor(row)}
      testId={evidenceAttachFileInputTestId(row.record_id, context)}
      disabledReason={
        !input.canRead || access.accessLost
          ? "Evidence access is unavailable."
          : (input.attachDisabledReason ??
            (owner ? null : "File attachment is unavailable."))
      }
      busy={retained.some(
        (entry) => entry.recordId === row.record_id && entry.busy,
      )}
      onAttach={(files) => attachFiles(row, files)}
    />
  );
  const metrics = workbookGridDensityMetrics(input.density);
  const renderFileRecovery = (
    entry: (typeof retained)[number],
    presentation: "grid" | "inspector",
  ) => {
    if (!owner || !input.canRead) return null;
    const row = input.rows.find((row) => row.record_id === entry.recordId);
    return (
      <EvidenceFileRecovery
        {...entry}
        key={entry.recordId}
        presentation={presentation}
        source={row ? titleFor(row) : "Original Evidence record"}
        onConfirmReview={() => owner.confirmReview(entry.recordId)}
        onReview={() => void owner.review(entry.recordId)}
        onResume={() => void owner.resume(entry.recordId)}
        onFreshSlot={() => owner.freshSlot(entry.recordId)}
        onNewId={() => owner.newRequestId(entry.recordId)}
        onDiscard={() => owner.discard(entry.recordId)}
        onRefresh={() => void owner.refresh(entry.recordId)}
      />
    );
  };
  const announcements = (
    <>
      {active && owner
        ? retained.map((entry) => renderFileRecovery(entry, "grid"))
        : null}
      {active ? (
        <div style={announcementStyle}>
          <div role="status" aria-live="polite" aria-atomic="true">
            {access.announcement?.priority === "polite"
              ? access.announcement.text
              : ""}
          </div>
          <div role="alert" aria-live="assertive" aria-atomic="true">
            {access.announcement?.priority === "assertive"
              ? access.announcement.text
              : ""}
          </div>
        </div>
      ) : null}
    </>
  );
  const overlay = access.overlay;
  const inspectorRegions = (
    row: WorkbookQueryRow,
  ): [WorkbookInspectorRegion, ...WorkbookInspectorRegion[]] | null => {
    if (!active) return null;
    if (!input.canRead || access.accessLost)
      return [
        {
          id: "evidence-access",
          kind: "snapshot",
          model: { access: "concealed" },
        },
      ];
    const operation =
      access.operations[row.record_id]?.identity === targetFor(row).identity
        ? access.operations[row.record_id]?.state
        : null;
    const presentation = buildEvidenceAccessPresentation(
      rowLifecycle(row),
      operation ?? null,
    );
    return [
      savedInspectorRegion("evidence-metadata", {
        kind: "populated",
        content: (
          <>
            <p style={evidenceMessageStyle}>Evidence information</p>
            <dl>
              <dt>Lifecycle</dt>
              <dd>{presentation.lifecycleLabel}</dd>
              <dt>File</dt>
              <dd>{presentation.uploadLabel}</dd>
            </dl>
          </>
        ),
      }),
      {
        id: "evidence-access",
        kind: "snapshot",
        model: {
          access: "readable",
          messageId: evidenceAccessMessageTestId(row.record_id, "inspector"),
          announcement: "owner",
          data:
            operation?.kind === "pending"
              ? { state: "initial_loading" }
              : operation?.kind === "accepted"
                ? {
                    state: "ready",
                    content: {
                      kind: "populated",
                      content: (
                        <p
                          id={evidenceAccessMessageTestId(
                            row.record_id,
                            "inspector",
                          )}
                          data-testid={evidenceAccessMessageTestId(
                            row.record_id,
                            "inspector",
                          )}
                        >
                          {presentation.message}
                        </p>
                      ),
                    },
                  }
                : {
                    state: "unavailable",
                    cause:
                      operation?.kind === "rejected"
                        ? "load_failed"
                        : presentation.canPreview || presentation.canDownload
                          ? "not_requested"
                          : "owner_blocked",
                    message: presentation.message,
                  },
          commands: renderActions(row, "inspector"),
        },
      },
      savedInspectorRegion("evidence-attachment", {
        kind: "populated",
        content: (
          <section aria-label="Evidence attachment and recovery">
            {renderAttachment(row, "inspector")}
            {retained
              .filter((entry) => entry.recordId === row.record_id)
              .map((entry) => renderFileRecovery(entry, "inspector"))}
          </section>
        ),
      }),
    ];
  };
  return {
    actionsWidth: active
      ? Math.ceil(
          metrics.fontSizeCssPx * 32 + metrics.cellPaddingInlineCssPx * 2,
        )
      : 76,
    hasRecordActions: active,
    renderRowActions: (row: WorkbookQueryRow) =>
      active ? (
        <div
          style={{
            display: "flex",
            gap: "var(--ct-spacing-xs)",
            alignItems: "center",
            blockSize: "100%",
          }}
        >
          {renderActions(row, "row")}
          {renderAttachment(row, "row")}
        </div>
      ) : null,
    inspectorRegions,
    overlay,
    announcements,
    closePreview: access.closePreview,
    resetLocalState: access.reset,
  };
}

const announcementStyle = {
  position: "absolute",
  inlineSize: 1,
  blockSize: 1,
  overflow: "hidden",
  clipPath: "inset(50%)",
  whiteSpace: "nowrap",
} satisfies CSSProperties;
