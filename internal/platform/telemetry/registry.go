package telemetry

const (
	EvidenceCleanupOperationsMetricName     = "cartulary.evidence.cleanup.operations"
	EvidenceCleanupSweepDurationMetricName  = "cartulary.evidence.cleanup.sweep.duration"
	EvidenceCleanupOverdueMetricName        = "cartulary.evidence.cleanup.overdue"
	EvidenceCleanupOldestAgeMetricName      = "cartulary.evidence.cleanup.oldest_eligible.age"
	JobsQueuedMetricName                    = "cartulary.jobs.queued"
	JobsQueueWaitDurationMetricName         = "cartulary.jobs.queue_wait.duration"
	NetworkFlowGraphPhaseDurationMetricName = "cartulary.network_flow.graph.phase.duration"
	NetworkFlowGraphRowsMetricName          = "cartulary.network_flow.graph.contributing_rows"
	NetworkFlowGraphObjectsMetricName       = "cartulary.network_flow.graph.result.objects"
	NetworkFlowCleanupOperationsMetricName  = "cartulary.network_flow.cleanup.operations"
	NetworkFlowCleanupDurationMetricName    = "cartulary.network_flow.cleanup.sweep.duration"
	NetworkFlowCleanupDeletedMetricName     = "cartulary.network_flow.cleanup.deleted"
	NetworkFlowCleanupEligibleMetricName    = "cartulary.network_flow.cleanup.eligible"
	NetworkFlowCleanupOldestAgeMetricName   = "cartulary.network_flow.cleanup.oldest_eligible_result.age"
)

type SpanRegistryRow struct {
	Family              string
	Name                string
	Kind                string
	Scope               Scope
	LifecycleBoundary   string
	StatusRule          string
	ParentRule          string
	LinkRule            string
	RequiredAttributes  []string
	OptionalAttributes  []string
	ForbiddenAttributes []string
}

type MetricRegistryRow struct {
	Name               string
	InstrumentKind     string
	Unit               string
	Description        string
	Aggregation        string
	Temporality        string
	Buckets            []float64
	AllowedAttributes  []string
	OptionalAttributes []string
	OverflowBehavior   string
}

