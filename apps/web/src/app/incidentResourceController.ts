import {
  type IncidentResource,
  incidentResourceOrder,
} from "../shared/incidentResource";

export type IncidentResourceAuthority = Readonly<{
  incidentId: string;
  actorId: string;
  lifetime: string;
}>;

/** One current incident resource, with no editing, fetching or replay policy. */
export class IncidentResourceController {
  private resource: IncidentResource | null = null;
  private listeners = new Set<() => void>();
  constructor(
    private readonly ports: {
      isCurrent: (authority: IncidentResourceAuthority) => boolean;
      accepted: (resource: IncidentResource) => void;
    },
  ) {}
  getSnapshot = () => this.resource;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private emit() {
    for (const listener of this.listeners) listener();
  }
  retire = () => {
    if (this.resource) {
      this.resource = null;
      this.emit();
    }
  };
  accept = (
    resource: IncidentResource,
    authority: IncidentResourceAuthority,
  ): boolean => {
    if (!this.ports.isCurrent(authority)) return false;
    const order = incidentResourceOrder(
      this.resource,
      resource,
      authority.incidentId,
    );
    if (order === "invalid") return false;
    if (order === "new") {
      this.resource = Object.freeze({ ...resource });
      this.ports.accepted(this.resource);
      this.emit();
    }
    return true;
  };
}
