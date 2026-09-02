import type { ViewFieldContract } from "@cartulary/view-contracts";
import {
  type GenericCollectionMode,
  isMultilineGenericField,
} from "../models/genericWorkbookModel";
import {
  type GenericReferenceOptions,
  genericFieldUsesReferenceOptions,
  referenceOptionsForField,
} from "../models/workbookReferenceOptions";

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
      readonly kind: "collection_reference";
      readonly options: readonly GenericMutationControlOption[];
      readonly size: number;
    })
  | (GenericMutationControlContext & {
      readonly emptyLabel: "None" | "Select";
      readonly kind: "direct_reference";
      readonly options: readonly GenericMutationControlOption[];
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
  referenceOptions,
  surface,
}: {
  readonly collectionItems: readonly {
    readonly displayText: string;
    readonly itemRef: string;
  }[];
  readonly collectionMode: GenericCollectionMode;
  readonly field: ViewFieldContract;
  readonly referenceOptions: GenericReferenceOptions;
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

  if (
    field.writeKind === "action_payload" &&
    genericFieldUsesReferenceOptions(field)
  ) {
    const options = referenceOptionsForField(field, referenceOptions).map(
      (option) => ({ label: option.label, value: option.recordId }),
    );
    return {
      ...context,
      kind: "collection_reference",
      options,
      size: collectionSelectSize(options.length, surface),
    };
  }

  if (genericFieldUsesReferenceOptions(field)) {
    return {
      ...context,
      emptyLabel: field.clearable ? "None" : "Select",
      kind: "direct_reference",
      options: referenceOptionsForField(field, referenceOptions).map(
        (option) => ({ label: option.label, value: option.recordId }),
      ),
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
  if (field.writeKind === "action_payload" || isMultilineGenericField(field)) {
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