func SpanRegistry() []SpanRegistryRow {
	return []SpanRegistryRow{
		{
			Family:             "http_server",
			Name:               "<HTTP_METHOD> <route_template>",
			Kind:               "server",
			Scope:              ScopeHTTPAPI,
			LifecycleBoundary:  "single_request",
			StatusRule:         "error_when_status_code_gte_400",
			ParentRule:         "new_root_remote_context_ignored",
			LinkRule:           "none",
			RequiredAttributes: []string{"http.request.method", "http.route", "http.response.status_code", "cartulary.route_family", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.error_code"},
			ForbiddenAttributes: []string{
				"url.full", "url.path", "url.query", "http.request.header.*", "http.response.header.*",
				"user.id", "session.id", "incident_id", "request.body", "response.body",
			},
		},
		{
			Family:             "workbook_query",
			Name:               "cartulary.workbook.query",
			Kind:               "internal",
			Scope:              ScopeWorkbook,
			LifecycleBoundary:  "single_query",
			StatusRule:         "error_when_result_not_success",
			ParentRule:         "current_context",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.view_schema_id", "cartulary.operation", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.error_code"},
			ForbiddenAttributes: []string{
				"saved_view_id", "filters", "search_text", "row.values", "projection_table_name",
			},
		},
		{
			Family:             "workbook_mutation",
			Name:               "cartulary.workbook.mutation",
			Kind:               "internal",
			Scope:              ScopeWorkbook,
			LifecycleBoundary:  "single_mutation",
			StatusRule:         "error_when_result_not_success",
			ParentRule:         "current_context",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.view_schema_id", "cartulary.record_type", "cartulary.operation", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.error_code"},
			ForbiddenAttributes: []string{
				"record_id", "row_version", "client_txn_id", "field_values",
			},
		},
		{
			Family:             "workbook_projection",
			Name:               "cartulary.workbook.projection",
			Kind:               "internal",
			Scope:              ScopeWorkbook,
			LifecycleBoundary:  "projection_rebuild",
			StatusRule:         "error_when_result_not_success",
			ParentRule:         "current_context",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.view_schema_id", "cartulary.operation", "cartulary.result"},
			ForbiddenAttributes: []string{
				"projection_table_name", "db.statement", "record_id",
			},
		},
		{
			Family:             "websocket_lifecycle",
			Name:               "cartulary.collaboration.websocket",
			Kind:               "internal",
			Scope:              ScopeCollaboration,
			LifecycleBoundary:  "accepted_socket_lifecycle",
			StatusRule:         "error_when_result_not_success",
			ParentRule:         "new_root_remote_context_ignored",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.operation", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.error_code"},
			ForbiddenAttributes: []string{
				"connection_id", "user_id", "incident_id", "payload",
			},
		},
		{
			Family:             "websocket_event_send",
			Name:               "cartulary.collaboration.event_send",
			Kind:               "internal",
			Scope:              ScopeCollaboration,
			LifecycleBoundary:  "single_event_send",
			StatusRule:         "error_when_result_failed_rejected_or_dropped",
			ParentRule:         "new_root",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.websocket.event_type", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.drop_reason"},
			ForbiddenAttributes: []string{
				"payload", "record_id", "user_id", "connection_id",
			},
		},
		{
			Family:             "job_enqueue",
			Name:               "cartulary.jobs.enqueue",
			Kind:               "internal",
			Scope:              ScopeJobs,
			LifecycleBoundary:  "single_enqueue",
			StatusRule:         "error_when_enqueue_error",
			ParentRule:         "current_context",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.job_kind", "cartulary.operation", "cartulary.result"},
			ForbiddenAttributes: []string{
				"job_id", "incident_id", "request.body",
			},
		},
		{
			Family:             "job_run",
			Name:               "cartulary.jobs.run",
			Kind:               "internal",
			Scope:              ScopeJobs,
			LifecycleBoundary:  "single_execution_attempt",
			StatusRule:         "error_when_run_error",
			ParentRule:         "current_context",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.job_kind", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.job_terminal_status", "cartulary.operation"},
			ForbiddenAttributes: []string{
				"job_id", "artifact_path", "incident_id", "evidence_id",
			},
		},
		{
			Family:             "network_flow_graph_phase",
			Name:               "cartulary.network_flow.graph.phase",
			Kind:               "internal",
			Scope:              ScopeNetworkFlow,
			LifecycleBoundary:  "single_graph_materialization_phase",
			StatusRule:         "error_when_result_not_success",
			ParentRule:         "current_context",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.operation", "cartulary.phase", "cartulary.graph_mode", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.error_class"},
			ForbiddenAttributes: []string{
				"incident_id", "table_id", "graph_view_id", "job_id", "projection_result_id", "selector", "digest",
				"endpoint", "label", "row", "property", "db.statement", "raw_error",
			},
		},
		{
			Family:             "network_flow_cleanup",
			Name:               "cartulary.network_flow.cleanup",
			Kind:               "internal",
			Scope:              ScopeNetworkFlow,
			LifecycleBoundary:  "single_cleanup_sweep",
			StatusRule:         "error_when_result_not_success",
			ParentRule:         "new_root",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.operation", "cartulary.phase", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.error_class"},
			ForbiddenAttributes: []string{
				"source_owner_id", "incident_id", "graph_view_id", "projection_result_id", "lease_id", "digest",
				"db.statement", "raw_error",
			},
		},
		{
			Family:             "postgres_dependency",
			Name:               "cartulary.postgres.operation",
			Kind:               "client",
			Scope:              ScopePostgres,
			LifecycleBoundary:  "single_db_operation",
			StatusRule:         "error_when_db_error",
			ParentRule:         "current_context",
			LinkRule:           "none",
			RequiredAttributes: []string{"db.system.name", "cartulary.operation", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.error_class"},
			ForbiddenAttributes: []string{
				"db.statement", "db.query.text", "db.query.summary", "db.namespace", "db.collection.name", "server.address", "server.port",
			},
		},
		{
			Family:             "objectstore_dependency",
			Name:               "cartulary.objectstore.operation",
			Kind:               "client",
			Scope:              ScopeObjectStore,
			LifecycleBoundary:  "single_objectstore_operation",
			StatusRule:         "error_when_objectstore_error",
			ParentRule:         "current_context",
			LinkRule:           "none",
			RequiredAttributes: []string{"cartulary.operation", "cartulary.result"},
			OptionalAttributes: []string{"cartulary.error_class"},
			ForbiddenAttributes: []string{
				"bucket", "key", "filename", "object_hash", "upload_id", "copy_source", "storage_ref", "aws.s3.*",
			},
		},
	}
}

