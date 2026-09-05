import { workbookActiveSurfaceFocusTargetTestId } from "@cartulary/ui-contracts";
import type { ReactNode, RefObject } from "react";
import { shellActiveSurfaceStyle } from "../layout/workbookShellStyles";
import type {
  WorkbookMutationRuntime,
  WorkbookStatusPresentation,
} from "../runtime/WorkbookMutationRuntime";
import { WorkbookEditRecoveryPanel } from "./WorkbookEditRecoveryPanel";
import { WorkbookQueueOverflowNotice } from "./WorkbookQueueOverflowNotice";
import { WorkbookSameFieldConflictResolver } from "./WorkbookSameFieldConflictResolver";

type WorkbookActiveSurfaceFrameProps = {
  readonly activeContent: ReactNode;
  readonly activeSurfaceRef: RefObject<HTMLElement | null>;
  readonly apiBase: string | undefined;
  readonly focus: {
    readonly resolverActivation: {
      readonly conflictKey: string;
      readonly sequence: number;
    } | null;
    readonly editRecoveryPanelRef: RefObject<HTMLElement | null>;
    readonly focusSameFieldSummary: () => void;
    readonly onFocusWithinChange: (focused: boolean) => void;
    readonly overflowNoticeRef: RefObject<HTMLElement | null>;
    readonly sameFieldSummaryRef: RefObject<HTMLDivElement | null>;
  };
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly mutationSnapshot: WorkbookStatusPresentation;
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
  const showBlocked =
    mutationSnapshot.action?.kind === "transaction_recovery" ||
    mutationSnapshot.action?.kind === "terminal_failure";
  const showOverflow = mutationSnapshot.action?.kind === "overflow";
  const conflictOnly =
    !showBlocked && !showOverflow && mutationSnapshot.conflictPanelOpen;
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
      {showBlocked && mutationSnapshot.blockedEdit !== null ? (
        <WorkbookEditRecoveryPanel
          blockedEdit={mutationSnapshot.blockedEdit}
          key={mutationSnapshot.blockedEdit.unitId}
          onDiscard={() => mutationRuntime.discardBlockedEdit()}
          onFocusWithinChange={focus.onFocusWithinChange}
          onRetry={() => mutationRuntime.retryBlockedEdit()}
          ref={focus.editRecoveryPanelRef}
        />
      ) : showOverflow && mutationSnapshot.overflowMessage !== null ? (
        <WorkbookQueueOverflowNotice
          message={mutationSnapshot.overflowMessage}
          onFocusWithinChange={focus.onFocusWithinChange}
          ref={focus.overflowNoticeRef}
        />
      ) : (
        <WorkbookSameFieldConflictResolver
          activation={focus.resolverActivation}
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
