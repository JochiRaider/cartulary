package telemetry

import (
	"fmt"
	"math"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"go.opentelemetry.io/otel/attribute"
)

type OTLPHTTPURLs struct {
	Traces  string
	Metrics string
	Logs    string
}

func BuildOTLPHTTPURLs(endpoint string) (OTLPHTTPURLs, error) {
	parsed, err := parseEndpoint(endpoint)
	if err != nil {
		return OTLPHTTPURLs{}, err
	}
	if err := validateOTLPHTTPPath(parsed); err != nil {
		return OTLPHTTPURLs{}, err
	}
	prefix := strings.TrimSuffix(parsed.EscapedPath(), "/")
	if prefix == "/" {
		prefix = ""
	}
	base := strings.ToLower(parsed.Scheme) + "://" + canonicalHostPort(parsed)
	return OTLPHTTPURLs{
		Traces:  base + prefix + "/v1/traces",
		Metrics: base + prefix + "/v1/metrics",
		Logs:    base + prefix + "/v1/logs",
	}, nil
}

type OTLPGRPCTarget struct {
	Target string
	Secure bool
}

func BuildOTLPGRPCTarget(endpoint string) (OTLPGRPCTarget, error) {
	parsed, err := parseEndpoint(endpoint)
	if err != nil {
		return OTLPGRPCTarget{}, err
	}
	if parsed.EscapedPath() != "" && parsed.EscapedPath() != "/" {
		return OTLPGRPCTarget{}, fmt.Errorf("OTLP/gRPC endpoint must not include a non-root path")
	}
	return OTLPGRPCTarget{
		Target: canonicalHostPort(parsed),
		Secure: strings.ToLower(parsed.Scheme) == "https",
	}, nil
}

func ExporterUserAgent(serviceVersion string, exporterVersion string) (string, error) {
	if resolveServiceVersion(serviceVersion) != serviceVersion {
		return "", fmt.Errorf("service version must be resolved before exporter User-Agent construction")
	}
	if !safeExporterVersion(exporterVersion) {
		return "", fmt.Errorf("exporter version evidence is required before exporter User-Agent construction")
	}
	return "Cartulary/" + serviceVersion + " OTel-OTLP-Exporter-go/" + exporterVersion, nil
}

type RetryClassification string

const (
	RetryTransient RetryClassification = "transient"
	RetryPermanent RetryClassification = "permanent"
)

const ExporterPermanentDiscardDropReason = "exporter_permanent_discard"

const (
	TelemetryExportFailureMetricName = "cartulary.telemetry.export.failure"
	TelemetryQueueDepthMetricName    = "cartulary.telemetry.queue.depth"
	QueueFullDropReason              = "queue_full"
	ShutdownTimeoutDropReason        = "shutdown_timeout"
	RecursionGuardDropReason         = "recursion_guard"
)

func ClassifyOTLPHTTPStatus(statusCode int) RetryClassification {
	switch statusCode {
	case 429, 502, 503, 504:
		return RetryTransient
	default:
		return RetryPermanent
	}
}

func ClassifyOTLPGRPCStatus(code string, retryInfo bool) RetryClassification {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "CANCELLED", "DEADLINE_EXCEEDED", "ABORTED", "OUT_OF_RANGE", "UNAVAILABLE", "DATA_LOSS":
		return RetryTransient
	case "RESOURCE_EXHAUSTED":
		if retryInfo {
			return RetryTransient
		}
		return RetryPermanent
	default:
		return RetryPermanent
	}
}

func DropReasonForRetryClassification(classification RetryClassification) string {
	if classification == RetryPermanent {
		return ExporterPermanentDiscardDropReason
	}
	return ""
}

type RetryPlan struct {
	BaseIntervalMS        int64
	StartRetry            bool
	ProductHotPathBlocked bool
}

func PlanRetry(retry config.TelemetryExporterRetryConfig, retryAttemptIndex int64, elapsedSinceFirstFailedAttemptStartMS int64, sampledDelayMS int64, shutdownStarted bool) RetryPlan {
	if shutdownStarted || !retry.Enabled || retry.MaxElapsedMS == 0 || retryAttemptIndex < 1 || sampledDelayMS < 0 {
		return RetryPlan{}
	}
	base := retry.InitialIntervalMS
	if retryAttemptIndex > 1 {
		baseFloat := float64(retry.InitialIntervalMS) * math.Pow(retry.Multiplier, float64(retryAttemptIndex-1))
		if baseFloat > float64(math.MaxInt64) {
			base = math.MaxInt64
		} else {
			base = int64(baseFloat)
		}
	}
	if base > retry.MaxIntervalMS {
		base = retry.MaxIntervalMS
	}
	if sampledDelayMS > base {
		return RetryPlan{BaseIntervalMS: base}
	}
	if elapsedSinceFirstFailedAttemptStartMS+sampledDelayMS > retry.MaxElapsedMS {
		return RetryPlan{BaseIntervalMS: base}
	}
	return RetryPlan{BaseIntervalMS: base, StartRetry: true}
}

