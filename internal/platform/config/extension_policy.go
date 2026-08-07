package config

// ClaimRegistration binds one opaque registration identity to one closed
// Boolean configuration path. Registrations are supplied by static application
// composition after their owning artifact catalog has been admitted.
type ClaimRegistration struct {
	ID   string
	Path string
}

// ExtensionPolicy is the neutral application-supplied boundary for extension
// claim collection and inactive-value admission. Implementations must be
// immutable, canonically ordered, and effect-free.
type ExtensionPolicy interface {
	ClaimRegistrations() []ClaimRegistration
	Keys() []string
	ClaimKey(key string) (string, bool)
	ParseOverlay(key string, raw string) (any, error)
	ValidateAndDiscard(values map[string]any) [][2]string
}
