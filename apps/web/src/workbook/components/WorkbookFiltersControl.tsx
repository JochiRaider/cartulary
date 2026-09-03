import {
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
} from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import { SlidersHorizontal } from "lucide-react";
import { useMemo, useRef } from "react";
import { useRegisteredOverlayNavigation } from "../focus/useRegisteredOverlayNavigation";
import {
  parseDeclaredFieldKey,
  parseWorkbookBooleanDraftValue,
  validateFilterDraft,
  type WorkbookGridQueryCommand,
  type WorkbookGridQueryControlProjection,
} from "../models/workbookGridQueryControls";
import { type FilterDraft, filterInputMode } from "../models/workbookQuery";
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
  filterCount,
  isOpen,
  onApply,
  onChangeDraft,
  onClose,
  onCommand,
  onToggle,
  projection,
  surface,
}: {
  readonly contract: ViewContract;
  readonly draft: FilterDraft;
  readonly filterCount: number;
  readonly isOpen: boolean;
  readonly onApply: (draft: FilterDraft) => void;
  readonly onChangeDraft: (draft: FilterDraft) => void;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly surface: string;
}) {
  const triggerRef = useRef<HTMLButtonElement>(null);
  const itemKeys = useMemo(
    () => [
      "field",
      "value",
      ...projection.chips.map((chip) => `chip:${chip.key}`),
      ...(projection.chips.length > 1 ? ["clear"] : []),
      "cancel",
      "apply",
    ],
    [projection.chips],
  );
  const navigation = useRegisteredOverlayNavigation({
    initialItemKey: "field",
    isOpen,
    itemKeys,
    onRequestClose: onClose,
    subjectKey: surface,
    triggerRef,
  });
  const validation = validateFilterDraft(contract, draft);
  const inputMode = filterInputMode(draft.fieldKey);
  const hiddenCount = projection.hiddenChips.length;

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
            ? `Filters, ${hiddenCount} hidden`
            : `Filters, ${filterCount} active filters`
        }
        data-testid={workbookFilterPopoverTriggerTestId(surface)}
        style={controlButtonStyle}
        type="button"
        onClick={() => {
          if (!isOpen) navigation.prepareOpen("field");
          onToggle();
        }}
      >
        <SlidersHorizontal aria-hidden="true" size={15} />
        Filters{filterCount > 0 ? ` ${filterCount}` : ""}
      </button>
      {isOpen ? (
        <div
          aria-label="Draft filters"
          data-testid={workbookFilterPopoverTestId(surface)}
          id={workbookFilterPopoverTestId(surface)}
          role="dialog"
          style={filterPopoverStyle}
          tabIndex={-1}
          onBlur={navigation.onOverlayBlur}
          onKeyDown={(event) => {
            if (event.key !== "Escape") return;
            navigation.onItemKeyDown(event, navigation.activeKey ?? "field");
          }}
        >
          <label style={stackedLabelStyle}>
            Field
            <select
              ref={navigation.registerItem("field")}
              data-testid={gridFilterFieldTestId(surface)}
              style={selectStyle}
              value={draft.fieldKey}
              onChange={(event) => {
                const fieldKey = parseDeclaredFieldKey(
                  event.currentTarget.value,
                  contract.filterFields,
                );
                if (fieldKey === null) return;
                onChangeDraft({ booleanValue: "", fieldKey, value: "" });
              }}
            >
              {contract.filterFields.map((fieldKey) => (
                <option key={fieldKey} value={fieldKey}>
                  {contract.fieldMap[fieldKey]?.label ?? fieldKey}
                </option>
              ))}
            </select>
          </label>
          <FilterValueControl
            draft={draft}
            inputMode={inputMode}
            navigation={navigation}
            onChangeDraft={onChangeDraft}
            surface={surface}
          />
          {validation.kind === "invalid" ? (
            <p role="status" style={filterValidationStyle}>
              {validation.message}
            </p>
          ) : null}
          <FilterOverflowActions
            navigation={navigation}
            onCommand={onCommand}
            projection={projection}
          />
          <div style={popoverActionsStyle}>
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

function FilterValueControl({
  draft,
  inputMode,
  navigation,
  onChangeDraft,
  surface,
}: {
  readonly draft: FilterDraft;
  readonly inputMode: ReturnType<typeof filterInputMode>;
  readonly navigation: ReturnType<typeof useRegisteredOverlayNavigation>;
  readonly onChangeDraft: (draft: FilterDraft) => void;
  readonly surface: string;
}) {
  if (inputMode === "boolean") {
    return (
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
            if (booleanValue === null) return;
            onChangeDraft({ ...draft, booleanValue });
          }}
        >
          <option value="">Select value</option>
          <option value="true">true</option>
          <option value="false">false</option>
        </select>
      </label>
    );
  }
  return (
    <label style={stackedLabelStyle}>
      Value
      <input
        ref={navigation.registerItem("value")}
        data-testid={gridFilterValueTestId(surface)}
        placeholder={placeholderForMode(inputMode)}
        style={inputStyle}
        type="text"
        value={draft.value}
        onChange={(event) => {
          onChangeDraft({ ...draft, value: event.currentTarget.value });
        }}
      />
    </label>
  );
}

function FilterOverflowActions({
  navigation,
  onCommand,
  projection,
}: {
  readonly navigation: ReturnType<typeof useRegisteredOverlayNavigation>;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly projection: WorkbookGridQueryControlProjection;
}) {
  if (projection.chips.length === 0) return null;
  return (
    <section aria-label="Active query overflow" style={queryListStyle}>
      <strong>Active query</strong>
      {projection.chips.map((chip, index) => {
        const itemKey = `chip:${chip.key}`;
        return (
          <button
            key={chip.key}
            ref={navigation.registerItem(itemKey)}
            aria-label={`Remove ${chip.label}${
              index >= projection.visibleChipCapacity
                ? ", hidden from the view bar"
                : ""
            }`}
            style={queryListButtonStyle}
            type="button"
            onClick={() => {
              onCommand(chip.command);
            }}
          >
            {chip.label}
          </button>
        );
      })}
      {projection.chips.length > 1 ? (
        <button
          ref={navigation.registerItem("clear")}
          style={clearButtonStyle}
          type="button"
          onClick={() => {
            onCommand({ kind: "query_clear" });
          }}
        >
          Clear all
        </button>
      ) : null}
    </section>
  );
}

function placeholderForMode(mode: ReturnType<typeof filterInputMode>) {
  switch (mode) {
    case "date":
      return "YYYY-MM-DD or YYYY-MM-DD..YYYY-MM-DD";
    case "tagset":
      return "comma-separated values";
    case "timestamp":
      return "RFC3339 or start..end";
    case "boolean":
    case "text":
      return "Value";
  }
}

const filterPopoverStyle = {
  ...menuStyle,
  inlineSize: "min(24rem, 88vw)",
  gap: "0.55rem",
};
const popoverActionsStyle = {
  display: "flex",
  justifyContent: "end",
  gap: "0.4rem",
};
