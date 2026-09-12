import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
} from "@cartulary/ui-contracts";
import {
  useEffect,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import { ContextualReferenceControl } from "./ContextualReferenceControl";
import { contextualReferenceKind } from "./contextualCreateModel";
import type { WorkbookContextualTaskDecisionCreateOwner } from "./WorkbookContextualTaskDecisionCreateOwner";

export function ContextualCreateForm({
  owner,
  attachment,
  onSubmit,
  disabled = false,
}: {
  readonly owner: WorkbookContextualTaskDecisionCreateOwner;
  readonly attachment: symbol;
  readonly onSubmit: () => void;
  readonly disabled?: boolean;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const form = useRef<HTMLElement>(null);
  const visibleDraftId =
    snapshot.attachment === attachment ? snapshot.draft?.id : undefined;
  useLayoutEffect(() => {
    if (visibleDraftId === undefined) return;
    const element = form.current;
    const origin = document.activeElement;
    element?.querySelector<HTMLElement>("input, select, textarea")?.focus();
    return () => {
      // Restore only while this presentation still owns focus. Detached or late
      // completion cannot pull focus out of the analyst's new context.
      if (
        element?.contains(document.activeElement) &&
        origin instanceof HTMLElement &&
        origin.isConnected &&
        origin.getClientRects().length &&
        !origin.matches(":disabled")
      )
        origin.focus({ preventScroll: true });
    };
  }, [visibleDraftId]);
  useEffect(() => () => owner.detach(attachment), [owner, attachment]);
  const draft = snapshot.draft,
    reader = owner.getReader();
  if (!draft || snapshot.attachment !== attachment || !reader) return null;
  return (
    <section
      ref={form}
      aria-label={draft.feature.label}
      style={{
        display: "grid",
        gap: "var(--ct-spacing-sm)",
        minWidth: 0,
        overflowWrap: "anywhere",
      }}
    >
      <p>
        Create in {draft.target.title}. Source:{" "}
        {draft.labels[draft.source.recordId] ?? draft.source.recordId} (
        {draft.presentation.surfaceLabel}).
      </p>
      <fieldset
        disabled={disabled || !owner.canSubmit()}
        style={{
          border: 0,
          padding: 0,
          margin: 0,
          minWidth: 0,
          display: "grid",
          gap: "var(--ct-spacing-sm)",
        }}
      >
        <legend>Related {draft.target.title}</legend>
        {draft.target.fields
          .filter((field) => field.createWritable)
          .map((field) => {
            const id = `contextual-create-${field.fieldKey}`,
              error = snapshot.errors[field.fieldKey];
            return (
              <div key={field.fieldKey} style={{ minWidth: 0 }}>
                {contextualReferenceKind(field) ? (
                  <ContextualReferenceControl
                    draft={draft}
                    field={field}
                    reader={reader}
                    revision={snapshot.candidateRevision}
                    onChange={(value, labels) =>
                      owner.update(field.fieldKey, value, labels)
                    }
                  />
                ) : (
                  <label
                    htmlFor={id}
                    style={{ display: "grid", gap: "0.25rem" }}
                  >
                    {field.label}
                    <GenericMutationControl
                      id={id}
                      field={
                        field.fieldKey === "decision.status"
                          ? {
                              ...field,
                              enumValues:
                                field.enumValues?.filter(
                                  (value) => value !== "superseded",
                                ) ?? null,
                            }
                          : field
                      }
                      collectionMode="add"
                      referenceOptions={emptyGenericReferenceOptions()}
                      value={draft.values[field.fieldKey] ?? ""}
                      testId={genericCreateFieldTestId(field.fieldKey)}
                      invalid={!!error}
                      describedBy={error ? `${id}-error` : undefined}
                      onChange={(value) => owner.update(field.fieldKey, value)}
                    />
                  </label>
                )}
                {error ? (
                  <p id={`${id}-error`} role="alert">
                    {error}
                  </p>
                ) : null}
              </div>
            );
          })}
        {draft.target.createInputs.map((input) => (
          <label key={input.inputKey}>
            {input.inputKey}
            <input
              value={draft.values[input.inputKey] ?? ""}
              required={input.required}
              onChange={(event) =>
                owner.update(input.inputKey, event.currentTarget.value)
              }
            />
          </label>
        ))}
      </fieldset>
      {draft.target.viewSchemaId.includes("decisions") ? (
        <p>An unset decision time uses the creation commit time.</p>
      ) : null}
      {snapshot.needsReview ? (
        <p role="status">
          Context changed. Your values are retained. Review them before
          creating.
        </p>
      ) : null}
      {snapshot.message ? <p role="status">{snapshot.message}</p> : null}
      <div style={{ display: "flex", flexWrap: "wrap", gap: "0.5rem" }}>
        {snapshot.needsReview ? (
          <WorkbookInspectorActionButton
            tone="secondary"
            type="button"
            disabled={disabled || !owner.canSubmit()}
            onClick={() => void owner.review()}
          >
            Review retained context
          </WorkbookInspectorActionButton>
        ) : null}
        <WorkbookInspectorActionButton
          tone="primary"
          type="button"
          data-testid={genericCreateSubmitTestId(draft.target.viewSchemaId)}
          disabled={disabled || snapshot.needsReview || !owner.canSubmit()}
          onClick={onSubmit}
        >
          Create related row
        </WorkbookInspectorActionButton>
        <WorkbookInspectorActionButton
          tone="secondary"
          type="button"
          onClick={() => owner.detach(attachment)}
        >
          Keep draft and close
        </WorkbookInspectorActionButton>
        <WorkbookInspectorActionButton
          tone="secondary"
          type="button"
          disabled={disabled}
          onClick={() => owner.discard()}
        >
          Discard draft
        </WorkbookInspectorActionButton>
      </div>
    </section>
  );
}
