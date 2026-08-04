import {
  type IncidentStreamMessage,
  incidentStreamMessageDecoder,
} from "@cartulary/protocol-ts/collaboration";

declare const message: IncidentStreamMessage;

export const collaborationDecodeResult =
  incidentStreamMessageDecoder.decode(message);
