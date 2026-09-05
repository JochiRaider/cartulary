import { errorRegistry } from "@cartulary/protocol-ts/errors";

type ReasonRegistry = (typeof errorRegistry)["reason_registries"][number];
export type PublicErrorReasonCode =
  ReasonRegistry["reason_codes"][number]["code"];
export type EvidenceAccessReasonCode = Extract<
  ReasonRegistry,
  {
    readonly error_code: "evidence_access_unavailable";
  }
>["reason_codes"][number]["code"];

/** Preserve only owner-projected code/reason pairs, never arbitrary details. */
export function validatedPublicErrorReason(
  code: string,
  value: unknown,
): PublicErrorReasonCode | undefined {
  const registry = errorRegistry.reason_registries.find(
    (entry) => entry.error_code === code,
  );
  return registry?.reason_codes.find((entry) => entry.code === value)?.code;
}
