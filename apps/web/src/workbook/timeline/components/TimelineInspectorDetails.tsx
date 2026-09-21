import {
  genericEditSubmitTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { useId, useState, useSyncExternalStore } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "../../inspector/presentation/WorkbookInspectorFeedback";
import { useWorkbookInspectorEditDraft } from "../../inspector/useWorkbookInspectorEditDraft";
import { useWorkbookInspectorFieldFeedback } from "../../inspector/useWorkbookInspectorFieldFeedback";
import { WorkbookExplicitPatchRecovery } from "../../inspector/WorkbookExplicitPatchRecovery";
import { WorkbookInspectorDetails } from "../../inspector/WorkbookInspectorDetails";
import { WorkbookInspectorDraftFeedback } from "../../inspector/WorkbookInspectorDraftFeedback";
import type { WorkbookInspectorDraftStore } from "../../inspector/WorkbookInspectorDraftStore";
import { WorkbookInspectorEditControl } from "../../inspector/WorkbookInspectorEditControl";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookExplicitPatchOwner } from "../../runtime/WorkbookExplicitPatchOwner";
import {
  normalizeTimelineFullRow,
  type WorkbookRow,
} from "../models/timelineRowModel";

const contract = requireViewContract(timelineViewSchemaId);
const editableFields = contract.fields.filter(
  (field) => field.patchWritable && field.writeKind === "direct_value",
);

export type TimelineInspectorDetailsOwner = {
  readonly patches: WorkbookExplicitPatchOwner;
  readonly drafts: WorkbookInspectorDraftStore;
  readonly sheetRef: SheetRef;
  readonly presentation: string;
  readonly accepted: (
    rowKey: string,
    mutation: Pick<WorkbookPendingMutationAccepted, "row" | "viewSchemaId">,
  ) => unknown;
};

/** Timeline supplies normalization and source-row effects to the retained PATCH owner. */
export function TimelineInspectorDetails({
  row,
  owner,
}: {
  readonly row: WorkbookRow;
  readonly owner: TimelineInspectorDetailsOwner;
}) {
  const [fieldKey, setFieldKey] = useState("");
  const snapshot = useSyncExternalStore(
    owner.patches.subscribe,
    owner.patches.getSnapshot,
  );
  const latest = row.recordId ? owner.patches.latestRow(row.recordId) : null;
  const saved =
    latest && latest.row_version > (row.rawRow?.row_version ?? 0)
      ? latest
      : row.rawRow;
  const field =
    editableFields.find((candidate) => candidate.fieldKey === fieldKey) ?? null;
  const edit = useWorkbookInspectorEditDraft({
    store: owner.drafts,
    row: saved,
    field,
    viewSchemaId: timelineViewSchemaId,
    presentation: owner.presentation,
    active: !!saved,
  });
  const feedback = useWorkbookInspectorFieldFeedback(edit);
  const feedbackId = useId();
  if (!saved || !snapshot.authority || !owner.drafts.canRead()) return null;
  const blocked = owner.patches.blocksRecord(saved.record_id);
  const canSubmit = edit.canSubmit && !blocked;
  const submit = async () => {
    if (
      !field ||
      !edit.canSubmit ||
      owner.patches.blocksRecord(saved.record_id)
    )
      return;
    // Timeline visible text preserves source bytes, whitespace and empty text.
    // Its owner contract intentionally differs from normalized ordinary titles.
    const change = { field_key: field.fieldKey, value: edit.value };
    const captured = edit.capture(),
      failure = feedback.capture();
    feedback.clear();
    const accept = (mutation: WorkbookPendingMutationAccepted | null) => {
      if (!mutation) return;
      const normalized = normalizeTimelineFullRow(
        mutation.row,
        "Timeline inspector accepted row",
      );
      owner.accepted(row.key, {
        row: normalized,
        viewSchemaId: timelineViewSchemaId,
      });
    };
    const result = await owner.patches.submit(
      {
        viewSchemaId: timelineViewSchemaId,
        baseline: edit.baseline ?? saved,
        changes: [change],
        purpose: "timeline-inspector-scalar",
        sheetRef: owner.sheetRef,
        surfaceLabel: "Timeline",
        presentationIdentity: captured.attachment,
        ...(captured.draft
          ? { authoringRevision: captured.draft.revision }
          : {}),
      },
      [
        {
          prepare: async () => {
            if (!edit.isCurrent(captured))
              throw new Error(
                "The inspector changed before dispatch. Resume the original field to submit it.",
              );
          },
          acknowledged: (entry) => {
            edit.complete(captured);
            accept(entry.receipt);
          },
          conflictResolved: (entry, _kind, resolvedRow) => {
            edit.complete(captured);
            if (resolvedRow) {
              owner.accepted(row.key, {
                row: normalizeTimelineFullRow(
                  resolvedRow,
                  "Timeline inspector conflict resolution",
                ),
                viewSchemaId: timelineViewSchemaId,
              });
            } else accept(entry.receipt);
          },
        },
      ],
    );
    if (result?.failure) failure(result.failure);
    else if (!result && edit.isCurrent(captured))
      feedback.rejectLocal(
        "This change cannot be submitted while an earlier operation needs recovery.",
        false,
      );
  };
  return (
    <WorkbookInspectorDetails
      contract={contract}
      row={saved}
      editableFields={editableFields}
      activeField={fieldKey}
      onEdit={setFieldKey}
      onDetach={() => setFieldKey("")}
      onSubmit={() => void submit()}
      canSubmit={canSubmit}
      disabledReason={
        owner.drafts.canAuthor() ? null : "Current access permits reading only."
      }
      editor={
        field ? (
          <>
            <WorkbookInspectorEditControl
              edit={edit}
              field={field}
              collectionMode="add"
              ariaLabel={field.label}
              invalid={feedback.message !== null}
              describedBy={feedback.message ? feedbackId : undefined}
              testId={timelineScalarEditorTestId({
                fieldKey,
                recordId: saved.record_id,
                surface: "inspector",
              })}
            />
            {feedback.message ? (
              <p role="alert" id={feedbackId}>
                {feedback.message}
              </p>
            ) : null}
            <WorkbookInspectorDraftFeedback
              edit={edit}
              contract={contract}
              row={saved}
            />
            <Button
              tone="primary"
              data-testid={genericEditSubmitTestId(timelineViewSchemaId)}
              disabled={!canSubmit}
              onClick={() => void submit()}
            >
              Update
            </Button>
            {feedback.actionError ? (
              <WorkbookInspectorPublicError error={feedback.actionError} />
            ) : null}
            <WorkbookExplicitPatchRecovery
              owner={owner.patches}
              viewSchemaId={timelineViewSchemaId}
              recordId={saved.record_id}
              fieldKey={fieldKey}
            />
          </>
        ) : null
      }
    />
  );
}
