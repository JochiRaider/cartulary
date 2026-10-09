import { requireViewContract } from "@cartulary/view-contracts";
import { type AriaAttributes, useId, useRef, useState } from "react";
import { isBooleanEqualityFilter } from "../models/workbookBooleanFilterOperand";
import { enumFilterChoices } from "../models/workbookEnumFilterOperand";
import {
  applyFilterDraft,
  changeFilterDraftOperandKind,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  type FilterDraftControl,
  filterDraftForField,
  formatFilterSetMembers,
  validateFilterDraft,
  type WorkbookFilterOperator,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { WorkbookBooleanFilterOperand } from "./WorkbookBooleanFilterOperand";
import {
  useEnumLiteralDisclosure,
  WorkbookEnumFilterOperand,
} from "./WorkbookEnumFilterOperand";
import { WorkbookLiteralSetFilterOperand } from "./WorkbookLiteralSetFilterOperand";
import {
  inputStyle,
  secondaryButtonStyle,
  stackedLabelStyle,
} from "./workbookGridControlStyles";

/** Query edits are local until explicitly applied; no discovery on keystrokes. */
export function WorkbookCandidateQueryControl({
  view,
  label,
  query,
  onApply,
  disclosureRef,
}: {
  readonly disclosureRef?: import("react").Ref<HTMLElement>;
  readonly view: string;
  readonly label: string;
  readonly query: WorkbookQueryState;
  readonly onApply: (query: WorkbookQueryState) => void;
}) {
  const contract = requireViewContract(view);
  const [draft, setDraft] = useState(() => defaultFilterDraft(contract));
  const [staged, setStaged] = useState(query);
  const enumChoices = enumFilterChoices(contract, draft);
  const enumDisclosure = useEnumLiteralDisclosure(
    `${view}:${draft.fieldKey}:${draft.op}:${draft.op === "eq" ? draft.operandKind : ""}`,
    draft.op === "eq" ? draft : null,
    enumChoices,
  );
  const validation = validateFilterDraft(contract, draft);
  const feedbackId = useId();
  const feedbackFor = (
    control: FilterDraftControl,
  ): Pick<AriaAttributes, "aria-invalid" | "aria-describedby"> =>
    validation.kind === "invalid" &&
    (validation.controls.includes(control) ||
      (control.startsWith("member:") &&
        validation.controls.includes("value") &&
        !validation.controls.some((key) => key.startsWith("member:"))))
      ? { "aria-invalid": true, "aria-describedby": feedbackId }
      : {};
  return (
    <details>
      <summary ref={disclosureRef}>{label} ordering and filters</summary>
      <div
        style={{ display: "grid", gap: "var(--ct-spacing-xs)", minWidth: 0 }}
      >
        <label style={stackedLabelStyle}>
          Order
          <select
            aria-label={`${label} order`}
            style={inputStyle}
            value={
              staged.sort[0]
                ? `${staged.sort[0].fieldKey}:${staged.sort[0].direction}`
                : ""
            }
            onChange={(event) => {
              const [fieldKey, direction] =
                event.currentTarget.value.split(":");
              setStaged({
                ...staged,
                sort: fieldKey
                  ? [
                      {
                        fieldKey,
                        direction: direction === "desc" ? "desc" : "asc",
                      },
                    ]
                  : [],
              });
            }}
          >
            <option value="">Default order</option>
            {contract.sortFields.flatMap((key) =>
              ["asc", "desc"].map((direction) => (
                <option
                  key={`${key}:${direction}`}
                  value={`${key}:${direction}`}
                >
                  {contract.fieldMap[key]?.label ?? key} ·{" "}
                  {direction === "asc" ? "ascending" : "descending"}
                </option>
              )),
            )}
          </select>
        </label>
        {contract.filterFields.length ? (
          <>
            <label style={stackedLabelStyle}>
              Filter field
              <select
                aria-label={`${label} filter field`}
                {...feedbackFor("field")}
                style={inputStyle}
                value={draft.fieldKey}
                onChange={(event) =>
                  setDraft(
                    filterDraftForField(contract, event.currentTarget.value),
                  )
                }
              >
                {contract.filterFields.map((key) => (
                  <option key={key} value={key}>
                    {contract.fieldMap[key]?.label ?? key}
                  </option>
                ))}
              </select>
            </label>
            <label style={stackedLabelStyle}>
              Match
              <select
                aria-label={`${label} filter operator`}
                {...feedbackFor("operator")}
                style={inputStyle}
                value={draft.op}
                onChange={(event) =>
                  setDraft(
                    filterDraftForField(
                      contract,
                      draft.fieldKey,
                      event.currentTarget.value as WorkbookFilterOperator,
                    ),
                  )
                }
              >
                {contract.fieldMap[draft.fieldKey]?.filterOps.map((op) => (
                  <option key={op} value={op}>
                    {operatorLabels[op] ?? op}
                  </option>
                ))}
              </select>
            </label>
            <Operand
              enumChoices={enumChoices}
              enumDisclosure={enumDisclosure}
              label={label}
              draft={draft}
              onChange={setDraft}
              feedbackFor={feedbackFor}
              isBoolean={isBooleanEqualityFilter(
                contract,
                draft.fieldKey,
                draft.op,
              )}
              isDate={contract.fieldMap[draft.fieldKey]?.readKind === "date"}
            />
            <button
              type="button"
              style={secondaryButtonStyle}
              disabled={validation.kind === "invalid"}
              onClick={() =>
                setStaged((current) =>
                  applyFilterDraft(contract, current, draft),
                )
              }
            >
              Add filter
            </button>
            <span id={feedbackId} role="status" aria-atomic="true">
              {validation.kind === "invalid" ? validation.message : ""}
            </span>
            {staged.filters.map((filter) => (
              <div key={filter.fieldKey} style={{ overflowWrap: "anywhere" }}>
                {contract.fieldMap[filter.fieldKey]?.label}:{" "}
                {operatorLabels[filter.op]}{" "}
                {Array.isArray(filter.arg.values)
                  ? formatFilterSetMembers(filter.arg.values)
                  : Object.values(filter.arg).map(String).join(", ")}{" "}
                <button
                  type="button"
                  style={secondaryButtonStyle}
                  aria-label={`Remove filter ${contract.fieldMap[filter.fieldKey]?.label}`}
                  onClick={() =>
                    setStaged({
                      ...staged,
                      filters: staged.filters.filter(
                        (item) => item.fieldKey !== filter.fieldKey,
                      ),
                    })
                  }
                >
                  Remove
                </button>
              </div>
            ))}
          </>
        ) : null}
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "var(--ct-spacing-xs)",
          }}
        >
          <button
            type="button"
            style={secondaryButtonStyle}
            onClick={() => onApply(staged)}
          >
            Apply candidate query
          </button>
          <button
            type="button"
            style={secondaryButtonStyle}
            onClick={() => {
              const empty = emptyWorkbookQueryState();
              setStaged(empty);
              setDraft(defaultFilterDraft(contract));
              onApply(empty);
            }}
          >
            Reset candidate query
          </button>
        </div>
      </div>
    </details>
  );
}
const operatorLabels: Readonly<Record<string, string>> = {
  eq: "Equals",
  contains_any: "Contains any",
  contains_all: "Contains all",
  prefix: "Starts with",
  full_text: "Text search",
  range: "Range",
};
function Operand({
  enumChoices,
  enumDisclosure,
  label,
  draft,
  onChange,
  feedbackFor,
  isDate,
  isBoolean,
}: {
  readonly enumChoices: readonly string[] | null;
  readonly enumDisclosure: ReturnType<typeof useEnumLiteralDisclosure>;
  readonly label: string;
  readonly draft: FilterDraft;
  readonly onChange: (draft: FilterDraft) => void;
  readonly feedbackFor: (
    control: FilterDraftControl,
  ) => Pick<AriaAttributes, "aria-invalid" | "aria-describedby">;
  readonly isDate: boolean;
  readonly isBoolean: boolean;
}) {
  const booleanControls = useRef<HTMLDivElement>(null);
  const text = (
    name: string,
    value: string,
    change: (value: string) => void,
    control: FilterDraftControl = "value",
  ) => (
    <label style={stackedLabelStyle}>
      {name}
      <input
        {...feedbackFor(control)}
        placeholder={isDate ? "YYYY-MM-DD" : undefined}
        aria-label={`${label} filter ${name.toLowerCase()}`}
        style={inputStyle}
        value={value}
        onChange={(event) => change(event.currentTarget.value)}
      />
    </label>
  );
  if (draft.op === "range")
    return (
      <>
        <label style={stackedLabelStyle}>
          Lower bound
          <select
            aria-label={`${label} lower bound`}
            {...feedbackFor("lower_kind")}
            style={inputStyle}
            value={draft.lowerKind}
            onChange={(event) =>
              onChange({
                ...draft,
                lowerKind: event.currentTarget.value === "gt" ? "gt" : "gte",
              })
            }
          >
            <option value="gte">At or after</option>
            <option value="gt">After</option>
          </select>
        </label>
        {text(
          "From",
          draft.lowerValue,
          (lowerValue) => onChange({ ...draft, lowerValue }),
          "lower_value",
        )}
        <label style={stackedLabelStyle}>
          Upper bound
          <select
            aria-label={`${label} upper bound`}
            {...feedbackFor("upper_kind")}
            style={inputStyle}
            value={draft.upperKind}
            onChange={(event) =>
              onChange({
                ...draft,
                upperKind: event.currentTarget.value === "lt" ? "lt" : "lte",
              })
            }
          >
            <option value="lte">At or before</option>
            <option value="lt">Before</option>
          </select>
        </label>
        {text(
          "To",
          draft.upperValue,
          (upperValue) => onChange({ ...draft, upperValue }),
          "upper_value",
        )}
      </>
    );
  if (draft.op === "eq")
    return (
      <>
        <label style={stackedLabelStyle}>
          Value matching
          <select
            aria-label={`${label} equality operand`}
            {...(draft.operandKind === "null" ? feedbackFor("value") : {})}
            style={inputStyle}
            value={draft.operandKind}
            onChange={(event) => {
              const value = event.currentTarget.value;
              if (
                isBoolean &&
                booleanControls.current?.contains(document.activeElement)
              )
                event.currentTarget.focus();
              if (value === "value" || value === "values" || value === "null")
                onChange(changeFilterDraftOperandKind(draft, value));
            }}
          >
            <option value="value">One value</option>
            <option value="values">Any of these values</option>
            <option value="null">Empty</option>
          </select>
        </label>
        {isBoolean ? (
          <div ref={booleanControls}>
            <WorkbookBooleanFilterOperand
              draft={draft}
              onChange={onChange}
              feedback={feedbackFor("value")}
              valueLabel={`${label} filter value`}
            />
          </div>
        ) : enumChoices ? (
          <WorkbookEnumFilterOperand
            draft={draft}
            choices={enumChoices}
            disclosure={enumDisclosure}
            onChange={onChange}
            feedbackFor={feedbackFor}
            valueLabel={`${label} filter value`}
          />
        ) : draft.operandKind === "null" ? null : draft.operandKind ===
          "values" ? (
          <WorkbookLiteralSetFilterOperand
            members={draft.values}
            onChange={(values) => onChange({ ...draft, values })}
            valueLabel={`${label} filter value`}
            placeholder={isDate ? "YYYY-MM-DD" : undefined}
            feedbackFor={feedbackFor}
          />
        ) : (
          text("Value", draft.value, (value) => onChange({ ...draft, value }))
        )}
      </>
    );
  if (draft.op === "prefix")
    return text("Value", draft.value, (value) => onChange({ ...draft, value }));
  if (draft.op === "full_text")
    return text("Text", draft.query, (query) => onChange({ ...draft, query }));
  return (
    <WorkbookLiteralSetFilterOperand
      members={draft.values}
      onChange={(values) => onChange({ ...draft, values })}
      valueLabel={`${label} filter value`}
      feedbackFor={feedbackFor}
    />
  );
}
