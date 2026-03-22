# Responsive Interface Design for Browser-Based Web UIs

## Thesis

For desktop browser interfaces that behave like spreadsheets or other structured-data workspaces, responsiveness is not just low latency. It is the preservation of a tight perception-action loop in which users act directly on domain objects on the working surface, immediately see how the system interpreted the action, encounter constraints that expose ambiguity or invalidity before costly errors accumulate, and remain insulated from network, storage, and synchronization internals. Across foundational HCI, modern visual analytics, spreadsheet-comprehension research, and local-first/reactive web-systems work, breakdowns in that loop lead to slower action, strategy shifts, more context-seeking, weaker confidence, and higher error risk. The strongest browser designs therefore combine client-centric reactive architecture with spreadsheet-specific supports for dependency visibility, layered structure, overview/detail navigation, and explicit provisional state. ([Shneiderman, 1983](https://www.cs.umd.edu/users/ben/papers/Shneiderman1983Direct.pdf))

## Executive summary

- The enduring core idea is still direct manipulation: interfaces are most cognitively continuous when the user works on a visible problem representation and the effect of each action is rapid, incremental, and legible. Modern descendants of this idea include semantic interaction, direct-manipulation data wrangling, spreadsheet-like web-app builders, and synchronized tabular control surfaces. ([Shneiderman, 1983](https://www.cs.umd.edu/users/ben/papers/Shneiderman1983Direct.pdf))

- Delay changes behavior, not just satisfaction. In exploratory visual analysis, an added 500 ms reduced activity, coverage, and rates of observation and hypothesis generation; latency effects also vary with task complexity, and anticipated delays can slow users before feedback even appears. The spreadsheet-specific latency literature is thinner, so exact timing targets should not be copied mechanically from adjacent domains. ([Liu & Heer, 2014](https://pubmed.ncbi.nlm.nih.gov/26356926/))

- When work cannot complete immediately, staged interfaces can still preserve flow if they return something useful quickly and clearly mark its status. Incremental visualization, optimistic visualization, and latency-aware prefetch/progressive-resolution systems show that partial or approximate feedback can support exploration when users can later verify or recover. Direct spreadsheet evidence here is indirect, so this is a careful transfer from visualization research rather than a settled spreadsheet result. ([Fisher et al., 2012](https://www.microsoft.com/en-us/research/publication/trust-me-im-partially-right-incremental-visualization-lets-analysts-explore-large-datasets-faster/))

- Spreadsheet comprehension problems are often problems of hidden structure. Users spend substantial time seeking labels, provenance, intent, and contextual meaning; they chase references across sheets, lose place, and often want to ask the original author. That makes responsiveness partly a problem of reducing context-switch cost, not just reducing render time. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

- Structure and flexibility are in tension, not opposition. Irregular layout, spacing, color, and annotations help people think, communicate, and coordinate work, while regular structure enables sort, filter, lookup, charting, validation, and automation. The implication is not “force everything into tables,” but “layer optional structure over a flexible grid.” ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

- Browser architecture is part of interaction design. Reactive dataflow, client-side persistent state, local-first sync, explicit integrity semantics, cached summaries, and prefetch/indexing are not backend niceties; they are what let the interface acknowledge, preview, and reconcile work fast enough to preserve cognitive flow. ([Satyanarayan et al., 2016](https://pubmed.ncbi.nlm.nih.gov/26390466/))

## Modern research synthesis

### Scope and evidence limits

This report treats “responsive interface design” primarily as interaction responsiveness, feedback timing, and preservation of cognitive continuity in desktop/laptop browser UIs using mouse, keyboard, and trackpad. It largely excludes breakpoint-driven layout adaptation except where navigation or overview mechanisms for large tables are relevant.

Three transfer limits matter. First, the strongest empirical latency papers are mostly from visual analytics and pointing/steering, not spreadsheet editing. Second, the strongest spreadsheet-comprehension studies were conducted in Excel-like environments rather than browser-native spreadsheet engines. Third, many local-first/reactive papers are architecture or systems papers with limited comparative end-user evaluation. Accordingly, the report treats spreadsheet comprehension and large-sheet navigation results as the strongest direct evidence for spreadsheet UI, while some browser-architecture recommendations are evidence-backed design interpretations rather than settled behavioral facts. ([Liu & Heer, 2014](https://pubmed.ncbi.nlm.nih.gov/26356926/))

### 1. Direct manipulation and visible system state

**Foundational evidence.** Shneiderman’s classic definition of direct manipulation emphasized continuous representation of the object of interest, physical action rather than command syntax, and rapid, incremental, reversible operations with immediately visible results. Hutchins, Hollan, and Norman then gave a cognitive account of why such systems feel direct, centering semantic and articulatory distance between the user’s intention, the interface expression, and its physical form. Endert’s semantic interaction work extends the same logic to analytic systems: users manipulate meaningful visual objects, while the system infers model changes under the surface. ([Shneiderman, 1983](https://www.cs.umd.edu/users/ben/papers/Shneiderman1983Direct.pdf))

**System evidence.** Modern structured-data tools repeatedly succeed when they keep the user’s domain representation visible and let system internals stay implicit. Wrangler couples direct selection in a data table with inferred transforms and an inspectable history of transformation scripts. Gneiss keeps spreadsheet data, interface bindings, and live output in one visible environment with two-way data flow. Wildcard synchronizes an application’s data with a spreadsheet view so that manipulating either surface updates the other. These systems do not eliminate computation or program structure; they relocate it behind a user-facing representation that remains semantically legible. ([Wrangler, 2011](https://idl.uw.edu/papers/wrangler))

**Design interpretation.** For a browser spreadsheet or spreadsheet-adjacent UI, responsiveness therefore means more than “fast repaint.” The interface should let users act on cells, columns, ranges, filters, formulas, and annotations directly on the workspace, and it should show the system’s interpretation at the same semantic level. Users should not need to reason about reducers, RPCs, sync queues, or parser internals to stay oriented. That interpretation can be explicit without becoming technical: “sorted this region by Date descending,” “interpreted 17 pasted values as dates,” or “filled formula pattern across 420 cells” is useful; “dispatched async mutation” is not. This last sentence is a design interpretation, not a direct empirical finding. ([Shneiderman, 1983](https://www.cs.umd.edu/users/ben/papers/Shneiderman1983Direct.pdf))

### 2. Interactive latency and cognitive continuity

**Empirical finding.** Liu and Heer’s controlled study remains the most important direct evidence for how delay alters exploratory interaction: adding 500 ms produced significant costs, decreasing activity and dataset coverage, and reducing the rate of observations, generalizations, and hypotheses. They also found a carryover effect, where initial exposure to higher latency degraded later performance even after latency dropped. Battle and colleagues showed that latency effects are not uniform but interact with task complexity and search behavior. ([Liu & Heer, 2014](https://pubmed.ncbi.nlm.nih.gov/26356926/))

**Adjacent empirical finding.** In low-latency motor tasks, Friston and colleagues found that latency affects pointing and steering performance in a non-linear way and can matter at very low ranges. More recently, Bogon and colleagues found that anticipated system delay slows user actions before delayed feedback is even seen. Taken together, these papers suggest that users internalize timing properties of an interface and adapt strategy, tempo, and confidence around them. The transfer from pointing tasks to spreadsheet work is indirect but relevant for mouse- and trackpad-mediated actions such as drag-selection, fill-handle moves, resizing, and scrub-like controls. ([Friston et al., 2016](https://pubmed.ncbi.nlm.nih.gov/27045915/))

**Design interpretation.** A browser spreadsheet should therefore optimize for a stable interaction beat, not just aggregate throughput. The hot path is not only keystroke echo; it includes selection movement, cross-reference peek, drag and fill, filter application, and navigation among distant regions. The literature does not justify a single spreadsheet-wide numeric cutoff, but it does justify designing so that common micro-actions receive immediate local acknowledgement and do not frequently stall behind network, remote compute, or whole-sheet recomputation. ([Liu & Heer, 2014](https://pubmed.ncbi.nlm.nih.gov/26356926/))

### 3. Progressive, approximate, optimistic, and staged interaction

**Empirical finding from adjacent visualization literature.** Fisher and colleagues argued that incremental visualization lets analysts explore large datasets faster by giving them progressively improving results rather than forcing them to wait for final output. Moritz and colleagues’ optimistic visualization work showed that approximate answers can support interactive exploration when the system provides a way to detect and recover from error later. Falcon adds a more architectural version of the same idea, using user-centered prefetching, indexing, and progressive resolution so brushing and view switching stay fast even on very large datasets. ([Fisher et al., 2012](https://www.microsoft.com/en-us/research/publication/trust-me-im-partially-right-incremental-visualization-lets-analysts-explore-large-datasets-faster/))

**Inference for spreadsheets.** Spreadsheet-specific literature on optimistic or approximate interaction is comparatively thin. The strongest direct bridge is therefore inferential: large browser sheets, remote tables, and grid-plus-database hybrids can likely preserve flow by separating immediate local acknowledgement from full validation or full-resolution computation. For example, a costly group-by, sort over remote data, or formula-heavy aggregate can show a preview or progressive refinement if the provisional state is clearly marked and the final verified result is easy to compare. This is an inference from adjacent literature, not direct spreadsheet evidence. ([Fisher et al., 2012](https://www.microsoft.com/en-us/research/publication/trust-me-im-partially-right-incremental-visualization-lets-analysts-explore-large-datasets-faster/))

**Design interpretation.** The trust condition is crucial. Staged interfaces help only when provisionality is legible. “Approximate,” “preview,” “stale,” “syncing,” and “verified” are different states and should not be conflated. If eventual correction happens silently, the user experiences the system as unstable or deceptive rather than responsive. ([Moritz et al., 2017](https://www.microsoft.com/en-us/research/publication/trust-verify-optimistic-visualizations-approximate-queries-exploring-big-data-2/))

### 4. Browser architectures that make responsiveness possible

**System evidence.** Reactive Vega treats input data, scene elements, and interaction events as first-class streaming sources in a dataflow graph, with runtime graph rewriting and performance optimizations. Riffle makes a similar move for general web apps: it keeps the entire application state in a client-side persistent relational database, expresses transformations as reactive relational queries, and synchronously updates dependent state transactionally so UI state and domain state remain consistent. The shared lesson is that visible consistency depends on explicit dependency structure and local propagation, not ad hoc chains of imperative updates. ([Satyanarayan et al., 2016](https://pubmed.ncbi.nlm.nih.gov/26390466/))

**System evidence.** Local-first software reframes responsiveness by making local state primary and cloud sync secondary. Kleppmann and colleagues define local-first around offline work, multi-device collaboration, improved privacy and preservation, and stronger user control. MyWebstrates shows what this looks like in a browser-centered system: local browser software can selectively sync over servers or peer-to-peer links while enabling offline collaborative use. This does not make consistency free, but it moves visible acknowledgement onto the client side. ([Kleppmann et al., 2019](https://www.repository.cam.ac.uk/items/525725e6-1a1d-46ce-a5a3-d0d9b1beeee1))

**System evidence and evidence gap.** Local-first availability creates integrity problems. Schiefer, Litt, and Jackson argue that some concurrent changes should merge automatically, while others should fork rather than silently converge if convergence would violate integrity constraints. Yanakieva, Bird, and Bieniusa show the same issue in collaborative spreadsheets specifically: spreadsheet operations need carefully chosen semantics if decentralized replicas are to preserve user intent. These papers are highly relevant to collaborative browser spreadsheets, but they are mostly semantic and systems contributions, not rich end-user studies. ([Schiefer et al., 2022](https://groups.csail.mit.edu/sdg/pubs/2022/Merge_What_You_Can_PaPoC_2022.pdf))

**Design interpretation.** For browser spreadsheet interfaces, the architecture implication is concrete: treat local interaction state as authoritative for immediate feedback; treat remote sync and server validation as downstream processes; encode dependencies and invariants explicitly; and surface only the states users need to reason about, such as pending, synced, conflicted, approximate, invalid, or stale. Do not leak implementation-layer concepts when a semantic state label will do. ([Litt et al., 2023](https://groups.csail.mit.edu/sdg/pubs/2023/riffle-uist-23.pdf))

### 5. Spreadsheet comprehension, hidden structure, and context switching

**Empirical finding.** Ragavan, Sarkar, and Gordon’s CHI 2021 study is the most important recent paper for understanding spreadsheet responsiveness as a cognitive issue. Participants spent only about 60% of their time directly reading and understanding spreadsheet contents; the remaining 40% went to information-seeking detours. They frequently needed to look for missing context, search labels, infer intent, and chase references. Twelve of fifteen participants wanted to go back to the original author for clarification. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

**Empirical finding.** Formula understanding in particular imposed working-memory and navigation costs. Participants often had to navigate away from the current formula to inspect referenced ranges and adjacent labels, and even tool-provided highlighting cues were limited when referenced cells or labels were off-screen or on another sheet. Ragavan and colleagues describe this as repeated reconnaissance and “rabbit holes” of formulas referencing other formulas. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

**Related systems evidence.** Other spreadsheet work supports the same diagnosis from different angles. UCheck shows that header and unit inference can make semantic inconsistency visible without requiring heavy user intervention. Hermans, Pinzger, and van Deursen’s leveled dataflow diagrams show that spreadsheet logic benefits from hierarchical, alternative representations. Cunha and colleagues’ explanation sheets argue that spreadsheet explanations can be embedded in spreadsheet form rather than requiring a wholly separate notation. ExceLint exploits rectangular regularity to identify anomalous formula regions quickly. ([UCheck, 2007](https://web.engr.oregonstate.edu/~erwig/papers/UCheck_JVLC07.pdf))

**Design interpretation.** In spreadsheet-like browser UIs, responsiveness therefore includes reducing the cost of “where is the meaning?” Navigation to references, provenance, author intent, structure, and downstream impact should be low-friction enough that users do not fall out of the task. Cross-sheet peeks, inline labels, reference previews, explanation sidecars, and anomaly lenses are not extras; they are part of keeping the interaction loop intact. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

### 6. Structure and flexibility should be layered, not polarized

**Empirical finding.** Chalhoub and Sarkar’s CHI 2022 study shows why spreadsheets remain compelling despite decades of calls for stricter structure. Users valued irregular layouts, coloring, annotations, spacing, and the freedom to place things “where my mind wants.” Those properties improved comprehension, readability, control, and communication. At the same time, users also needed structured operations, and the study found that column-based abstractions are important for features like conditional formatting, data validation, and formula authoring. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

**Empirical finding.** The same paper makes the tension explicit: unstructured data hampers charts, sorting, filtering, pivot tables, formulas, and rearrangement. In other words, flexibility supports sensemaking and collaborative expression, while structure supports systematic operations and automation. There is no research basis for treating one as categorically superior. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

**Design interpretation.** Strong browser spreadsheet design should therefore support layered structure. The grid should be allowed to stay expressive and irregular where users benefit from spatial rhetoric, annotations, grouping, and visual emphasis. On top of that, the system can offer typed columns, named regions, optional schema overlays, and structure-aware operations. The goal is not to force early normalization, but to let users introduce structure where it pays for itself. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

### 7. Large tabular workspaces need overview, semantic zoom, and history

**Foundational evidence.** Rao and Card’s Table Lens remains an important precursor because it showed how focus+context methods can preserve labels and local detail while still giving an interpretable overview of large tables. The important enduring idea is not the exact visualization form, but the combination of tabular semantics with scalable overview and distal context. ([Rao & Card, 1994](https://www.researchgate.net/publication/2541647_The_Table_Lens_Merging_Graphical_and_Symbolic_Representations_in_an_Interactive_FocusContext_Visualization_for_Tabular_Information))

**Empirical finding.** NOAH provides the strongest direct modern spreadsheet evidence. Rahman and colleagues embed dynamic hierarchical overviews alongside spreadsheets, preserve spreadsheet semantics and look-and-feel, support semantic zoom, and couple overview interaction with raw-data highlighting. In user studies, NOAH yielded fewer mistakes and faster task completion than Excel and Pivot Table on the tasks studied, and navigation history materially helped users avoid repeated work. ([Rahman et al., 2021](https://www.vldb.org/pvldb/vol14/p970-rahman.pdf))

**Design interpretation.** Large browser sheets should not rely on endless scrolling and raw range selection as the only navigation grammar. Hierarchical overviews, minimaps, semantic zoom, linked highlighting, and persistent navigation history belong in the core model. This is especially true for browser UIs, where viewport size is limited and context loss from tab switching or async operations is common. ([Rao & Card, 1994](https://www.researchgate.net/publication/2541647_The_Table_Lens_Merging_Graphical_and_Symbolic_Representations_in_an_Interactive_FocusContext_Visualization_for_Tabular_Information))

### 8. Spreadsheet surfaces can act as control surfaces for web applications

**System evidence.** Gneiss uses a spreadsheet model to build interactive web applications with two-way flow between spreadsheet state and interface state. Wildcard applies the same basic idea in reverse: augment an existing web application with a synchronized spreadsheet so users can sort, annotate, join, and customize data without traditional programming. These systems treat the table not just as a storage format but as a control surface for behavior. ([Chang & Myers, 2014](https://www.cs.cmu.edu/~shihpinc/pdf/Gneiss_UIST14_final.pdf))

**Design interpretation.** This is highly relevant for browser-based enterprise products whose users already think in tables. A spreadsheet-adjacent surface can become the place where users define workflow logic, data enrichment, filters, presentation rules, or private annotations. The literature supports this as feasible and learnable, but it also implies a requirement: the mapping between the table and the surrounding application must stay live, visible, and semantically stable, or the control surface becomes yet another hidden system layer. ([Chang & Myers, 2014](https://www.cs.cmu.edu/~shihpinc/pdf/Gneiss_UIST14_final.pdf))

## Synthesis table: evidence to browser UI and architecture

| Evidence | Browser UI implication | Architectural implication |
|---|---|---|
| Direct manipulation works best when users act on visible domain objects and see immediate effects. | Keep cells, ranges, filters, summaries, and annotations on the working surface; avoid frequent modal indirection. | Keep input, state mutation, and render feedback local; model dependencies explicitly. |
| Delay changes exploration strategy, action tempo, and confidence. | Optimize the most common micro-actions for a stable interaction beat. | Separate local acknowledgement from remote durability and heavy compute. |
| Progressive and optimistic interfaces can preserve flow when provisionality is legible. | Return previews, partial results, and approximate answers with clear status. | Support staged execution, incremental recompute, and progressive refinement. |
| Reactive dataflow and client-side persistent state reduce visible inconsistency. | Avoid UI states that drift from domain state; update what the user sees atomically. | Use a reactive graph or client-side relational state layer with transactional propagation. |
| Spreadsheet understanding is limited by hidden references and context switching. | Provide low-cost peeks, label abstraction, provenance, and alternative logic views. | Precompute dependency maps, reference previews, and explanation artifacts. |
| Freeform layout aids thinking, while regular structure enables operations. | Layer optional structure onto the grid rather than forcing early normalization. | Represent layout, schema overlays, and validation rules simultaneously. |
| Overviews, semantic zoom, and history reduce large-sheet navigation cost. | Add hierarchical overviews, minimaps, coupled highlighting, and navigation history. | Maintain cached summaries and navigation state separately from raw cell storage. |
| Local-first collaboration needs explicit integrity semantics. | Show syncing and conflict states and provide safe reconciliation workflows. | Distinguish mergeable from non-mergeable operations; fork or block unsafe convergence. |

The table synthesizes foundational direct-manipulation work, latency studies, progressive-visualization research, reactive/local-first systems, and spreadsheet-comprehension/navigation papers. The strongest direct empirical support is for latency effects and for large-sheet navigation/comprehension; some architecture rows are design interpretations grounded in systems papers rather than user-behavior experiments. ([Shneiderman, 1983](https://www.cs.umd.edu/users/ben/papers/Shneiderman1983Direct.pdf))

## Spreadsheet-specific implications

Spreadsheet-like browser interfaces should treat hidden structure as a first-class interaction problem. The grid shows values, but many of the user’s real tasks involve inferring labels, intent, provenance, dependency, and scope. Browser implementations should therefore minimize the distance between a visible value and the contextual information needed to interpret it. A formula should not force a full navigation jump when a hover, side peek, or keyboard-invoked preview would suffice. A large sheet should not force repeated scroll excursions when a coordinated overview and history model can preserve place. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

They should also preserve user-authored spatial rhetoric. Chalhoub and Sarkar show that “messiness” often encodes meaning: emphasis, grouping, workflow, collaboration, reminders, and readability. Over-normalizing these surfaces into rigid tables can destroy valuable cues. The better move is layered structure: let users keep spacing, color, and annotations while progressively attaching column semantics, named regions, units, validation rules, and dependency summaries where those afford better operations. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

Finally, spreadsheet logic should be visible in more than one representation. Alternative logic views such as leveled dataflow diagrams, explanation sheets, and region-regularity/anomaly lenses address different comprehension burdens than the grid does. Browser spreadsheets that remain “grid only” for complex structured work will continue to externalize too much of the user’s comprehension load into memory, notes, or author consultation. ([Hermans et al., 2011](https://repository.tudelft.nl/file/File_a9ae41e9-af48-4ba2-832b-6f6f033785e6))

## Synthesis table: spreadsheet findings to UI response

| Spreadsheet research finding | Design risk or opportunity | Recommended UI response |
|---|---|---|
| Users spend large amounts of time on information-seeking detours and reference chasing. | Risk of losing context, abandoning formulas, or guessing incorrectly. | Cross-sheet peek, inline labels, hover previews, breadcrumb/history, nearby documentation. |
| Flexible irregular layouts improve comprehension and control. | Risk of destroying meaning by forced normalization. | Preserve spacing, color, and annotations; add structure as overlays, not replacements. |
| Unstructured data impedes sorting, filtering, charting, and lookup. | Opportunity to add just enough structure when users need operations. | Column-based operations, named regions, lightweight schema suggestions, typed validation. |
| Spreadsheet logic is hidden and hard to summarize from the grid alone. | Risk of author dependency and comprehension error. | Dataflow diagrams, explanation sheets, formula cards, provenance notes. |
| Formula errors often appear as breaks in regional regularity. | Opportunity for low-cost anomaly detection. | Region-level pattern highlighting and anomaly badges. |
| Header and unit inference can surface invalid operations early. | Opportunity to reduce ambiguity before errors propagate. | Typed paste/import, unit warnings, semantic validation, disambiguating receipts. |
| Hierarchical overview improves speed and accuracy on navigation/exploration tasks. | Opportunity to cut scrolling and repeated range selection. | Embedded overview pane with semantic zoom and overview-to-grid coupling. |
| Spreadsheet-like surfaces can control surrounding web applications. | Opportunity to let users customize behavior without learning internals. | Live synchronized table view plus direct manipulation of app data. |

These rows are supported most directly by Ragavan et al., Chalhoub and Sarkar, UCheck, ExceLint, Hermans et al., Cunha et al., NOAH, and Wildcard/Gneiss. The weakest evidence is around typed paste/import receipts, which is a design interpretation extrapolated from validation, direct-manipulation, and comprehension work rather than a direct controlled study. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

## Browser architecture implications

A browser spreadsheet or spreadsheet-adjacent structured-data UI should treat its architecture as an interaction contract.

First, the interface should separate **local acknowledgement** from **remote durability**. A user edit, fill, sort request, annotation, or filter change should become visible locally as soon as the system can represent the intended semantic action safely. Remote sync, authorization, reconciliation, or server-side materialization can follow. This is the most direct architectural consequence of the latency, local-first, and reactive-state literature. ([Liu & Heer, 2014](https://pubmed.ncbi.nlm.nih.gov/26356926/))

Second, keep **UI state and domain state in one reactive dependency model** when feasible. Riffle’s key contribution is not only local persistence; it is avoiding a state architecture in which the database, cache, view model, and visible DOM each lag or contradict the others. For spreadsheet interfaces, that means selections, derived ranges, validation status, visible summaries, and formula results should propagate through one explicit dependency structure rather than through unrelated imperative subsystems. ([Litt et al., 2023](https://groups.csail.mit.edu/sdg/pubs/2023/riffle-uist-23.pdf))

Third, add **staged execution paths** for expensive operations. Large-range sorts, remote joins, wide aggregations, large formula recomputations, dependency analysis, and hierarchical summary generation should have hot-path behavior distinct from cold-path completion. The hot path returns preview or acknowledgement; the cold path verifies, refines, or syncs. Falcon’s prefetching/indexing and progressive resolution provide a particularly concrete pattern here. ([Moritz et al., 2019](https://www.domoritz.de/papers/2019-Falcon-CHI.pdf))

Fourth, maintain **semantic caches and indexes**, not just render caches. For spreadsheet work, the useful cached objects are often region summaries, dependency slices, label maps, structure overlays, and navigation history. Those are what let hover peeks, anomaly lenses, and overview panes appear fast enough to preserve flow. ([Moritz et al., 2019](https://www.domoritz.de/papers/2019-Falcon-CHI.pdf))

Fifth, define **integrity-aware collaboration semantics** early. Some edits can merge automatically; others should block, warn, or fork. A collaborative browser sheet that silently collapses semantically incompatible edits may feel fast but will not be trustworthy. This is especially relevant for validations, formula edits, and structure-changing operations. ([Schiefer et al., 2022](https://groups.csail.mit.edu/sdg/pubs/2022/Merge_What_You_Can_PaPoC_2022.pdf))

Sixth, make the browser’s internal state categories visible in **semantic, user-facing terms**. Pending, synced, approximate, conflicted, stale, invalid, and author-verified are meaningful. Queue depth, cache invalidation, or CRDT merge stage are not. Shielding users from internals does not mean hiding state; it means presenting the right state abstraction. ([Hutchins et al., 1985](https://worrydream.com/refs/Hutchins_1985_-_Direct_Manipulation_Interfaces.pdf))

## Design principles for browser-based spreadsheet or spreadsheet-adjacent interfaces

1. **Keep the user’s problem representation on the working surface.** Users should manipulate cells, ranges, summaries, and data-linked artifacts directly, not primarily through detached configuration panels. ([Shneiderman, 1983](https://www.cs.umd.edu/users/ben/papers/Shneiderman1983Direct.pdf))

2. **Show the system’s interpretation, not only its outcome.** After ambiguous actions such as paste, fill, sort, import, or auto-typing, tell users what the system inferred in domain terms. ([Wrangler, 2011](https://idl.uw.edu/papers/wrangler))

3. **Acknowledge actions locally before remote work completes.** Immediate local visibility preserves the user’s rhythm even when validation or sync continues in the background. ([Liu & Heer, 2014](https://pubmed.ncbi.nlm.nih.gov/26356926/))

4. **Prefer incremental, reversible operations over opaque batch transitions.** Reversible micro-steps help users maintain confidence and repair mistakes without falling out of task context. ([Shneiderman, 1983](https://www.cs.umd.edu/users/ben/papers/Shneiderman1983Direct.pdf))

5. **Make provisional states explicit and inspectable.** Preview, approximate, syncing, stale, and conflicted are different states and should be distinguishable. ([Moritz et al., 2017](https://www.microsoft.com/en-us/research/publication/trust-verify-optimistic-visualizations-approximate-queries-exploring-big-data-2/))

6. **Surface hidden dependencies with very low access cost.** Formula references, downstream impacts, units, provenance, and labels should be peekable without destructive navigation jumps. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

7. **Layer structure over flexibility.** Let users preserve freeform layout while opting into typed columns, named regions, and validation where those pay off. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

8. **Treat columns and regions as first-class abstractions, not just cell ranges.** Many operations are easier to specify and understand at that level. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

9. **Provide multiple coordinated representations of logic and data.** The grid alone is insufficient for large or complex logic. Pair it with diagrams, explanation sheets, anomaly regions, or summaries. ([Hermans et al., 2011](https://repository.tudelft.nl/file/File_a9ae41e9-af48-4ba2-832b-6f6f033785e6))

10. **Use overview/detail and semantic zoom for large workspaces.** Navigation should not depend on repeated scrolling and re-selection. ([Rao & Card, 1994](https://www.researchgate.net/publication/2541647_The_Table_Lens_Merging_Graphical_and_Symbolic_Representations_in_an_Interactive_FocusContext_Visualization_for_Tabular_Information))

11. **Preserve navigation history as a first-class artifact.** Users should be able to revisit prior semantic locations and comparisons cheaply. ([Rahman et al., 2021](https://www.vldb.org/pvldb/vol14/p970-rahman.pdf))

12. **Keep visible state consistent by using explicit reactive dependencies.** UI drift is a responsiveness failure even when rendering is fast. ([Satyanarayan et al., 2016](https://pubmed.ncbi.nlm.nih.gov/26390466/))

13. **Cache semantic summaries near the interaction loop.** Precomputed structure, overview, and dependency artifacts often matter more than raw render speed. ([Moritz et al., 2019](https://www.domoritz.de/papers/2019-Falcon-CHI.pdf))

14. **Do not silently merge semantically unsafe collaborative edits.** Fast convergence is not worth corrupted intent. ([Schiefer et al., 2022](https://groups.csail.mit.edu/sdg/pubs/2022/Merge_What_You_Can_PaPoC_2022.pdf))

## Recommended UI patterns

- **Inline interpretation receipts.** After paste, parse, fill, sort, or import, show a short inline receipt such as “parsed 14 values as dates” or “filled relative references across selected region,” with a one-click inspect or undo affordance. This is a design interpretation grounded in direct manipulation, validation, and comprehension research. ([Shneiderman, 1983](https://www.cs.umd.edu/users/ben/papers/Shneiderman1983Direct.pdf))

- **Peekable reference and provenance cards.** Hover, focus, or keyboard invocation should open a nearby preview of referenced cells, labels, units, notes, and downstream dependencies without requiring a full jump. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

- **Region-regularity and anomaly lens.** Show contiguous formula-pattern regions and highlight cells that break the local pattern. This turns structural inconsistency into visible state. ([Barowy et al., 2018](https://arxiv.org/abs/1901.11100))

- **Layered structure overlays.** Let users optionally declare a column type, named region, or schema overlay on top of a visually irregular grid, rather than forcing conversion to a fully structured table. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

- **Explanation sidecars.** For a selected region, provide alternate representations such as an explanation sheet, dataflow diagram, or formula card view that uses labels instead of bare addresses where possible. ([Cunha et al., 2018](https://web.engr.oregonstate.edu/~erwig/papers/ExplainingSpreadsheets_GPCE18.pdf))

- **Embedded hierarchical overview with semantic zoom.** Keep a small but always-available overview pane linked to the main grid, with click-to-jump, coupled highlighting, and preserved history. ([Rahman et al., 2021](https://www.vldb.org/pvldb/vol14/p970-rahman.pdf))

- **Local-first pending/sync/conflict markers.** Show whether a change is local-only, synced, conflicted, or blocked by integrity rules, and make reconciliation explicit. ([Kleppmann et al., 2019](https://www.repository.cam.ac.uk/items/525725e6-1a1d-46ce-a5a3-d0d9b1beeee1))

- **Live synchronized table control surfaces.** For spreadsheet-adjacent web apps, provide a tabular view that remains synchronized with the application and can drive customization or automation. ([Chang & Myers, 2014](https://www.cs.cmu.edu/~shihpinc/pdf/Gneiss_UIST14_final.pdf))

## Anti-patterns and tradeoffs

### Anti-patterns

- **Blocking spinners after routine actions.** If common edits, filters, or navigation actions wait behind full remote completion, users change strategy and pace in ways that degrade cognition, not just speed. ([Liu & Heer, 2014](https://pubmed.ncbi.nlm.nih.gov/26356926/))

- **Silent reinterpretation.** Parsing, coercion, sorting, or repair without visible explanation leaves users unsure what the system understood. ([Wrangler, 2011](https://idl.uw.edu/papers/wrangler))

- **Grid-only logic for complex sheets.** For large, formula-heavy, or hierarchical work, a pure cell grid hides too much structure. ([Hermans et al., 2011](https://repository.tudelft.nl/file/File_a9ae41e9-af48-4ba2-832b-6f6f033785e6))

- **Forced normalization that destroys layout semantics.** Eliminating spacing, annotations, and freeform layout can improve machine operation at the expense of human comprehension. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

- **Server-authoritative hot paths for every micro-action.** This exposes users to latency variability and couples cognitive continuity to network conditions. ([Kleppmann et al., 2019](https://www.repository.cam.ac.uk/items/525725e6-1a1d-46ce-a5a3-d0d9b1beeee1))

- **Silent collaborative convergence under integrity-sensitive rules.** Auto-merge is not always the right user-facing behavior. ([Schiefer et al., 2022](https://groups.csail.mit.edu/sdg/pubs/2022/Merge_What_You_Can_PaPoC_2022.pdf))

### Tradeoffs

- **More structure vs. more expressive layout.** Structure improves operations and validation; expressive layout improves sensemaking and communication. The literature supports layered coexistence, not one-way replacement. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

- **More local autonomy vs. more consistency complexity.** Local-first interaction preserves flow, but it pushes more responsibility onto conflict semantics, integrity handling, and browser storage. ([Kleppmann et al., 2019](https://www.repository.cam.ac.uk/items/525725e6-1a1d-46ce-a5a3-d0d9b1beeee1))

- **More progressive feedback vs. more trust work.** Previews and approximations can preserve flow, but only if the verification path is clear and cheap. ([Moritz et al., 2017](https://www.microsoft.com/en-us/research/publication/trust-verify-optimistic-visualizations-approximate-queries-exploring-big-data-2/))

- **More explanatory overlays vs. more workspace clutter.** Alternative views reduce hidden-structure costs but can overwhelm users if always-on. This is a product tradeoff suggested by the literature rather than directly settled by it. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

## Open research questions and product hypotheses

1. **Operation-specific latency budgets for browser spreadsheets remain under-measured.** The literature shows that delay matters, but it does not yet tell us the right thresholds for typing, range drag, fill-handle use, filter apply, cross-sheet peek, or recalculation in browser-native spreadsheet work. ([Liu & Heer, 2014](https://pubmed.ncbi.nlm.nih.gov/26356926/))

2. **Visible interpretation receipts may reduce spreadsheet guesswork.** Product hypothesis: brief, inspectable feedback after paste, fill, parse, or auto-structure operations will lower rework and hesitation compared with silent inference. Direct evidence is currently indirect. ([Wrangler, 2011](https://idl.uw.edu/papers/wrangler))

3. **Local-first spreadsheets with explicit provisional state may preserve flow better than server-authoritative models under ordinary network degradation.** The architecture literature strongly suggests this, but direct comparative user studies for spreadsheet tasks are scarce. ([Kleppmann et al., 2019](https://www.repository.cam.ac.uk/items/525725e6-1a1d-46ce-a5a3-d0d9b1beeee1))

4. **Explanation layers may reduce the 40% information-seeking burden seen in spreadsheet comprehension.** A concrete hypothesis is that cross-sheet previews plus explanation sheets/dataflow sidecars will reduce context-switch cost and author dependency. ([Ragavan et al., 2021](https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf))

5. **Layered structure overlays may outperform both rigid tables and fully freeform grids.** The literature supports the underlying tension, but not yet the optimal browser UI for negotiating it. ([Chalhoub & Sarkar, 2022](https://www.cs.ox.ac.uk/files/13563/chalhoub_2022_data_structuring.pdf))

6. **Collaborative spreadsheet conflict semantics need better user-facing representations.** We have semantics work on CRDT-based spreadsheets and integrity-aware local-first systems, but less evidence on what ordinary spreadsheet users can understand and manage reliably. ([Schiefer et al., 2022](https://groups.csail.mit.edu/sdg/pubs/2022/Merge_What_You_Can_PaPoC_2022.pdf))

7. **Hierarchical overviews plus history may generalize beyond exploration to editing and auditing.** NOAH shows strong navigation benefits, but there is room to test whether similar mechanisms help formula debugging, review, and data-entry workflows in browser products. ([Rahman et al., 2021](https://www.vldb.org/pvldb/vol14/p970-rahman.pdf))

## Annotated bibliography

- **Shneiderman, B. (1983). _Direct Manipulation: A Step Beyond Programming Languages_. Computer, 16(8), 57-69. DOI: 10.1109/MC.1983.1654471.** Foundational statement of continuous representation, physical action, and immediate visible effect; still the clearest conceptual anchor for “responsive” as interactional directness. Stable URL: https://www.cs.umd.edu/users/ben/papers/Shneiderman1983Direct.pdf

- **Hutchins, E. L., Hollan, J. D., & Norman, D. A. (1985). _Direct Manipulation Interfaces_. Human-Computer Interaction, 1(4), 311-338. DOI: 10.1207/s15327051hci0104_2.** Provides the cognitive framing of directness via semantic and articulatory distance, which remains useful when evaluating whether a browser UI exposes domain semantics or system machinery. Stable URL: https://worrydream.com/refs/Hutchins_1985_-_Direct_Manipulation_Interfaces.pdf

- **Endert, A. (2014). _Semantic Interaction for Visual Analytics: Toward Coupling Cognition and Computation_. IEEE Computer Graphics and Applications, 34(4), 8-15. DOI: 10.1109/MCG.2014.73.** Useful bridge from classic direct manipulation to modern analytic interfaces in which domain-surface interaction steers hidden models. Stable URL: https://pubmed.ncbi.nlm.nih.gov/25051565/

- **Liu, Z., & Heer, J. (2014). _The Effects of Interactive Latency on Exploratory Visual Analysis_. IEEE Transactions on Visualization and Computer Graphics, 20(12), 2122-2131. DOI: 10.1109/TVCG.2014.2346452.** The strongest empirical evidence in this report for how delay changes strategy, activity, and knowledge discovery. Stable URL: https://pubmed.ncbi.nlm.nih.gov/26356926/

- **Battle, L., Crouser, R. J., Nakeshimana, A., Montoly, A., Chang, R., & Stonebraker, M. (2020). _The Role of Latency and Task Complexity in Predicting Visual Search Behavior_. IEEE Transactions on Visualization and Computer Graphics, 26(1), 1246-1255. DOI: 10.1109/TVCG.2019.2934556.** Important nuance paper showing that latency effects are mediated by task demands, not captured by one universal threshold. Stable URL: https://pubmed.ncbi.nlm.nih.gov/31442990/

- **Friston, S., Karlström, P., & Steed, A. (2016). _The Effects of Low Latency on Pointing and Steering Tasks_. IEEE Transactions on Visualization and Computer Graphics, 22(5), 1605-1615. DOI: 10.1109/TVCG.2015.2446467.** Adjacent motor-control evidence for mouse/trackpad interactions; useful but not directly about spreadsheet editing. Stable URL: https://pubmed.ncbi.nlm.nih.gov/27045915/

- **Bogon, J., Hößl, S., Wolff, C., Henze, N., & Halbhuber, D. (2025). _Cognitive Integration of Delays: Anticipated System Delays Slow Down User Actions_. Proceedings of the 2025 CHI Conference on Human Factors in Computing Systems. DOI: 10.1145/3706598.3713475.** Important recent result showing that users adapt their action tempo based on expected delay, not only observed delay. Stable URL: https://dl.acm.org/doi/10.1145/3706598.3713475

- **Fisher, D., Popov, I. O., Drucker, S. M., & schraefel, m. c. (2012). _Trust Me, I’m Partially Right: Incremental Visualization Lets Analysts Explore Large Datasets Faster_. Proceedings of CHI 2012. DOI: 10.1145/2207676.2208294.** Foundational evidence for progressive results as a way to preserve exploratory tempo. Stable URL: https://danyelfisher.info/ux-of-big-data

- **Moritz, D., Fisher, D., Ding, B., & Wang, C. (2017). _Trust, but Verify: Optimistic Visualizations of Approximate Queries for Exploring Big Data_. Proceedings of CHI 2017. DOI: 10.1145/3025453.3025456.** Stronger treatment of trust conditions for provisional feedback than the incremental-visualization literature alone. Stable URL: https://www.microsoft.com/en-us/research/publication/trust-verify-optimistic-visualizations-approximate-queries-exploring-big-data-2/

- **Moritz, D., Howe, B., & Heer, J. (2019). _Falcon: Balancing Interactive Latency and Resolution Sensitivity for Scalable Linked Visualizations_. Proceedings of CHI 2019. DOI: 10.1145/3290605.3300924.** Valuable because it translates responsiveness principles into concrete browser-friendly prefetching, indexing, and progressive-resolution mechanisms. Stable URL: https://www.domoritz.de/papers/2019-Falcon-CHI.pdf

- **Satyanarayan, A., Russell, R., Hoffswell, J., & Heer, J. (2016). _Reactive Vega: A Streaming Dataflow Architecture for Declarative Interactive Visualization_. IEEE Transactions on Visualization and Computer Graphics, 22(1), 659-668. DOI: 10.1109/TVCG.2015.2467091.** Key reference for treating interaction events, data, and visual state as one reactive graph. Stable URL: https://pubmed.ncbi.nlm.nih.gov/26390466/

- **Kleppmann, M., Wiggins, A., van Hardenberg, P., & McGranaghan, M. (2019). _Local-first software: you own your data, in spite of the cloud_. Onward! 2019. DOI: 10.1145/3359591.3359737.** The clearest articulation of why local acknowledgement, offline operation, and user control are design issues, not merely storage issues. Stable URL: https://www.repository.cam.ac.uk/items/525725e6-1a1d-46ce-a5a3-d0d9b1beeee1

- **Litt, G., Schiefer, N., Schickling, J., & Jackson, D. (2023). _Riffle: Reactive Relational State for Local-First Applications_. UIST 2023. DOI: 10.1145/3586183.3606801.** Particularly relevant for browser spreadsheets because it unifies client persistence, reactive queries, and consistent visible state. Stable URL: https://groups.csail.mit.edu/sdg/pubs/2023/riffle-uist-23.pdf

- **Klokmose, C. N., Eagan, J. R., & van Hardenberg, P. (2024). _MyWebstrates: Webstrates as Local-first Software_. UIST 2024. DOI: 10.1145/3654777.3676445.** Strong browser-specific example of local software that selectively syncs, rather than treating the browser as a thin client. Stable URL: https://cs.au.dk/~clemens/files/MyWebstrates-UIST2024.pdf

- **Schiefer, N., Litt, G., & Jackson, D. (2022). _Merge What You Can, Fork What You Can’t: Managing Data Integrity in Local-First Software_. PaPoC 2022. DOI: 10.1145/3517209.3524041.** Essential caution against equating responsiveness with automatic convergence when integrity constraints matter. Stable URL: https://groups.csail.mit.edu/sdg/pubs/2022/Merge_What_You_Can_PaPoC_2022.pdf

- **Yanakieva, E., Bird, P., & Bieniusa, A. (2023). _A Study of Semantics for CRDT-based Collaborative Spreadsheets_. PaPoC 2023. DOI: 10.1145/3578358.3591324.** One of the few sources directly addressing collaborative spreadsheet semantics in decentralized settings. Stable URL: https://dl.acm.org/doi/10.1145/3578358.3591324

- **Kandel, S., Paepcke, A., Hellerstein, J. M., & Heer, J. (2011). _Wrangler: Interactive Visual Specification of Data Transformation Scripts_. Proceedings of CHI 2011. DOI: 10.1145/1978942.1979444.** Canonical example of direct manipulation plus visible inferred interpretation in tabular work. Stable URL: https://idl.uw.edu/papers/wrangler

- **Abraham, R., & Erwig, M. (2007). _UCheck: A Spreadsheet Type Checker for End Users_. Journal of Visual Languages & Computing, 18(1), 71-95. DOI: 10.1016/j.jvlc.2006.06.001.** Still highly relevant for making invalid operations and semantic inconsistency visible to end users. Stable URL: https://web.engr.oregonstate.edu/~erwig/papers/UCheck_JVLC07.pdf

- **Hermans, F., Pinzger, M., & van Deursen, A. (2011). _Supporting Professional Spreadsheet Users by Generating Leveled Dataflow Diagrams_. ICSE 2011. DOI: 10.1145/1985793.1985855.** Important evidence that spreadsheet logic often needs hierarchical, non-grid representations for comprehension. Stable URL: https://www.researchgate.net/publication/221553716_Supporting_professional_spreadsheet_users_by_generating_leveled_dataflow_diagrams

- **Barowy, D. W., Berger, E. D., & Zorn, B. G. (2018). _ExceLint: Automatically Finding Spreadsheet Formula Errors_. Proceedings of the ACM on Programming Languages, 2(OOPSLA). DOI: 10.1145/3276518.** Valuable because it operationalizes spreadsheet regularity as a visible anomaly-detection aid. Stable URL: https://arxiv.org/abs/1901.11100

- **Ragavan, S. S., Sarkar, A., & Gordon, A. D. (2021). _Spreadsheet Comprehension: Guesswork, Giving Up and Going Back to the Author_. CHI 2021. DOI: 10.1145/3411764.3445634.** The strongest recent direct evidence on hidden-structure costs, context-switching, and author dependency in spreadsheet reading. Stable URL: https://advait.org/files/ragavan_2021_spreadsheet_comprehension.pdf

- **Chalhoub, G., & Sarkar, A. (2022). _“It’s Freedom to Put Things Where My Mind Wants”: Understanding and Improving the User Experience of Structuring Data in Spreadsheets_. CHI 2022. DOI: 10.1145/3491102.3501833.** Best recent paper on the tension between freeform spatial expression and structured table operations. Stable URL: https://www.researchgate.net/publication/360263426_It%27s_Freedom_to_Put_Things_Where_My_Mind_Wants_Understanding_and_Improving_the_User_Experience_of_Structuring_Data_in_Spreadsheets

- **Cunha, J., Dan, M., Erwig, M., Fedorin, D., & Grejuc, A. (2018). _Explaining Spreadsheets with Spreadsheets_. GPCE 2018. DOI: 10.1145/3278122.3278136.** Useful for thinking about explanation layers that stay inside spreadsheet idioms rather than requiring external notations. Stable URL: https://web.engr.oregonstate.edu/~erwig/papers/ExplainingSpreadsheets_GPCE18.pdf

- **Rahman, S., Bendre, M., Liu, Y., Zhu, S., Su, Z., Karahalios, K., & Parameswaran, A. G. (2021). _NOAH: Interactive Spreadsheet Exploration with Dynamic Hierarchical Overviews_. Proceedings of the VLDB Endowment, 14(6), 970-983. DOI: 10.14778/3447689.3447701.** The strongest direct evidence in this report for overview, semantic zoom, and history in spreadsheet navigation. Stable URL: https://www.vldb.org/pvldb/vol14/p970-rahman.pdf

- **Rao, R., & Card, S. K. (1994). _The Table Lens: Merging Graphical and Symbolic Representations in an Interactive Focus + Context Visualization for Tabular Information_. Proceedings of CHI 1994. DOI: 10.1145/191666.191776.** Older but still influential foundation for focus+context navigation in large tables. Stable URL: https://www.researchgate.net/publication/2541647_The_Table_Lens_Merging_Graphical_and_Symbolic_Representations_in_an_Interactive_FocusContext_Visualization_for_Tabular_Information

- **Chang, K. S.-P., & Myers, B. A. (2014). _Creating Interactive Web Data Applications with Spreadsheets_. UIST 2014. DOI: 10.1145/2642918.2647371.** Directly relevant precursor for spreadsheet-driven web application logic and live two-way bindings. Stable URL: https://www.cs.cmu.edu/~shihpinc/pdf/Gneiss_UIST14_final.pdf

- **Litt, G., & Jackson, D. (2020). _Wildcard: Spreadsheet-Driven Customization of Web Applications_. <Programming’20> Companion. DOI: 10.1145/3397537.3397541.** Most direct evidence here for spreadsheet-like control surfaces layered onto existing web applications. Stable URL: https://www.geoffreylitt.com/wildcard/salon2020/paper.pdf
