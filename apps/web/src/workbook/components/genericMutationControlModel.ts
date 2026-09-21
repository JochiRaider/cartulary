import type { ViewFieldContract } from "@cartulary/view-contracts";
import {
  type GenericCollectionMode,
  isMultilineGenericField,
} from "../models/genericWorkbookModel";
export type GenericMutationControlSurface = "form" | "grid";

type GenericMutationControlOption = {
  readonly label: string;
  readonly value: string;
};

type GenericMutationControlContext = {
  readonly ariaLabel: string;
  readonly surface: GenericMutationControlSurface;
};

export type GenericMutationControlDescriptor =
  | (GenericMutationControlContext & {
      readonly kind: "collection_removal";
      readonly options: readonly GenericMutationControlOption[];
      readonly size: number;
    })
  | (GenericMutationControlContext & {
      readonly kind: "enumerated_value";
      readonly options: readonly string[];
    })
  | (GenericMutationControlContext & { readonly kind: "boolean" })
  | (GenericMutationControlContext & {
      readonly inputType: "number" | "text";
      readonly kind: "number";
    })
  | (GenericMutationControlContext & {
      readonly kind: "multiline_text";
      readonly rows: number;
    })
  | (GenericMutationControlContext & {
      readonly kind: "text";
      readonly placeholder?: "RFC3339 timestamp" | undefined;
    });

export function resolveGenericMutationControl({
  collectionItems,
  collectionMode,
  field,
  surface,
}: {
  readonly collectionItems: readonly {
    readonly displayText: string;
    readonly itemRef: string;
  }[];
  readonly collectionMode: GenericCollectionMode;
  readonly field: ViewFieldContract;
  readonly surface: GenericMutationControlSurface;
}): GenericMutationControlDescriptor {
  const context = { ariaLabel: `${field.label} value`, surface };
  if (field.writeKind === "action_payload" && collectionMode === "remove") {
    const options = collectionItems.map((item) => ({
      label: item.displayText,
      value: item.itemRef,
    }));
    return {
      ...context,
      kind: "collection_removal",
      options,
      size: collectionSelectSize(options.length, surface),
    };
  }

  if (field.enumValues !== null && field.enumValues.length > 0) {
    return { ...context, kind: "enumerated_value", options: field.enumValues };
  }
  if (field.readKind === "boolean") return { ...context, kind: "boolean" };
  if (field.readKind === "number") {
    return {
      ...context,
      inputType: surface === "grid" ? "text" : "number",
      kind: "number",
    };
  }
  if (
    field.writeKind === "action_payload" ||
    isMultilineGenericField(field) ||
    (surface === "form" &&
      field.stringContractId === "timeline_visible_text_v1")
  ) {
    return {
      ...context,
      kind: "multiline_text",
      rows: surface === "grid" ? 1 : 3,
    };
  }
  return {
    ...context,
    kind: "text",
    placeholder:
      field.directScalarContractId === "timestamp_instant_v1"
        ? "RFC3339 timestamp"
        : undefined,
  };
}

function collectionSelectSize(
  optionCount: number,
  surface: GenericMutationControlSurface,
): number {
  return surface === "grid" ? 1 : Math.min(Math.max(optionCount, 2), 6);
}
