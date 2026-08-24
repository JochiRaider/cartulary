package parties

import partysource "github.com/JochiRaider/cartulary/internal/modules/parties/internal/source"

const (
	PartyMatchAmbiguousExactMatch = partysource.MatchAmbiguousExactMatch
	PartyMatchCrossKeyExactMatch  = partysource.MatchCrossKeyExactMatch
	PartyMatchExactKeyClaimed     = partysource.MatchExactKeyClaimed
)

// PartyMatchConflictError is the closed, value-free source conflict exposed to
// application adapters.
type PartyMatchConflictError = partysource.MatchConflictError
