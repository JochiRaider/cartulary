import { useEffect, useLayoutEffect, useRef } from "react";
import type { InspectorRecordHistoryAction } from "../inspector/inspectorCapabilityResolver";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import { useWorkbookRecordHistoryController } from "../inspector/useWorkbookRecordHistoryController";
import { WorkbookRecordHistoryPanel } from "../inspector/WorkbookInspectorRecordHistory";
import { workbookRecordHistoryLoadedData } from "../inspector/workbookRecordHistoryModel";
import { HistoryLookupFeedback } from "./HistoryLookupFeedback";
import { useWorkbookHistoryRuntime } from "./WorkbookHistoryContext";
import type { HistoryReviewLocator } from "./workbookHistoryReview";

const actions: ReadonlySet<InspectorRecordHistoryAction> = new Set([
  "rollback",
  "restore",
  "delete",
]);

/** Shared History presentation bound to a read locator, never to grid membership. */
export function WorkbookHistoryReview({
  locator,
  onClose,
}: {
  readonly locator: HistoryReviewLocator;
  readonly onClose: () => void;
}) {
  const runtime = useWorkbookHistoryRuntime();
  const controller = useWorkbookRecordHistoryController({
    subject: null,
    locator,
    canMutate: true,
    ownerEffects: {
      deleteAccepted: () => {},
      restoreAccepted: () => {},
      rollbackAccepted: () => {},
      refresh: () => runtime?.history.refreshSurface(locator.viewSchemaId),
    },
  });
  const open = controller.commands.open;
  useEffect(() => {
    open();
  }, [open]);
  const heading = useRef<HTMLHeadingElement>(null);
  const review = useRef<HTMLElement>(null);
  const hasLoadedMatch =
    workbookRecordHistoryLoadedData(controller.snapshot)?.items.some(
      (item) => item.change_set_id === locator.changeSetId,
    ) ?? false;
  useLayoutEffect(() => {
    // Only the explicit activation owns focus; responses never reclaim it.
    if (document.activeElement?.closest("#workbook-recovery-panel"))
      heading.current?.focus({ preventScroll: true });
  }, []);
  return (
    <section ref={review} aria-label="Review this change">
      <h3 ref={heading} tabIndex={-1}>
        Timeline change review
      </h3>
      <p>Record at completion: {locator.label}</p>
      <p>Marked entries belong to this change. Older pages may contain more.</p>
      <WorkbookInspectorActionButton onClick={onClose}>
        Close change review
      </WorkbookInspectorActionButton>
      <HistoryLookupFeedback
        purpose="change"
        state={controller.locationLookup}
        onContinue={controller.commands.continueLocation}
        onRestart={controller.commands.restartLocation}
        onCancel={controller.commands.cancelLocation}
      />
      {controller.locationLookup?.phase === "matched" ? (
        <p role="status">Requested change found in this record’s history.</p>
      ) : null}
      {hasLoadedMatch ? (
        <WorkbookInspectorActionButton
          onClick={() => {
            const target = review.current?.querySelector<HTMLElement>(
              "[data-history-requested-change]",
            );
            target?.scrollIntoView({ block: "nearest" });
            target?.focus({ preventScroll: true });
          }}
        >
          Show requested entries
        </WorkbookInspectorActionButton>
      ) : null}
      <p>Only the last three pages stay loaded. Refresh history to restart.</p>
      <WorkbookRecordHistoryPanel
        requestedChangeSetId={locator.changeSetId}
        locatingChange={controller.locationLookup?.phase === "checking"}
        actions={actions}
        canMutate={controller.locationLookup?.phase !== "checking"}
        idleRecordId={locator.recordId}
        state={controller.snapshot}
        browsingControls={controller.commands}
        onOpenHistory={controller.commands.open}
        onCancelPendingAction={controller.commands.cancel}
        onConfirmPendingAction={() => void controller.commands.confirm()}
        onPreviewDeleteRestore={controller.commands.previewDeleteRestore}
        onPreviewRollback={controller.commands.previewRollback}
      />
    </section>
  );
}
