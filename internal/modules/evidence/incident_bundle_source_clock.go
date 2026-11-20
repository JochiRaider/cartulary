package evidence

import (
	"time"

	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
)

// portableSourceClock admits a canonical UTC portable clock.
func portableSourceClock(raw any) (time.Time, bool) {
	return evidenceblob.ParseClock(raw)
}
