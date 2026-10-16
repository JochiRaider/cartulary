package renderprocess

import "errors"

var ErrNetworkObservation = errors.New("render process network observation unavailable")

type NetworkObservation struct {
	Kind             string
	DestinationClass string
	Allowed          bool
	Count            int
}

// NetworkError contains only closed classifications; never addresses or paths.
type NetworkError struct{ Kind, DestinationClass string }

func (e *NetworkError) Error() string {
	return "render process network policy violation: " + e.Kind + ": " + e.DestinationClass
}
