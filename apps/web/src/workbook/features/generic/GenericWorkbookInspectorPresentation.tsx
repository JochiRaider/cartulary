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
import {
  type Dispatch,
  type ReactNode,
  type SetStateAction,
  useId,
} from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import {
  workbookFormFieldStackStyle,
  workbookFormFieldsStyle,
  workbookFormInputStyle,
  workbookFormMessageStyle,
} from "../../components/workbookFormStyles";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "../../inspector/presentation/WorkbookInspectorFeedback";
import type { WorkbookInspectorEditDraft } from "../../inspector/useWorkbookInspectorEditDraft";
import { WorkbookInspectorDraftFeedback } from "../../inspector/WorkbookInspectorDraftFeedback";
import { WorkbookInspectorEditControl } from "../../inspector/WorkbookInspectorEditControl";
import type { WorkbookInspectorErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import type { GenericCollectionMode } from "../../models/genericWorkbookModel";
import {
  genericCellLabel,
  genericCollectionSupportsRemove,
} from "../../models/genericWorkbookModel";
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
        <Button
          data-testid={genericCreateSubmitTestId(props.contract.viewSchemaId)}
          disabled={
            props.draftDisabled ||
            props.mutation.ordinaryCreate.busy(props.contract.viewSchemaId)
          }
          tone="secondary"
          type="button"
          onClick={() => void props.submitCreate()}
        >
          Commit draft row
        </Button>
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
  readonly fieldFeedback: string | null;
  readonly actionError: WorkbookInspectorErrorPresentation | null;
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
  readonly rows: readonly WorkbookQueryRow[];
  readonly selectedEdit: SelectedEdit;
  readonly selectedRecordId: string;
  readonly setCollectionMode: (mode: GenericCollectionMode) => void;
  readonly setEditFieldKey: (fieldKey: string) => void;
  readonly submitEdit: () => Promise<void>;
};

function GenericDetails(props: GenericDetailsProps) {
  const feedbackId = useId();
  if (props.selectedEdit.row === null || props.editableFields.length === 0)
    return props.selectedEdit.row ? (
      <dl>
        {props.contract.fields.map((field) => (
          <div key={field.fieldKey}>
            <dt>{field.label}</dt>
            <dd>
              {genericCellLabel(
                props.selectedEdit.row?.cells[field.fieldKey]?.value,
              )}
            </dd>
          </div>
        ))}
      </dl>
    ) : null;
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
          invalid={props.fieldFeedback !== null}
          describedBy={props.fieldFeedback ? feedbackId : undefined}
          edit={props.edit}
          ariaLabel={field.label}
          id={`generic-edit-${props.selectedRecordId}-${field.fieldKey}`}
          collectionItems={props.collectionItems}
          collectionMode={props.collectionMode}
          field={field}
          testId={genericEditValueTestId(props.contract.viewSchemaId)}
        />
      ) : (
        <span role="status">Select an available field.</span>
      )}
      {props.fieldFeedback ? (
        <p id={feedbackId} role="alert" style={workbookFormMessageStyle}>
          {props.fieldFeedback}
        </p>
      ) : null}
      <WorkbookInspectorDraftFeedback
        edit={props.edit}
        contract={props.contract}
        row={props.selectedEdit.row}
      />
      <Button
        data-testid={genericEditSubmitTestId(props.contract.viewSchemaId)}
        disabled={props.mutationPending || !props.edit.canSubmit}
        tone="primary"
        type="button"
        onClick={() => void props.submitEdit()}
      >
        Update
      </Button>
      {props.actionError ? (
        <WorkbookInspectorPublicError error={props.actionError} />
      ) : null}
    </fieldset>
  );
}

type GenericRelationshipsProps = {
  readonly noteAssociations?: ReactNode;
  readonly referenceSummary: ReactNode;
  readonly party: ReturnType<typeof useGenericPartyLinkWorkflow>;
};

function GenericRelationships(props: GenericRelationshipsProps) {
  return (
    props.noteAssociations ?? (
      <>
        {props.referenceSummary}
        <PartyLinkPanel workflow={props.party} />
      </>
    )
  );
}

const editRowStyle = {
  ...workbookFormFieldsStyle,
  gridTemplateColumns: "minmax(0, 1fr)",
  alignItems: "stretch",
};
const draftInspectorFieldsStyle = {
  ...workbookFormFieldsStyle,
  gridTemplateColumns: "repeat(auto-fit, minmax(min(12rem, 100%), 1fr))",
  alignItems: "end",
};
const labelStyle = workbookFormFieldStackStyle;
const selectStyle = { ...workbookFormInputStyle, appearance: "auto" as const };
