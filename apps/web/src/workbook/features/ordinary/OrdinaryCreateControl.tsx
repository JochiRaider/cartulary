import type { ViewContract } from "@cartulary/view-contracts";
import { type ComponentProps, useContext, useSyncExternalStore } from "react";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import { WorkbookAuthoringReferenceControl } from "../../components/WorkbookAuthoringReferenceControl";
import { WorkbookReferenceContext } from "../../components/WorkbookReferenceControl";
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
  const actor = useContext(WorkbookReferenceContext)?.actorPresentation;
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
          return {
            recordId,
            displayText:
              views[0] === "incident_members" && actor?.userId === recordId
                ? `${actor.displayName} (${recordId})`
                : recordId,
            viewSchemaId: views[0] ?? "",
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
            ? selected
                .map((item) => item.displayText || item.recordId)
                .join("\n")
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
  if (views.length && reader)
    return (
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "minmax(0, 1fr) auto",
          ...(props.surface === "grid"
            ? ({ position: "absolute", inset: 0 } as const)
            : {}),
          alignItems: "center",
          gap: "var(--ct-spacing-xs)",
          minWidth: 0,
        }}
      >
        <div
          style={{ position: "relative", minWidth: 0, alignSelf: "stretch" }}
        >
          <GenericMutationControl {...props} disabled={false} />
        </div>
        <WorkbookAuthoringReferenceControl
          targetKey={`ordinary:${contract.viewSchemaId}:${schema.draft.id}:${props.field.fieldKey}`}
          maximum={props.field.writeKind === "action_payload" ? 64 : 1}
          compact={props.surface === "grid"}
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
