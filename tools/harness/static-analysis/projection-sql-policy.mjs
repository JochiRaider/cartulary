// Declaration capabilities for SQL-bearing mixed packages. Dedicated source
// packages use provider defaults; new mixed-package SQL requires review.
export const providerPackageDefaults = {
 "internal/modules/evidence/timelinefacts":"timeline",
 "internal/modules/links/facts":"shared_read",
 "internal/modules/records/enveloperead":"shared_read",
  "internal/modules/timeline/sourcerepository": "timeline",
  "internal/modules/entities/timelinefacts": "timeline",
  "internal/modules/indicators/internal/providers/projection": "indicator",
  "internal/modules/assessments/internal/providers/projection": "assessment",
  "internal/modules/artifacts/internal/providers/projection": "artifact",
  "internal/modules/evidence/internal/providers/projection": "evidence",
  "internal/modules/parties/internal/providers/projection": "party"
};
export const declarationBindings = {
  "internal/modules/entities/hostidentity/projectionprovider": {
    "source.ListHostProjectionInputsTx": "host",
"source.LoadHostProjectionInputTx": "host",
"hostProjectionSourceSQL": "host",
    "source.ListIdentityProjectionInputsTx": "identity",
"source.LoadIdentityProjectionInputTx": "identity",
"identityProjectionSourceSQL": "identity"
  },
  "internal/modules/projections/internal/runtime": {
"projectionCountQuery": "projection_count",
    "listRestoreProjectionIncidentIDs": "restore_incidents"
  },
  "internal/modules/projections/internal/queryengine": {
    "TaskReader.CollectTaskDerivedFactsTx": "task_request",
    "TaskRequestPlans": "task_request",
    "taskRecordRefCollectionExpr": "task_request",
    "IdentityReader.CollectIdentityDerivedFactsTx": "identity",
    "buildIdentityQueryPageSQL": "identity",
    "HostReader.CollectHostDerivedFactsTx": "host",
    "buildHostQueryPageSQL": "host",
    "DecisionReader.CollectDecisionDerivedFactsTx": "decision",
    "DecisionPlans": "decision",
    "ArtifactReader.CollectDerivedFactsTx": "artifact",
    "artifactSurface": "artifact",
    "recordRefCollectionExprFor": "artifact",
    "tagCollectionExprFor": "artifact",
    "partyRefCollectionExpr": "artifact",
    "riskRefCollectionExpr": "artifact",
    "TimelinePlans": "timeline",
    "appendTagFilter": "compiled_projection",
    "IndicatorPlans": "indicator",
    "EvidencePlans": "evidence",
    "AssessmentPlans": "assessment",
    "PartyPlans": "party"
  },
  "internal/modules/projections/internal/storage": {
    "Store.InsertTaskRequestTx": "task_request",
    "Store.DeleteTaskRequestRowTx": "task_request",
    "Store.DeleteTaskRequestIncidentTx": "task_request",
    "Store.UpsertIdentityTx": "identity",
    "Store.DeleteIdentityRowTx": "identity",
    "Store.DeleteIdentityIncidentTx": "identity",
    "Store.LoadIdentityEvidenceAssociationStateTx": "identity",
    "Store.UpsertHostTx": "host",
    "Store.DeleteHostRowTx": "host",
    "Store.DeleteHostIncidentTx": "host",
    "Store.LoadHostEvidenceAssociationStateTx": "host",
    "Store.InsertDecisionTx": "decision",
    "Store.DeleteDecisionRowTx": "decision",
    "Store.DeleteDecisionIncidentTx": "decision",
    "Store.InsertArtifactTx": "artifact",
    "Store.DeleteArtifactRowTx": "artifact",
    "Store.DeleteArtifactIncidentTx": "artifact",
    "Store.UpsertTimelineRowTx": "timeline",
    "Store.InsertTimelineFixtureBatchTx": "timeline",
    "countTimelineFixtureRows": "timeline",
    "Store.DeleteTimelineRowTx": "timeline",
    "Store.DeleteTimelineIncidentTx": "timeline",
    "Store.UpsertIndicatorTx": "indicator",
    "Store.DeleteIndicatorRowTx": "indicator",
    "Store.DeleteIndicatorIncidentTx": "indicator",
    "Store.InsertEvidenceTx": "evidence",
    "Store.DeleteEvidenceRowTx": "evidence",
    "Store.DeleteEvidenceIncidentTx": "evidence",
    "Store.UpsertAssessmentTx": "assessment",
    "Store.DeleteAssessmentRowTx": "assessment",
    "Store.DeleteAssessmentIncidentTx": "assessment",
    "Store.InsertPartyTx": "party",
    "Store.DeletePartyRowTx": "party",
    "Store.DeletePartyIncidentTx": "party"
  },
  "internal/modules/tasksdecisions/internal/providers/projection": {
    "TaskRequestSource.ListTaskRequestProjectionInputsTx": "task_request",
"TaskRequestSource.LoadTaskRequestProjectionInputTx": "task_request",
"taskRequestProjectionSourceSQL": "task_request",
    "DecisionSource.ListDecisionProjectionInputsTx": "decision",
"DecisionSource.LoadDecisionProjectionInputTx": "decision",
"decisionProjectionSourceSQL": "decision"
  },
};
// Implementation entries are capabilities, not recursive source ceilings.
export const providerEntries = {
  timeline: ["internal/modules/timeline/projectionprovider", "internal/app/timelinefactassembly"],
  host: ["internal/modules/entities/hostidentity/projectionprovider"],
  identity: ["internal/modules/entities/hostidentity/projectionprovider"],
  artifact: ["internal/modules/artifacts/internal/providers/projection"],
  assessment: ["internal/modules/assessments/internal/providers/projection", "internal/app/assessmentassembly/projectionfacts"],
  evidence: ["internal/modules/evidence/internal/providers/projection"],
  indicator: ["internal/modules/indicators/internal/providers/projection"],
  party: ["internal/modules/parties/internal/providers/projection"],
  task_request: ["internal/modules/tasksdecisions/internal/providers/projection"],
  decision: ["internal/modules/tasksdecisions/internal/providers/projection"]
};
export const infrastructureEntries = [
  "internal/modules/projections/internal/queryengine",
  "internal/modules/projections/internal/storage",
  "internal/modules/projections/internal/runtime"
];
// Foreign implementation is outside this repository's SQL ownership policy.
// Exact import boundaries remain reviewable; new dependencies require admission.
export const foreignPackages = new Set([
 "github.com/BurntSushi/toml", "github.com/google/uuid", "github.com/jackc/pgx/v5", "github.com/jackc/pgx/v5/pgconn", "github.com/jackc/pgx/v5/pgtype",
 "go.opentelemetry.io/otel", "go.opentelemetry.io/otel/attribute", "go.opentelemetry.io/otel/codes",
 "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc", "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp",
 "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc", "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp",
 "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc", "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp",
 "go.opentelemetry.io/otel/log", "go.opentelemetry.io/otel/log/global", "go.opentelemetry.io/otel/log/noop",
 "go.opentelemetry.io/otel/metric", "go.opentelemetry.io/otel/metric/noop", "go.opentelemetry.io/otel/sdk/log",
 "go.opentelemetry.io/otel/sdk/metric", "go.opentelemetry.io/otel/sdk/metric/exemplar", "go.opentelemetry.io/otel/sdk/metric/metricdata",
 "go.opentelemetry.io/otel/sdk/resource", "go.opentelemetry.io/otel/sdk/trace", "go.opentelemetry.io/otel/trace", "go.opentelemetry.io/otel/trace/noop",
 "golang.org/x/sys/unix", "golang.org/x/text/cases", "golang.org/x/text/unicode/norm",
 "google.golang.org/genproto/googleapis/rpc/errdetails", "google.golang.org/grpc", "google.golang.org/grpc/codes", "google.golang.org/grpc/credentials", "google.golang.org/grpc/status"
]);
export const standardPackages = new Set([
 "bytes", "context", "crypto", "crypto/aes", "crypto/cipher", "crypto/ecdsa", "crypto/elliptic", "crypto/fips140", "crypto/hkdf", "crypto/hmac", "crypto/rand", "crypto/rsa", "crypto/sha256", "crypto/tls", "crypto/x509",
 "encoding/binary", "encoding/hex", "encoding/json", "encoding/pem", "errors", "fmt", "io", "io/fs", "math", "math/big", "net", "net/http", "net/url", "os", "path", "path/filepath", "reflect", "regexp", "runtime", "runtime/debug", "slices", "sort", "strconv", "strings", "sync", "sync/atomic", "time", "unicode", "unicode/utf16", "unicode/utf8"
]);
