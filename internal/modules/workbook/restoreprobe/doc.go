// Package restoreprobe owns Workbook's restore-verification query registry and
// executor. Timeline declares registrations, application assembly supplies the
// complete registration set and a Projections-backed query, and Recovery
// selects the restored incident and invokes the resulting executor.
//
// The package does not select incidents, rebuild projections, define source
// semantics, publish overall restore readiness, or participate in Graph
// restoration. Graph Projection remains a separate Recovery participant.
package restoreprobe
