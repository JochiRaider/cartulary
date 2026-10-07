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
import { type AriaAttributes, type RefObject, useId } from "react";
import { useRegisteredOverlayNavigation } from "../../shared/useRegisteredOverlayNavigation";
import {
  booleanFilterControlKeys,
  isBooleanEqualityFilter,
} from "../models/workbookBooleanFilterOperand";
import {
  enumFilterChoices,
  enumFilterControlKeys,
} from "../models/workbookEnumFilterOperand";
import {
  parseDeclaredFieldKey,
  type WorkbookGridQueryCommand,
  type WorkbookGridQueryControlProjection,
  type WorkbookRequestedFilterChange,
} from "../models/workbookGridQueryControls";
import {
  changeFilterDraftOperandKind,
  type FilterDraft,
  type FilterDraftControl,
  type FilterDraftValidation,
  filterDraftForField,
  filterMemberControlKeys,
  isWorkbookFilterOperator,
  validateFilterDraft,
  type WorkbookFilter,
} from "../models/workbookQuery";
import { WorkbookBooleanFilterOperand } from "./WorkbookBooleanFilterOperand";
import {
  useEnumLiteralDisclosure,
  WorkbookEnumFilterOperand,
} from "./WorkbookEnumFilterOperand";
import { WorkbookLiteralSetFilterOperand } from "./WorkbookLiteralSetFilterOperand";
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
  editingRequestedFilter,
  filterCount,
  requestedFilterCount,
  requestedChanges,
  isOpen,
  onApply,
  onChangeDraft,
  onClose,
  onCommand,
  onComplete,
  onEditFilter,
  onEditRequestedFilter,
  onEditQueryEntry,
  onRestoreFilter,
  onToggle,
  projection,
  returnFocusRef,
  surface,
  triggerRef,
}: {
  readonly contract: ViewContract;
  readonly draft: FilterDraft;
  readonly editingFieldKey: string | null;
  readonly editingRequestedFilter: boolean;
  readonly filterCount: number;
  readonly requestedFilterCount: number;
  readonly requestedChanges: readonly WorkbookRequestedFilterChange[];
  readonly isOpen: boolean;
  readonly onApply: (draft: FilterDraft) => FilterDraftValidation;
  readonly onChangeDraft: (draft: FilterDraft) => void;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onComplete: (draft: FilterDraft) => void;
  readonly onEditFilter: (fieldKey: string) => void;
  readonly onEditRequestedFilter: (fieldKey: string) => void;
  readonly onEditQueryEntry: (
    entry: WorkbookGridQueryControlProjection["chips"][number],
  ) => void;
  readonly onRestoreFilter: (filter: WorkbookFilter) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly returnFocusRef: RefObject<HTMLElement | null>;
  readonly surface: string;
  readonly triggerRef: RefObject<HTMLButtonElement | null>;
}) {
  const feedbackId = useId();
  const isBoolean = isBooleanEqualityFilter(contract, draft.fieldKey, draft.op);
  const enumChoices = enumFilterChoices(contract, draft);
  const enumDisclosure = useEnumLiteralDisclosure(
    `${surface}:${isOpen}:${editingFieldKey}:${draft.fieldKey}:${draft.op}:${draft.op === "eq" ? draft.operandKind : ""}`,
    draft.op === "eq" ? draft : null,
    enumChoices,
  );
  const itemKeys = [
    "field",
    "operator",
    ...(draft.op === "range"
      ? ["lower_kind", "lower_value", "upper_kind", "upper_value"]
      : draft.op === "eq"
        ? [
            "operand_kind",
            ...(isBoolean
              ? booleanFilterControlKeys(draft.operandKind)
              : enumChoices
                ? enumFilterControlKeys(draft, enumChoices, enumDisclosure.open)
                : draft.operandKind === "null"
                  ? []
                  : draft.operandKind === "values"
                    ? filterMemberControlKeys(draft.values)
                    : ["value"]),
          ]
        : draft.op === "contains_any" || draft.op === "contains_all"
          ? filterMemberControlKeys(draft.values)
          : ["value"]),
    ...requestedChanges.flatMap((change) =>
      change.kind === "removed"
        ? [`restore:${change.filter.fieldKey}`]
        : change.kind === "added"
          ? [
              `edit_requested:${change.filter.fieldKey}`,
              `remove_requested:${change.filter.fieldKey}`,
            ]
          : [`edit_requested:${change.filter.fieldKey}`],
    ),
    ...projection.chips
      .filter((chip) => chip.identity.kind === "filter")
      .flatMap((chip) => [`edit:${chip.key}`, `remove:${chip.key}`]),
    ...projection.hiddenChips
      .filter((chip) => chip.identity.kind !== "filter")
      .map((chip) => `overflow:${chip.key}`),
    ...(requestedFilterCount > 0 ? ["clear"] : []),
    ...(editingFieldKey === null ? [] : ["remove_editing"]),
    "cancel",
    "apply",
  ];
  const initialKey = editingFieldKey === null ? "field" : "operator";
  const navigation = useRegisteredOverlayNavigation({
    initialItemKey: initialKey,
    isOpen,
    itemKeys,
    keyboardMode: "form",
    onRequestClose: onClose,
    preferredReturnFocusRef: returnFocusRef,
    reconcileItems: true,
    reconcileItemKey: (key, previous, eligible) => {
      if (key.startsWith("member:") || key.startsWith("member_remove:")) {
        const index = previous.indexOf(key);
        const isMember = (candidate: string) =>
          candidate.startsWith("member:") && eligible.includes(candidate);
        return (
          previous.slice(index + 1).find(isMember) ??
          previous.slice(0, index).reverse().find(isMember) ??
          (eligible.includes("member_add") ? "member_add" : null)
        );
      }
      return isBoolean &&
        (key === "value" || key.startsWith("boolean_choice:")) &&
        eligible.includes("operand_kind")
        ? "operand_kind"
        : null;
    },
    subjectKey: surface,
    trapTab: true,
    triggerRef,
  });
  const validation = validateFilterDraft(contract, draft);
  const feedbackFor: FilterFeedbackFor = (control) =>
    validation.kind === "invalid" &&
    (validation.controls.includes(control) ||
      (control.startsWith("member:") &&
        validation.controls.includes("value") &&
        !validation.controls.some((key) => key.startsWith("member:"))))
      ? { "aria-invalid": true, "aria-describedby": feedbackId }
      : {};
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
            ? `Filters, ${filterCount} active filters, ${hiddenCount} hidden query entries${requestedChanges.length > 0 ? ", Unapplied changes" : ""}`
            : `Filters, ${filterCount} active filters${requestedChanges.length > 0 ? ", Unapplied changes" : ""}`
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
        {requestedChanges.length > 0 ? " · Unapplied" : ""}
      </button>
      {isOpen ? (
        <div
          aria-label={
            editingFieldKey === null
              ? "Add filter"
              : editingRequestedFilter
                ? "Edit unapplied filter"
                : "Edit filter"
          }
          data-testid={workbookFilterPopoverTestId(surface)}
          id={workbookFilterPopoverTestId(surface)}
          role="dialog"
          style={filterPopoverStyle}
          tabIndex={-1}
          onBlur={navigation.onOverlayBlur}
          onFocusCapture={navigation.onOverlayFocus}
          onKeyDown={navigation.onOverlayKeyDown}
        >
          <strong>
            {editingFieldKey === null
              ? "Add filter"
              : editingRequestedFilter
                ? "Edit unapplied filter"
                : "Edit filter"}
          </strong>
          <label style={stackedLabelStyle}>
            Field
            <select
              ref={navigation.registerItem("field")}
              {...feedbackFor("field")}
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
              {...feedbackFor("operator")}
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
            enumChoices={enumChoices}
            enumDisclosure={enumDisclosure}
            isBoolean={isBoolean}
            isDate={field?.readKind === "date"}
            feedbackFor={feedbackFor}
            draft={draft}
            navigation={navigation}
            onChangeDraft={onChangeDraft}
            surface={surface}
          />
          <p
            id={feedbackId}
            role="status"
            aria-atomic="true"
            style={filterValidationStyle}
          >
            {validation.kind === "invalid" ? validation.message : ""}
          </p>
          <FilterQueryActions
            filterCount={filterCount}
            requestedFilterCount={requestedFilterCount}
            requestedChanges={requestedChanges}
            navigation={navigation}
            onCommand={onCommand}
            onEditFilter={onEditFilter}
            onEditRequestedFilter={onEditRequestedFilter}
            onEditQueryEntry={onEditQueryEntry}
            onRestoreFilter={onRestoreFilter}
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
                  navigation.close({ restoreTriggerFocus: true });
                }}
              >
                Remove filter
              </button>
            )}
            <button
              ref={navigation.registerItem("cancel")}
              style={secondaryButtonStyle}
              type="button"
              onClick={() => navigation.close({ restoreTriggerFocus: true })}
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
                if (
                  validation.kind === "valid" &&
                  onApply(draft).kind === "valid"
                ) {
                  navigation.close({ restoreTriggerFocus: true });
                  onComplete(draft);
                }
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

