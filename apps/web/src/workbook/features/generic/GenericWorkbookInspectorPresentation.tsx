import {
  coordinationWorkflowTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditActionSelectTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  genericWorkbookTestId,
} from "@cartulary/ui-contracts";
import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import type { Dispatch, SetStateAction } from "react";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import type { GenericCollectionMode } from "../../models/genericWorkbookModel";
import {
  genericCollectionSupportsRemove,
  genericRowLabel,
} from "../../models/genericWorkbookModel";
import type { GenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { WorkbookMutationCommandPorts } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOwnerBinding } from "../../policies/workbookSurfacePolicy";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { CoordinationWorkflowBindings } from "../coordination/CoordinationWorkflowBindings";
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
  readonly canCreateRows: boolean;
  readonly contract: ViewContract;
  readonly createDraft: Record<string, string>;
  readonly draftInspectorFields: readonly ViewFieldContract[];
  readonly invalidationKey: string;
  readonly linkedNoteSourceRecordId: string;
  readonly mutation: GenericSurfaceMutationController;
  readonly mutationCommands: WorkbookMutationCommandPorts;
  readonly ownerBindings: readonly WorkbookOwnerBinding[];
  readonly referenceOptions: GenericReferenceOptions;
  readonly rows: readonly WorkbookQueryRow[];
  readonly setCreateDraft: Dispatch<SetStateAction<Record<string, string>>>;
  readonly setLinkedNoteSourceRecordId: (value: string) => void;
  readonly submitCreate: () => Promise<void>;
  readonly subjectPresent: boolean;
};