func MetricRegistry() []MetricRegistryRow {
	durationBuckets := []float64{0.005, 0.010, 0.025, 0.050, 0.100, 0.250, 0.500, 1.000, 2.500, 5.000, 10.000, 30.000}
	byteBuckets := []float64{1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456}
	rowBuckets := []float64{0, 1, 5, 10, 25, 50, 100, 250, 500}
	return []MetricRegistryRow{
		{Name: "cartulary.http.server.request.duration", InstrumentKind: "Histogram", Unit: "s", Description: "HTTP server request duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"http.request.method", "http.response.status_code", "http.route", "cartulary.route_family", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_code"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.workbook.query.duration", InstrumentKind: "Histogram", Unit: "s", Description: "Workbook query duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"cartulary.view_schema_id", "cartulary.operation", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_code"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.workbook.mutation.duration", InstrumentKind: "Histogram", Unit: "s", Description: "Workbook mutation duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"cartulary.view_schema_id", "cartulary.record_type", "cartulary.operation", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_code"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.workbook.rows.returned", InstrumentKind: "Histogram", Unit: "{row}", Description: "Serialized rows returned by one successful workbook view-query response.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: rowBuckets, AllowedAttributes: []string{"cartulary.view_schema_id", "cartulary.result"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.collaboration.connections.active", InstrumentKind: "ObservableGauge", Unit: "{connection}", Description: "Active accepted WebSocket connections.", Aggregation: "last_value", Temporality: "cumulative_equivalent_current_observation", AllowedAttributes: nil, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.collaboration.events.sent", InstrumentKind: "Counter", Unit: "{event}", Description: "WebSocket events sent.", Aggregation: "monotonic_sum", Temporality: "cumulative", AllowedAttributes: []string{"cartulary.websocket.event_type", "cartulary.result"}, OptionalAttributes: []string{"cartulary.drop_reason"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: EvidenceCleanupOperationsMetricName, InstrumentKind: "Counter", Unit: "{operation}", Description: "Evidence cleanup sweep operations by closed result.", Aggregation: "monotonic_sum", Temporality: "cumulative", AllowedAttributes: []string{"cartulary.operation", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_class"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: EvidenceCleanupSweepDurationMetricName, InstrumentKind: "Histogram", Unit: "s", Description: "Evidence cleanup sweep duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"cartulary.operation", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_class"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: EvidenceCleanupOverdueMetricName, InstrumentKind: "ObservableGauge", Unit: "{blob}", Description: "Current failed unattached Evidence blobs beyond the deletion deadline.", Aggregation: "last_value", Temporality: "cumulative_equivalent_current_observation", AllowedAttributes: nil, OverflowBehavior: "drop_metric_overflow"},
		{Name: EvidenceCleanupOldestAgeMetricName, InstrumentKind: "ObservableGauge", Unit: "s", Description: "Age of the oldest currently eligible failed unattached Evidence blob.", Aggregation: "last_value", Temporality: "cumulative_equivalent_current_observation", AllowedAttributes: nil, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.jobs.active", InstrumentKind: "ObservableGauge", Unit: "{job}", Description: "Active background jobs by kind.", Aggregation: "last_value", Temporality: "cumulative_equivalent_current_observation", AllowedAttributes: []string{"cartulary.job_kind"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: JobsQueuedMetricName, InstrumentKind: "ObservableGauge", Unit: "{job}", Description: "Queued background jobs, including retry-delayed jobs, by kind.", Aggregation: "last_value", Temporality: "cumulative_equivalent_current_observation", AllowedAttributes: []string{"cartulary.job_kind"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: JobsQueueWaitDurationMetricName, InstrumentKind: "Histogram", Unit: "s", Description: "Eligibility-to-successful-claim duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"cartulary.job_kind"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.jobs.duration", InstrumentKind: "Histogram", Unit: "s", Description: "Background job runtime duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"cartulary.job_kind", "cartulary.job_terminal_status", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_code"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.jobs.attempts", InstrumentKind: "Counter", Unit: "{attempt}", Description: "Completed handler attempts by kind and closed outcome.", Aggregation: "monotonic_sum", Temporality: "cumulative", AllowedAttributes: []string{"cartulary.job_kind", "cartulary.result"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.jobs.expired", InstrumentKind: "Counter", Unit: "{job}", Description: "Job resources compacted after logical expiry.", Aggregation: "monotonic_sum", Temporality: "cumulative", AllowedAttributes: []string{"cartulary.job_kind"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.jobs.lease_renewal.failures", InstrumentKind: "Counter", Unit: "{failure}", Description: "Failed lease renewals by kind and closed outcome.", Aggregation: "monotonic_sum", Temporality: "cumulative", AllowedAttributes: []string{"cartulary.job_kind", "cartulary.result"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: NetworkFlowGraphPhaseDurationMetricName, InstrumentKind: "Histogram", Unit: "s", Description: "Network Flow graph materialization phase duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"cartulary.phase", "cartulary.graph_mode", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_class"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: NetworkFlowGraphRowsMetricName, InstrumentKind: "Histogram", Unit: "{row}", Description: "Accepted source rows contributing to one graph result.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: rowBuckets, AllowedAttributes: []string{"cartulary.graph_mode", "cartulary.result"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: NetworkFlowGraphObjectsMetricName, InstrumentKind: "Histogram", Unit: "{object}", Description: "Network Flow graph result objects by closed kind.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: rowBuckets, AllowedAttributes: []string{"cartulary.graph_mode", "cartulary.graph_object_kind", "cartulary.result"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: NetworkFlowCleanupOperationsMetricName, InstrumentKind: "Counter", Unit: "{operation}", Description: "Network Flow cleanup sweep operations by closed result.", Aggregation: "monotonic_sum", Temporality: "cumulative", AllowedAttributes: []string{"cartulary.operation", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_class"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: NetworkFlowCleanupDurationMetricName, InstrumentKind: "Histogram", Unit: "s", Description: "Network Flow cleanup sweep duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"cartulary.operation", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_class"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: NetworkFlowCleanupDeletedMetricName, InstrumentKind: "Counter", Unit: "{object}", Description: "Expired leases and eligible projection results deleted by Network Flow cleanup.", Aggregation: "monotonic_sum", Temporality: "cumulative", AllowedAttributes: []string{"cartulary.graph_object_kind", "cartulary.result"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: NetworkFlowCleanupEligibleMetricName, InstrumentKind: "ObservableGauge", Unit: "{result}", Description: "Current eligible projection-result backlog.", Aggregation: "last_value", Temporality: "cumulative_equivalent_current_observation", AllowedAttributes: nil, OverflowBehavior: "drop_metric_overflow"},
		{Name: NetworkFlowCleanupOldestAgeMetricName, InstrumentKind: "ObservableGauge", Unit: "s", Description: "Age from published_at of the oldest currently eligible projection result.", Aggregation: "last_value", Temporality: "cumulative_equivalent_current_observation", AllowedAttributes: nil, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.postgres.operation.duration", InstrumentKind: "Histogram", Unit: "s", Description: "Postgres dependency operation duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"db.system.name", "cartulary.operation", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_class"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.objectstore.operation.duration", InstrumentKind: "Histogram", Unit: "s", Description: "Object-store dependency operation duration.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: durationBuckets, AllowedAttributes: []string{"cartulary.operation", "cartulary.result"}, OptionalAttributes: []string{"cartulary.error_class"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: "cartulary.objectstore.transfer.bytes", InstrumentKind: "Histogram", Unit: "By", Description: "Safe object-store transfer size.", Aggregation: "explicit_bucket_histogram", Temporality: "cumulative", Buckets: byteBuckets, AllowedAttributes: []string{"cartulary.operation", "cartulary.result"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: TelemetryExportFailureMetricName, InstrumentKind: "Counter", Unit: "{failure}", Description: "Telemetry export failures.", Aggregation: "monotonic_sum", Temporality: "cumulative", AllowedAttributes: []string{"cartulary.signal_kind", "cartulary.telemetry.exporter_kind", "cartulary.error_class"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: TelemetryItemDroppedMetricName, InstrumentKind: "Counter", Unit: "{item}", Description: "Telemetry items dropped before or during export.", Aggregation: "monotonic_sum", Temporality: "cumulative", AllowedAttributes: []string{"cartulary.signal_kind", "cartulary.drop_reason"}, OverflowBehavior: "drop_metric_overflow"},
		{Name: TelemetryQueueDepthMetricName, InstrumentKind: "ObservableGauge", Unit: "{item}", Description: "Current processor queue depth by signal.", Aggregation: "last_value", Temporality: "cumulative_equivalent_current_observation", AllowedAttributes: []string{"cartulary.signal_kind"}, OverflowBehavior: "drop_metric_overflow"},
	}
}
