import { getReferenceFieldContract } from "@cartulary/view-contracts";
import type { ComponentProps } from "react";
import { GenericMutationControl } from "../components/GenericMutationControl";
import { WorkbookReferenceControl } from "../components/WorkbookReferenceControl";
import { WorkbookInspectorActionButton as Button } from "./presentation/WorkbookInspectorActions";
import type { WorkbookInspectorEditDraft } from "./useWorkbookInspectorEditDraft";

/** Retains selected identities independently from the current option page. */
export function WorkbookInspectorEditControl({
  edit,
  ...props
}: Omit<
  ComponentProps<typeof GenericMutationControl>,
  "value" | "onChange" | "disabled"
> & { edit: WorkbookInspectorEditDraft }) {
  const reference = getReferenceFieldContract(
    edit.identity.viewSchemaId,
    props.field.fieldKey,
  );
  return (
    <>
      {reference && props.collectionMode !== "remove" ? (
        <WorkbookReferenceControl
          field={reference}
          label={props.field.label}
          value={edit.value ?? ""}
          disabled={!edit.canEdit}
          sourceRecordId={edit.identity.recordId}
          retained={edit.draft?.references}
          testId={props.testId}
          invalid={props.invalid}
          describedBy={props.describedBy}
          focusTargetRef={(element) => {
            edit.controlRef.current = element;
          }}
          onChange={edit.update}
          onAccept={(items) =>
            edit.selectReferences(
              items.map((item) => ({
                recordId: item.identity.id,
                displayText: item.displayText,
                viewSchemaId: item.viewSchemaId,
              })),
            )
          }
        />
      ) : (
        <GenericMutationControl
          {...props}
          disabled={!edit.canEdit}
          readOnly={!edit.canEdit && !edit.needsResume}
          value={edit.value ?? ""}
          focusTargetRef={(element) => {
            edit.controlRef.current = element;
          }}
          onChange={edit.update}
        />
      )}
      {props.field.clearable && props.field.writeKind === "direct_value" ? (
        <Button
          type="button"
          disabled={!edit.canEdit}
          onClick={() => edit.update(null)}
        >
          Clear {props.field.label}
        </Button>
      ) : null}
      {edit.draft?.value === null ? (
        <span role="status">This field will be cleared when you update.</span>
      ) : null}
    </>
  );
}
