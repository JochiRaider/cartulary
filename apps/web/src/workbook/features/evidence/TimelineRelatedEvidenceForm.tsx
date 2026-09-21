import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
} from "@cartulary/ui-contracts";
import { useLayoutEffect, useRef, useSyncExternalStore } from "react";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import {
  workbookFormActionsStyle,
  workbookFormFieldStackStyle,
  workbookFormFieldsStyle,
  workbookFormGroupStyle,
  workbookFormHeadingStyle,
  workbookFormInputStyle,
  workbookFormMessageStyle,
} from "../../components/workbookFormStyles";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { RelatedEvidencePartyControl } from "./RelatedEvidencePartyControl";
import { metadataEvidenceStates } from "./timelineRelatedEvidenceModel";
import type { WorkbookTimelineRelatedEvidenceOwner } from "./WorkbookTimelineRelatedEvidenceOwner";

export function TimelineRelatedEvidenceForm({
  owner,
  attachment,
  onSubmit,
  onReview,
}: {
  readonly owner: WorkbookTimelineRelatedEvidenceOwner;
  readonly attachment: symbol;
  readonly onSubmit: () => void;
  readonly onReview: () => void;
}) {
  const state = useSyncExternalStore(owner.subscribe, owner.getSnapshot),
    form = useRef<HTMLElement>(null);
  const visible = state.attachment === attachment ? state.draft?.id : undefined;
  useLayoutEffect(() => {
    if (visible === undefined) return;
    const element = form.current,
      origin = document.activeElement;
    element
      ?.querySelector<HTMLElement>("input, select, textarea")
      ?.focus({ preventScroll: true });
    return () => {
      if (
        element?.contains(document.activeElement) &&
        origin instanceof HTMLElement &&
        origin.isConnected &&
        origin.getClientRects().length &&
        !origin.matches(":disabled")
      )
        origin.focus({ preventScroll: true });
    };
  }, [visible]);
  const draft = state.draft,
    reader = owner.getReader();
  if (!draft || state.attachment !== attachment || !reader) return null;
  return (
    <section
      ref={form}
      aria-label={draft.feature.label}
      style={workbookFormFieldsStyle}
    >
      <p>
        Create Evidence metadata, then link the new record to the original
        Timeline row. Each step saves separately.
      </p>
      <details>
        <summary>Original Timeline record</summary>
        <p>{draft.presentation.label}</p>
        <p>{draft.source.recordId}</p>
      </details>
      <p style={workbookFormMessageStyle}>
        Closing keeps unfinished work in this session. Only the explicit create
        action saves it.
      </p>
      <fieldset
        disabled={owner.busy || !owner.canSubmit()}
        style={workbookFormGroupStyle}
      >
        <legend style={workbookFormHeadingStyle}>Evidence metadata</legend>
        {draft.target.fields
          .filter((field) => field.createWritable)
          .map((field) => {
            const id = `related-evidence-${field.fieldKey}`,
              error = state.errors[field.fieldKey],
              value = draft.values[field.fieldKey] ?? "";
            return (
              <div key={field.fieldKey}>
                {field.directReferenceContractId ===
                "same_incident_party_ref_v1" ? (
                  <RelatedEvidencePartyControl
                    disabled={owner.busy || !owner.canSubmit()}
                    targetKey={`related-evidence:${draft.id}`}
                    field={field}
                    value={value}
                    labels={draft.labels}
                    reader={reader}
                    revision={state.candidateRevision}
                    errorId={error ? `${id}-error` : undefined}
                    onChange={(next, labels) =>
                      owner.update(field.fieldKey, next, labels)
                    }
                  />
                ) : (
                  <label htmlFor={id} style={workbookFormFieldStackStyle}>
                    {field.label}
                    {field.fieldKey === "evidence.lifecycle_state" ? (
                      <select
                        id={id}
                        value={value}
                        aria-invalid={!!error}
                        aria-describedby={error ? `${id}-error` : undefined}
                        data-testid={genericCreateFieldTestId(field.fieldKey)}
                        onChange={(event) =>
                          owner.update(
                            field.fieldKey,
                            event.currentTarget.value,
                          )
                        }
                        style={workbookFormInputStyle}
                      >
                        <option value="">Use server default (Requested)</option>
                        {metadataEvidenceStates.map((state) => (
                          <option key={state} value={state}>
                            {state === "pending_receipt"
                              ? "Pending receipt"
                              : state[0]?.toUpperCase() + state.slice(1)}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <GenericMutationControl
                        id={id}
                        field={field}
                        collectionMode="add"
                        value={value}
                        testId={genericCreateFieldTestId(field.fieldKey)}
                        invalid={!!error}
                        describedBy={error ? `${id}-error` : undefined}
                        onChange={(next) => owner.update(field.fieldKey, next)}
                      />
                    )}
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
      </fieldset>
      <p>
        Title is optional when other metadata is entered. Party references do
        not replace Collector or Source text. Metadata creation does not make a
        file available.
      </p>
      {state.errors.payload ? <p role="alert">{state.errors.payload}</p> : null}
      {state.message ? <p role="status">{state.message}</p> : null}
      {state.needsReview ? (
        <p role="status">
          Context changed. Review the original source and retained metadata
          before creating.
        </p>
      ) : null}
      <div style={workbookFormActionsStyle}>
        {state.needsReview ? (
          <Button
            type="button"
            tone="secondary"
            disabled={!owner.canSubmit()}
            aria-disabled={owner.busy || !owner.canSubmit()}
            onClick={onReview}
          >
            Review Evidence draft
          </Button>
        ) : null}
        <Button
          type="button"
          tone="primary"
          data-testid={genericCreateSubmitTestId(draft.target.viewSchemaId)}
          disabled={!owner.canSubmit() || state.needsReview}
          aria-disabled={owner.busy || !owner.canSubmit() || state.needsReview}
          onClick={onSubmit}
        >
          Create Evidence and link
        </Button>
        <Button
          type="button"
          tone="secondary"
          onClick={() => owner.detach(attachment)}
        >
          Keep draft and close
        </Button>
        <Button
          type="button"
          tone="secondary"
          disabled={owner.busy}
          onClick={() => owner.discard()}
        >
          Discard Evidence draft
        </Button>
      </div>
    </section>
  );
}
