import { requireViewContract } from "@cartulary/view-contracts";
import { useState } from "react";
import { validateFilterDraft } from "../models/workbookGridQueryControls";
import {
  applyFilterDraft,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  filterDraftForField,
  type WorkbookFilterOperator,
  type WorkbookQueryState,
} from "../models/workbookQuery";
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
}: {
  readonly view: string;
  readonly label: string;
  readonly query: WorkbookQueryState;
  readonly onApply: (query: WorkbookQueryState) => void;
}) {
  const contract = requireViewContract(view);
  const [draft, setDraft] = useState(() => defaultFilterDraft(contract));
  const [staged, setStaged] = useState(query);
  const validation = validateFilterDraft(contract, draft);
  return (
    <details>
      <summary>{label} ordering and filters</summary>
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
            <Operand label={label} draft={draft} onChange={setDraft} />
            <button
              type="button"
              style={secondaryButtonStyle}
              disabled={validation.kind === "invalid"}
              onClick={() => setStaged(applyFilterDraft(staged, draft))}
            >
              Add filter
            </button>
            {validation.kind === "invalid" ? (
              <span>{validation.message}</span>
            ) : null}
            {staged.filters.map((filter) => (
              <div key={filter.fieldKey} style={{ overflowWrap: "anywhere" }}>
                {contract.fieldMap[filter.fieldKey]?.label}:{" "}
                {operatorLabels[filter.op]}{" "}
                {Object.values(filter.arg).map(String).join(", ")}{" "}
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
  label,
  draft,
  onChange,
}: {
  readonly label: string;
  readonly draft: FilterDraft;
  readonly onChange: (draft: FilterDraft) => void;
}) {
  const text = (
    name: string,
    value: string,
    change: (value: string) => void,
  ) => (
    <label style={stackedLabelStyle}>
      {name}
      <input
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
        {text("From", draft.lowerValue, (lowerValue) =>
          onChange({ ...draft, lowerValue }),
        )}
        <label style={stackedLabelStyle}>
          Upper bound
          <select
            aria-label={`${label} upper bound`}
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
        {text("To", draft.upperValue, (upperValue) =>
          onChange({ ...draft, upperValue }),
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
            style={inputStyle}
            value={draft.operandKind}
            onChange={(event) => {
              const value = event.currentTarget.value;
              if (value === "value" || value === "values" || value === "null")
                onChange({ ...draft, operandKind: value });
            }}
          >
            <option value="value">One value</option>
            <option value="values">Any of these values</option>
            <option value="null">Empty</option>
          </select>
        </label>
        {draft.operandKind === "null" ? null : draft.operandKind ===
          "values" ? (
          text("Values (comma separated)", draft.values, (values) =>
            onChange({ ...draft, values }),
          )
        ) : draft.valueType === "boolean" ? (
          <label style={stackedLabelStyle}>
            Value
            <select
              aria-label={`${label} filter value`}
              style={inputStyle}
              value={draft.booleanValue}
              onChange={(event) => {
                const booleanValue = event.currentTarget.value;
                if (
                  booleanValue === "" ||
                  booleanValue === "true" ||
                  booleanValue === "false"
                )
                  onChange({ ...draft, booleanValue });
              }}
            >
              <option value="">Choose a value</option>
              <option value="true">true</option>
              <option value="false">false</option>
            </select>
          </label>
        ) : (
          text("Value", draft.value, (value) => onChange({ ...draft, value }))
        )}
      </>
    );
  if (draft.op === "prefix")
    return text("Value", draft.value, (value) => onChange({ ...draft, value }));
  if (draft.op === "full_text")
    return text("Text", draft.query, (query) => onChange({ ...draft, query }));
  return text("Values (comma separated)", draft.values, (values) =>
    onChange({ ...draft, values }),
  );
}
