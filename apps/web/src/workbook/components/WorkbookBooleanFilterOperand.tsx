import type { AriaAttributes, RefCallback } from "react";
import { emptyBooleanFilterOperand } from "../models/workbookBooleanFilterOperand";
import type { FilterDraft } from "../models/workbookQuery";
import { inputStyle, stackedLabelStyle } from "./workbookGridControlStyles";

/** Native controls project typed choices; changing them never applies a query. */
export function WorkbookBooleanFilterOperand({
  draft,
  onChange,
  valueLabel = "Value",
  valueTestId,
  feedback,
  registerItem,
}: {
  readonly draft: Extract<FilterDraft, { readonly op: "eq" }>;
  readonly onChange: (
    draft: Extract<FilterDraft, { readonly op: "eq" }>,
  ) => void;
  readonly valueLabel?: string;
  readonly valueTestId?: string;
  readonly feedback: Pick<AriaAttributes, "aria-invalid" | "aria-describedby">;
  readonly registerItem?: (key: string) => RefCallback<HTMLElement>;
}) {
  if (draft.operandKind === "null") return null;
  const operand = draft.booleanOperand ?? emptyBooleanFilterOperand();
  const replace = (next: typeof operand) =>
    onChange({ ...draft, valueType: "boolean", booleanOperand: next });
  if (draft.operandKind === "value")
    return (
      <label style={stackedLabelStyle}>
        Value
        <select
          ref={registerItem?.("value")}
          aria-label={valueLabel}
          data-testid={valueTestId}
          {...feedback}
          style={{ ...inputStyle, minInlineSize: 0, inlineSize: "100%" }}
          value={
            typeof operand.value === "boolean" ? String(operand.value) : ""
          }
          onChange={(event) => {
            const value = event.currentTarget.value;
            if (value !== "" && value !== "true" && value !== "false") return;
            replace({
              value: value === "" ? undefined : value === "true",
              values: operand.values,
            });
          }}
        >
          <option value="">Choose a value</option>
          <option value="true">true</option>
          <option value="false">false</option>
        </select>
      </label>
    );
  return (
    <fieldset
      style={{
        minInlineSize: 0,
        margin: 0,
        padding: "var(--ct-spacing-xs)",
        border: "var(--ct-border-hairline)",
      }}
    >
      <legend>Values</legend>
      {[true, false].map((value) => (
        <label
          key={String(value)}
          style={{
            display: "flex",
            alignItems: "center",
            gap: "var(--ct-spacing-xs)",
          }}
        >
          <input
            type="checkbox"
            ref={registerItem?.(`boolean_choice:${value}`)}
            {...feedback}
            checked={
              Array.isArray(operand.values) && operand.values.includes(value)
            }
            onChange={(event) => {
              const members =
                operand.invalidRestoredArg === undefined ? operand.values : [];
              replace({
                value: operand.value,
                values: event.currentTarget.checked
                  ? members.includes(value)
                    ? members
                    : [...members, value]
                  : members.filter((member) => member !== value),
              });
            }}
          />
          {String(value)}
        </label>
      ))}
    </fieldset>
  );
}