type ExporterHeaderPlan struct {
	RequestHeaders    map[string]string
	DiagnosticHeaders map[string]string
}

func BuildExporterRequestHeaders(configured map[string]string, userAgent string) (ExporterHeaderPlan, error) {
	if !safeExporterUserAgent(userAgent) {
		return ExporterHeaderPlan{}, fmt.Errorf("exporter User-Agent is invalid")
	}
	request := map[string]string{
		"content-type": "application/x-protobuf",
		"user-agent":   userAgent,
	}
	diagnostic := map[string]string{
		"content-type": "application/x-protobuf",
		"user-agent":   userAgent,
	}
	seen := map[string]struct{}{
		"content-type": {},
		"user-agent":   {},
	}
	for name, value := range configured {
		lowerName := strings.ToLower(strings.TrimSpace(name))
		if !safeExporterHeaderName(lowerName) {
			return ExporterHeaderPlan{}, fmt.Errorf("exporter header name %q is invalid", name)
		}
		if _, owned := protocolOwnedExporterHeaders[lowerName]; owned {
			return ExporterHeaderPlan{}, fmt.Errorf("exporter header %q is protocol-owned", name)
		}
		if _, exists := seen[lowerName]; exists {
			return ExporterHeaderPlan{}, fmt.Errorf("exporter header %q is duplicated after canonicalization", name)
		}
		if !safeExporterHeaderValue(value) {
			return ExporterHeaderPlan{}, fmt.Errorf("exporter header %q value is invalid", name)
		}
		seen[lowerName] = struct{}{}
		request[lowerName] = value
		diagnostic[lowerName] = "[redacted]"
	}
	return ExporterHeaderPlan{RequestHeaders: request, DiagnosticHeaders: diagnostic}, nil
}

func ExporterDiagnosticsContainSecret(plan ExporterHeaderPlan, secretValues ...string) bool {
	for _, secret := range secretValues {
		if secret == "" {
			continue
		}
		for name, value := range plan.DiagnosticHeaders {
			if strings.Contains(name, secret) || strings.Contains(value, secret) {
				return true
			}
		}
	}
	return false
}

type ExportFailure struct {
	SignalKind   string
	ExporterKind string
	ErrorClass   string
	Recursive    bool
}

func ExportFailureMetric(failure ExportFailure) (string, []attribute.KeyValue, bool) {
	if failure.Recursive {
		return "", nil, false
	}
	attrs := SafeAttributes(
		attribute.String("cartulary.signal_kind", failure.SignalKind),
		attribute.String("cartulary.telemetry.exporter_kind", failure.ExporterKind),
		attribute.String("cartulary.error_class", failure.ErrorClass),
	)
	if len(attrs) != 3 {
		return "", nil, false
	}
	return TelemetryExportFailureMetricName, attrs, true
}

type ExportAttemptTimeoutPlan struct {
	TimedOut              bool
	Classification        RetryClassification
	ProductHotPathBlocked bool
}

func PlanExporterAttemptTimeout(exportTimeoutMS int64, elapsedMS int64, transportClassification RetryClassification) ExportAttemptTimeoutPlan {
	if exportTimeoutMS <= 0 || elapsedMS < exportTimeoutMS {
		return ExportAttemptTimeoutPlan{}
	}
	classification := RetryPermanent
	if transportClassification == RetryTransient {
		classification = RetryTransient
	}
	return ExportAttemptTimeoutPlan{
		TimedOut:       true,
		Classification: classification,
	}
}

type ProcessorQueue struct {
	SignalKind string
	MaxSize    int
	Depth      int
	Recursive  bool
}

type QueueOfferResult struct {
	Accepted          bool
	Depth             int
	RetainedQueued    int
	DroppedNewItem    bool
	OverflowPolicy    string
	DropReason        string
	MetricName        string
	Attributes        []attribute.KeyValue
	ProductBlocked    bool
	QueueDepthMetric  string
	QueueDepthAttrs   []attribute.KeyValue
	QueueDepthValue   int
	QueueDepthCurrent bool
}

