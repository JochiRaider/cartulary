import { useLayoutEffect, useRef, useState } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import type { ObservationDraftStore } from "./ObservationDraftStore";
import { ObservationTargetPicker } from "./ObservationTargetPicker";
import {
  type ObservationDraft,
  observationCanTransition,
} from "./observationModel";
import type {
  ObservationIntent,
  ObservationReadPort,
} from "./observationOperation";
import { observationStack, observationText } from "./observationStyles";

export function ObservationDetails({
  item,
  reader,
  generation,
  draft,
  drafts,
  manage,
  disabled,
  onSubmit,
  targetLabel,
}: {
  item: IndicatorObservation;
  reader: ObservationReadPort;
  generation: number;
  draft: ObservationDraft;
  drafts: ObservationDraftStore;
  manage: boolean;
  disabled: boolean;
  onSubmit: (intent: ObservationIntent, draft: ObservationDraft) => void;
  targetLabel?: string | undefined;
}) {
  const [resolving, setResolving] = useState(false);
  const article = useRef<HTMLElement>(null),
    title = useRef<HTMLParagraphElement>(null),
    version = useRef(item.row_version);
  const retainFocus =
    version.current !== item.row_version &&
    article.current?.contains(document.activeElement);
  useLayoutEffect(() => {
    version.current = item.row_version;
    if (retainFocus && document.activeElement === document.body)
      title.current?.focus();
  }, [item.row_version, retainFocus]);
  return (
    <article
      ref={article}
      aria-label={`Observation: ${item.observed_text}`}
      style={{ ...observationStack, overflowWrap: "anywhere" }}
    >
      <p ref={title} tabIndex={-1} style={observationText}>
        <strong>{item.observed_text}</strong>
      </p>
      <p style={observationText}>
        {item.resolution_status}
        {item.resolved_indicator_record_id
          ? ` · ${targetLabel ?? "Indicator details unavailable"}`
          : ""}
      </p>
      <details>
        <summary>Observation provenance</summary>
        <dl>
          <dt>Source record</dt>
          <dd>{item.source_record_id}</dd>
          <dt>Source field</dt>
          <dd>{item.source_field_key}</dd>
          <dt>Origin</dt>
          <dd>{item.origin_kind}</dd>
          <dt>Source locator</dt>
          <dd>{item.origin_locator}</dd>
          <dt>Parsed type</dt>
          <dd>{item.parsed_indicator_type ?? "Not parsed"}</dd>
          <dt>Normalized candidate</dt>
          <dd>{item.normalized_candidate ?? "None"}</dd>
          <dt>Captured by</dt>
          <dd>{item.created_by_user_id}</dd>
          <dt>Captured at</dt>
          <dd>{item.created_at}</dd>
          <dt>Resolution actor</dt>
          <dd>{item.resolved_by_user_id ?? "None"}</dd>
          <dt>Resolution time</dt>
          <dd>{item.resolved_at ?? "None"}</dd>
          <dt>Resolution method</dt>
          <dd>{item.resolution_method ?? "None"}</dd>
          <dt>Resolved Indicator record</dt>
          <dd>{item.resolved_indicator_record_id ?? "None"}</dd>
        </dl>
        <p>Observed text is historical; later source edits do not change it.</p>
      </details>
      {manage ? (
        item.resolution_status === "dismissed" ? (
          <div>
            <WorkbookInspectorActionButton
              disabled={disabled}
              onClick={() =>
                onSubmit({ action: "restore", observation: item }, draft)
              }
            >
              Restore observation
            </WorkbookInspectorActionButton>
            <p>
              Restore returns this observation to unresolved without its former
              target.
            </p>
          </div>
        ) : (
          <>
            <div>
              <WorkbookInspectorActionButton
                disabled={disabled}
                onClick={() => setResolving((value) => !value)}
                aria-expanded={resolving}
              >
                {item.resolution_status === "resolved" ? "Reassign" : "Resolve"}
              </WorkbookInspectorActionButton>{" "}
              <WorkbookInspectorActionButton
                disabled={disabled}
                onClick={() =>
                  onSubmit({ action: "dismiss", observation: item }, draft)
                }
              >
                Dismiss observation
              </WorkbookInspectorActionButton>
            </div>
            {resolving ? (
              <div style={observationStack}>
                <ObservationTargetPicker
                  reader={reader}
                  generation={generation}
                  selected={draft.target}
                  disabled={disabled}
                  label="Resolve this observation"
                  onChange={(target) => drafts.update(draft.key, { target })}
                />
                {draft.target?.recordId ===
                item.resolved_indicator_record_id ? (
                  <p>
                    Select a different Indicator to reassign this observation.
                  </p>
                ) : null}
                <WorkbookInspectorActionButton
                  disabled={
                    disabled ||
                    !observationCanTransition(
                      item,
                      "resolve",
                      draft.target?.recordId,
                    )
                  }
                  onClick={() => {
                    if (draft.target)
                      onSubmit(
                        {
                          action: "resolve",
                          observation: item,
                          targetId: draft.target.recordId,
                        },
                        draft,
                      );
                  }}
                >
                  {item.resolution_status === "resolved"
                    ? "Reassign observation"
                    : "Resolve observation"}
                </WorkbookInspectorActionButton>
              </div>
            ) : null}
          </>
        )
      ) : null}
    </article>
  );
}
