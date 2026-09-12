import {
  assessmentsViewSchemaId,
  requireViewContract,
} from "@cartulary/view-contracts";
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import { initialAssessmentDraft } from "../../models/assessmentWorkbookModel";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { WorkbookAssessmentAuthoringOwner } from "./WorkbookAssessmentAuthoringOwner";

const emptyDraft = initialAssessmentDraft(
  requireViewContract(assessmentsViewSchemaId),
);

/** React owns attachment only. Intent, admission, dispatch, and receipts belong to the runtime owner. */
export function useAssessmentCreationController({
  owner,
  lifecycleResetKey,
  sheetRef,
}: {
  readonly owner: WorkbookAssessmentAuthoringOwner;
  readonly lifecycleResetKey: string;
  readonly sheetRef: SheetRef;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const [attachment, setAttachment] = useState<{
    key: string;
    review: number;
  } | null>(null);
  const current = useRef({ attachment, key: lifecycleResetKey });
  current.current = { attachment, key: lifecycleResetKey };
  const detach = useCallback(() => {
    current.current.attachment = null;
    setAttachment(null);
  }, []);
  useEffect(
    () => () => {
      current.current.attachment = null;
    },
    [],
  );
  const attach = () =>
    setAttachment({
      key: lifecycleResetKey,
      review: owner.getSnapshot().reviewRevision,
    });
  const attached =
    !!attachment &&
    attachment.key === lifecycleResetKey &&
    attachment.review === snapshot.reviewRevision;
  return {
    commands: {
      cancel: detach,
      reset: detach,
      resume: attach,
      discard: () => {
        if (owner.discardDraft()) detach();
      },
      openStandalone: () => {
        if (owner.openStandalone()) attach();
      },
      openFollowOn: (row: WorkbookQueryRow | null) => {
        const opened = owner.openFollowOn(row);
        if (opened) attach();
        return opened;
      },
      rejectStart: (message: string) => owner.reject(message),
      updateDraft: owner.updateDraft.bind(owner),
      submit: async (canCreate: boolean) => {
        const origin = current.current.attachment;
        if (!canCreate || !attached || !origin) return;
        await owner.submit({
          sheetRef,
          isCurrent: () =>
            current.current.attachment === origin &&
            current.current.key === origin.key &&
            owner.getSnapshot().reviewRevision === origin.review,
        });
      },
    },
    snapshot: {
      ...snapshot,
      feedback: attached ? snapshot.feedback : null,
      draft: snapshot.draft?.values ?? emptyDraft,
      draftMode: snapshot.draft?.mode ?? "standalone",
      hasDraft: !!snapshot.draft,
      attached,
      isSubmitting: owner.busy,
    },
  };
}
