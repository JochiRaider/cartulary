// The pinned public run promise settles before isolated fork teardown. Observe
// the public PoolWorker boundary; execution and IPC remain Vitest's responsibility.
export function settledForkPool(ForksPoolWorker) {
  let current;
  return {
    beginFile() {
      if (current) throw new Error("previous Vitest file has not settled");
      const workers = [];
      current = workers;
      return async () => {
        if (workers.length !== 1) throw new Error("Vitest file requires exactly one isolated fork");
        const result = await workers[0];
        if (!result.acknowledged || result.terminationError) throw new Error("Vitest fork termination is uncertain");
        current = undefined;
        return result.teardownError;
      };
    },
    policy: {
      name: "cartulary-settled-forks",
      createPoolWorker(options) {
        if (!current) throw new Error("Vitest fork has no selected file owner");
        let settle;
        current.push(new Promise((resolve) => { settle = resolve; }));
        const result = { acknowledged: false, teardownError: null, terminationError: null };
        return new class extends ForksPoolWorker {
          async start() {
            try {
              await super.start();
              this.on("message", (message) => {
                const response = this.deserialize(message);
                if (response?.__vitest_worker_response__ && response.type === "stopped") {
                  result.acknowledged = true;
                  result.teardownError = response.error ?? null;
                }
              });
            } catch (error) { result.terminationError = error; settle(result); throw error; }
          }
          async stop() {
            try { await super.stop(); }
            catch (error) { result.terminationError = error; throw error; }
            finally { settle(result); }
          }
        }(options);
      },
    },
  };
}
