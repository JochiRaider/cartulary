import { type AriaAttributes, type RefCallback, useId, useState } from "react";
import {
  type EqualityFilterDraft,
  enumFilterHasCustomLiteral,
  enumFilterMembers,
  toggleEnumFilterChoice,
} from "../models/workbookEnumFilterOperand";
import type { FilterDraftControl } from "../models/workbookQuery";
import { WorkbookLiteralSetFilterOperand } from "./WorkbookLiteralSetFilterOperand";
import {
  inputStyle,
  secondaryButtonStyle,
  stackedLabelStyle,
} from "./workbookGridControlStyles";

/** Presentation disclosure is scoped; all operands remain in FilterDraft. */
export function useEnumLiteralDisclosure(
  scope: string,
  draft: EqualityFilterDraft | null,
  choices: readonly string[] | null,
) {
  const [disclosure, setDisclosure] = useState<{
    scope: string;
    open: boolean;
  } | null>(null);
  const initiallyOpen =
    draft !== null &&
    choices !== null &&
    enumFilterHasCustomLiteral(draft, choices);
  if (disclosure?.scope !== scope)
    setDisclosure({ scope, open: initiallyOpen });
  const open = disclosure?.scope === scope ? disclosure.open : initiallyOpen;
  return { open, setOpen: (open: boolean) => setDisclosure({ scope, open }) };
}

export function WorkbookEnumFilterOperand({
  draft,
  choices,
  onChange,
  disclosure,
  valueLabel = "Value",
  valueTestId,
  feedbackFor,
  registerItem,
}: {
  readonly draft: EqualityFilterDraft;
  readonly choices: readonly string[];
  readonly onChange: (draft: EqualityFilterDraft) => void;
  readonly disclosure: ReturnType<typeof useEnumLiteralDisclosure>;
  readonly valueLabel?: string;
  readonly valueTestId?: string;
  readonly feedbackFor: (
    control: FilterDraftControl,
  ) => Pick<AriaAttributes, "aria-invalid" | "aria-describedby">;
  readonly registerItem?: (key: string) => RefCallback<HTMLElement>;
}) {
  const literalsId = useId();
  const feedback = feedbackFor("value");
  const members = enumFilterMembers(draft);
  if (draft.operandKind === "null") return null;
  return (
    <>
      {draft.operandKind === "value" ? (
        <label style={stackedLabelStyle}>
          Value
          <select
            ref={registerItem?.("value")}
            aria-label={valueLabel}
            data-testid={valueTestId}
            {...feedback}
            style={{ ...inputStyle, minInlineSize: 0, inlineSize: "100%" }}
            value={choices.includes(draft.value) ? draft.value : ""}
            onChange={(event) =>
              onChange({
                ...draft,
                value: event.currentTarget.value,
                valueType: "string",
              })
            }
          >
            <option value="">
              {draft.value && !choices.includes(draft.value)
                ? "Custom literal (see below)"
                : "Choose a value"}
            </option>
            {choices.map((value) => (
              <option key={value} value={value}>
                {value}
              </option>
            ))}
          </select>
        </label>
      ) : (
        <fieldset
          style={{
            minInlineSize: 0,
            margin: 0,
            padding: "var(--ct-spacing-xs)",
            border: "var(--ct-border-hairline)",
          }}
        >
          <legend>Values</legend>
          {choices.map((value) => (
            <label
              key={value}
              style={{
                display: "flex",
                alignItems: "start",
                gap: "var(--ct-spacing-xs)",
                overflowWrap: "anywhere",
              }}
            >
              <input
                type="checkbox"
                ref={registerItem?.(`enum_choice:${value}`)}
                {...feedback}
                checked={members.includes(value)}
                onChange={(event) =>
                  onChange(
                    toggleEnumFilterChoice(
                      draft,
                      value,
                      event.currentTarget.checked,
                    ),
                  )
                }
              />
              {value}
            </label>
          ))}
        </fieldset>
      )}
      <button
        type="button"
        ref={registerItem?.("enum_literals")}
        aria-expanded={disclosure.open}
        aria-controls={literalsId}
        style={secondaryButtonStyle}
        onClick={() => disclosure.setOpen(!disclosure.open)}
      >
        Custom literals
      </button>
      {disclosure.open ? (
        <div
          id={literalsId}
          style={{
            display: "grid",
            gap: "var(--ct-spacing-xs)",
            minInlineSize: 0,
          }}
        >
          {draft.operandKind === "value" ? (
            <label style={stackedLabelStyle}>
              Literal value
              <input
                ref={registerItem?.("enum_literal")}
                aria-label={`${valueLabel} literal`}
                {...feedback}
                style={{ ...inputStyle, minInlineSize: 0, inlineSize: "100%" }}
                value={draft.value}
                onChange={(event) =>
                  onChange({
                    ...draft,
                    value: event.currentTarget.value,
                    valueType: "string",
                  })
                }
              />
            </label>
          ) : (
            <WorkbookLiteralSetFilterOperand
              members={draft.values}
              onChange={(values) => onChange({ ...draft, values })}
              valueLabel={`${valueLabel} literal`}
              feedbackFor={feedbackFor}
              registerItem={registerItem}
              addLabel="Add literal"
              removeLabel="Remove literal"
            />
          )}
        </div>
      ) : null}
    </>
  );
}
