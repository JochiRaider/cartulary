import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditActionSelectTestId,
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
  workbookFormGroupStyle,
  workbookFormHeadingStyle,
  workbookFormInputStyle,
  workbookFormMessageStyle,
} from "../../components/workbookFormStyles";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "../../inspector/presentation/WorkbookInspectorFeedback";
import {
  savedInspectorRegion,
  type WorkbookInspectorRegion,
} from "../../inspector/presentation/WorkbookInspectorPanelContent";
import type { WorkbookInspectorEditDraft } from "../../inspector/useWorkbookInspectorEditDraft";
import { WorkbookExplicitPatchRecovery } from "../../inspector/WorkbookExplicitPatchRecovery";
import { WorkbookInspectorDetails } from "../../inspector/WorkbookInspectorDetails";
import { WorkbookInspectorDraftFeedback } from "../../inspector/WorkbookInspectorDraftFeedback";
import { WorkbookInspectorEditControl } from "../../inspector/WorkbookInspectorEditControl";
import type { WorkbookInspectorErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import type { GenericCollectionMode } from "../../models/genericWorkbookModel";
import { genericCollectionSupportsRemove } from "../../models/genericWorkbookModel";
import type { WorkbookMutationCommandPorts } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOwnerBinding } from "../../policies/workbookSurfacePolicy";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { WorkbookExplicitPatchOwner } from "../../runtime/WorkbookExplicitPatchOwner";
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
      relationshipsContent={
        relationships.noteAssociations ?? [
          savedInspectorRegion("references", {
            kind: "populated",
            content: <GenericRelationships {...relationships} />,
          }),
        ]
      }
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
      {props.canCreateRows ? (
        <fieldset style={workbookFormGroupStyle}>
          <legend style={workbookFormHeadingStyle}>
            New {props.contract.title} draft
          </legend>
          <p style={workbookFormMessageStyle}>
            Create a new record. Source links are saved only when explicitly
            supplied in this draft.
          </p>
          {props.ownerBindings.includes("linked_note_create") ? (
            <NoteSheetAuthoring />
          ) : null}
          <GenericDraftFields {...props} />
          {props.canCreateRows ? (
            <Button
              data-testid={genericCreateSubmitTestId(
                props.contract.viewSchemaId,
              )}
              disabled={
                props.draftDisabled ||
                props.mutation.ordinaryCreate.busy(props.contract.viewSchemaId)
              }
              tone="primary"
              type="button"
              onClick={() => void props.submitCreate()}
            >
              Create {props.contract.title}
            </Button>
          ) : null}
        </fieldset>
      ) : (
        <GenericDraftFields {...props} />
      )}
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
  readonly patches: WorkbookExplicitPatchOwner;
  readonly disabledReason: string | null;
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
  if (props.selectedEdit.row === null) return null;
  const field = props.selectedEdit.field;
  return (
    <WorkbookInspectorDetails
      contract={props.contract}
      row={props.selectedEdit.row}
      editableFields={props.editableFields}
      activeField={props.editFieldKey}
      onEdit={props.setEditFieldKey}
      onDetach={() => props.setEditFieldKey("")}
      disabledReason={props.disabledReason}
      canSubmit={!props.mutationPending && props.edit.canSubmit}
      onSubmit={() => void props.submitEdit()}
      retainedWork={props.edit.retainedWork}
      onReviewDraft={(identity) => {
        props.setCollectionMode(
          identity.action === "remove" ? "remove" : "add",
        );
        props.setEditFieldKey(identity.fieldKey);
      }}
      editor={{
        content: field ? (
          <fieldset
            style={{ ...editRowStyle, border: 0, padding: 0, minWidth: 0 }}
          >
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
          </fieldset>
        ) : null,
        actions: field ? (
          <Button
            data-testid={genericEditSubmitTestId(props.contract.viewSchemaId)}
            disabled={props.mutationPending || !props.edit.canSubmit}
            tone="primary"
            type="button"
            onClick={() => void props.submitEdit()}
          >
            Update
          </Button>
        ) : null,
        feedback: (
          <>
            {props.fieldFeedback ? (
              <p id={feedbackId} role="alert" style={workbookFormMessageStyle}>
                {props.fieldFeedback}
              </p>
            ) : null}
            {props.actionError ? (
              <WorkbookInspectorPublicError error={props.actionError} />
            ) : null}
            <WorkbookExplicitPatchRecovery
              owner={props.patches}
              viewSchemaId={props.contract.viewSchemaId}
              recordId={props.selectedEdit.row.record_id}
              fieldKey={field?.fieldKey ?? ""}
            />
          </>
        ),
        retainedDraft: (
          <WorkbookInspectorDraftFeedback
            edit={props.edit}
            contract={props.contract}
            row={props.selectedEdit.row}
          />
        ),
      }}
    />
  );
}

type GenericRelationshipsProps = {
  readonly noteAssociations?:
    | readonly [WorkbookInspectorRegion, ...WorkbookInspectorRegion[]]
    | undefined;
  readonly referenceSummary: ReactNode;
  readonly party: ReturnType<typeof useGenericPartyLinkWorkflow>;
};

function GenericRelationships(props: GenericRelationshipsProps) {
  return (
    <>
      {props.referenceSummary}
      <PartyLinkPanel workflow={props.party} />
    </>
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
