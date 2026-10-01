export type TimelineBulkTagAuthoringSnapshot = Readonly<{
  raw: string;
  revision: number;
  generation: number;
  canEdit: boolean;
}>;

/** One presentation borrows commands; raw authoring survives its attachment. */
export interface TimelineBulkTagAuthoringPort {
  subscribe(listener: () => void): () => void;
  getSnapshot(): TimelineBulkTagAuthoringSnapshot | null;
  attach(): () => void;
  isCurrent(generation: number | undefined): boolean;
  update(raw: string, generation: number): boolean;
  clear(generation: number): boolean;
}
