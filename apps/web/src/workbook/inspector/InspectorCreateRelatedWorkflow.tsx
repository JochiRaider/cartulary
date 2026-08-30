import type { CSSProperties } from "react";
import { GenericMutationControl } from "../components/GenericMutationControl";
import type { GenericReferenceOptions } from "../models/workbookReferenceOptions";
import type { InspectorCreateRelatedWorkflowState } from "./useInspectorCreateRelatedWorkflow";

export function InspectorCreateRelatedWorkflow({
  referenceOptions,
  state,
  onCancel,
  onSubmit,
  onUpdateDraft,
}: {
  readonly referenceOptions: GenericReferenceOptions;
  readonly state: InspectorCreateRelatedWorkflowState;
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
              testId={controlId}
              value={state.draft[field.fieldKey] ?? ""}
              onChange={(value) => onUpdateDraft(field.fieldKey, value)}
            />
          </label>
        );
      })}
      {state.message === null ? null : (
        <p aria-live="assertive" role="alert" style={messageStyle}>
          {state.message}
        </p>
      )}
      <div style={actionsStyle}>
        <button disabled={state.isSubmitting} type="button" onClick={onSubmit}>
          Create related row
        </button>
        <button disabled={state.isSubmitting} type="button" onClick={onCancel}>
          Cancel
        </button>
      </div>
    </section>
  );
}

const workflowStyle = {
  display: "grid",
  gap: "0.6rem",
  paddingBlock: "0.5rem",
} satisfies CSSProperties;

const labelStyle = {
  display: "grid",
  gap: "0.3rem",
} satisfies CSSProperties;

const messageStyle = { margin: 0 } satisfies CSSProperties;

const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "0.5rem",
} satisfies CSSProperties;