func OfferProcessorQueue(queue ProcessorQueue) QueueOfferResult {
	depth := queue.Depth
	if depth < 0 {
		depth = 0
	}
	result := QueueOfferResult{
		Depth:            depth,
		RetainedQueued:   depth,
		OverflowPolicy:   "drop_new",
		QueueDepthMetric: TelemetryQueueDepthMetricName,
		QueueDepthValue:  depth,
	}
	if attrs := queueDepthAttributes(queue.SignalKind); len(attrs) == 1 {
		result.QueueDepthAttrs = attrs
		result.QueueDepthCurrent = true
	}
	if queue.MaxSize <= 0 || depth >= queue.MaxSize {
		result.DroppedNewItem = true
		result.DropReason = QueueFullDropReason
		if name, attrs, ok := dropMetric(queue.SignalKind, QueueFullDropReason, queue.Recursive); ok {
			result.MetricName = name
			result.Attributes = attrs
		}
		return result
	}
	result.Accepted = true
	result.Depth = depth + 1
	result.RetainedQueued = depth + 1
	result.QueueDepthValue = depth + 1
	return result
}

type ShutdownPlan struct {
	SignalKind      string
	FlushTimeoutMS  int64
	ElapsedMS       int64
	ActiveProvider  bool
	AlreadyShutdown bool
	Recursive       bool
}

type ShutdownResult struct {
	ContinueShutdown  bool
	CallShutdown      bool
	ShutdownCallCount int
	TimedOut          bool
	DropReason        string
	MetricName        string
	Attributes        []attribute.KeyValue
	ProductBlocked    bool
}

func PlanShutdown(plan ShutdownPlan) ShutdownResult {
	result := ShutdownResult{ContinueShutdown: true}
	if !plan.ActiveProvider || plan.AlreadyShutdown {
		return result
	}
	result.CallShutdown = true
	result.ShutdownCallCount = 1
	if plan.FlushTimeoutMS > 0 && plan.ElapsedMS > plan.FlushTimeoutMS {
		result.TimedOut = true
		result.DropReason = ShutdownTimeoutDropReason
		if name, attrs, ok := dropMetric(plan.SignalKind, ShutdownTimeoutDropReason, plan.Recursive); ok {
			result.MetricName = name
			result.Attributes = attrs
		}
	}
	return result
}

type SelfDiagnosticPlan struct {
	SignalKind    string
	Recursive     bool
	MetricAllowed bool
}

type SelfDiagnosticResult struct {
	Record         bool
	DropReason     string
	MetricName     string
	Attributes     []attribute.KeyValue
	RecursionBound bool
}

func PlanSelfDiagnostic(plan SelfDiagnosticPlan) SelfDiagnosticResult {
	if !plan.Recursive {
		return SelfDiagnosticResult{Record: true, RecursionBound: true}
	}
	result := SelfDiagnosticResult{DropReason: RecursionGuardDropReason, RecursionBound: true}
	if plan.MetricAllowed {
		if name, attrs, ok := dropMetric(plan.SignalKind, RecursionGuardDropReason, false); ok {
			result.MetricName = name
			result.Attributes = attrs
		}
	}
	return result
}

type RuntimeSurface string

const (
	RuntimeSurfaceHTTPRequest             RuntimeSurface = "http_request"
	RuntimeSurfaceWorkbookQuery           RuntimeSurface = "workbook_query"
	RuntimeSurfaceWorkbookMutation        RuntimeSurface = "workbook_mutation"
	RuntimeSurfaceWebSocketSend           RuntimeSurface = "websocket_send"
	RuntimeSurfaceEvidenceAccess          RuntimeSurface = "evidence_access"
	RuntimeSurfaceBackgroundJobTransition RuntimeSurface = "background_job_transition"
)

type RuntimeFailureMode string

const (
	RuntimeFailureExporterFailure    RuntimeFailureMode = "exporter_failure"
	RuntimeFailureExporterTimeout    RuntimeFailureMode = "exporter_timeout"
	RuntimeFailureQueueOverflow      RuntimeFailureMode = "queue_overflow"
	RuntimeFailureRedactionRejection RuntimeFailureMode = "redaction_rejection"
)

type RuntimeInvarianceCase struct {
	Surface         RuntimeSurface
	FailureMode     RuntimeFailureMode
	ProductResponse string
	CommittedState  string
	ProductBlocked  bool
}

func RuntimeInvarianceMatrix() []RuntimeInvarianceCase {
	surfaces := []RuntimeSurface{
		RuntimeSurfaceHTTPRequest,
		RuntimeSurfaceWorkbookQuery,
		RuntimeSurfaceWorkbookMutation,
		RuntimeSurfaceWebSocketSend,
		RuntimeSurfaceEvidenceAccess,
		RuntimeSurfaceBackgroundJobTransition,
	}
	failures := []RuntimeFailureMode{
		RuntimeFailureExporterFailure,
		RuntimeFailureExporterTimeout,
		RuntimeFailureQueueOverflow,
		RuntimeFailureRedactionRejection,
	}
	cases := make([]RuntimeInvarianceCase, 0, len(surfaces)*len(failures))
	for _, surface := range surfaces {
		for _, failure := range failures {
			cases = append(cases, RuntimeInvarianceCase{
				Surface:         surface,
				FailureMode:     failure,
				ProductResponse: "match_no_export_baseline",
				CommittedState:  "match_no_export_baseline",
			})
		}
	}
	return cases
}

