import {
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  workbookFilterClearButtonTestId,
  workbookFilterOperatorTestId,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
  workbookQueryOverflowEntryTestId,
} from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import { SlidersHorizontal } from "lucide-react";
import { type RefObject, useMemo } from "react";
import { useRegisteredOverlayNavigation } from "../../shared/useRegisteredOverlayNavigation";
import {
  parseDeclaredFieldKey,
  parseWorkbookBooleanDraftValue,
  validateFilterDraft,
  type WorkbookGridQueryCommand,
  type WorkbookGridQueryControlProjection,
} from "../models/workbookGridQueryControls";
import {
  type FilterDraft,
  filterDraftForField,
  isWorkbookFilterOperator,
} from "../models/workbookQuery";
import {
  clearButtonStyle,
  controlButtonStyle,
  filterValidationStyle,
  fixedMenuFrameStyle,
  inputStyle,
  menuStyle,
  primaryButtonStyle,
  queryListButtonStyle,
  queryListStyle,
  secondaryButtonStyle,
  selectStyle,
  stackedLabelStyle,
} from "./workbookGridControlStyles";

export function WorkbookFiltersControl({
  contract,
  draft,
  editingFieldKey,
  filterCount,
  isOpen,
  onApply,
  onChangeDraft,
  onClose,
  onCommand,
  onEditFilter,
  onEditQueryEntry,
  onToggle,
  projection,
  returnFocusRef,
  surface,
  triggerRef,
}: {
  readonly contract: ViewContract;
  readonly draft: FilterDraft;
  readonly editingFieldKey: string | null;
  readonly filterCount: number;
  readonly isOpen: boolean;
  readonly onApply: (draft: FilterDraft) => void;
  readonly onChangeDraft: (draft: FilterDraft) => void;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onEditFilter: (fieldKey: string) => void;
  readonly onEditQueryEntry: (
    entry: WorkbookGridQueryControlProjection["chips"][number],
  ) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly returnFocusRef: RefObject<HTMLElement | null>;
  readonly surface: string;
  readonly triggerRef: RefObject<HTMLButtonElement | null>;
}) {
  const itemKeys = useMemo(
    () => [
      "field",
      "operator",
      "operand_kind",
      "value",
      ...projection.chips
        .filter((chip) => chip.identity.kind === "filter")
        .flatMap((chip) => [`edit:${chip.key}`, `remove:${chip.key}`]),
      ...projection.hiddenChips
        .filter((chip) => chip.identity.kind !== "filter")
        .map((chip) => `overflow:${chip.key}`),
      ...(filterCount > 0 ? ["clear"] : []),
      ...(editingFieldKey === null ? [] : ["remove_editing"]),
      "cancel",
      "apply",
    ],
    [editingFieldKey, filterCount, projection.chips, projection.hiddenChips],
  );
  const initialKey = editingFieldKey === null ? "field" : "operator";
  const navigation = useRegisteredOverlayNavigation({
    initialItemKey: initialKey,
    isOpen,
    itemKeys,
    onRequestClose: onClose,
    preferredReturnFocusRef: returnFocusRef,
    subjectKey: surface,
    trapTab: true,
    triggerRef,
  });
  const validation = validateFilterDraft(contract, draft);
  const hiddenCount = projection.hiddenChips.length;
  const field = contract.fieldMap[draft.fieldKey];
  const declaredOperators =
    field?.filterOps.filter(isWorkbookFilterOperator) ?? [];

  return (
    <div style={fixedMenuFrameStyle}>
      <button
        ref={triggerRef}
        aria-controls={
          isOpen ? workbookFilterPopoverTestId(surface) : undefined
        }
        aria-expanded={isOpen}
        aria-haspopup="dialog"
        aria-label={
          hiddenCount > 0
            ? `Filters, ${filterCount} active filters, ${hiddenCount} hidden query entries`
            : `Filters, ${filterCount} active filters`
        }
        data-testid={workbookFilterPopoverTriggerTestId(surface)}
        style={controlButtonStyle}
        type="button"
        onClick={() => {
          if (!isOpen) navigation.prepareOpen(initialKey);
          onToggle();
        }}
      >
        <SlidersHorizontal aria-hidden="true" size={15} />
        Filters{filterCount > 0 ? ` ${filterCount}` : ""}
      </button>
      {isOpen ? (
        <div
          aria-label={editingFieldKey === null ? "Add filter" : "Edit filter"}
          data-testid={workbookFilterPopoverTestId(surface)}
          id={workbookFilterPopoverTestId(surface)}
          role="dialog"
          style={filterPopoverStyle}
          tabIndex={-1}
          onBlur={navigation.onOverlayBlur}
          onKeyDown={(event) => {
            if (navigation.activeKey === null) return;
            navigation.onItemKeyDown(event, navigation.activeKey);
          }}
        >
          <strong>
            {editingFieldKey === null ? "Add filter" : "Edit filter"}
          </strong>
          <label style={stackedLabelStyle}>
            Field
            <select
              ref={navigation.registerItem("field")}
              data-testid={gridFilterFieldTestId(surface)}
              disabled={editingFieldKey !== null}
              style={selectStyle}
              value={draft.fieldKey}
              onChange={(event) => {
                const fieldKey = parseDeclaredFieldKey(
                  event.currentTarget.value,
                  contract.filterFields,
                );
                if (fieldKey === null) return;
                onChangeDraft(filterDraftForField(contract, fieldKey));
              }}
            >
              {contract.filterFields.map((fieldKey) => (
                <option key={fieldKey} value={fieldKey}>
                  {contract.fieldMap[fieldKey]?.label ?? fieldKey}
                </option>
              ))}
            </select>
          </label>
          <label style={stackedLabelStyle}>
            Operator
            <select
              ref={navigation.registerItem("operator")}
              data-testid={workbookFilterOperatorTestId(surface)}
              style={selectStyle}
              value={draft.op}
              onChange={(event) => {
                const op = event.currentTarget.value;
                if (!isWorkbookFilterOperator(op)) return;
                onChangeDraft(
                  filterDraftForField(contract, draft.fieldKey, op),
                );
              }}
            >
              {declaredOperators.map((op) => (
                <option key={op} value={op}>
                  {operatorLabel(op)}
                </option>
              ))}
            </select>
          </label>
          <FilterOperandControl
            draft={draft}
            navigation={navigation}
            onChangeDraft={onChangeDraft}
            surface={surface}
          />
          {validation.kind === "invalid" ? (
            <p role="status" style={filterValidationStyle}>
              {validation.message}
            </p>
          ) : null}
          <AppliedFilterActions
            filterCount={filterCount}
            navigation={navigation}
            onCommand={onCommand}
            onEditFilter={onEditFilter}
            onEditQueryEntry={onEditQueryEntry}
            projection={projection}
            surface={surface}
          />
          <div style={popoverActionsStyle}>
            {editingFieldKey === null ? null : (
              <button
                ref={navigation.registerItem("remove_editing")}
                style={clearButtonStyle}
                type="button"
                onClick={() => {
                  onCommand({
                    kind: "filter_remove",
                    fieldKey: editingFieldKey,
                  });
                  onClose();
                }}
              >
                Remove filter
              </button>
            )}
            <button
              ref={navigation.registerItem("cancel")}
              style={secondaryButtonStyle}
              type="button"
              onClick={onClose}
            >
              Cancel
            </button>
            <button
              ref={navigation.registerItem("apply")}
              data-testid={gridFilterApplyTestId(surface)}
              disabled={validation.kind === "invalid"}
              style={primaryButtonStyle}
              type="button"
              onClick={() => {
                if (validation.kind === "valid") onApply(draft);
              }}
            >
              Apply
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function FilterOperandControl({
  draft,
  navigation,
  onChangeDraft,
  surface,
}: {
  readonly draft: FilterDraft;
  readonly navigation: ReturnType<typeof useRegisteredOverlayNavigation>;
  readonly onChangeDraft: (draft: FilterDraft) => void;
  readonly surface: string;
}) {
  if (draft.op === "eq") {
    return (
      <>
        <label style={stackedLabelStyle}>
          Match
          <select
            ref={navigation.registerItem("operand_kind")}
            aria-label="Equality operand kind"
            style={selectStyle}
            value={draft.operandKind}
            onChange={(event) => {
              const operandKind = event.currentTarget.value;
              if (
                operandKind !== "value" &&
                operandKind !== "values" &&
                operandKind !== "null"
              ) {
                return;
              }
              onChangeDraft({ ...draft, operandKind });
            }}
          >
            <option value="value">Equals value</option>
            <option value="values">Is one of</option>
            <option value="null">Is empty</option>
          </select>
        </label>
        {draft.operandKind === "null" ? null : draft.operandKind ===
          "values" ? (
          <TextOperand
            draft={draft}
            label="Values"
            navigation={navigation}
            onValue={(values) => onChangeDraft({ ...draft, values })}
            placeholder="Comma-separated values"
            surface={surface}
            value={draft.values}
          />
        ) : draft.valueType === "boolean" ? (
          <label style={stackedLabelStyle}>
            Value
            <select
              ref={navigation.registerItem("value")}
              data-testid={gridFilterValueTestId(surface)}
              style={selectStyle}
              value={draft.booleanValue}
              onChange={(event) => {
                const booleanValue = parseWorkbookBooleanDraftValue(
                  event.currentTarget.value,
                );
                if (booleanValue !== null) {
                  onChangeDraft({ ...draft, booleanValue });
                }
              }}
            >
              <option value="">Select value</option>
              <option value="true">true</option>
              <option value="false">false</option>
            </select>
          </label>
        ) : (
          <TextOperand
            draft={draft}
            label="Value"
            navigation={navigation}
            onValue={(value) => onChangeDraft({ ...draft, value })}
            placeholder="Value"
            surface={surface}
            value={draft.value}
          />
        )}
      </>
    );
  }
  if (draft.op === "range") {
    return (
      <div style={rangeStyle}>
        <label style={stackedLabelStyle}>
          Lower bound
          <span style={boundStyle}>
            <select
              aria-label="Lower-bound comparison"
              value={draft.lowerKind}
              onChange={(event) => {
                const lowerKind = event.currentTarget.value;
                if (lowerKind === "gt" || lowerKind === "gte") {
                  onChangeDraft({ ...draft, lowerKind });
                }
              }}
            >
              <option value="gte">At least</option>
              <option value="gt">Greater than</option>
            </select>
            <input
              ref={navigation.registerItem("value")}
              data-testid={gridFilterValueTestId(surface)}
              style={inputStyle}
              value={draft.lowerValue}
              onChange={(event) =>
                onChangeDraft({
                  ...draft,
                  lowerValue: event.currentTarget.value,
                })
              }
            />
          </span>
        </label>
        <label style={stackedLabelStyle}>
          Upper bound
          <span style={boundStyle}>
            <select
              aria-label="Upper-bound comparison"
              value={draft.upperKind}
              onChange={(event) => {
                const upperKind = event.currentTarget.value;
                if (upperKind === "lt" || upperKind === "lte") {
                  onChangeDraft({ ...draft, upperKind });
                }
              }}
            >
              <option value="lte">At most</option>
              <option value="lt">Less than</option>
            </select>
            <input
              aria-label="Upper-bound value"
              style={inputStyle}
              value={draft.upperValue}
              onChange={(event) =>
                onChangeDraft({
                  ...draft,
                  upperValue: event.currentTarget.value,
                })
              }
            />
          </span>
        </label>
      </div>
    );
  }
  if (draft.op === "contains_any" || draft.op === "contains_all") {
    return (
      <TextOperand
        draft={draft}
        label="Values"
        navigation={navigation}
        onValue={(values) => onChangeDraft({ ...draft, values })}
        placeholder="Comma-separated values"
        surface={surface}
        value={draft.values}
      />
    );
  }
  if (draft.op === "full_text") {
    return (
      <TextOperand
        draft={draft}
        label="Query"
        navigation={navigation}
        onValue={(query) => onChangeDraft({ ...draft, query })}
        placeholder="Search tokens"
        surface={surface}
        value={draft.query}
      />
    );
  }
  if (draft.op === "prefix") {
    return (
      <TextOperand
        draft={draft}
        label="Value"
        navigation={navigation}
        onValue={(value) => onChangeDraft({ ...draft, value })}
        placeholder="Prefix"
        surface={surface}
        value={draft.value}
      />
    );
  }
  return null;
}

function TextOperand({
  label,
  navigation,
  onValue,
  placeholder,
  surface,
  value,
}: {
  readonly draft: FilterDraft;
  readonly label: string;
  readonly navigation: ReturnType<typeof useRegisteredOverlayNavigation>;
  readonly onValue: (value: string) => void;
  readonly placeholder: string;
  readonly surface: string;
  readonly value: string;
}) {
  return (
    <label style={stackedLabelStyle}>
      {label}
      <input
        ref={navigation.registerItem("value")}
        data-testid={gridFilterValueTestId(surface)}
        placeholder={placeholder}
        style={inputStyle}
        type="text"
        value={value}
        onChange={(event) => onValue(event.currentTarget.value)}
      />
    </label>
  );
}

function AppliedFilterActions({
  filterCount,
  navigation,
  onCommand,
  onEditFilter,
  onEditQueryEntry,
  projection,
  surface,
}: {
  readonly filterCount: number;
  readonly navigation: ReturnType<typeof useRegisteredOverlayNavigation>;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onEditFilter: (fieldKey: string) => void;
  readonly onEditQueryEntry: (
    entry: WorkbookGridQueryControlProjection["chips"][number],
  ) => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly surface: string;
}) {
  const filterEntries = projection.chips.filter(
    (chip) => chip.identity.kind === "filter",
  );
  const otherOverflow = projection.hiddenChips.filter(
    (chip) => chip.identity.kind !== "filter",
  );
  if (filterEntries.length === 0 && otherOverflow.length === 0) return null;
  return (
    <section aria-label="Applied query overflow" style={queryListStyle}>
      {filterEntries.length === 0 ? null : <strong>Applied filters</strong>}
      {filterEntries.map((chip) => (
        <div key={chip.key} style={appliedRowStyle}>
          <button
            ref={navigation.registerItem(`edit:${chip.key}`)}
            aria-label={`Edit ${chip.accessibleName}`}
            data-testid={workbookQueryOverflowEntryTestId(
              surface,
              "filter",
              chip.identity.fieldKey,
            )}
            style={queryListButtonStyle}
            type="button"
            onClick={() => onEditFilter(chip.identity.fieldKey)}
          >
            {chip.label}
          </button>
          <button
            ref={navigation.registerItem(`remove:${chip.key}`)}
            aria-label={`Remove ${chip.accessibleName}`}
            type="button"
            onClick={() => onCommand(chip.removeCommand)}
          >
            ×
          </button>
        </div>
      ))}
      {filterCount > 0 ? (
        <button
          ref={navigation.registerItem("clear")}
          data-testid={workbookFilterClearButtonTestId(surface)}
          style={clearButtonStyle}
          type="button"
          onClick={() => onCommand({ kind: "filters_clear" })}
        >
          Clear filters
        </button>
      ) : null}
      {otherOverflow.length === 0 ? null : <strong>More applied query</strong>}
      {otherOverflow.map((chip) => (
        <button
          key={chip.key}
          ref={navigation.registerItem(`overflow:${chip.key}`)}
          aria-label={`Edit ${chip.accessibleName}, hidden from the view bar`}
          data-testid={workbookQueryOverflowEntryTestId(
            surface,
            chip.identity.kind,
            chip.identity.fieldKey,
          )}
          style={queryListButtonStyle}
          type="button"
          onClick={() => onEditQueryEntry(chip)}
        >
          {chip.label}
        </button>
      ))}
    </section>
  );
}

function operatorLabel(op: FilterDraft["op"]): string {
  return (
    {
      contains_all: "Contains all",
      contains_any: "Contains any",
      eq: "Equals",
      full_text: "Full text",
      prefix: "Starts with",
      range: "Range",
    } as const
  )[op];
}

const filterPopoverStyle = {
  ...menuStyle,
  inlineSize: "min(var(--ct-layout-viewBarOverlayMaxInlineSize), 92vw)",
  gap: "var(--ct-spacing-sm)",
};
const popoverActionsStyle = {
  display: "flex",
  justifyContent: "end",
  gap: "var(--ct-spacing-xs)",
};
const rangeStyle = { display: "grid", gap: "var(--ct-spacing-sm)" };
const boundStyle = {
  display: "grid",
  gridTemplateColumns: "max-content minmax(0, 1fr)",
  gap: "var(--ct-spacing-xs)",
};
const appliedRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) max-content",
  gap: "var(--ct-spacing-xs)",
};
