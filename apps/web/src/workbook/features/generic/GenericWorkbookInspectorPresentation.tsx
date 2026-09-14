import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditActionSelectTestId,
  genericEditFieldSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorDisabledCondition,
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import type { Dispatch, SetStateAction } from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import type { WorkbookInspectorEditDraft } from "../../inspector/useWorkbookInspectorEditDraft";
import { WorkbookInspectorDraftFeedback } from "../../inspector/WorkbookInspectorDraftFeedback";
import { WorkbookInspectorEditControl } from "../../inspector/WorkbookInspectorEditControl";
import type { GenericCollectionMode } from "../../models/genericWorkbookModel";
import { genericCollectionSupportsRemove } from "../../models/genericWorkbookModel";
import type { GenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { WorkbookMutationCommandPorts } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOwnerBinding } from "../../policies/workbookSurfacePolicy";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { CoordinationWorkflowBindings } from "../coordination/CoordinationWorkflowBindings";
import { NoteSheetAuthoring } from "../notes/NoteSheetAuthoring";
import { OrdinaryCreateControl } from "../ordinary/OrdinaryCreateControl";
import { PartyLinkPanel } from "../parties/PartyLinkPanel";
import type { useGenericPartyLinkWorkflow } from "../parties/useGenericPartyLinkWorkflow";
import { GenericWorkbookInspector } from "./GenericWorkbookInspector";

type InspectorProps = Parameters<typeof GenericWorkbookInspector>[0];
type SelectedEdit = {
  readonly field: ViewFieldContract | null;
  readonly row: WorkbookQueryRow | null;
};

export function GenericWorkbookInspectorPresentation({
  details,
  inspector,
  isOpen,
  relationships,
  workflow,
}: {
  readonly details: GenericDetailsProps;
  readonly inspector: Omit<
    InspectorProps,
    "detailsContent" | "relationshipsContent" | "workflowContent"
  >;
  readonly isOpen: boolean;
  readonly relationships: GenericRelationshipsProps;
  readonly workflow: GenericWorkflowProps;
}) {
  if (!isOpen) return undefined;
  return (
    <GenericWorkbookInspector
      {...inspector}
      detailsContent={<GenericDetails {...details} />}
      relationshipsContent={<GenericRelationships {...relationships} />}
      workflowContent={<GenericWorkflow {...workflow} />}
    />
  );
}

type GenericWorkflowProps = {
  readonly subjectRow: WorkbookQueryRow | null;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly lifecycleDisabled: boolean;
  readonly canCreateRows: boolean;
  readonly contract: ViewContract;
  readonly createDraft: Record<string, string>;
  readonly draftDisabled: boolean;
  readonly draftInspectorFields: readonly ViewFieldContract[];
  readonly invalidationKey: string;
  readonly mutation: GenericSurfaceMutationController;
  readonly mutationCommands: WorkbookMutationCommandPorts;
  readonly ownerBindings: readonly WorkbookOwnerBinding[];
  readonly referenceOptions: GenericReferenceOptions;
  readonly rows: readonly WorkbookQueryRow[];
  readonly setCreateDraft: Dispatch<SetStateAction<Record<string, string>>>;
  readonly submitCreate: () => Promise<void>;
  readonly subjectPresent: boolean;
};

function GenericWorkflow(props: GenericWorkflowProps) {
  return (
    <>
      {props.ownerBindings.includes("linked_note_create") ? (
        <NoteSheetAuthoring />
      ) : null}
      <GenericDraftFields {...props} />
      {props.canCreateRows ? (
        <button
          data-testid={genericCreateSubmitTestId(props.contract.viewSchemaId)}
          disabled={
            props.draftDisabled ||
            props.mutation.ordinaryCreate.busy(props.contract.viewSchemaId)
          }
          style={secondaryActionButtonStyle}
          type="button"
          onClick={() => void props.submitCreate()}
        >
          Commit draft row
        </button>
      ) : null}
      {props.subjectRow ? (
        <CoordinationWorkflowBindings
          contract={props.contract}
          disabled={props.mutation.mutationPending || props.lifecycleDisabled}
          mutation={props.mutation}
          drafts={props.mutation.taskDrafts}
          row={props.subjectRow}
          currentIncidentRole={props.currentIncidentRole}
          disabledTokens={props.disabledTokens}
          referenceOptions={props.referenceOptions}
        />
      ) : null}
    </>
  );
}

