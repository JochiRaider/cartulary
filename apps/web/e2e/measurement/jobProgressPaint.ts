export type JobProgressPaintInput = {
  jobPurpose:
    | "discovery"
    | "apply"
    | "snapshot"
    | "report"
    | "evidence_verification"
    | "reference_pack_import"
    | "reference_pack_reverify"
    | "reference_pack_refresh";
  gridLocation?: "current" | "opener";
  recordId: string;
  rowVersion: number;
  fieldKey: string;
  draftValue: string;
  gridSelector: string;
  editorSelector: string;
  progressSurfaceSelector: string;
  startMark: string;
  admissionMark: string;
  dispatchMark: string;
  timeoutMs: number;
};
export type JobProgressCompletion = {
  elapsed: number;
  jobId: string;
  jobStatus: "queued" | "running";
  blobId?: string;
};

// The DOM regression runs the same function serialized into Chromium.
export function observeJobProgressPaintInBrowser(
  input: JobProgressPaintInput,
): Promise<JobProgressCompletion> {
  const origin = performance.timeOrigin;
  const deadline = performance.now() + input.timeoutMs;
  const companion = input.gridLocation === "opener";
  const gridWindow: Window | null = companion ? window.opener : window;
  if (!gridWindow || (companion && gridWindow === window))
    return Promise.reject(new Error("Job measurement companion is missing"));
  let gridDocument: Document;
  let gridOrigin: number;
  try {
    gridDocument = gridWindow.document;
    gridOrigin = gridWindow.performance.timeOrigin;
  } catch {
    return Promise.reject(
      new Error("Job measurement companion is not same-origin"),
    );
  }
  const grid = gridDocument.querySelector<HTMLElement>(input.gridSelector);
  const assistant = document.querySelector<HTMLElement>(
    input.progressSurfaceSelector,
  );
  if (!grid || !assistant)
    return Promise.reject(new Error("Job measurement surfaces are missing"));
  const fixed = new Set<HTMLElement>();
  for (const root of [grid, assistant]) {
    fixed.add(root);
    for (const node of root.querySelectorAll<HTMLElement>("*")) fixed.add(node);
    for (let parent = root.parentElement; parent; parent = parent.parentElement)
      fixed.add(parent);
  }
  const scroll = [...fixed]
    .filter(
      (node) =>
        node.scrollHeight > node.clientHeight ||
        node.scrollWidth > node.clientWidth,
    )
    .map((node) => ({ node, left: node.scrollLeft, top: node.scrollTop }));
  function visible(node: HTMLElement) {
    const owner = node.ownerDocument;
    const view = owner.defaultView;
    if (!view) return false;
    let left = 0,
      top = 0,
      right = view.innerWidth,
      bottom = view.innerHeight;
    for (
      let parent: HTMLElement | null = node;
      parent;
      parent = parent.parentElement
    ) {
      const style = view.getComputedStyle(parent);
      if (
        !parent.isConnected ||
        parent.hidden ||
        parent.hasAttribute("inert") ||
        parent.getAttribute("aria-hidden") === "true" ||
        style.display === "none" ||
        style.visibility === "hidden" ||
        Number(style.opacity || 1) <= 0
      )
        return false;
      if (
        parent !== node &&
        style.display !== "contents" &&
        /auto|scroll|hidden|clip/u.test(
          `${style.overflow} ${style.overflowX} ${style.overflowY}`,
        )
      ) {
        const b = parent.getBoundingClientRect();
        left = Math.max(left, b.left);
        top = Math.max(top, b.top);
        right = Math.min(right, b.right);
        bottom = Math.min(bottom, b.bottom);
      }
    }
    const b = node.getBoundingClientRect();
    const hit = owner.elementFromPoint(
      b.left + b.width / 2,
      b.top + b.height / 2,
    );
    return (
      hit !== null &&
      node.contains(hit) &&
      b.width > 0 &&
      b.height > 0 &&
      b.left >= left &&
      b.right <= right &&
      b.top >= top &&
      b.bottom <= bottom
    );
  }
  return new Promise((resolveResult, rejectResult) => {
    let settled = false;
    let timer: number | undefined;
    const reject = (error: Error) => {
      settled = true;
      if (timer !== undefined) window.clearTimeout(timer);
      rejectResult(error);
    };
    const resolve = (value: JobProgressCompletion) => {
      settled = true;
      if (timer !== undefined) window.clearTimeout(timer);
      resolveResult(value);
    };
    if (companion)
      timer = window.setTimeout(
        () =>
          reject(
            new Error(
              "Companion animation frames did not qualify before the observation deadline",
            ),
          ),
        input.timeoutMs,
      );
    let previous: HTMLElement[] = [],
      frames = 0,
      previousJob = "";
    const schedule = () =>
      requestAnimationFrame(() => {
        if (settled) return;
        // Each qualification cycle includes a frame in both actual documents.
        // One document's DOM state must never stand in for the other's paint.
        if (companion) gridWindow.requestAnimationFrame(tick);
        else tick();
      });
    const tick = () => {
      if (settled) return;
      let gridChanged = true;
      try {
        gridChanged =
          (companion && window.opener !== gridWindow) ||
          gridWindow.closed ||
          gridWindow.document !== gridDocument ||
          gridWindow.performance.timeOrigin !== gridOrigin ||
          gridWindow.innerWidth !== 1920 ||
          gridWindow.innerHeight !== 1080 ||
          gridWindow.devicePixelRatio !== 1;
      } catch {
        // A cross-origin navigation invalidates the prepared document as well.
      }
      if (
        performance.timeOrigin !== origin ||
        innerWidth !== 1920 ||
        innerHeight !== 1080 ||
        devicePixelRatio !== 1 ||
        gridChanged ||
        gridDocument.querySelector(input.gridSelector) !== grid ||
        document.querySelector(input.progressSurfaceSelector) !== assistant ||
        scroll.some(
          ({ node, left, top }) =>
            node.isConnected &&
            (node.scrollLeft !== left || node.scrollTop !== top),
        )
      ) {
        reject(
          new Error("Job measurement document, geometry or scroll changed"),
        );
        return;
      }
      const starts = performance.getEntriesByName(
        input.startMark,
        "mark",
      ) as PerformanceMark[];
      const admissions = performance.getEntriesByName(
        input.admissionMark,
        "mark",
      ) as PerformanceMark[];
      const dispatches = performance.getEntriesByName(
        input.dispatchMark,
        "mark",
      ) as PerformanceMark[];
      if (starts.length > 1 || admissions.length > 1 || dispatches.length > 1) {
        reject(new Error("Job action or admission was repeated"));
        return;
      }
      const start = starts[0],
        admission = admissions[0],
        dispatch = dispatches[0];
      if (
        admission &&
        (!start?.detail?.transactionId ||
          admission.detail?.transactionId !== start.detail.transactionId ||
          admission.startTime < start.startTime ||
          typeof admission.detail?.jobId !== "string" ||
          !admission.detail.jobId)
      ) {
        reject(new Error("Job admission does not match the accepted action"));
        return;
      }
      if (
        dispatch &&
        (!start ||
          dispatch.detail?.transactionId !== start.detail?.transactionId ||
          dispatch.detail?.replay !== false ||
          dispatch.startTime < start.startTime ||
          (admission && admission.startTime < dispatch.startTime))
      ) {
        reject(
          new Error("Job dispatch does not match one fresh accepted action"),
        );
        return;
      }
      const referencePack = input.jobPurpose.startsWith("reference_pack_");
      if (
        referencePack &&
        start &&
        (start.detail?.purpose !== input.jobPurpose ||
          (dispatch && dispatch.detail?.purpose !== input.jobPurpose) ||
          (admission && admission.detail?.purpose !== input.jobPurpose))
      ) {
        reject(new Error("Reference Pack action purpose changed"));
        return;
      }
      const jobs = [
        ...assistant.querySelectorAll<HTMLElement>("[data-job-id]"),
      ].filter(
        (node) =>
          !referencePack || node.dataset.jobId === admission?.detail?.jobId,
      );
      const job = jobs.length === 1 ? jobs[0] : undefined;
      if (
        input.jobPurpose === "evidence_verification" &&
        admission &&
        (!start?.detail?.blobId ||
          admission.detail?.blobId !== start.detail.blobId ||
          dispatch?.detail?.blobId !== start.detail.blobId)
      ) {
        reject(
          new Error("Evidence verification admission changed its blob binding"),
        );
        return;
      }
      const status = job?.dataset.jobStatus;
      const progress = job?.querySelector<HTMLElement>(
        '[data-job-progress="true"]',
      );
      const cancel = job?.querySelector<HTMLButtonElement>(
        'button[data-job-cancel="true"]',
      );
      const editors = grid.querySelectorAll<
        HTMLInputElement | HTMLTextAreaElement
      >(input.editorSelector);
      const editor = editors.length === 1 ? editors[0] : undefined;
      const row = editor?.closest<HTMLElement>("[data-grid-record-id]");
      const ready =
        !!start &&
        !!dispatch &&
        !!admission &&
        !!job &&
        job.dataset.jobId === admission.detail.jobId &&
        job.dataset.jobPurpose === input.jobPurpose &&
        (input.jobPurpose !== "evidence_verification" ||
          job.dataset.jobBlobId === admission.detail.blobId) &&
        (status === "queued" || status === "running") &&
        job.dataset.jobCancelable === "true" &&
        !!progress?.textContent?.trim() &&
        visible(progress) &&
        !!cancel &&
        !cancel.disabled &&
        cancel.getAttribute("aria-disabled") !== "true" &&
        visible(cancel) &&
        !!editor &&
        !editor.disabled &&
        !editor.readOnly &&
        gridDocument.activeElement === editor &&
        visible(editor) &&
        editor.value === input.draftValue &&
        row?.dataset.gridRecordId === input.recordId &&
        Number(row.dataset.gridRowVersion) === input.rowVersion &&
        editor.closest<HTMLElement>("[data-grid-field-key]")?.dataset
          .gridFieldKey === input.fieldKey;
      const nodes =
        ready && job && progress && cancel && editor && row
          ? [job, progress, cancel, editor, row]
          : [];
      const identity = ready ? (job?.dataset.jobId ?? "") : "";
      frames = ready
        ? identity === previousJob &&
          nodes.every((node, index) => node === previous[index])
          ? frames + 1
          : 1
        : 0;
      previous = nodes;
      previousJob = identity;
      if (
        frames === 2 &&
        start &&
        admission &&
        (status === "queued" || status === "running")
      ) {
        const elapsed = performance.now() - start.startTime;
        if (!Number.isFinite(elapsed) || elapsed < 0) {
          reject(new Error("Job elapsed interval is invalid"));
          return;
        }
        resolve({
          elapsed,
          jobId: admission.detail.jobId,
          jobStatus: status,
          ...(input.jobPurpose === "evidence_verification"
            ? { blobId: admission.detail.blobId as string }
            : {}),
        });
        return;
      }
      if (performance.now() >= deadline) {
        reject(
          new Error(
            "Job progress, cancel and grid edit did not qualify across two visible frames",
          ),
        );
        return;
      }
      schedule();
    };
    schedule();
  });
}
