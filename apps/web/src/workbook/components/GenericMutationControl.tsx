import type { ViewFieldContract } from "@cartulary/view-contracts";
import type { CSSProperties, RefCallback } from "react";
import {
  type GenericCollectionMode,
  splitDraftValues,
} from "../models/genericWorkbookModel";
import type { GenericReferenceOptions } from "../models/workbookReferenceOptions";
import {
  type GenericMutationControlDescriptor,
  type GenericMutationControlSurface,
  resolveGenericMutationControl,
} from "./genericMutationControlModel";

type GenericMutationControlRef = RefCallback<
  HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
>;

type GenericMutationControlElementProps = {
  readonly descriptor: GenericMutationControlDescriptor;
  readonly focusTargetRef?: GenericMutationControlRef | undefined;
  readonly id?: string | undefined;
  readonly testId: string;
  readonly value: string;
  readonly onChange: (value: string) => void;
};

export function GenericMutationControl({
  collectionItems = [],
  collectionMode,
  field,
  focusTargetRef,
  id,
  referenceOptions,
  surface = "form",
  testId,
  value,
  onChange,
}: {
  collectionItems?: readonly { itemRef: string; displayText: string }[];
  collectionMode: GenericCollectionMode;
  field: ViewFieldContract;
  focusTargetRef?: GenericMutationControlRef | undefined;
  id?: string;
  referenceOptions: GenericReferenceOptions;
  surface?: GenericMutationControlSurface;
  testId: string;
  value: string;
  onChange: (value: string) => void;
}) {
  const descriptor = resolveGenericMutationControl({
    collectionItems,
    collectionMode,
    field,
    referenceOptions,
    surface,
  });
  const props = {
    descriptor,
    focusTargetRef,
    id,
    testId,
    value,
    onChange,
  };
  switch (descriptor.kind) {
    case "collection_removal":
    case "collection_reference":
      return <GenericMultiSelectControl {...props} descriptor={descriptor} />;
    case "direct_reference":
    case "enumerated_value":
      return <GenericSingleSelectControl {...props} descriptor={descriptor} />;
    case "boolean":
      return <GenericBooleanControl {...props} descriptor={descriptor} />;
    case "number":
    case "text":
      return <GenericTextInputControl {...props} descriptor={descriptor} />;
    case "multiline_text":
      return <GenericTextareaControl {...props} descriptor={descriptor} />;
  }
}

