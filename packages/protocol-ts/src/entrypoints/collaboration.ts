import type { IncidentStreamMessage } from "../generated/collaboration-types.js";
import { validateCartularyWsIncidentStreamMessageV1 } from "../generated/collaboration-validators.js";
import { createDecoder } from "../internal/decoder.js";

export type { IncidentStreamMessage };

export const incidentStreamMessageDecoder =
  createDecoder<IncidentStreamMessage>(
    "cartulary.ws.incident_stream_message.v1",
    validateCartularyWsIncidentStreamMessageV1,
  );