func dropMetric(signalKind string, reason string, recursive bool) (string, []attribute.KeyValue, bool) {
	if recursive {
		return "", nil, false
	}
	attrs := SafeAttributes(
		attribute.String("cartulary.signal_kind", signalKind),
		attribute.String("cartulary.drop_reason", reason),
	)
	if len(attrs) != 2 {
		return "", nil, false
	}
	return TelemetryItemDroppedMetricName, attrs, true
}

func queueDepthAttributes(signalKind string) []attribute.KeyValue {
	return SafeAttributes(attribute.String("cartulary.signal_kind", signalKind))
}

func parseEndpoint(endpoint string) (*url.URL, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed == nil {
		return nil, fmt.Errorf("parse OTLP endpoint: %w", err)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("OTLP endpoint scheme must be http or https")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("OTLP endpoint must not contain userinfo, query, or fragment")
	}
	if parsed.Hostname() == "" || parsed.Port() == "" {
		return nil, fmt.Errorf("OTLP endpoint must include host and explicit port")
	}
	if strings.Contains(parsed.Hostname(), "%") {
		return nil, fmt.Errorf("OTLP endpoint host must not contain zone or percent encoding")
	}
	if !validEndpointHost(parsed.Hostname()) {
		return nil, fmt.Errorf("OTLP endpoint host must be an ASCII hostname or bracketed IPv6 address")
	}
	return parsed, nil
}

func canonicalHostPort(parsed *url.URL) string {
	host := strings.ToLower(parsed.Hostname())
	if ip := net.ParseIP(host); ip != nil && strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return host + ":" + parsed.Port()
}

func validateOTLPHTTPPath(parsed *url.URL) error {
	if parsed.Path == "" || parsed.Path == "/" {
		return nil
	}
	if strings.Contains(parsed.EscapedPath(), "%") || strings.Contains(parsed.Path, "//") {
		return fmt.Errorf("OTLP/HTTP endpoint path must not contain percent-encoded or duplicate slash segments")
	}
	for _, segment := range strings.Split(strings.Trim(parsed.Path, "/"), "/") {
		if segment == "" || segment == "." || segment == ".." || !endpointPathSegmentPattern.MatchString(segment) {
			return fmt.Errorf("OTLP/HTTP endpoint path contains an unsupported segment")
		}
	}
	return nil
}

func validEndpointHost(host string) bool {
	if host == "" || strings.ContainsAny(host, "\t\n\r ") {
		return false
	}
	if strings.Contains(host, ":") {
		return net.ParseIP(host) != nil
	}
	lowerHost := strings.ToLower(host)
	if strings.Contains(lowerHost, "xn--") {
		return false
	}
	for _, r := range host {
		if r > 127 {
			return false
		}
	}
	return endpointHostPattern.MatchString(host)
}

func safeExporterUserAgent(value string) bool {
	if strings.Count(value, " ") != 1 || strings.ContainsAny(value, "()\t\n\r") {
		return false
	}
	segments := strings.Split(value, " ")
	if len(segments) != 2 || !strings.HasPrefix(segments[0], "Cartulary/") || !strings.HasPrefix(segments[1], "OTel-OTLP-Exporter-go/") {
		return false
	}
	serviceVersion := strings.TrimPrefix(segments[0], "Cartulary/")
	exporterVersion := strings.TrimPrefix(segments[1], "OTel-OTLP-Exporter-go/")
	return safeServiceVersion(serviceVersion) && safeExporterVersion(exporterVersion)
}

func safeExporterHeaderName(name string) bool {
	return exporterHeaderNamePattern.MatchString(name)
}

func safeExporterHeaderValue(value string) bool {
	if value == "" || value != strings.TrimSpace(value) || len(value) > 4096 {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r > 0x7e {
			return false
		}
	}
	return true
}

func safeExporterVersion(version string) bool {
	version = strings.TrimSpace(version)
	return len(version) <= 64 && exporterVersionPattern.MatchString(version)
}

var protocolOwnedExporterHeaders = map[string]struct{}{
	"host":              {},
	"content-type":      {},
	"content-length":    {},
	"transfer-encoding": {},
	"connection":        {},
	"te":                {},
	"user-agent":        {},
	"traceparent":       {},
	"tracestate":        {},
	"baggage":           {},
}

var (
	endpointHostPattern        = regexp.MustCompile(`^[A-Za-z0-9.-]+$`)
	endpointPathSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]{1,64}$`)
	exporterVersionPattern     = regexp.MustCompile(`^v[0-9]+(\.[0-9]+)+$`)
	exporterHeaderNamePattern  = regexp.MustCompile(`^[a-z0-9_.-]{1,64}$`)
)
