import type { ViewContract } from "@cartulary/view-contracts";
import {
  appendFilterDraftMember,
  type FilterDraft,
  filterMemberControlKeys,
} from "./workbookQuery";

export type EqualityFilterDraft = Extract<FilterDraft, { readonly op: "eq" }>;

/** Declared choices guide authoring; they do not constrain query operands. */
export function enumFilterChoices(
  contract: ViewContract,
  draft: FilterDraft,
): readonly string[] | null {
  const field = contract.fieldMap[draft.fieldKey];
  return draft.op === "eq" &&
    contract.filterFields.includes(draft.fieldKey) &&
    field?.readKind === "enum" &&
    field.filterOps.includes("eq") &&
    field.enumValues?.length
    ? field.enumValues
    : null;
}

export function enumFilterMembers(
  draft: EqualityFilterDraft,
): readonly string[] {
  return draft.values.map((member) => member.value);
}

export function enumFilterHasCustomLiteral(
  draft: EqualityFilterDraft,
  choices: readonly string[],
): boolean {
  return draft.operandKind === "value"
    ? draft.value.length > 0 && !choices.includes(draft.value)
    : draft.operandKind === "values" &&
        enumFilterMembers(draft).some((value) => !choices.includes(value));
}

export function toggleEnumFilterChoice(
  draft: EqualityFilterDraft,
  value: string,
  checked: boolean,
): EqualityFilterDraft {
  const members = enumFilterMembers(draft);
  return {
    ...draft,
    values: checked
      ? members.includes(value)
        ? draft.values
        : appendFilterDraftMember(draft.values, value)
      : draft.values.filter((member) => member.value !== value),
  };
}

/** Order includes only mounted controls, including every literal row action. */
export function enumFilterControlKeys(
  draft: EqualityFilterDraft,
  choices: readonly string[],
  literalsOpen: boolean,
): readonly string[] {
  if (draft.operandKind === "null") return [];
  return [
    ...(draft.operandKind === "value"
      ? ["value"]
      : choices.map((value) => `enum_choice:${value}`)),
    "enum_literals",
    ...(literalsOpen
      ? draft.operandKind === "value"
        ? ["enum_literal"]
        : filterMemberControlKeys(draft.values)
      : []),
  ];
}
