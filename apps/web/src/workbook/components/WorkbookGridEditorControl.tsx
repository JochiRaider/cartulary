import type {
  GridCellTarget,
  GridEditCommitOutcome,
  GridEditorAdapter,
} from "@cartulary/grid-adapter";
import type { ViewFieldContract } from "@cartulary/view-contracts";
import {
  type KeyboardEvent,
  type ReactNode,
  useEffect,
  useSyncExternalStore,
} from "react";
import type { WorkbookCollaborationCoordinator } from "../collaboration/WorkbookCollaborationCoordinator";
import {
  taskGuardFields,
  taskViewId,
} from "../features/coordination/taskLifecycleModel";
import type { WorkbookGridDraftStore } from "../models/WorkbookGridDraftStore";
import type { GenericReferenceOptions } from "../models/workbookReferenceOptions";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { GenericMutationControl } from "./GenericMutationControl";
import {
  WorkbookCellPresenceMarker,
  WorkbookPresenceCellLayout,
} from "./WorkbookPresenceMarkers";

export function workbookGridEditorAdapter<Row>({
  commit,
  field,
  readValue,
  referenceOptions,
  collaboration,
  drafts,
  readRow,
  readCurrentRow,
  viewSchemaId,
}: {
  readonly commit: (
    draftValue: string | null,
    target: GridCellTarget,
    row: Row,
  ) => Promise<GridEditCommitOutcome>;
  readonly field: ViewFieldContract;
  readonly readValue: (row: Row) => unknown;
  readonly referenceOptions: GenericReferenceOptions;
  readonly collaboration?: WorkbookCollaborationCoordinator | undefined;
  readonly drafts: WorkbookGridDraftStore;
  readonly readRow: (row: Row) => WorkbookQueryRow;
  readonly readCurrentRow?: (row: Row) => WorkbookQueryRow;
  readonly viewSchemaId: string;
}): GridEditorAdapter<Row> {
  const identity = (row: Row) => ({
    viewSchemaId,
    recordId: readRow(row).record_id,
    fieldKey: field.fieldKey,
  });
  return {
    ...(field.clearable ? { clearDraftValue: null } : {}),
    retainDraft: (row, value) =>
      drafts.update(
        identity(row),
        readRow(row),
        value === null ? null : String(value ?? ""),
      ),
    discardDraft: (row) => drafts.discard(identity(row)),
    commit: async (intent) => {
      drafts.update(
        identity(intent.row),
        readRow(intent.row),
        intent.draftValue === null ? null : String(intent.draftValue ?? ""),
      );
      const captured = drafts.capture(identity(intent.row));
      const outcome = await commit(
        intent.draftValue === null ? null : String(intent.draftValue ?? ""),
        intent.target,
        intent.row,
      );
      if (outcome.kind !== "accepted")
        drafts.setValidation(captured, outcome.message);
      return outcome;
    },
    initialDraftValue: (row) => {
      const retained = drafts.read(identity(row));
      if (retained) return retained.value;
      const value = readValue(row);
      return value === null || value === undefined ? "" : String(value);
    },
    renderEditor: (context) => {
      const draftValue = String(context.draftValue ?? "");
      const commitOnEnter = (event: KeyboardEvent<HTMLFieldSetElement>) => {
        if (event.nativeEvent.isComposing) return;
        if (
          event.target instanceof Element &&
          event.target.closest("[data-grid-editor-toolbar]")
        )
          return;
        if (event.key === "Escape") {
          event.preventDefault();
          context.cancel();
          return;
        }
        if (event.key === "Enter" && !event.shiftKey) {
          event.preventDefault();
          void context.commit();
        }
      };
      const content = (
        <>
          <GenericMutationControl
            invalid={context.outcome?.kind === "validation_error"}
            collectionMode="add"
            field={field}
            focusTargetRef={context.focusTargetRef}
            referenceOptions={referenceOptions}
            surface="grid"
            testId={`grid-editor-${
              context.target.rowIdentity.kind === "core_record"
                ? context.target.rowIdentity.recordId
                : "unsupported"
            }-${field.fieldKey}`}
            value={draftValue}
            onChange={(value) => context.setDraftValue(value)}
          />
          <fieldset data-grid-editor-toolbar="true" aria-label="Cell actions">
            {field.readKind === "boolean" &&
            draftValue !== "true" &&
            draftValue !== "false" ? (
              <span style={{ whiteSpace: "pre-wrap" }}>
                Unfinished value: {JSON.stringify(draftValue)}
              </span>
            ) : null}
            <WorkbookGridDraftFeedback
              drafts={drafts}
              pending={context.pending}
              field={field}
              row={(readCurrentRow ?? readRow)(context.row)}
              viewSchemaId={viewSchemaId}
            />
            {field.clearable ? (
              <button type="button" onClick={() => context.setDraftValue(null)}>
                Clear {field.label}
              </button>
            ) : null}
            {context.draftValue === null ? <span>Clear on commit</span> : null}
            <button type="button" onClick={() => void context.commit()}>
              Commit
            </button>
            <button type="button" onClick={context.cancel}>
              Cancel
            </button>
          </fieldset>
        </>
      );
      return (
        <fieldset
          aria-label={`Edit ${field.label}`}
          aria-keyshortcuts="Alt+ArrowDown"
          data-grid-editor-kind={workbookGridEditorKind(field)}
          onKeyDown={commitOnEnter}
        >
          {collaboration === undefined ||
          context.target.rowIdentity.kind !== "core_record" ? (
            content
          ) : (
            <WorkbookEditorPresence
              collaboration={collaboration}
              field={field}
              recordId={context.target.rowIdentity.recordId}
            >
              {content}
            </WorkbookEditorPresence>
          )}
        </fieldset>
      );
    },
  };
}