function GenericMultiSelectControl({
  descriptor,
  focusTargetRef,
  id,
  testId,
  value,
  onChange,
}: GenericMutationControlElementProps & {
  readonly descriptor: Extract<
    GenericMutationControlDescriptor,
    { readonly kind: "collection_reference" | "collection_removal" }
  >;
}) {
  return (
    <select
      aria-label={descriptor.ariaLabel}
      data-testid={testId}
      id={id}
      multiple
      ref={focusTargetRef}
      size={descriptor.size}
      style={selectControlStyle(descriptor.surface)}
      value={splitDraftValues(value)}
      onChange={(event) => {
        onChange(
          Array.from(event.currentTarget.selectedOptions)
            .map((option) => option.value)
            .join("\n"),
        );
      }}
    >
      {descriptor.options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}

function GenericSingleSelectControl({
  descriptor,
  focusTargetRef,
  id,
  testId,
  value,
  onChange,
}: GenericMutationControlElementProps & {
  readonly descriptor: Extract<
    GenericMutationControlDescriptor,
    { readonly kind: "direct_reference" | "enumerated_value" }
  >;
}) {
  const options =
    descriptor.kind === "direct_reference"
      ? descriptor.options
      : descriptor.options.map((option) => ({ label: option, value: option }));
  return (
    <select
      aria-label={descriptor.ariaLabel}
      data-testid={testId}
      id={id}
      ref={focusTargetRef}
      style={selectControlStyle(descriptor.surface)}
      value={value}
      onChange={(event) => onChange(event.target.value)}
    >
      <option value="">
        {descriptor.kind === "direct_reference"
          ? descriptor.emptyLabel
          : "Select"}
      </option>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}

function GenericBooleanControl({
  descriptor,
  focusTargetRef,
  id,
  testId,
  value,
  onChange,
}: GenericMutationControlElementProps & {
  readonly descriptor: Extract<
    GenericMutationControlDescriptor,
    { readonly kind: "boolean" }
  >;
}) {
  return (
    <input
      aria-label={descriptor.ariaLabel}
      checked={value === "true"}
      data-testid={testId}
      id={id}
      ref={focusTargetRef}
      style={descriptor.surface === "grid" ? gridCheckboxStyle : inputStyle}
      type="checkbox"
      onChange={(event) => onChange(event.target.checked ? "true" : "false")}
    />
  );
}

function GenericTextInputControl({
  descriptor,
  focusTargetRef,
  id,
  testId,
  value,
  onChange,
}: GenericMutationControlElementProps & {
  readonly descriptor: Extract<
    GenericMutationControlDescriptor,
    { readonly kind: "number" | "text" }
  >;
}) {
  return (
    <input
      aria-label={descriptor.ariaLabel}
      data-testid={testId}
      id={id}
      inputMode={descriptor.kind === "number" ? "numeric" : undefined}
      placeholder={
        descriptor.kind === "text" ? descriptor.placeholder : undefined
      }
      ref={focusTargetRef}
      style={inputControlStyle(descriptor.surface)}
      type={descriptor.kind === "number" ? descriptor.inputType : "text"}
      value={value}
      onChange={(event) => onChange(event.target.value)}
    />
  );
}

function GenericTextareaControl({
  descriptor,
  focusTargetRef,
  id,
  testId,
  value,
  onChange,
}: GenericMutationControlElementProps & {
  readonly descriptor: Extract<
    GenericMutationControlDescriptor,
    { readonly kind: "multiline_text" }
  >;
}) {
  return (
    <textarea
      aria-label={descriptor.ariaLabel}
      data-testid={testId}
      id={id}
      ref={focusTargetRef}
      rows={descriptor.rows}
      style={textareaControlStyle(descriptor.surface)}
      value={value}
      onChange={(event) => onChange(event.target.value)}
    />
  );
}

function inputControlStyle(surface: GenericMutationControlSurface) {
  return surface === "grid" ? gridInputStyle : inputStyle;
}

function textareaControlStyle(surface: GenericMutationControlSurface) {
  return surface === "grid" ? gridTextareaStyle : textareaStyle;
}

function selectControlStyle(surface: GenericMutationControlSurface) {
  return surface === "grid" ? gridSelectStyle : selectStyle;
}

const inputStyle = {
  boxSizing: "border-box",
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
} satisfies CSSProperties;

const textareaStyle = {
  ...inputStyle,
  resize: "vertical",
} satisfies CSSProperties;

const selectStyle = {
  ...inputStyle,
  appearance: "auto",
} satisfies CSSProperties;

const gridInputStyle = {
  ...inputStyle,
  position: "absolute" as const,
  inset: 0,
  inlineSize: "100%",
  blockSize: "100%",
  minHeight: 0,
  width: "100%",
  border: "none",
  borderRadius: 0,
  background: "transparent",
  padding: "var(--cartulary-grid-cell-padding)",
  fontSize: "var(--cartulary-grid-font-size)",
  lineHeight: "var(--cartulary-grid-line-height)",
  color: "var(--ct-colors-ink)",
} satisfies CSSProperties;

const gridTextareaStyle = {
  ...gridInputStyle,
  resize: "none",
  overflow: "auto",
} satisfies CSSProperties;

const gridSelectStyle = {
  ...gridInputStyle,
  appearance: "auto",
} satisfies CSSProperties;

const gridCheckboxStyle = {
  ...gridInputStyle,
  inlineSize: "100%",
  margin: 0,
  padding: "calc(var(--cartulary-grid-row-height) / 4)",
} satisfies CSSProperties;
