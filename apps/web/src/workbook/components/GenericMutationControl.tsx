import type { ViewFieldContract } from "@cartulary/view-contracts";
import type { CSSProperties, RefCallback } from "react";
import type { GenericCollectionMode } from "../models/genericWorkbookModel";
import {
  type GenericMutationControlDescriptor,
  type GenericMutationControlSurface,
  resolveGenericMutationControl,
} from "./genericMutationControlModel";
import { workbookFormInputStyle as inputStyle } from "./workbookFormStyles";

type GenericMutationControlRef = RefCallback<
  HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement | HTMLButtonElement
>;

type GenericMutationControlFeedback = {
  readonly disabled?: boolean | undefined;
  readonly readOnly?: boolean | undefined;
  readonly ariaLabel?: string | undefined;
  readonly describedBy?: string | undefined;
  readonly invalid?: boolean | undefined;
};

type GenericMutationControlElementProps = GenericMutationControlFeedback & {
  readonly descriptor: GenericMutationControlDescriptor;
  readonly focusTargetRef?: GenericMutationControlRef | undefined;
  readonly id?: string | undefined;
  readonly testId: string;
  readonly value: string;
  readonly onChange: (value: string) => void;
};

export function GenericMutationControl({
  disabled,
  readOnly,
  ariaLabel,
  describedBy,
  invalid,
  collectionItems = [],
  collectionMode,
  field,
  focusTargetRef,
  id,
  retainedOptions = [],
  surface = "form",
  testId,
  value,
  onChange,
}: GenericMutationControlFeedback & {
  collectionItems?: readonly { itemRef: string; displayText: string }[];
  collectionMode: GenericCollectionMode;
  field: ViewFieldContract;
  focusTargetRef?: GenericMutationControlRef | undefined;
  id?: string;
  retainedOptions?: readonly { value: string; label: string }[] | undefined;
  surface?: GenericMutationControlSurface;
  testId: string;
  value: string;
  onChange: (value: string) => void;
}) {
  let descriptor = resolveGenericMutationControl({
    collectionItems,
    collectionMode,
    field,
    surface,
  });
  if (descriptor.kind === "collection_removal") {
    const selected = value ? value.split("\n") : [];
    const available = descriptor.options;
    descriptor = {
      ...descriptor,
      options: [
        ...available,
        ...selected
          .filter((id) => !available.some((option) => option.value === id))
          .map(
            (id) =>
              retainedOptions.find((option) => option.value === id) ?? {
                value: id,
                label:
                  surface === "grid"
                    ? id
                    : "Selected item (outside current options)",
              },
          ),
      ],
    };
  }
  const props = {
    disabled,
    readOnly,
    ariaLabel,
    describedBy,
    invalid,
    descriptor,
    focusTargetRef,
    id,
    testId,
    value,
    onChange,
  };
  switch (descriptor.kind) {
    case "collection_removal":
      return <GenericMultiSelectControl {...props} descriptor={descriptor} />;
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
  disabled,
  ariaLabel,
  describedBy,
  invalid,
  descriptor,
  focusTargetRef,
  id,
  testId,
  value,
  onChange,
}: GenericMutationControlElementProps & {
  readonly descriptor: Extract<
    GenericMutationControlDescriptor,
    { readonly kind: "collection_removal" }
  >;
}) {
  return (
    <select
      disabled={disabled}
      aria-label={ariaLabel ?? descriptor.ariaLabel}
      aria-describedby={describedBy}
      aria-invalid={invalid}
      data-testid={testId}
      id={id}
      multiple
      ref={focusTargetRef}
      size={descriptor.size}
      style={selectControlStyle(descriptor.surface)}
      value={value ? value.split("\n") : []}
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
  disabled,
  ariaLabel,
  describedBy,
  invalid,
  descriptor,
  focusTargetRef,
  id,
  testId,
  value,
  onChange,
}: GenericMutationControlElementProps & {
  readonly descriptor: Extract<
    GenericMutationControlDescriptor,
    { readonly kind: "enumerated_value" }
  >;
}) {
  const options = descriptor.options.map((option) => ({
    label: option,
    value: option,
  }));
  return (
    <select
      disabled={disabled}
      aria-label={ariaLabel ?? descriptor.ariaLabel}
      aria-describedby={describedBy}
      aria-invalid={invalid}
      data-testid={testId}
      id={id}
      ref={focusTargetRef}
      style={selectControlStyle(descriptor.surface)}
      value={value}
      onChange={(event) => onChange(event.target.value)}
    >
      <option value="">Select</option>
      {descriptor.surface === "grid" &&
      value !== "" &&
      !options.some((option) => option.value === value) ? (
        <option value={value}>{value}</option>
      ) : null}
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}

function GenericBooleanControl({
  disabled,
  ariaLabel,
  describedBy,
  invalid,
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
      disabled={disabled}
      aria-label={ariaLabel ?? descriptor.ariaLabel}
      aria-describedby={describedBy}
      aria-invalid={invalid}
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
  disabled,
  readOnly,
  ariaLabel,
  describedBy,
  invalid,
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
      disabled={disabled && !readOnly}
      readOnly={readOnly}
      aria-label={ariaLabel ?? descriptor.ariaLabel}
      aria-describedby={describedBy}
      aria-invalid={invalid}
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
  disabled,
  readOnly,
  ariaLabel,
  describedBy,
  invalid,
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
      disabled={disabled && !readOnly}
      readOnly={readOnly}
      aria-label={ariaLabel ?? descriptor.ariaLabel}
      aria-describedby={describedBy}
      aria-invalid={invalid}
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
