import type { GetIncidentResponse } from "@cartulary/protocol-ts/http";

/** Current incident resource, independent of any editing workflow. */
export type IncidentResource = GetIncidentResponse["data"];
