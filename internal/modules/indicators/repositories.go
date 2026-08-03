package indicators

// Repositories contain only fixed, Indicator-owned SQL and always operate on
// a transaction supplied by the workflow façade. Transaction boundaries,
// revisions, projections, and idempotency remain Store responsibilities.
type sourceRepository struct{}

type observationRepository struct{}

type lifecycleRepository struct{}
