package networkflow

import (
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
)

// External integration fixtures use only these test-build names. Ordinary
// application builds expose the Module facade and explicit owner contributions.
type CreateTableParams = createTableParams

var DefaultEffectiveLimits = defaultEffectiveLimits
var DeriveTableDisplayName = deriveTableDisplayName
var EndpointID = endpointID
var ErrGraphViewDeclarationInvalid = errGraphViewDeclarationInvalid
var ErrGraphViewPublicationStale = errGraphViewPublicationStale
var ErrSavedGraphCutoverIncompatible = errSavedGraphCutoverIncompatible
var ErrTableNotActive = errTableNotActive
var ErrTableVersionConflict = errTableVersionConflict

const FieldDstIP = fieldDstIP
const FieldSrcIP = fieldSrcIP
const FieldSrcPort = fieldSrcPort

var FlowEdgeID = flowEdgeID

type FlowRow = flowRow
type GraphViewDeclaration = graphViewDeclaration

const GraphViewDeclarationStateActive = graphViewDeclarationStateActive
const GraphViewDeclarationStateRetired = graphViewDeclarationStateRetired

type GraphViewSelectedResultBinding = graphViewSelectedResultBinding

var GraphViewSemanticQuerySHA256 = graphViewSemanticQuerySHA256

const GraphViewWorkerKind = graphViewWorkerKind

type InvalidDisplayNameError = invalidDisplayNameError

var LifecycleStates = lifecycleStates
var NewReportingGraphSource = newReportingGraphSource
var NewStore = newStore
var ParseKeyRings = parseKeyRingsWithDefaultRegistry

const ParserProfileRFC4180HeaderedCSV = parserProfileRFC4180HeaderedCSV

type RejectedRowDiagnostic = rejectedRowDiagnostic
type RenameTableParams = renameTableParams

var SafeDigest = safeDigest
var SanitizeSourceFilenameDisplay = sanitizeSourceFilenameDisplay

type SoftDeleteTableParams = softDeleteTableParams

const SourceProfileCiscoSNANetFlowCSV = sourceProfileCiscoSNANetFlowCSV

type Store = store
type StoreOption = storeOption
type TableLimitError = tableLimitError
type TableRecord = tableRecord

const TableStatusActive = tableStatusActive
const TableStatusSoftDeleted = tableStatusSoftDeleted

var WithEffectiveLimits = withEffectiveLimits
var WithOwnerParticipants = withOwnerParticipants
var WithResourceIntentAppender = withResourceIntentAppender
var WithSafeDigester = withSafeDigester

const WorkspaceKeyNetworkAnalysis = workspaceKeyNetworkAnalysis

func parseKeyRingsWithDefaultRegistry(raw []byte, env map[string]string, now time.Time) (*KeyRings, error) {
	registry := secretpurpose.NewRegistry()
	if err := authn.RegisterMasterSecretPurpose(registry, env); err != nil {
		return nil, err
	}
	return parseKeyRings(raw, env, now, registry)
}

func lifecycleStates() []string {
	return []string{tableStatusActive, tableStatusSoftDeleted}
}

func defaultLimits() EffectiveLimits {
	return defaultEffectiveLimits()
}

func withEffectiveLimits(limits EffectiveLimits) storeOption {
	return func(s *store) {
		s.limits = limits
	}
}
