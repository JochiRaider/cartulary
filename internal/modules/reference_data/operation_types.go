package reference_data

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"time"
)

func referenceCatalogJobKind(kind string) string {
	switch kind {
	case "import":
		return ImportJobKind
	case "reverify":
		return ReverifyJobKind
	case "refresh":
		return RefreshJobKind
	default:
		return ""
	}
}

var ErrNotFound = errors.New("reference_data: not found")

type JobAcceptedResult struct {
	Job      jobs.Resource
	Replayed bool
}

type ActionParams struct {
	ActorUserID uuid.UUID
	PackKey     string
	PackVersion string
	Request     ActionRequest
	Now         time.Time
}

type ActionResult struct {
	Payload  map[string]any
	Replayed bool
}

func intPtr(value int) *int { return &value }
