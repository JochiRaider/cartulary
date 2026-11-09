// Package packformat coordinates application policy around the independent format library.
package packformat

import "github.com/JochiRaider/cartulary/referencepacks/format/envelope"

type Failure struct {
	Code    string
	CheckID string
	Summary *ValidationSummary
	Details SafeDetails
}

func (e *Failure) Error() string { return "reference pack: " + e.Code }
func fail(code string) error     { return &Failure{Code: code} }

var ErrStorage = envelope.ErrStorage

type ArchiveLimits = envelope.ArchiveLimits
type Destination = envelope.Destination
type Member = envelope.Member
type Inventory = envelope.Inventory

func DefaultArchiveLimits() ArchiveLimits { return envelope.DefaultArchiveLimits() }
func ValidatePath(name string, directory bool) error {
	return formatError(envelope.ValidatePath(name, directory))
}
