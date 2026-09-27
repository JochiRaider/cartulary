import {
  decisionsViewSchemaId,
  listViewContracts,
  partiesViewSchemaId,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { useContext, useEffect, useRef, useState } from "react";
import { useSelectedReferenceRemovalFocus } from "../../components/useSelectedReferenceRemovalFocus";
import { WorkbookAuthoringReferencePicker } from "../../components/WorkbookAuthoringReferencePicker";
import { WorkbookCandidateAuthorityContext } from "../../hooks/useWorkbookCandidateDiscovery";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import {
  type ContextualCreateDraft,
  contextualReferenceIds,
  contextualReferenceKind,
} from "./contextualCreateModel";
import type { ContextualCreateReader } from "./contextualCreateOperation";

export function ContextualReferenceControl({
  disabled,
  draft,
  field,
  reader,
  revision,
  onChange,
}: {
  readonly disabled: boolean;
  readonly draft: ContextualCreateDraft;
  readonly field: ViewFieldContract;
  readonly reader: ContextualCreateReader;
  readonly revision: number;
  readonly onChange: (
    value: string,
    labels: Readonly<Record<string, string>>,
  ) => void;
}) {
  const [open, setOpen] = useState(false);
  useEffect(() => {
    if (disabled) setOpen(false);
  }, [disabled]);
  const trigger = useRef<HTMLButtonElement>(null);
  const group = useRef<HTMLFieldSetElement>(null);
  const ids = contextualReferenceIds(draft.values[field.fieldKey] ?? "");
  const authority = useContext(WorkbookCandidateAuthorityContext);
  const removalFocus = useSelectedReferenceRemovalFocus({
    ids,
    scopeKey: `${draft.id}:${field.fieldKey}:${authority.identity}`,
    disabled: disabled || !authority.canRead,
    fallback: () => trigger.current,
    groupRef: group,
  });
  const close = () => {
    setOpen(false);
    trigger.current?.focus({ preventScroll: true });
  };
  return (
    <fieldset
      ref={group}
      aria-label={`${field.label} selected references`}
      data-reference-focus-fallback
      tabIndex={-1}
      style={{
        display: "grid",
        gap: "0.5rem",
        minWidth: 0,
        border: 0,
        margin: 0,
        padding: 0,
      }}
    >
      <span>{field.label}</span>
      {ids.length ? (
        <ul>
          {ids.map((id) => (
            <li key={id} style={{ overflowWrap: "anywhere" }}>
              {draft.labels[id] ??
                (contextualReferenceKind(field) === "members" &&
                id === draft.actorId
                  ? "Current actor"
                  : id)}
              {draft.seeds[field.fieldKey]?.split("\n").includes(id)
                ? " (source context)"
                : ""}{" "}
              <WorkbookInspectorActionButton
                ref={removalFocus.buttonRef(id)}
                tone="secondary"
                type="button"
                disabled={disabled || !authority.canRead}
                aria-label={`Remove ${field.label} ${draft.labels[id] ?? id}${ids.filter((item) => (draft.labels[item] ?? item) === (draft.labels[id] ?? id)).length > 1 ? ` (${id})` : ""}`}
                onClick={(event) =>
                  removalFocus.remove(id, event.currentTarget, () =>
                    onChange(ids.filter((item) => item !== id).join("\n"), {}),
                  )
                }
              >
                Remove
              </WorkbookInspectorActionButton>
              <details>
                <summary>Reference ID</summary>
                <span>{id}</span>
              </details>
            </li>
          ))}
        </ul>
      ) : (
        <span>
          {contextualReferenceKind(field) === "members"
            ? "Uses the current actor as owner when unset."
            : "No references selected."}
        </span>
      )}
      <WorkbookInspectorActionButton
        tone="secondary"
        ref={trigger}
        type="button"
        disabled={disabled || !authority.canRead}
        onClick={() => setOpen(true)}
      >
        Choose {field.label}
      </WorkbookInspectorActionButton>
      {open && !disabled ? (
        <ReferencePicker
          key={`${draft.id}:${field.fieldKey}`}
          draft={draft}
          field={field}
          reader={reader}
          revision={revision}
          onCancel={close}
          onApply={(value, labels) => {
            onChange(value, labels);
            close();
          }}
        />
      ) : null}
    </fieldset>
  );
}
function ReferencePicker({
  draft,
  field,
  reader,
  revision,
  onApply,
  onCancel,
}: {
  readonly draft: ContextualCreateDraft;
  readonly field: ViewFieldContract;
  readonly reader: ContextualCreateReader;
  readonly revision: number;
  readonly onApply: (
    value: string,
    labels: Readonly<Record<string, string>>,
  ) => void;
  readonly onCancel: () => void;
}) {
  const kind = contextualReferenceKind(field);
  const initial =
    kind === "members"
      ? "incident_members"
      : kind === "parties"
        ? partiesViewSchemaId
        : kind === "decisions"
          ? decisionsViewSchemaId
          : draft.source.viewSchemaId;
  const views =
    kind === "records"
      ? listViewContracts().map((contract) => contract.viewSchemaId)
      : [initial];
  const ids = contextualReferenceIds(draft.values[field.fieldKey] ?? "");
  return (
    <WorkbookAuthoringReferencePicker
      regionLabel={`Choose ${field.label}`}
      targetKey={`contextual:${draft.id}:${field.fieldKey}`}
      label={field.label}
      testId={`contextual-reference-${field.fieldKey}`}
      views={views}
      initialView={initial}
      multiple={field.readKind === "collection"}
      maximum={field.readKind === "collection" ? 64 : 1}
      selected={ids.map((recordId) => ({
        recordId,
        displayText: draft.labels[recordId] ?? recordId,
        viewSchemaId: initial,
      }))}
      reader={reader}
      revision={revision}
      disabled={false}
      onCancel={onCancel}
      onApply={(items) =>
        onApply(
          items.map((item) => item.recordId).join("\n"),
          Object.fromEntries(
            items.map((item) => [item.recordId, item.displayText]),
          ),
        )
      }
    />
  );
}
