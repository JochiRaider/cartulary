// Owns lifecycle and bounded collection facts. Adapters own I/O and transport.
export function createCollectionEngine({ policy, identity, adapter, gate, store, clock, observer, availabilityFor = () => "unsupported", onStatus = () => {} }) {
  const processes = new Map(), leases = new Map();
  let quiescenceError;
  let intent = "active", busy = null, timer, watching, pendingAdmission, stopCompletion;
  let controls = Promise.resolve(), quietWaiting = false, truncated = false;
  let sweeps = 0, omitted = 0, discoveryTruncations = 0, failures = 0, segment = 0;
  let lastContext = -Infinity, maximumSweepMs = 0, collectorHeapPeak = 0, latest = null;
  const publish = (value) => { latest = value; onStatus(value); };
  const enqueue = (operation) => { const completion = controls.then(operation); controls = completion.catch(() => {}); return completion; };
  function stopWatching() { watching?.(); watching = null; quietWaiting = false; }
  function cancel() { clock.clearTimeout(timer); stopWatching(); pendingAdmission?.abort(); }
  function waitForQuietEnd() {
    quietWaiting = true;
    watching = gate.waitUntilAvailable(() => { stopWatching(); if (intent === "active") schedule(); }, () => {
      failures += 1; watching?.(); watching = null;
      // Failure of the wakeup channel leaves collection quiescent.
      publish({ ...latest, availability: "collector_failed", failed_sweeps: failures });
    });
  }
  function register(proof, unitID, owner, allocationRef = null) {
    if (processes.has(proof.identity)) {
      // A direct lifecycle registration is more precise than discovery ancestry.
      if (owner === "registered" || processes.get(proof.identity).attribution !== "registered") Object.assign(processes.get(proof.identity), { unit_id: unitID, attribution: owner, allocation_ref: allocationRef });
      return;
    }
    if (processes.size >= policy.maximum_lifetime_processes) { omitted += 1; truncated = true; return; }
    processes.set(proof.identity, { proof, process_ref: `process:${processes.size + 1}`,
      unit_id: unitID, attribution: owner, allocation_ref: allocationRef, gone: false });
  }
  async function sweep() {
    const sweepSegment = segment;
    const active = () => intent === "active" && segment === sweepSegment;
    if (!active() || !adapter || truncated || store.truncated) return;
    if (gate.quiet()) { segment += 1; publish({ omitted_observations: omitted, discovery_truncations: discoveryTruncations, failed_sweeps: failures, availability: "paused_measurement", sampled_elapsed_ms: latest?.sampled_elapsed_ms ?? null, rss_bytes: null, process_count: null }); waitForQuietEnd(); return; }
    let lease;
    pendingAdmission = new AbortController();
    try {
      lease = await gate.acquire({ signal: pendingAdmission.signal });
      if (!active()) return;
      const start = clock.now(), deadline = start + policy.sweep_budget_ms;
      const discovered = await adapter.discover(policy.maximum_discovery_entries, deadline);
      if (discovered.truncated) discoveryTruncations += 1;
      const byPID = new Map(discovered.found.map((proof) => [proof.pid, proof]));
      const ancestors = new Map([...processes.values()].filter((entry) => !entry.gone).map((entry) => [entry.proof.pid, entry]));
      for (const proof of discovered.found) {
        if (processes.get(proof.identity)?.attribution === "registered") continue;
        let ancestor = proof, seen = new Set();
        while (ancestor && !seen.has(ancestor.pid)) {
          seen.add(ancestor.pid);
          const owned = ancestors.get(ancestor.stat.parent);
          const observedParent = byPID.get(ancestor.stat.parent);
          if (owned && observedParent?.identity === owned.proof.identity && BigInt(proof.start) >= BigInt(owned.proof.start)) {
            register(proof, owned.unit_id, "observed_descendant", owned.allocation_ref); break;
          }
          ancestor = observedParent;
        }
      }
      const sampledAt = clock.elapsed();
      let live = 0, rss = 0, completeRSS = true;
      for (const entry of processes.values()) {
        if (!active()) break;
        if (entry.gone) continue;
        if (live >= policy.maximum_live_processes || clock.now() >= deadline) { omitted += 1; completeRSS = false; continue; }
        let sample;
        try { sample = await adapter.processSample(entry.proof); }
        catch (error) { sample = { availability: availabilityFor(error), metrics: {} }; }
        if (!active()) break;
        if (sample.availability === "process_gone") entry.gone = true;
        else if (sample.availability === "available") live += 1;
        const memory = sample.metrics.rss_bytes;
        if (!entry.gone && memory?.availability === "available") rss += memory.value;
        else if (!entry.gone) completeRSS = false;
        await store.append({ schema_id: "cartulary.harness_resource_sample.v1", seq: store.samples + 1,
          elapsed_ms: sampledAt, segment: sweepSegment, identity_digest: entry.proof.identity, scope: "process", scope_ref: entry.process_ref,
          availability: sample.availability, metrics: sample.metrics });
      }
      if (active() && sampledAt - lastContext >= policy.context_cadence_ms && clock.now() < deadline) {
        lastContext = sampledAt;
        for (const record of await adapter.context()) {
          if (!active() || clock.now() >= deadline) { omitted += 1; break; }
          await store.append({ schema_id: "cartulary.harness_resource_sample.v1", seq: store.samples + 1,
            elapsed_ms: sampledAt, segment: sweepSegment, identity_digest: record.identity_digest, scope: record.scope, scope_ref: record.scope_ref,
            availability: "available", metrics: record.metrics });
        }
      }
      maximumSweepMs = Math.max(maximumSweepMs, clock.now() - start);
      collectorHeapPeak = Math.max(collectorHeapPeak, observer.heap());
      sweeps += 1;
      latest = { omitted_observations: omitted, discovery_truncations: discoveryTruncations, failed_sweeps: failures, availability: live > 0 ? "available" : "not_observed", sampled_elapsed_ms: sampledAt,
        rss_bytes: live > 0 && completeRSS ? rss : null, process_count: live > 0 ? live : null };
      publish(latest);
    } catch {
      if (!active()) return;
      failures += 1;
      publish({
        availability: latest?.availability ?? "not_observed",
        sampled_elapsed_ms: latest?.sampled_elapsed_ms ?? null,
        rss_bytes: latest?.rss_bytes ?? null, process_count: latest?.process_count ?? null,
        omitted_observations: omitted, discovery_truncations: discoveryTruncations, failed_sweeps: failures,
      });
    }
    finally {
      pendingAdmission = null;
      try { await lease?.release(); }
      catch (error) { quiescenceError = error; cancel(); throw error; }
    }
  }

  function schedule() {
    clock.clearTimeout(timer);
    if (quiescenceError || busy || intent !== "active" || quietWaiting || !adapter || truncated || store.truncated) return;
    timer = clock.setTimeout(() => {
      busy = sweep().catch(() => { failures += 1; }).finally(() => { busy = null; schedule(); });
    }, policy.process_cadence_ms);
  }
  async function finish() {
    await busy;
    gate.close();
    if (quiescenceError) throw quiescenceError;
    const shutdownStart = clock.now();
    const cpu = observer.cpu();
    await store.finish({
      schema_id: "cartulary.harness_resource_index.v1", ...identity,
      status: !adapter ? "unsupported" : truncated ? "truncated" : (failures || omitted || discoveryTruncations) ? "partial" : "complete",
      clock: "graph_process_monotonic", coverage: "observed_partial",
      omitted_observations: omitted, discovery_truncations: discoveryTruncations, failed_sweeps: failures, sweeps,
      processes: [...processes.values()].map((entry) => ({ process_ref: entry.process_ref,
        identity_digest: entry.proof.identity, unit_id: entry.unit_id, attribution: entry.attribution, allocation_ref: entry.allocation_ref })),
      leases: [...leases.values()],
      observer: { cpu_user_us: cpu.user, cpu_system_us: cpu.system, heap_peak_bytes: collectorHeapPeak,
        read_bytes: adapter?.bytesRead ?? 0, write_bytes: store.bytes, maximum_sweep_ms: maximumSweepMs,
        shutdown_ms: clock.now() - shutdownStart, gate_cpu: "not_observed" },
    }, shutdownStart, clock.now);
  }
  schedule();
  return {
    register(proof, unitID, allocationRef = null) { if (intent !== "stopped" && adapter) register(proof, unitID, "registered", allocationRef); },
    lease(record) {
      if (intent === "stopped") return;
      if (leases.size < policy.maximum_lifetime_processes || leases.has(record.lease_ref)) leases.set(record.lease_ref, record);
      else { omitted += 1; truncated = true; }
    },
    pause() {
      if (intent === "stopped") return stopCompletion;
      intent = "paused"; segment += 1; cancel();
      return enqueue(async () => { await busy; if (quiescenceError) throw quiescenceError; });
    },
    resume() {
      if (intent === "stopped") return stopCompletion;
      intent = "active"; segment += 1;
      return enqueue(() => { if (intent === "active") schedule(); });
    },
    stop(omittedRegistrations = 0) {
      if (stopCompletion) return stopCompletion;
      intent = "stopped"; cancel();
      omitted += omittedRegistrations; truncated ||= omittedRegistrations > 0;
      stopCompletion = enqueue(finish);
      return stopCompletion;
    },
  };
}
