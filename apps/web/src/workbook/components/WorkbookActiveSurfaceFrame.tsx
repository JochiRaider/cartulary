import { workbookActiveSurfaceFocusTargetTestId } from "@cartulary/ui-contracts";
import type { ReactNode, RefObject } from "react";
import { shellActiveSurfaceStyle } from "../layout/workbookShellStyles";
import type {
  WorkbookMutationRuntime,
  WorkbookMutationSnapshot,
} from "../runtime/WorkbookMutationRuntime";
import { WorkbookEditRecoveryPanel } from "./WorkbookEditRecoveryPanel";
import { WorkbookQueueOverflowNotice } from "./WorkbookQueueOverflowNotice";
import { WorkbookSameFieldConflictResolver } from "./WorkbookSameFieldConflictResolver";

type WorkbookActiveSurfaceFrameProps = {
  readonly activeContent: ReactNode;
  readonly activeSurfaceRef: RefObject<HTMLElement | null>;
  readonly apiBase: string | undefined;
  readonly focus: {
    readonly editRecoveryPanelRef: RefObject<HTMLElement | null>;
    readonly focusSameFieldSummary: () => void;
    readonly onFocusWithinChange: (focused: boolean) => void;
    readonly overflowNoticeRef: RefObject<HTMLElement | null>;
    readonly sameFieldSummaryRef: RefObject<HTMLDivElement | null>;
  };
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly mutationSnapshot: WorkbookMutationSnapshot;
  readonly onActivateOrigin: (viewSchemaId: string) => void;
};

/** Presents the active surface behind the mutually exclusive recovery layer. */
export function WorkbookActiveSurfaceFrame({
  activeContent,
  activeSurfaceRef,
  apiBase,
  focus,
  mutationRuntime,
  mutationSnapshot,
  onActivateOrigin,
}: WorkbookActiveSurfaceFrameProps) {
  const conflictOnly =
    mutationSnapshot.blockedEdit === null &&
    mutationSnapshot.overflowMessage === null &&
    mutationSnapshot.conflictPanelOpen;
  return (
    <section
      aria-label="Active workbook surface focus target"
      data-testid={workbookActiveSurfaceFocusTargetTestId()}
      ref={activeSurfaceRef}
      style={shellActiveSurfaceStyle}
      tabIndex={-1}
    >
      <div
        aria-hidden={conflictOnly ? true : undefined}
        inert={conflictOnly ? true : undefined}
        style={{ display: "contents" }}
      >
        {activeContent}
      </div>
      {mutationSnapshot.blockedEdit !== null ? (
        <WorkbookEditRecoveryPanel
          blockedEdit={mutationSnapshot.blockedEdit}
          key={mutationSnapshot.blockedEdit.unitId}
          onDiscard={() => mutationRuntime.discardBlockedEdit()}
          onFocusWithinChange={focus.onFocusWithinChange}
          onRetry={() => mutationRuntime.retryBlockedEdit()}
          ref={focus.editRecoveryPanelRef}
        />
      ) : mutationSnapshot.overflowMessage !== null ? (
        <WorkbookQueueOverflowNotice
          message={mutationSnapshot.overflowMessage}
          onFocusWithinChange={focus.onFocusWithinChange}
          ref={focus.overflowNoticeRef}
        />
      ) : (
        <WorkbookSameFieldConflictResolver
          apiBase={apiBase}
          focusSummary={focus.focusSameFieldSummary}
          mutationRuntime={mutationRuntime}
          onActivateOrigin={onActivateOrigin}
          snapshot={mutationSnapshot}
          summaryRef={focus.sameFieldSummaryRef}
        />
      )}
    </section>
  );
}
