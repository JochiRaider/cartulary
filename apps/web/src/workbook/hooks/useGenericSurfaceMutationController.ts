import {
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookProtocolPatchRecordRequest } from "../adapters/workbookProtocolTypes";
import {
  type WorkbookInspectorErrorPresentation,
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
} from "../inspector/workbookInspectorErrorModel";
import type { GenericViewMutationAccepted } from "../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { ExplicitPatchContribution } from "../runtime/WorkbookExplicitPatchOwner";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";

type GenericPatchMutationRequest = {
  readonly baseline: WorkbookQueryRow;
  readonly contributions?: readonly ExplicitPatchContribution[];
  readonly isCurrent?: () => boolean;
  readonly authoringRevision?: number;
  readonly presentationIdentity?: string;
  readonly baseRowVersion: number;
  readonly changes: readonly WorkbookProtocolPatchRecordRequest["changes"][number][];
  readonly purpose: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
};
export type GenericSurfaceMutationController = {
  readonly ordinaryCreate: WorkbookMutationRuntime["ordinaryCreate"];
  readonly partyLinks: WorkbookMutationRuntime["partyLinks"];
  readonly explicitPatches: WorkbookMutationRuntime["explicitPatches"];
  readonly inspectorDrafts: WorkbookMutationRuntime["inspectorDrafts"];
  readonly taskDrafts: WorkbookMutationRuntime["taskDrafts"];
  readonly beginMutation: () => () => void;
  readonly beginMutationReport: () => () => void;
  readonly clearMutationError: () => void;
  readonly mutationError: WorkbookInspectorErrorPresentation | null;
  readonly mutationPending: boolean;
  readonly rejectMutationFailure: (failure: WorkbookOperationFailure) => void;
  readonly setValidationError: (message: string) => void;
  readonly submitPatchMutation: (
    request: GenericPatchMutationRequest,
  ) => Promise<GenericViewMutationAccepted | null>;
};

export function useGenericSurfaceMutationController({
  mutationRuntime,
  surfaceLabel,
  sheetRef,
  selectedRecordId = "",
}: {
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly surfaceLabel: string;
  readonly sheetRef: SheetRef;
  readonly selectedRecordId?: string;
}): GenericSurfaceMutationController {
  useSyncExternalStore(
    mutationRuntime.explicitPatches.subscribe,
    mutationRuntime.explicitPatches.getSnapshot,
  );
  const generation = useRef(0);
  useLayoutEffect(() => {
    void sheetRef;
    return () => {
      generation.current++;
    };
  }, [sheetRef]);
  const subjectRef = useRef(selectedRecordId);
  subjectRef.current = selectedRecordId;
  const [errorState, setErrorState] = useState<{
    subject: string;
    error: WorkbookInspectorErrorPresentation | null;
  }>({ subject: selectedRecordId, error: null });
  const setMutationError = useCallback(
    (error: WorkbookInspectorErrorPresentation | null) => {
      setErrorState({ subject: subjectRef.current, error });
    },
    [],
  );
  const mutationError =
    errorState.subject === selectedRecordId ? errorState.error : null;
  const [pendingCount, setPendingCount] = useState(0);
  const beginMutationReport = useCallback(
    () => mutationRuntime.beginExplicitMutation(),
    [mutationRuntime],
  );
  const beginMutation = useCallback(() => {
    const finish = beginMutationReport();
    setPendingCount((count) => count + 1);
    setMutationError(null);
    let finished = false;
    return () => {
      if (finished) return;
      finished = true;
      finish();
      setPendingCount((count) => count - 1);
    };
  }, [beginMutationReport, setMutationError]);
  const clearMutationError = useCallback(
    () => setMutationError(null),
    [setMutationError],
  );
  const rejectMutationFailure = useCallback(
    (failure: WorkbookOperationFailure) =>
      setMutationError(workbookInspectorErrorPresentation(failure)),
    [setMutationError],
  );
  const setValidationError = useCallback(
    (message: string) =>
      setMutationError(workbookInspectorLocalErrorPresentation(message)),
    [setMutationError],
  );
  const submitPatchMutation = useCallback(
    async (request: GenericPatchMutationRequest) => {
      const started = generation.current;
      const subject = subjectRef.current;
      const result = await mutationRuntime.explicitPatches.submit(
        {
          viewSchemaId: request.viewSchemaId,
          baseline: request.baseline,
          changes: request.changes,
          purpose: request.purpose,
          compound: request.purpose === "task-lifecycle",
          sheetRef,
          surfaceLabel,
          ...(request.authoringRevision !== undefined
            ? { authoringRevision: request.authoringRevision }
            : {}),
          ...(request.presentationIdentity !== undefined
            ? { presentationIdentity: request.presentationIdentity }
            : {}),
        },
        request.contributions,
      );
      if (
        started !== generation.current ||
        subject !== subjectRef.current ||
        request.isCurrent?.() === false ||
        !mutationRuntime.explicitPatches.getSnapshot().authority
      )
        return result?.receipt ?? null;
      if (!result) {
        setValidationError(
          "This record cannot be submitted while an operation needs recovery or authorization is unavailable. Your draft is retained.",
        );
        return null;
      }
      if (result.failure) rejectMutationFailure(result.failure);
      return result.receipt;
    },
    [
      mutationRuntime,
      rejectMutationFailure,
      sheetRef,
      surfaceLabel,
      setValidationError,
    ],
  );
  return {
    ordinaryCreate: mutationRuntime.ordinaryCreate,
    explicitPatches: mutationRuntime.explicitPatches,
    inspectorDrafts: mutationRuntime.inspectorDrafts,
    partyLinks: mutationRuntime.partyLinks,
    taskDrafts: mutationRuntime.taskDrafts,
    beginMutation,
    beginMutationReport,
    clearMutationError,
    mutationError,
    mutationPending: pendingCount > 0,
    rejectMutationFailure,
    setValidationError,
    submitPatchMutation,
  };
}
