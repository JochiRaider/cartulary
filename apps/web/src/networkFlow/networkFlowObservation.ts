import type {
  NetworkFlowGraphSelector,
  NetworkFlowRow,
} from "../services/networkFlowContractAdapter";

export function networkFlowEndpointIP(
  row: NetworkFlowRow,
  field: "network_flow.endpoint_a_ip" | "network_flow.endpoint_b_ip",
): string {
  return row[
    field === "network_flow.endpoint_a_ip"
      ? "network_flow.endpoint_a"
      : "network_flow.endpoint_b"
  ].ip;
}

// Labels consume the admitted direction variant. They never inspect source roles.
function networkFlowDirectionLabel(row: NetworkFlowRow): string {
  const direction = row["network_flow.direction"];
  if (direction.kind === "unknown")
    return direction.reason === "missing_orientation"
      ? "Unknown (missing orientation)"
      : "Unknown (incomplete orientation)";
  const from = direction.from_endpoint.toUpperCase();
  const to = direction.from_endpoint === "a" ? "B" : "A";
  return direction.kind === "client_to_server"
    ? `Client ${from} → server ${to}`
    : `Source ${from} → destination ${to}`;
}

export function networkFlowObservationValue(
  row: object,
  field: string,
): unknown {
  const values = row as Record<string, unknown>;
  const endpoint = /^network_flow\.endpoint_([ab])_(ip|port)$/u.exec(field);
  if (endpoint && endpoint[2] !== undefined)
    return (
      values[`network_flow.endpoint_${endpoint[1]}`] as
        | Record<string, unknown>
        | undefined
    )?.[endpoint[2]];
  if (field === "network_flow.direction")
    return networkFlowDirectionLabel(row as NetworkFlowRow);
  const direction = (row as NetworkFlowRow)["network_flow.direction"];
  if (field === "network_flow.direction_kind") return direction.kind;
  if (field === "network_flow.direction_reason")
    return direction.kind === "unknown" ? direction.reason : null;
  return values[field];
}

export function networkFlowSemanticEdgeLabel(
  selector: Exclude<NetworkFlowGraphSelector, { kind: "vertex" }>,
): string {
  if (selector.direction_kind === "unknown")
    return `Endpoint A ${selector.source_endpoint_value}:${selector.source_port ?? "—"} · Endpoint B ${selector.destination_endpoint_value}:${selector.destination_port_present ? selector.destination_port : "—"} · unknown direction · protocol ${selector.protocol}`;
  const roles =
    selector.direction_kind === "client_to_server"
      ? "client → server"
      : "source → destination";
  return `${selector.source_endpoint_value} → ${selector.destination_endpoint_value} · ${roles} · protocol ${selector.protocol} · destination port ${selector.destination_port_present ? selector.destination_port : "—"}`;
}