function WorkbookGridDraftFeedback({
  drafts,
  pending,
  field,
  row,
  viewSchemaId,
}: {
  drafts: WorkbookGridDraftStore;
  pending: boolean;
  field: ViewFieldContract;
  row: WorkbookQueryRow;
  viewSchemaId: string;
}) {
  useSyncExternalStore(drafts.subscribe, drafts.getSnapshot);
  const identity = {
    viewSchemaId,
    recordId: row.record_id,
    fieldKey: field.fieldKey,
  };
  const dependencies =
    viewSchemaId === taskViewId &&
    taskGuardFields.some((key) => key === field.fieldKey)
      ? taskGuardFields
      : [];
  const stale = pending ? [] : drafts.staleFields(identity, row, dependencies);
  const draft = drafts.read(identity);
  if (!draft) return null;
  return (
    <span role="status">
      {stale.length
        ? "Saved values changed. Review before committing this draft. "
        : (draft.validation ??
          (pending
            ? "Submission pending; current text retained. "
            : "Unsaved "))}
      {stale.length ? (
        <button type="button" onClick={() => drafts.review(identity, row)}>
          Keep draft {field.label}
        </button>
      ) : null}
    </span>
  );
}

function WorkbookEditorPresence({
  children,
  collaboration,
  field,
  recordId,
}: {
  readonly children: ReactNode;
  readonly collaboration: WorkbookCollaborationCoordinator;
  readonly field: ViewFieldContract;
  readonly recordId: string;
}) {
  useEffect(
    () => collaboration.beginEditingPresence(recordId, field.fieldKey),
    [collaboration, field.fieldKey, recordId],
  );
  useSyncExternalStore(
    collaboration.subscribe,
    collaboration.getSnapshot,
    collaboration.getSnapshot,
  );
  return (
    <WorkbookPresenceCellLayout
      editing
      marker={
        <WorkbookCellPresenceMarker
          fieldKey={field.fieldKey}
          fieldLabel={field.label}
          recordId={recordId}
          presences={collaboration.editingPresenceForCell(
            recordId,
            field.fieldKey,
          )}
        />
      }
    >
      {children}
    </WorkbookPresenceCellLayout>
  );
}

export type WorkbookGridEditorKind =
  | "boolean"
  | "enum"
  | "multiline"
  | "number"
  | "single-line"
  | "stable-reference"
  | "timestamp";

export function workbookGridEditorKind(
  field: ViewFieldContract,
): WorkbookGridEditorKind | null {
  if (!field.gridEditable || field.writeKind !== "direct_value") return null;
  if (field.enumValues && field.enumValues.length > 0) return "enum";
  if (field.readKind === "boolean") return "boolean";
  if (field.readKind === "number") return "number";
  if (field.directReferenceContractId) return "stable-reference";
  if (field.directScalarContractId === "timestamp_instant_v1") {
    return "timestamp";
  }
  if (field.stringContractId === "multiline_body_v1") return "multiline";
  return "single-line";
}
