import type { GridHandle } from "@cartulary/grid-adapter";
import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import {
  createContext,
  type ReactNode,
  useContext,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import {
  workbookFormInputStyle,
  workbookQuietCommandStyle,
} from "../components/workbookFormStyles";
import { useWorkbookMenuPlacement } from "../layout/useWorkbookMenuPlacement";
import {
  searchWorkbookCommands,
  type WorkbookCommandDescriptor,
  type WorkbookCommandTarget,
} from "./workbookCommandIndex";

class CommandIndex {
  private commands = new Map<string, WorkbookCommandDescriptor>();
  private listeners = new Set<() => void>();
  private revision = 0;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.revision;
  list = () => [...this.commands.values()];
  private publish() {
    this.revision++;
    for (const listener of this.listeners) listener();
  }
  register(command: WorkbookCommandDescriptor) {
    if (this.commands.has(command.id))
      throw new Error(`Duplicate command identity: ${command.id}`);
    this.commands.set(command.id, command);
    this.publish();
    return () => {
      if (this.commands.get(command.id) === command) {
        this.commands.delete(command.id);
        this.publish();
      }
    };
  }
}
type ContextValue = {
  readonly index: CommandIndex;
  readonly surface: string;
  readonly grid: () => GridHandle | null;
  readonly readable: boolean;
};
const Context = createContext<ContextValue | null>(null);
export function WorkbookCommandsProvider({
  children,
  ...value
}: Omit<ContextValue, "index"> & { readonly children: ReactNode }) {
  const index = useMemo(() => new CommandIndex(), []);
  return <Context value={{ ...value, index }}>{children}</Context>;
}
export function useWorkbookCommand(command: WorkbookCommandDescriptor | null) {
  const context = useContext(Context);
  const current = useRef(command);
  current.current = command;
  const key = command
    ? JSON.stringify([
        command.id,
        command.label,
        command.family,
        command.terms,
        command.targetKind,
      ])
    : null;
  // biome-ignore lint/correctness/useExhaustiveDependencies: Static descriptor identity determines registration lifetime; callbacks use the current owner binding.
  useLayoutEffect(() => {
    const contribution = current.current;
    if (!context?.index || !contribution) return;
    let attached = true;
    const unregister = context.index.register({
      ...contribution,
      availability: (target) =>
        attached && current.current
          ? current.current.availability(target)
          : "This command is no longer available.",
      invoke: (target) =>
        attached && current.current ? current.current.invoke(target) : false,
    });
    return () => {
      attached = false;
      unregister();
    };
  }, [context?.index, key]);
}

export function WorkbookCommandsControl() {
  const context = useContext(Context);
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(0);
  const [message, setMessage] = useState<string | null>(null);
  const panel = useRef<HTMLElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const input = useRef<HTMLInputElement>(null);
  const origin = useRef<HTMLElement | null>(null);
  useWorkbookMenuPlacement(open, panel);
  const captured = useRef<{
    surface: string;
    selection: readonly string[];
    cell: Extract<WorkbookCommandTarget, { kind: "record" | "cell" }> | null;
  } | null>(null);
  useSyncExternalStore(
    context?.index.subscribe ?? emptySubscribe,
    context?.index.getSnapshot ?? emptySnapshot,
  );
  useLayoutEffect(() => {
    if (!context?.readable) {
      setOpen(false);
      captured.current = null;
      origin.current = null;
    }
  }, [context?.readable]);
  useLayoutEffect(() => {
    if (!open) return;
    input.current?.focus();
    const outside = (event: Event) => {
      if (
        event.target instanceof Node &&
        !panel.current?.contains(event.target) &&
        !trigger.current?.contains(event.target)
      ) {
        setOpen(false);
        origin.current = null;
      }
    };
    document.addEventListener("pointerdown", outside, true);
    document.addEventListener("focusin", outside, true);
    return () => {
      document.removeEventListener("pointerdown", outside, true);
      document.removeEventListener("focusin", outside, true);
    };
  }, [open]);
  if (!context) return null;
  const capture = () => {
    origin.current =
      document.activeElement instanceof HTMLElement
        ? document.activeElement
        : null;
    const anchor = context.grid()?.getActiveCell?.();
    captured.current = {
      surface: context.surface,
      selection: context.grid()?.getSelectedRecordIds?.() ?? [],
      cell:
        anchor?.rowIdentity.kind === "core_record"
          ? {
              kind: "cell",
              viewSchemaId: context.surface,
              recordId: anchor.rowIdentity.recordId,
              fieldKey: anchor.fieldKey,
            }
          : null,
    };
  };
  const targetFor = (
    command: WorkbookCommandDescriptor,
  ): WorkbookCommandTarget | null => {
    if (command.targetKind === "shell") return { kind: "shell" };
    const target = captured.current;
    if (!target) return null;
    if (command.targetKind === "selection")
      return target.selection.length
        ? {
            kind: "selection",
            viewSchemaId: target.surface,
            recordIds: target.selection,
          }
        : null;
    if (command.targetKind === "surface")
      return { kind: "surface", viewSchemaId: target.surface };
    if (
      target.cell &&
      (command.targetKind === "record" || command.targetKind === "cell")
    )
      return { ...target.cell, kind: command.targetKind };
    return null;
  };
  const availability = (command: WorkbookCommandDescriptor) => {
    const target = targetFor(command);
    if (!target || !context.readable)
      return "Incident access is being checked.";
    if (target.kind !== "shell" && target.viewSchemaId !== context.surface)
      return "The selected surface changed.";
    if (target.kind === "cell" || target.kind === "record") {
      const active = context.grid()?.getActiveCell?.();
      if (
        active?.rowIdentity.kind !== "core_record" ||
        active.rowIdentity.recordId !== target.recordId ||
        (target.kind === "cell" && active.fieldKey !== target.fieldKey)
      )
        return "The selection changed. Reopen Commands to use the new selection.";
    }
    if (target.kind === "selection") {
      const current = context.grid()?.getSelectedRecordIds?.() ?? [];
      if (
        current.length !== target.recordIds.length ||
        !target.recordIds.every((id) => current.includes(id))
      )
        return "The bounded selection changed. Reopen Commands.";
    }
    return command.availability(target);
  };
  const commands = searchWorkbookCommands(
    context.index.list().filter((command) => targetFor(command) !== null),
    query,
  );
  const size = cartularyDesignPresentation.workbookWorkbench.command_page_size;
  const pages = Math.max(1, Math.ceil(commands.length / size));
  const currentPage = Math.min(page, pages - 1);
  const cancel = () => {
    setOpen(false);
    const before = origin.current;
    origin.current = null;
    (before?.isConnected && !before.closest("[inert],[hidden],:disabled")
      ? before
      : trigger.current
    )?.focus({ preventScroll: true });
  };
  return (
    <div
      style={{ position: "relative" }}
      data-grid-editor-external-action="true"
    >
      <button
        style={workbookQuietCommandStyle}
        ref={trigger}
        type="button"
        disabled={!context.readable}
        aria-expanded={open}
        onPointerDown={capture}
        onKeyDown={(event) => {
          if (event.key === "Enter" || event.key === " ") capture();
        }}
        onClick={() => {
          if (open) {
            cancel();
            return;
          }
          if (!captured.current) capture();
          setQuery("");
          setPage(0);
          setMessage(null);
          setOpen(true);
        }}
      >
        Commands
      </button>
      {open ? (
        <section
          ref={panel}
          aria-label="Commands"
          style={{
            // Escape the below-minimum header's horizontal scrollport.
            position: "fixed",
            insetInlineEnd: "var(--ct-spacing-sm)",
            top: "var(--ct-layout-topBarHeight)",
            width: "min(420px, 90vw)",
            boxSizing: "border-box",
            maxHeight: "70vh",
            overflow: "auto",
            padding: "var(--ct-spacing-md)",
            background: "var(--ct-colors-surface-1)",
            border: "var(--ct-border-hairline)",
            zIndex: 30,
          }}
          onKeyDown={(event) => {
            if (
              event.key === "Escape" &&
              !event.defaultPrevented &&
              !event.nativeEvent.isComposing
            ) {
              event.preventDefault();
              event.stopPropagation();
              cancel();
            }
          }}
        >
          <label>
            Search commands
            <input
              style={workbookFormInputStyle}
              ref={input}
              value={query}
              onChange={(event) => {
                setQuery(event.target.value);
                setPage(0);
                setMessage(null);
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter") event.preventDefault();
              }}
            />
          </label>
          <p>
            Commands search actions. Find in loaded rows searches cell content.
          </p>
          <p>
            {captured.current?.selection.length
              ? `Target: ${captured.current.selection.length} selected committed records`
              : captured.current?.cell
                ? "Target: selected committed cell"
                : "Target: current workbook surface"}
          </p>
          <ul style={{ padding: 0, listStyle: "none" }}>
            {commands
              .slice(currentPage * size, (currentPage + 1) * size)
              .map((command) => (
                <li key={command.id}>
                  <button
                    style={workbookQuietCommandStyle}
                    type="button"
                    aria-disabled={!!availability(command)}
                    onClick={() => {
                      const target = targetFor(command);
                      if (!target || !context.readable) return;
                      const reason = availability(command);
                      if (reason) {
                        setMessage(reason);
                        return;
                      }
                      if (command.invoke(target)) {
                        origin.current = null;
                        setOpen(false);
                      } else
                        setMessage("This command is not currently available.");
                    }}
                  >
                    {command.family} · {command.label}
                  </button>
                  {availability(command) ? (
                    <small style={{ display: "block" }}>
                      {availability(command)}
                    </small>
                  ) : null}
                </li>
              ))}
          </ul>
          <p role="status">
            {message ??
              (commands.length
                ? `${commands.length} commands. Page ${currentPage + 1} of ${pages}.`
                : "No matching commands.")}
          </p>
          <button
            style={workbookQuietCommandStyle}
            type="button"
            disabled={currentPage === 0}
            onClick={() => setPage(currentPage - 1)}
          >
            Previous commands
          </button>
          <button
            style={workbookQuietCommandStyle}
            type="button"
            disabled={currentPage + 1 >= pages}
            onClick={() => setPage(currentPage + 1)}
          >
            Next commands
          </button>
          <button
            style={workbookQuietCommandStyle}
            type="button"
            onClick={cancel}
          >
            Cancel
          </button>
        </section>
      ) : null}
    </div>
  );
}
const emptySubscribe = () => () => {};
const emptySnapshot = () => 0;