function GenericWorkflow(props: GenericWorkflowProps) {
  return (
    <>
      {props.ownerBindings.includes("linked_note_create") ? (
        <label
          htmlFor={genericWorkbookTestId("note-source-record")}
          style={labelStyle}
        >
          Linked source for draft row
          <select
            data-testid={genericWorkbookTestId("note-source-record")}
            id={genericWorkbookTestId("note-source-record")}
            style={selectStyle}
            value={props.linkedNoteSourceRecordId}
            onChange={(event) =>
              props.setLinkedNoteSourceRecordId(event.target.value)
            }
          >
            <option value="">None</option>
            {props.referenceOptions.noteSourceRecords.map((option) => (
              <option key={option.recordId} value={option.recordId}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      ) : null}
      <GenericDraftFields {...props} />
      {props.canCreateRows ? (
        <button
          data-testid={genericCreateSubmitTestId(props.contract.viewSchemaId)}
          disabled={props.mutation.mutationPending}
          style={secondaryActionButtonStyle}
          type="button"
          onClick={() => void props.submitCreate()}
        >
          Commit draft row
        </button>
      ) : null}
      {props.subjectPresent ? (
        <CoordinationWorkflowBindings
          contract={props.contract}
          disabled={props.mutation.mutationPending}
          mutation={props.mutation}
          mutationCommands={props.mutationCommands.coordination}
          ownerBindings={props.ownerBindings}
          referenceOptions={props.referenceOptions}
          resetKey={props.invalidationKey}
          rows={props.rows}
        />
      ) : null}
    </>
  );
}

function GenericDraftFields(props: GenericWorkflowProps) {
  if (!props.canCreateRows || props.draftInspectorFields.length === 0)
    return null;
  return (
    <div style={draftInspectorFieldsStyle}>
      {props.draftInspectorFields.map((field) => {
        const controlId = `generic-create-inspector-${field.fieldKey}`;
        return (
          <label htmlFor={controlId} key={field.fieldKey} style={labelStyle}>
            {field.label}
            <GenericMutationControl
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
  readonly collectionItems: readonly {
    readonly displayText: string;
    readonly itemRef: string;
  }[];
  readonly collectionMode: GenericCollectionMode;
  readonly contract: ViewContract;
  readonly editableFields: readonly ViewFieldContract[];
  readonly editFieldKey: string;
  readonly editValue: string;
  readonly mutationPending: GenericSurfaceMutationController["mutationPending"];
  readonly onSelectRecord: (recordId: string) => void;
  readonly referenceOptions: GenericReferenceOptions;
  readonly rows: readonly WorkbookQueryRow[];
  readonly selectedEdit: SelectedEdit;
  readonly selectedRecordId: string;
  readonly setCollectionMode: (mode: GenericCollectionMode) => void;
  readonly setEditFieldKey: (fieldKey: string) => void;
  readonly setEditValue: (value: string) => void;
  readonly submitEdit: () => Promise<void>;
};

function GenericDetails(props: GenericDetailsProps) {
  if (props.rows.length === 0 || props.selectedEdit.field === null) return null;
  const field = props.selectedEdit.field;
  return (
    <div style={editRowStyle}>
      <select
        data-testid={genericEditRecordSelectTestId(props.contract.viewSchemaId)}
        style={selectStyle}
        value={props.selectedRecordId}
        onChange={(event) => props.onSelectRecord(event.target.value)}
      >
        <option value="">Row</option>
        {props.rows.map((row) => (
          <option key={row.record_id} value={row.record_id}>
            {genericRowLabel(props.contract, row)}
          </option>
        ))}
      </select>
      <select
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
      {field.writeKind === "action_payload" &&
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
            props.setEditValue("");
          }}
        >
          <option value="add">Add</option>
          <option value="remove">Remove</option>
        </select>
      ) : null}
      <GenericMutationControl
        collectionItems={props.collectionItems}
        collectionMode={props.collectionMode}
        field={field}
        referenceOptions={props.referenceOptions}
        testId={genericEditValueTestId(props.contract.viewSchemaId)}
        value={props.editValue}
        onChange={props.setEditValue}
      />
      <button
        data-testid={genericEditSubmitTestId(props.contract.viewSchemaId)}
        disabled={props.mutationPending}
        style={actionButtonStyle}
        type="button"
        onClick={() => void props.submitEdit()}
      >
        Update
      </button>
    </div>
  );
}

type GenericRelationshipsProps = {
  readonly disabled: boolean;
  readonly party: ReturnType<typeof useGenericPartyLinkWorkflow>;
  readonly partyLinkPairs: readonly {
    readonly key: string;
    readonly label: string;
  }[];
  readonly referenceOptions: GenericReferenceOptions;
  readonly rowSelected: boolean;
};

function GenericRelationships(props: GenericRelationshipsProps) {
  if (props.partyLinkPairs.length === 0 || !props.rowSelected) return null;
  return (
    <div style={editRowStyle}>
      <select
        aria-label="Party link field"
        data-testid={coordinationWorkflowTestId("party-pair")}
        style={selectStyle}
        value={props.party.selectedPartyLinkPair?.key ?? ""}
        onChange={(event) =>
          props.party.setPartyLinkPairKey(event.target.value)
        }
      >
        {props.partyLinkPairs.map((pair) => (
          <option key={pair.key} value={pair.key}>
            {pair.label}
          </option>
        ))}
      </select>
      <select
        aria-label="Existing party"
        data-testid={coordinationWorkflowTestId("party-existing")}
        style={selectStyle}
        value={props.party.partyLinkExistingPartyId}
        onChange={(event) =>
          props.party.setPartyLinkExistingPartyId(event.target.value)
        }
      >
        <option value="">Party</option>
        {props.referenceOptions.parties.map((option) => (
          <option key={option.recordId} value={option.recordId}>
            {option.label}
          </option>
        ))}
      </select>
      <PartyButtons party={props.party} disabled={props.disabled} />
    </div>
  );
}

function PartyButtons({
  disabled,
  party,
}: {
  readonly disabled: boolean;
  readonly party: ReturnType<typeof useGenericPartyLinkWorkflow>;
}) {
  const buttons = [
    [
      "party-create-from-text",
      "Create party from text",
      party.createPartyFromText,
    ],
    ["party-link-existing", "Link existing party", party.linkExistingParty],
    ["party-clear-link", "Clear party link", party.clearPartyLink],
    ["party-clear-text", "Clear party text", party.clearPartyText],
    ["party-clear-both", "Clear both", party.clearPartyBoth],
  ] as const;
  return (
    <>
      {buttons.map(([id, label, action]) => (
        <button
          data-testid={coordinationWorkflowTestId(id)}
          disabled={disabled}
          key={id}
          style={secondaryActionButtonStyle}
          type="button"
          onClick={() => void action()}
        >
          {label}
        </button>
      ))}
      {party.partialCompletionMessage === null ? null : (
        <div>
          <p
            data-testid={coordinationWorkflowTestId("party-partial-completion")}
            role="status"
          >
            {party.partialCompletionMessage}
          </p>
          <button
            data-testid={coordinationWorkflowTestId("party-retry-created-link")}
            disabled={disabled}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => void party.retryCreatedPartyLink()}
          >
            Retry link to created party
          </button>
        </div>
      )}
    </>
  );
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