type FilterFeedbackFor = (
  control: FilterDraftControl,
) => Pick<AriaAttributes, "aria-invalid" | "aria-describedby">;

function FilterOperandControl({
  enumChoices,
  enumDisclosure,
  draft,
  isDate,
  isBoolean,
  feedbackFor,
  navigation,
  onChangeDraft,
  surface,
}: {
  readonly enumChoices: readonly string[] | null;
  readonly enumDisclosure: ReturnType<typeof useEnumLiteralDisclosure>;
  readonly draft: FilterDraft;
  readonly isDate: boolean;
  readonly isBoolean: boolean;
  readonly feedbackFor: FilterFeedbackFor;
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
            {...(draft.operandKind === "null" ? feedbackFor("value") : {})}
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
              onChangeDraft(changeFilterDraftOperandKind(draft, operandKind));
            }}
          >
            <option value="value">Equals value</option>
            <option value="values">Is one of</option>
            <option value="null">Is empty</option>
          </select>
        </label>
        {isBoolean ? (
          <WorkbookBooleanFilterOperand
            draft={draft}
            onChange={onChangeDraft}
            feedback={feedbackFor("value")}
            registerItem={navigation.registerItem}
            valueTestId={gridFilterValueTestId(surface)}
          />
        ) : enumChoices ? (
          <WorkbookEnumFilterOperand
            draft={draft}
            choices={enumChoices}
            disclosure={enumDisclosure}
            onChange={onChangeDraft}
            feedbackFor={feedbackFor}
            registerItem={navigation.registerItem}
            valueTestId={gridFilterValueTestId(surface)}
          />
        ) : draft.operandKind === "null" ? null : draft.operandKind ===
          "values" ? (
          <WorkbookLiteralSetFilterOperand
            members={draft.values}
            onChange={(values) => onChangeDraft({ ...draft, values })}
            feedbackFor={feedbackFor}
            registerItem={navigation.registerItem}
            placeholder={isDate ? "YYYY-MM-DD" : undefined}
            valueTestId={gridFilterValueTestId(surface)}
          />
        ) : (
          <TextOperand
            draft={draft}
            feedbackFor={feedbackFor}
            label={isDate ? "Date value" : "Value"}
            navigation={navigation}
            onValue={(value) => onChangeDraft({ ...draft, value })}
            placeholder={isDate ? "YYYY-MM-DD" : "Value"}
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
        <div style={stackedLabelStyle}>
          Lower bound
          <span style={boundStyle}>
            <select
              ref={navigation.registerItem("lower_kind")}
              {...feedbackFor("lower_kind")}
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
              ref={navigation.registerItem("lower_value")}
              {...feedbackFor("lower_value")}
              aria-label="Lower-bound value"
              placeholder={isDate ? "YYYY-MM-DD" : undefined}
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
        </div>
        <div style={stackedLabelStyle}>
          Upper bound
          <span style={boundStyle}>
            <select
              ref={navigation.registerItem("upper_kind")}
              {...feedbackFor("upper_kind")}
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
              ref={navigation.registerItem("upper_value")}
              {...feedbackFor("upper_value")}
              aria-label="Upper-bound value"
              placeholder={isDate ? "YYYY-MM-DD" : undefined}
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
        </div>
      </div>
    );
  }
  if (draft.op === "contains_any" || draft.op === "contains_all") {
    return (
      <WorkbookLiteralSetFilterOperand
        members={draft.values}
        onChange={(values) => onChangeDraft({ ...draft, values })}
        feedbackFor={feedbackFor}
        registerItem={navigation.registerItem}
        valueTestId={gridFilterValueTestId(surface)}
      />
    );
  }
  if (draft.op === "full_text") {
    return (
      <TextOperand
        draft={draft}
        feedbackFor={feedbackFor}
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
        feedbackFor={feedbackFor}
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
  feedbackFor,
  navigation,
  onValue,
  placeholder,
  surface,
  value,
}: {
  readonly draft: FilterDraft;
  readonly feedbackFor: FilterFeedbackFor;
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
        {...feedbackFor("value")}
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

function FilterQueryActions({
  filterCount,
  requestedFilterCount,
  requestedChanges,
  navigation,
  onCommand,
  onEditFilter,
  onEditRequestedFilter,
  onEditQueryEntry,
  onRestoreFilter,
  projection,
  surface,
}: {
  readonly filterCount: number;
  readonly requestedFilterCount: number;
  readonly requestedChanges: readonly WorkbookRequestedFilterChange[];
  readonly navigation: ReturnType<typeof useRegisteredOverlayNavigation>;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onEditFilter: (fieldKey: string) => void;
  readonly onEditRequestedFilter: (fieldKey: string) => void;
  readonly onEditQueryEntry: (
    entry: WorkbookGridQueryControlProjection["chips"][number],
  ) => void;
  readonly onRestoreFilter: (filter: WorkbookFilter) => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly surface: string;
}) {
  const filterEntries = projection.chips.filter(
    (chip) => chip.identity.kind === "filter",
  );
  const otherOverflow = projection.hiddenChips.filter(
    (chip) => chip.identity.kind !== "filter",
  );
  if (
    filterEntries.length === 0 &&
    otherOverflow.length === 0 &&
    requestedChanges.length === 0
  )
    return null;
  return (
    <section
      aria-label="Filters and applied query overflow"
      style={queryListStyle}
    >
      {requestedChanges.length === 0 ? null : (
        <strong>Unapplied changes</strong>
      )}
      {requestedChanges.map((change) =>
        change.kind === "removed" ? (
          <div key={change.filter.fieldKey} style={appliedRowStyle}>
            <span>Remove {change.label} · Unapplied</span>
            <button
              ref={navigation.registerItem(`restore:${change.filter.fieldKey}`)}
              aria-label={`Restore ${change.label}`}
              style={clearButtonStyle}
              type="button"
              onClick={() => onRestoreFilter(change.filter)}
            >
              Restore
            </button>
          </div>
        ) : (
          <div key={change.filter.fieldKey} style={appliedRowStyle}>
            <button
              ref={navigation.registerItem(
                `edit_requested:${change.filter.fieldKey}`,
              )}
              aria-label={`Edit unapplied ${change.label}`}
              style={queryListButtonStyle}
              type="button"
              onClick={() => {
                onEditRequestedFilter(change.filter.fieldKey);
                navigation.focusItem("operator");
              }}
            >
              {change.kind === "added" ? "Add" : "Change"} {change.label} ·
              Unapplied
            </button>
            {change.kind === "added" ? (
              <button
                ref={navigation.registerItem(
                  `remove_requested:${change.filter.fieldKey}`,
                )}
                aria-label={`Remove unapplied ${change.label}`}
                style={clearButtonStyle}
                type="button"
                onClick={() =>
                  onCommand({
                    kind: "filter_remove",
                    fieldKey: change.filter.fieldKey,
                  })
                }
              >
                ×
              </button>
            ) : null}
          </div>
        ),
      )}
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
            onClick={() => {
              onEditFilter(chip.identity.fieldKey);
              navigation.focusItem("operator");
            }}
          >
            {chip.label}
          </button>
          <button
            ref={navigation.registerItem(`remove:${chip.key}`)}
            aria-label={`Remove ${chip.accessibleName}`}
            style={clearButtonStyle}
            type="button"
            onClick={() => onCommand(chip.removeCommand)}
          >
            ×
          </button>
        </div>
      ))}
      {filterCount > 0 || requestedFilterCount > 0 ? (
        <button
          ref={navigation.registerItem("clear")}
          data-testid={workbookFilterClearButtonTestId(surface)}
          disabled={requestedFilterCount === 0}
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
  insetInlineStart: "auto",
  insetInlineEnd: 0,
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
