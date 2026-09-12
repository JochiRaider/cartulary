import { indicatorCreateTestId } from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import { useId, useRef, useState } from "react";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import {
  indicatorCreateConstraints as constraints,
  type IndicatorCreateDraft,
  type IndicatorCreateErrors,
  indicatorCreateErrors,
  indicatorCreateTypes,
} from "./indicatorCreateModel";
import {
  observationField,
  observationStack,
  observationText,
} from "./observationStyles";

const labels: Readonly<Record<string, string>> = {
  "indicator.indicator_type": "Indicator type",
  "indicator.value_kind": "Value kind",
  "indicator.display_value": "Canonical value",
  "indicator.normalized_value": "Normalized value",
  "indicator.hash_algorithm": "Hash algorithm",
  "indicator.hash_value": "Hash value",
  "indicator.defanged_value": "Defanged presentation",
  "indicator.stix_pattern": "STIX pattern",
};
const references = emptyGenericReferenceOptions();
export function IndicatorCanonicalAuthoring({
  observation,
  contract,
  draft,
  disabled,
  onChange,
  onSubmit,
}: {
  observation: IndicatorObservation;
  contract: ViewContract;
  draft: IndicatorCreateDraft;
  disabled: boolean;
  onChange: (key: string, value: string) => void;
  onSubmit: () => void;
}) {
  const id = useId(),
    details = useRef<HTMLDetailsElement>(null),
    form = useRef<HTMLFormElement>(null);
  const [submitted, setSubmitted] = useState(false);
  const errors: IndicatorCreateErrors = submitted
    ? indicatorCreateErrors(contract, draft.values)
    : {};
  const type = draft.values["indicator.indicator_type"] ?? "";
  const field = (key: string) => {
    const entry = contract.fieldMap[key];
    if (
      !entry?.createWritable ||
      ((constraints.hashFields as readonly string[]).includes(key) &&
        (constraints.hashForbiddenTypes as readonly string[]).includes(type))
    )
      return null;
    const values =
      key === "indicator.indicator_type"
        ? indicatorCreateTypes
        : key === "indicator.value_kind"
          ? (constraints.atomicTypes as readonly string[]).includes(type)
            ? ["atomic"]
            : constraints.valueKinds
          : entry.enumValues;
    const errorId = `${id}-${key}-error`;
    return (
      <div key={key} style={observationField}>
        <label htmlFor={`${id}-${key}`}>{labels[key] ?? entry.label}</label>
        <GenericMutationControl
          field={{ ...entry, enumValues: values }}
          collectionMode="add"
          referenceOptions={references}
          id={`${id}-${key}`}
          ariaLabel={labels[key] ?? entry.label}
          describedBy={errors[key] ? errorId : undefined}
          invalid={!!errors[key]}
          testId={indicatorCreateTestId("field", key)}
          value={draft.values[key] ?? ""}
          onChange={(value) => onChange(key, value)}
        />
        {errors[key] ? (
          <p id={errorId} style={observationText}>
            {errors[key]}
          </p>
        ) : null}
      </div>
    );
  };
  return (
    <form
      ref={form}
      data-testid={indicatorCreateTestId("editor")}
      aria-label="Canonical Indicator proposal"
      style={observationStack}
      onSubmit={(event) => {
        event.preventDefault();
        if (disabled) return;
        setSubmitted(true);
        const errors = indicatorCreateErrors(contract, draft.values);
        if (Object.keys(errors).length) {
          if (details.current) details.current.open = true;
          const key = Object.keys(errors)[0];
          if (key)
            form.current
              ?.querySelector<HTMLElement>(`[id="${id}-${key}"]`)
              ?.focus();
          return;
        }
        onSubmit();
      }}
    >
      <p style={observationText}>
        Original observed text: <strong>{observation.observed_text}</strong>
      </p>
      <p style={observationText}>
        Review this canonical proposal. The server validates and normalizes its
        identity. Changing type clears the proposal and details.
      </p>
      <fieldset
        disabled={disabled}
        style={{ ...observationStack, border: 0, padding: 0, margin: 0 }}
      >
        <legend>Canonical identity</legend>
        {constraints.requiredFields.map(field)}
        <details ref={details}>
          <summary>Additional canonical details</summary>
          <div style={observationStack}>
            {contract.fields
              .filter(
                (entry) =>
                  entry.createWritable &&
                  !(constraints.requiredFields as readonly string[]).includes(
                    entry.fieldKey,
                  ),
              )
              .map((entry) => field(entry.fieldKey))}
            <p style={observationText}>
              Presentation metadata starts blank. A matching Indicator is
              returned unchanged; supplied details do not enrich it.
            </p>
          </div>
        </details>
        {Object.keys(errors).length ? (
          <p role="alert">
            {errors.contract ?? "Review the indicated canonical fields."}
          </p>
        ) : null}
        <WorkbookInspectorActionButton type="submit" disabled={disabled}>
          Create canonical Indicator
        </WorkbookInspectorActionButton>
      </fieldset>
      <p style={observationText}>
        This makes an Indicator available. Linking this observation is a
        separate action.
      </p>
    </form>
  );
}
