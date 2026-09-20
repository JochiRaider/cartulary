import {
  type GridCellAnchor,
  type GridCellNavigationOptions,
  type GridCellNavigationResult,
  type GridPresentationSnapshot,
  gridAnchorKey,
} from "@cartulary/grid-adapter";
import { cartularyDesignPresentation } from "@cartulary/ui-contracts";

/** Source-owned committed text; membership includes offscreen cells, never hidden rows.
 * Source object identity is the content revision: replace it on committed value
 * changes while retaining navigationKey for an unchanged accepted window/order.
 */
export type WorkbookFindSource = {
  readonly lifetimeKey: string;
  readonly navigationKey: string;
  readonly presentation: GridPresentationSnapshot;
  readonly readable: boolean;
  readonly unavailableReason: string | null;
  readonly stale: boolean;
  readonly readText: (anchor: GridCellAnchor) => readonly string[];
  readonly isCurrent?: () => boolean;
  readonly isNavigationCurrent?: () => boolean;
};

export type WorkbookFindSnapshot = {
  readonly active: boolean;
  readonly open: boolean;
  readonly term: string;
  readonly matchCase: boolean;
  readonly status: "idle" | "computing" | "ready" | "unavailable";
  readonly matches: readonly GridCellAnchor[];
  readonly current: GridCellAnchor | null;
  readonly navigating: boolean;
  readonly message: string;
  readonly stale: boolean;
};

const initial = (): WorkbookFindSnapshot => ({
  active: false,
  open: false,
  term: "",
  matchCase: false,
  status: "idle",
  matches: [],
  current: null,
  navigating: false,
  message: "",
  stale: false,
});

export function normalizeFindText(text: string, matchCase: boolean): string {
  const normalized = text.normalize("NFC");
  return matchCase ? normalized : normalized.toLowerCase().normalize("NFC");
}

export function findCellMatches(
  fragments: readonly string[],
  term: string,
  matchCase: boolean,
): boolean {
  return (
    term !== "" &&
    fragments.some((text) => normalizeFindText(text, matchCase).includes(term))
  );
}

export function workbookFindStatus(snapshot: WorkbookFindSnapshot): string {
  if (snapshot.status === "unavailable") return snapshot.message;
  if (snapshot.navigating) return "Waiting for the edit to save.";
  if (snapshot.status === "computing") return "Finding in loaded rows…";
  if (snapshot.term === "") return "Enter text to find in loaded rows.";
  if (snapshot.matches.length === 0)
    return cartularyDesignPresentation.workbookFind.zeroMatchesMessage;
  const index =
    snapshot.current === null
      ? -1
      : snapshot.matches.findIndex(
          (match) =>
            gridAnchorKey(match) ===
            gridAnchorKey(snapshot.current as GridCellAnchor),
        );
  return `${index < 0 ? "" : `${index + 1} of `}${snapshot.matches.length} matching ${snapshot.matches.length === 1 ? "cell" : "cells"} in loaded rows.${snapshot.message ? ` ${snapshot.message}` : ""}`;
}

/** Owns search intent only. Query data, authorization, drafts and writes remain with sources. */
export class WorkbookFindController {
  private snapshot = initial();
  private source: WorkbookFindSource | null = null;
  private origin: GridCellAnchor | null = null;
  private listeners = new Set<() => void>();
  private scanGeneration = 0;
  private scheduled: ReturnType<typeof setTimeout> | null = null;
  private destination: AbortController | null = null;
  private destinationTarget: GridCellAnchor | null = null;

