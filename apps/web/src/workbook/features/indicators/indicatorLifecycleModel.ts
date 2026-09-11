import {
  type IndicatorLifecycleValues,
  indicatorLifecycleConstraints,
} from "../../adapters/indicatorLifecycleProtocol";
import type { GenericReferenceOption } from "../../models/workbookReferenceOptions";

export const indicatorLifecycleViewId = "cartulary.view.indicators.v1";
export type LifecycleDraftValues = Readonly<{
  state: string;
  validFrom: string;
  validTo: string;
  confidence: string;
  rationale: string;
  assessor: string;
  support: readonly GenericReferenceOption[];
}>;
export type LifecycleDraft = Readonly<{
  recordId: string;
  label: string;
  baseRowVersion: number;
  revision: number;
  values: LifecycleDraftValues;
}>;
export type LifecycleFieldError = Readonly<{
  field: keyof LifecycleDraftValues;
  message: string;
}>;
export const emptyLifecycleValues: LifecycleDraftValues = Object.freeze({
  state: "active",
  validFrom: "",
  validTo: "",
  confidence: "",
  rationale: "",
  assessor: "",
  support: [],
});

export function canonicalLifecycleUUID(value: string): boolean {
  return (
    /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u.test(
      value,
    ) && value !== "00000000-0000-0000-0000-000000000000"
  );
}

/** UTC wall fields: never parse using the browser timezone or normalize an invalid calendar. */
export function lifecycleUTCTimestamp(raw: string): string | null {
  const match =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2})(?:\.(\d{1,9}))?)?Z?$/u.exec(
      raw,
    );
  if (!match) return null;
  const [year, month, day, hour, minute, second] = match
    .slice(1, 7)
    .map((part) => Number(part ?? 0));
  if (
    year === undefined ||
    month === undefined ||
    day === undefined ||
    hour === undefined ||
    minute === undefined ||
    second === undefined ||
    year < 1 ||
    month < 1 ||
    month > 12 ||
    day < 1 ||
    hour > 23 ||
    minute > 59 ||
    second > 59
  )
    return null;
  const date = new Date(0);
  date.setUTCFullYear(year, month - 1, day);
  date.setUTCHours(hour, minute, second, 0);
  if (
    date.getUTCFullYear() !== year ||
    date.getUTCMonth() !== month - 1 ||
    date.getUTCDate() !== day
  )
    return null;
  const fraction = (match[7] ?? "").replace(/0+$/u, "");
  return `${match[1]}-${match[2]}-${match[3]}T${match[4]}:${match[5]}:${match[6] ?? "00"}${fraction ? `.${fraction}` : ""}Z`;
}

/** Comparison retains submillisecond precision, unlike Date.parse. */
export function lifecycleTimeKey(canonical: string): string {
  const [seconds = "", fraction = ""] = canonical.slice(0, -1).split(".");
  return `${seconds}.${fraction.padEnd(9, "0")}`;
}

/** Public reads allow RFC3339 offsets. Present the same instant in UTC without losing fractional precision. */
export function lifecycleReadTimestamp(raw: string): string | null {
  const match = /^(.*?)(Z|([+-])(\d{2}):(\d{2}))$/u.exec(raw);
  if (!match) return null;
  const local = lifecycleUTCTimestamp(match[1] ?? "");
  if (!local) return null;
  if (match[2] === "Z") return local;
  const hours = Number(match[4]),
    minutes = Number(match[5]);
  if (hours > 23 || minutes > 59) return null;
  const fraction = /\.\d+Z$/u.exec(local)?.[0].slice(0, -1) ?? "";
  const seconds = local.replace(/\.\d+Z$/u, "Z");
  const instant = new Date(
    Date.parse(seconds) -
      (match[3] === "+" ? 1 : -1) * (hours * 60 + minutes) * 60000,
  );
  if (
    !Number.isFinite(instant.getTime()) ||
    instant.getUTCFullYear() < 1 ||
    instant.getUTCFullYear() > 9999
  )
    return null;
  return `${instant.toISOString().slice(0, 19)}${fraction}Z`;
}

export function validateLifecycleDraft(values: LifecycleDraftValues): {
  values: IndicatorLifecycleValues | null;
  errors: readonly LifecycleFieldError[];
} {
  const errors: LifecycleFieldError[] = [];
  const state = indicatorLifecycleConstraints.states.find(
    (candidate) => candidate === values.state,
  );
  if (!state)
    errors.push({
      field: "state",
      message: "Select an Indicator lifecycle state.",
    });
  const from = lifecycleUTCTimestamp(values.validFrom);
  const to =
    values.validTo === "" ? null : lifecycleUTCTimestamp(values.validTo);
  if (!from)
    errors.push({
      field: "validFrom",
      message: "Enter a valid effective start in UTC.",
    });
  if (
    (values.validTo !== "" && !to) ||
    (from && to && lifecycleTimeKey(to) < lifecycleTimeKey(from))
  )
    errors.push({
      field: "validTo",
      message:
        "Enter an effective end in UTC at or after the start, or leave it empty.",
    });
  const confidence =
    values.confidence === "" ? null : Number(values.confidence);
  const { minimum, maximum } = indicatorLifecycleConstraints.confidence;
  if (
    confidence !== null &&
    (!/^\d+$/u.test(values.confidence) ||
      !Number.isSafeInteger(confidence) ||
      confidence < minimum ||
      confidence > maximum)
  )
    errors.push({
      field: "confidence",
      message: `Confidence must be a whole number from ${minimum} through ${maximum}, or empty.`,
    });
  for (const field of ["rationale", "assessor"] as const)
    if (values[field].includes("\0"))
      errors.push({ field, message: "Remove the unsupported null character." });
  const refs = values.support.map((item) => item.recordId);
  if (
    refs.length > indicatorLifecycleConstraints.supportMaximum ||
    refs.some((id) => !canonicalLifecycleUUID(id)) ||
    new Set(refs).size !== refs.length
  )
    errors.push({
      field: "support",
      message: `Choose up to ${indicatorLifecycleConstraints.supportMaximum} distinct supporting records.`,
    });
  return {
    errors,
    values:
      errors.length || !state || !from
        ? null
        : {
            lifecycle_state: state,
            valid_from: from,
            valid_to: to,
            confidence,
            rationale: values.rationale === "" ? null : values.rationale,
            assessor: values.assessor === "" ? null : values.assessor,
            support_refs: refs,
          },
  };
}

export function freezeLifecycle<T>(value: T): T {
  if (value && typeof value === "object") {
    for (const child of Object.values(value)) freezeLifecycle(child);
    Object.freeze(value);
  }
  return value;
}
