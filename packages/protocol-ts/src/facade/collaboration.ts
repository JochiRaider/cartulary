import type { IncidentStreamMessage } from "../generated/collaboration-types.js";

import { createGeneratedDecoder } from "./runtimeValidation.js";

export type { IncidentStreamMessage };

export const incidentStreamMessageDecoder =
  createGeneratedDecoder<IncidentStreamMessage>(
    "cartulary.ws.incident_stream_message.v1",
  );
