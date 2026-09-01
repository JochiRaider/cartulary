import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import { GenericMutationControl } from "../components/GenericMutationControl";
import type { GenericReferenceOptions } from "../models/workbookReferenceOptions";
import type { InspectorRelatedRecordFormModel } from "./inspectorRelatedRecordModel";
import { WorkbookInspectorActionButton } from "./presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "./presentation/WorkbookInspectorFeedback";

export function InspectorCreateRelatedWorkflow({
  referenceOptions,
  state,
  onCancel,
  onSubmit,
  onUpdateDraft,
}: {
  readonly referenceOptions: GenericReferenceOptions;
  readonly state: InspectorRelatedRecordFormModel;
  readonly onCancel: () => void;
  readonly onSubmit: () => void;
  readonly onUpdateDraft: (fieldKey: string, value: string) => void;
}) {
  const createFields = state.targetContract.fields.filter(
    (field) => field.createWritable,
  );
  return (
    <section aria-label={state.featureGroup.label} style={workflowStyle}>
      <p style={messageStyle}>Create in {state.targetContract.title}</p>
      {createFields.map((field) => {
        const controlId = `inspector-related-${field.fieldKey}`;
        return (
          <label htmlFor={controlId} key={field.fieldKey} style={labelStyle}>
            {field.label}
            <GenericMutationControl
              collectionMode="add"
              field={field}
              id={controlId}
              referenceOptions={referenceOptions}
              testId={genericCreateFieldTestId(field.fieldKey)}
              value={state.draft[field.fieldKey] ?? ""}
              onChange={(value) => onUpdateDraft(field.fieldKey, value)}
            />
          </label>
        );
      })}
      {state.error === null ? null : (
        <WorkbookInspectorPublicError error={state.error} />
      )}
      <div style={actionsStyle}>
        <WorkbookInspectorActionButton
          data-testid={genericCreateSubmitTestId(
            state.targetContract.viewSchemaId,
          )}
          disabled={state.isSubmitting}
          tone="primary"
          onClick={onSubmit}
        >
          Create related row
        </WorkbookInspectorActionButton>
        <WorkbookInspectorActionButton
          disabled={state.isSubmitting}
          tone="secondary"
          onClick={onCancel}
        >
          Cancel
        </WorkbookInspectorActionButton>
      </div>
    </section>
  );
}

const workflowStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  paddingBlock: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const labelStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;

const messageStyle = { margin: 0 } satisfies CSSProperties;

const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
