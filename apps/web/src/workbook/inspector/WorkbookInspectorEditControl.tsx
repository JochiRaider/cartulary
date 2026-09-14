import type { ComponentProps } from "react";
import { GenericMutationControl } from "../components/GenericMutationControl";
import { referenceOptionsForField } from "../models/workbookReferenceOptions";
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
  const options = referenceOptionsForField(props.field, props.referenceOptions);
  return (
    <>
      <GenericMutationControl
        {...props}
        focusTargetRef={(element) => {
          edit.controlRef.current = element;
        }}
        disabled={!edit.canEdit}
        value={edit.value ?? ""}
        retainedOptions={edit.draft?.references.map((item) => ({
          value: item.recordId,
          label: item.displayText,
        }))}
        onChange={(value) => {
          if (props.field.directReferenceContractId || options.length) {
            edit.selectReferences(
              value
                ? value.split("\n").map((recordId) => {
                    const item = options.find(
                      (item) => item.recordId === recordId,
                    );
                    return (
                      edit.draft?.references.find(
                        (item) => item.recordId === recordId,
                      ) ?? {
                        recordId,
                        displayText: item?.label ?? "Selected reference",
                        viewSchemaId: item?.viewSchemaId ?? "",
                      }
                    );
                  })
                : [],
            );
          } else edit.update(value);
        }}
      />
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
