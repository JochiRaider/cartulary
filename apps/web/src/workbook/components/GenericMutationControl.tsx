import type { ViewFieldContract } from "@cartulary/view-contracts";
import type { CSSProperties } from "react";
import {
  type GenericCollectionMode,
  isMultilineGenericField,
  splitDraftValues,
} from "../models/genericWorkbookModel";
import {
  type GenericReferenceOptions,
  genericFieldUsesReferenceOptions,
  referenceOptionsForField,
} from "../models/workbookReferenceOptions";

function GenericMultiSelectControl({
  ariaLabel,
  id,
  options,
  size,
  style,
  testId,
  value,
  onChange,
}: {
  readonly ariaLabel: string;
  readonly id: string | undefined;
  readonly options: readonly {
    readonly label: string;
    readonly value: string;
  }[];
  readonly size: number;
  readonly style: CSSProperties;
  readonly testId: string;
  readonly value: string;
  readonly onChange: (value: string) => void;
}) {
  return (
    <select
      aria-label={ariaLabel}
      data-testid={testId}
      id={id}
      multiple
      size={size}
      style={style}
      value={splitDraftValues(value)}
      onChange={(event) => {
        onChange(
          Array.from(event.currentTarget.selectedOptions)
            .map((option) => option.value)
            .join("\n"),
        );
      }}
    >
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}

export function GenericMutationControl({
  collectionItems = [],
  collectionMode,
  field,
  id,
  referenceOptions,
  surface = "form",
  testId,
  value,
  onChange,
}: {
  collectionItems?: Array<{ itemRef: string; displayText: string }>;
  collectionMode: GenericCollectionMode;
  field: ViewFieldContract;
  id?: string;
  referenceOptions: GenericReferenceOptions;
  surface?: "form" | "grid";
  testId: string;
  value: string;
  onChange: (value: string) => void;
}) {
  const controlLabel = `${field.label} value`;
  const inputControlStyle = surface === "grid" ? gridInputStyle : inputStyle;
  const textareaControlStyle =
    surface === "grid" ? gridTextareaStyle : textareaStyle;
  const selectControlStyle = surface === "grid" ? gridSelectStyle : selectStyle;
  if (field.writeKind === "action_payload") {
    if (collectionMode === "remove") {
      return (
        <GenericMultiSelectControl
          ariaLabel={controlLabel}
          id={id}
          options={collectionItems.map((item) => ({
            label: item.displayText,
            value: item.itemRef,
          }))}
          size={
            surface === "grid"
              ? 1
              : Math.min(Math.max(collectionItems.length, 2), 6)
          }
          style={selectControlStyle}
          testId={testId}
          value={value}
          onChange={onChange}
        />
      );
    }

    const options = referenceOptionsForField(field, referenceOptions);
    if (genericFieldUsesReferenceOptions(field)) {
      return (
        <GenericMultiSelectControl
          ariaLabel={controlLabel}
          id={id}
          options={options.map((option) => ({
            label: option.label,
            value: option.recordId,
          }))}
          size={
            surface === "grid" ? 1 : Math.min(Math.max(options.length, 2), 6)
          }
          style={selectControlStyle}
          testId={testId}
          value={value}
          onChange={onChange}
        />
      );
    }

    return (
      <textarea
        aria-label={controlLabel}
        data-testid={testId}
        id={id}
        rows={surface === "grid" ? 1 : 3}
        style={textareaControlStyle}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
    );
  }

  const referenceChoices = referenceOptionsForField(field, referenceOptions);
  if (genericFieldUsesReferenceOptions(field)) {
    return (
      <select
        aria-label={controlLabel}
        data-testid={testId}
        id={id}
        style={selectControlStyle}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      >
        <option value="">{field.clearable ? "None" : "Select"}</option>
        {referenceChoices.map((option) => (
          <option key={option.recordId} value={option.recordId}>
            {option.label}
          </option>
        ))}
      </select>
    );
  }

  if (field.enumValues && field.enumValues.length > 0) {
    return (
      <select
        aria-label={controlLabel}
        data-testid={testId}
        id={id}
        style={selectControlStyle}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      >
        <option value="">Select</option>
        {field.enumValues.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
    );
  }

  if (field.readKind === "boolean") {
    return (
      <input
        aria-label={controlLabel}
        data-testid={testId}
        id={id}
        style={surface === "grid" ? gridCheckboxStyle : inputStyle}
        type="checkbox"
        checked={value === "true"}
        onChange={(event) => {
          onChange(event.target.checked ? "true" : "false");
        }}
      />
    );
  }

  if (field.readKind === "number") {
    return (
      <input
        aria-label={controlLabel}
        data-testid={testId}
        id={id}
        style={inputControlStyle}
        type="number"
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
    );
  }

  if (isMultilineGenericField(field)) {
    return (
      <textarea
        aria-label={controlLabel}
        data-testid={testId}
        id={id}
        rows={surface === "grid" ? 1 : 3}
        style={textareaControlStyle}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
    );
  }

  return (
    <input
      aria-label={controlLabel}
      data-testid={testId}
      id={id}
      placeholder={
        field.directScalarContractId === "timestamp_instant_v1"
          ? "RFC3339 timestamp"
          : undefined
      }
      style={inputControlStyle}
      type="text"
      value={value}
      onChange={(event) => {
        onChange(event.target.value);
      }}
    />
  );
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
