package config

// InactivePolicy is the neutral application-supplied boundary for Extensions
// admission. Implementations must be immutable and effect-free.
type InactivePolicy interface {
	Keys() []string
	ClaimKey(key string) (string, bool)
	ParseOverlay(key string, raw string) (any, error)
	ValidateAndDiscard(values map[string]any) [][2]string
}
