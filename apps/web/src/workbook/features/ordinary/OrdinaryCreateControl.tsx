import type { ViewContract } from "@cartulary/view-contracts";
import { type ComponentProps, useSyncExternalStore } from "react";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import { WorkbookAuthoringReferenceControl } from "../../components/WorkbookAuthoringReferenceControl";
import { referenceOptionsForField } from "../../models/workbookReferenceOptions";
import type { WorkbookOrdinaryCreateOwner } from "./WorkbookOrdinaryCreateOwner";

/** Borrowed authoring controls. Selected identity is retained by the owner, never a query page. */
export function OrdinaryCreateControl({
  owner,
  contract,
  ...props
}: ComponentProps<typeof GenericMutationControl> & {
  owner: WorkbookOrdinaryCreateOwner;
  contract: ViewContract;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  if (!owner.supports(contract.viewSchemaId))
    return <GenericMutationControl {...props} />;
  const schema = snapshot.schemas[contract.viewSchemaId];
  if (!schema) return null;
  const views =
    owner.contribution(contract.viewSchemaId)?.referenceViews(props.field) ??
    [];
  const retained = schema.draft.references[props.field.fieldKey];
  const selected =
    retained ??
    (props.value
      ? props.value.split("\n").map((recordId) => {
          const known = referenceOptionsForField(
            props.field,
            props.referenceOptions,
          ).find((item) => item.recordId === recordId);
          return {
            recordId,
            displayText: known?.label ?? "Selected reference",
            viewSchemaId: known?.viewSchemaId ?? views[0] ?? "",
          };
        })
      : []);
  if (!owner.canAuthor())
    return (
      <textarea
        aria-label={props.ariaLabel ?? `${props.field.label} retained draft`}
        data-testid={props.testId}
        readOnly
        rows={1}
        ref={props.focusTargetRef}
        value={
          views.length
            ? selected.map((item) => item.displayText).join("\n")
            : props.value
        }
        style={{
          width: "100%",
          boxSizing: "border-box",
          color: "inherit",
          background: "transparent",
        }}
      />
    );
  const reader = owner.getReader();
  if (views.length && !reader)
    return (
      <button
        type="button"
        disabled
        aria-label={`${props.field.label}: references unavailable`}
        data-testid={props.testId}
      >
        References unavailable
      </button>
    );
  if (views.length && reader)
    return (
      <div data-testid={props.testId}>
        <WorkbookAuthoringReferenceControl
          compact={props.surface === "grid"}
          focusTargetRef={props.focusTargetRef}
          label={props.field.label}
          testId={`${props.testId}-options`}
          views={views}
          multiple={props.field.writeKind === "action_payload"}
          selected={selected}
          reader={reader}
          revision={schema.referenceRevision}
          disabled={false}
          onApply={(items) =>
            owner.selectReferences(
              contract.viewSchemaId,
              props.field.fieldKey,
              items,
            )
          }
        />
      </div>
    );
  return <GenericMutationControl {...props} disabled={false} />;
}