  constructor(
    private readonly move: (
      anchor: GridCellAnchor,
      options: GridCellNavigationOptions,
    ) => Promise<GridCellNavigationResult>,
  ) {}

  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(patch: Partial<WorkbookFindSnapshot>) {
    this.snapshot = { ...this.snapshot, ...patch };
    for (const listener of this.listeners) listener();
  }
  cancelNavigation = () => {
    this.destination?.abort();
    this.destination = null;
    this.destinationTarget = null;
    if (this.snapshot.navigating) this.publish({ navigating: false });
  };
  private cancelScan() {
    this.scanGeneration += 1;
    if (this.scheduled !== null) clearTimeout(this.scheduled);
    this.scheduled = null;
  }
  setSource(source: WorkbookFindSource | null) {
    if (source === null || !source.readable) {
      this.source = null;
      this.close();
      return;
    }
    if (this.source === source) return;
    if (this.source && this.source.lifetimeKey !== source.lifetimeKey)
      this.close();
    if (this.source?.navigationKey !== source.navigationKey)
      this.cancelNavigation();
    this.source = source;
    if (
      this.destinationTarget &&
      !findCellMatches(
        source.readText(this.destinationTarget),
        normalizeFindText(this.snapshot.term, this.snapshot.matchCase),
        this.snapshot.matchCase,
      )
    )
      this.cancelNavigation();
    if (this.snapshot.active) this.scan();
  }
  open(origin: GridCellAnchor | null) {
    const starting = !this.snapshot.active;
    this.cancelNavigation();
    if (starting) this.origin = origin;
    this.publish({ active: true, open: true, message: "" });
    if (starting) this.scan();
  }
  collapse = () => {
    this.cancelNavigation();
    if (this.snapshot.open) this.publish({ open: false });
  };
  close = () => {
    this.cancelScan();
    this.cancelNavigation();
    this.origin = null;
    this.snapshot = initial();
    for (const listener of this.listeners) listener();
  };
  dispose = () => {
    this.source = null;
    this.close();
  };
  setTerm(term: string) {
    if (term === this.snapshot.term) return;
    this.cancelNavigation();
    this.publish({ term, message: "" });
    this.scan();
  }
  setMatchCase(matchCase: boolean) {
    if (matchCase === this.snapshot.matchCase) return;
    this.cancelNavigation();
    this.publish({ matchCase, message: "" });
    this.scan();
  }
  private scan() {
    this.cancelScan();
    const source = this.source;
    if (!source || source.unavailableReason !== null) {
      this.publish({
        status: "unavailable",
        matches: [],
        current: null,
        message: source?.unavailableReason ?? "Loaded rows are unavailable.",
      });
      return;
    }
    const { term, matchCase } = this.snapshot;
    if (term === "") {
      this.publish({
        status: "idle",
        matches: [],
        current: null,
        stale: source.stale,
      });
      return;
    }
    const normalized = normalizeFindText(term, matchCase);
    const generation = this.scanGeneration;
    const model = source.presentation;
    const matches: GridCellAnchor[] = [];
    let index = 0;
    this.publish({ status: "computing", stale: source.stale, message: "" });
    const work = () => {
      if (generation !== this.scanGeneration || source.isCurrent?.() === false)
        return;
      const count = model.rowIdentities.length * model.fieldKeys.length;
      // Yield between bounded groups of cells; never truncate source strings or results.
      const end = Math.min(count, index + 32);
      for (; index < end; index += 1) {
        const rowIdentity =
          model.rowIdentities[Math.floor(index / model.fieldKeys.length)];
        const fieldKey = model.fieldKeys[index % model.fieldKeys.length];
        if (!rowIdentity || !fieldKey) continue;
        const anchor = { surface: model.surface, rowIdentity, fieldKey };
        if (findCellMatches(source.readText(anchor), normalized, matchCase))
          matches.push(anchor);
      }
      if (generation !== this.scanGeneration || source.isCurrent?.() === false)
        return;
      if (index < count) {
        this.scheduled = setTimeout(work, 0);
        return;
      }
      this.scheduled = null;
      const current = this.snapshot.current;
      const retained =
        current !== null &&
        matches.some(
          (match) => gridAnchorKey(match) === gridAnchorKey(current),
        );
      if (current && !retained) this.origin = null;
      this.publish({
        status: "ready",
        matches,
        current: retained ? current : null,
      });
    };
    this.scheduled = setTimeout(work, 0);
  }
  async navigate(
    direction: 1 | -1,
    active: GridCellAnchor | null = null,
  ): Promise<GridCellNavigationResult> {
    const state = this.snapshot;
    const source = this.source;
    if (!source || state.status !== "ready" || state.matches.length === 0)
      return "unavailable";
    this.cancelNavigation();
    const abort = new AbortController();
    this.destination = abort;
    const currentIndex =
      state.current === null
        ? -1
        : state.matches.findIndex(
            (match) =>
              gridAnchorKey(match) ===
              gridAnchorKey(state.current as GridCellAnchor),
          );
    let index = direction === 1 ? 0 : state.matches.length - 1;
    let wrapped = false;
    if (currentIndex >= 0) {
      index = currentIndex + direction;
      wrapped = index < 0 || index === state.matches.length;
      index = (index + state.matches.length) % state.matches.length;
    } else {
      const origin = this.origin ?? active;
      const position = (anchor: GridCellAnchor) => {
        const row = source.presentation.rowIdentities.findIndex(
          (identity) =>
            gridAnchorKey({ ...anchor, rowIdentity: identity }) ===
            gridAnchorKey(anchor),
        );
        const column = source.presentation.fieldKeys.indexOf(anchor.fieldKey);
        return row < 0 || column < 0
          ? -1
          : row * source.presentation.fieldKeys.length + column;
      };
      const start = origin === null ? -1 : position(origin);
      if (start >= 0) {
        const candidates = state.matches.map((match, matchIndex) => ({
          matchIndex,
          position: position(match),
        }));
        const candidate =
          direction === 1
            ? candidates.find((entry) => entry.position >= start)
            : candidates.reverse().find((entry) => entry.position <= start);
        if (candidate) index = candidate.matchIndex;
        else wrapped = true;
      }
    }
    const target = state.matches[index];
    if (!target) return "unavailable";
    this.destinationTarget = target;
    const navigationKey = source.navigationKey;
    const { term, matchCase } = state;
    const isCurrent = () =>
      !abort.signal.aborted &&
      this.destination === abort &&
      this.source?.navigationKey === navigationKey &&
      this.source.isNavigationCurrent?.() !== false &&
      this.snapshot.term === term &&
      this.snapshot.matchCase === matchCase &&
      findCellMatches(
        this.source.readText(target),
        normalizeFindText(term, matchCase),
        matchCase,
      );
    this.publish({ navigating: true, message: "" });
    let result: GridCellNavigationResult;
    try {
      result = await this.move(target, {
        signal: abort.signal,
        isCurrent,
        beforeFocus: () => {
          if (isCurrent()) this.publish({ open: false });
        },
      });
    } catch {
      result = "unavailable";
    }
    if (!isCurrent()) {
      if (this.destination === abort) {
        this.destination = null;
        this.publish({
          navigating: false,
          message: "Matching cell no longer available.",
        });
      }
      return "cancelled";
    }
    this.destination = null;
    this.publish({
      navigating: false,
      ...(result === "focused"
        ? {
            current: target,
            open: false,
            message: wrapped
              ? direction === 1
                ? "Wrapped to beginning."
                : "Wrapped to end."
              : "",
          }
        : {
            message:
              result === "rejected"
                ? "Correct the original edit before navigating."
                : "Matching cell no longer available.",
          }),
    });
    return result;
  }
}
