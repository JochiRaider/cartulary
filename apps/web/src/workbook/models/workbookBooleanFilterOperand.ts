import type { ViewContract } from "@cartulary/view-contracts";

/** Boolean choices live in the existing filter draft, never in presentation state. */
export type BooleanFilterOperandDraft = {
  readonly value: boolean | undefined;
  readonly values: readonly boolean[];
  /** A malformed restored argument requires an explicit replacement, never coercion. */
  readonly invalidRestoredArg?: Readonly<Record<string, unknown>>;
};

export function isBooleanEqualityFilter(
  contract: ViewContract,
  fieldKey: string,
  op: string,
): boolean {
  const field = contract.fieldMap[fieldKey];
  return (
    op === "eq" &&
    contract.filterFields.includes(fieldKey) &&
    field?.readKind === "boolean" &&
    field.filterOps.includes("eq")
  );
}

export function emptyBooleanFilterOperand(): BooleanFilterOperandDraft {
  return { value: undefined, values: [] };
}

export function restoreBooleanFilterOperand(
  arg: Readonly<Record<string, unknown>>,
): BooleanFilterOperandDraft {
  if (Object.keys(arg).length === 1) {
    if (arg.value === null) return emptyBooleanFilterOperand();
    if (typeof arg.value === "boolean") return { value: arg.value, values: [] };
    if (
      Array.isArray(arg.values) &&
      arg.values.length > 0 &&
      [...arg.values].every((value) => typeof value === "boolean")
    ) {
      return { value: undefined, values: [...arg.values] };
    }
  }
  return { ...emptyBooleanFilterOperand(), invalidRestoredArg: { ...arg } };
}

export function booleanFilterOperandDecision(
  operandKind: string,
  operand: BooleanFilterOperandDraft | undefined,
):
  | { readonly kind: "valid"; readonly arg: Record<string, unknown> }
  | { readonly kind: "invalid"; readonly message: string } {
  if (operand?.invalidRestoredArg !== undefined) {
    return {
      kind: "invalid",
      message:
        "This restored filter is not a valid boolean operand. Choose true, false, or empty to replace it.",
    };
  }
  if (operandKind === "null") return { kind: "valid", arg: { value: null } };
  if (operandKind === "value" && typeof operand?.value === "boolean") {
    return { kind: "valid", arg: { value: operand.value } };
  }
  if (
    operandKind === "values" &&
    Array.isArray(operand?.values) &&
    operand.values.length > 0 &&
    [...operand.values].every((value) => typeof value === "boolean")
  ) {
    // The server owns logical-set normalization and canonical ordering.
    return { kind: "valid", arg: { values: [...operand.values] } };
  }
  return {
    kind: "invalid",
    message:
      operandKind === "values"
        ? "Select at least one boolean value: true or false. Every selected value must be a boolean."
        : "Choose true or false before applying this filter.",
  };
}

export function booleanFilterControlKeys(
  operandKind: string,
): readonly string[] {
  return operandKind === "null"
    ? []
    : operandKind === "values"
      ? ["boolean_choice:true", "boolean_choice:false"]
      : ["value"];
}
