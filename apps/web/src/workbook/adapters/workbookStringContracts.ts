import { timezoneNameRegistry } from "@cartulary/protocol-ts/http";

const admittedTimezones: ReadonlySet<string> = new Set(
  timezoneNameRegistry.identifiers.map((item) => item.name),
);
/** Exact packaged membership; host Intl/tzdb is not an admission authority. */
export const isWorkbookTimezoneName = (value: string) =>
  admittedTimezones.has(value);