function GenericDraftFields(props: GenericWorkflowProps) {
  const retained =
    props.mutation.ordinaryCreate.getSnapshot().schemas[
      props.contract.viewSchemaId
    ]?.draft;
  if (
    (!props.canCreateRows && !Object.keys(retained?.values ?? {}).length) ||
    props.draftInspectorFields.length === 0
  )
    return null;
  return (
    <div style={draftInspectorFieldsStyle}>
      {props.draftInspectorFields.map((field) => {
        const controlId = `generic-create-inspector-${field.fieldKey}`;
        return (
          <label htmlFor={controlId} key={field.fieldKey} style={labelStyle}>
            {field.label}
            <OrdinaryCreateControl
              owner={props.mutation.ordinaryCreate}
              contract={props.contract}
              disabled={props.draftDisabled}
              collectionMode="add"
              field={field}
              id={controlId}
              referenceOptions={props.referenceOptions}
              testId={genericCreateFieldTestId(field.fieldKey)}
              value={props.createDraft[field.fieldKey] ?? ""}
              onChange={(value) =>
                props.setCreateDraft((current) => ({
                  ...current,
                  [field.fieldKey]: value,
                }))
              }
            />
          </label>
        );
      })}
    </div>
  );
}

type GenericDetailsProps = {
  readonly edit: WorkbookInspectorEditDraft;
  readonly collectionItems: readonly {
    readonly displayText: string;
    readonly itemRef: string;
  }[];
  readonly collectionMode: GenericCollectionMode;
  readonly contract: ViewContract;
  readonly editableFields: readonly ViewFieldContract[];
  readonly editFieldKey: string;
  readonly mutationPending: GenericSurfaceMutationController["mutationPending"];
  readonly referenceOptions: GenericReferenceOptions;
  readonly rows: readonly WorkbookQueryRow[];
  readonly selectedEdit: SelectedEdit;
  readonly selectedRecordId: string;
  readonly setCollectionMode: (mode: GenericCollectionMode) => void;
  readonly setEditFieldKey: (fieldKey: string) => void;
  readonly submitEdit: () => Promise<void>;
};

function GenericDetails(props: GenericDetailsProps) {
  if (props.selectedEdit.row === null || props.editableFields.length === 0)
    return null;
  const field = props.selectedEdit.field;
  return (
    <fieldset style={{ ...editRowStyle, border: 0, padding: 0, minWidth: 0 }}>
      <select
        aria-label="Edit field"
        data-testid={genericEditFieldSelectTestId(props.contract.viewSchemaId)}
        style={selectStyle}
        value={props.editFieldKey}
        onChange={(event) => props.setEditFieldKey(event.target.value)}
      >
        <option value="">Field</option>
        {props.editableFields.map((candidate) => (
          <option key={candidate.fieldKey} value={candidate.fieldKey}>
            {candidate.label}
          </option>
        ))}
      </select>
      {field?.writeKind === "action_payload" &&
      genericCollectionSupportsRemove(field.fieldKey) ? (
        <select
          aria-label="Collection edit action"
          data-testid={genericEditActionSelectTestId(
            props.contract.viewSchemaId,
          )}
          style={selectStyle}
          value={props.collectionMode}
          onChange={(event) => {
            props.setCollectionMode(
              event.target.value === "remove" ? "remove" : "add",
            );
          }}
        >
          <option value="add">Add</option>
          <option value="remove">Remove</option>
        </select>
      ) : null}
      {field ? (
        <WorkbookInspectorEditControl
          edit={props.edit}
          ariaLabel={field.label}
          id={`generic-edit-${props.selectedRecordId}-${field.fieldKey}`}
          collectionItems={props.collectionItems}
          collectionMode={props.collectionMode}
          field={field}
          referenceOptions={props.referenceOptions}
          testId={genericEditValueTestId(props.contract.viewSchemaId)}
        />
      ) : (
        <span role="status">Select an available field.</span>
      )}
      <WorkbookInspectorDraftFeedback
        edit={props.edit}
        contract={props.contract}
        row={props.selectedEdit.row}
      />
      <button
        data-testid={genericEditSubmitTestId(props.contract.viewSchemaId)}
        disabled={props.mutationPending || !props.edit.canSubmit}
        style={actionButtonStyle}
        type="button"
        onClick={() => void props.submitEdit()}
      >
        Update
      </button>
    </fieldset>
  );
}

type GenericRelationshipsProps = {
  readonly party: ReturnType<typeof useGenericPartyLinkWorkflow>;
};

function GenericRelationships(props: GenericRelationshipsProps) {
  return <PartyLinkPanel workflow={props.party} />;
}

const editRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr)",
  gap: "0.6rem",
  alignItems: "stretch",
};
const draftInspectorFieldsStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.75rem",
  alignItems: "end",
};
const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
};
const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};
const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};
const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
};
const selectStyle = { ...inputStyle, appearance: "auto" as const };
