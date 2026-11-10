package networkflow

import (
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

type GraphResultCleanupService = graphResultCleanupService
type GraphResultCleanupSweepResult = graphResultCleanupSweepResult

func NewGraphResultCleanupService(db dbport.DB, declarations *store) (*GraphResultCleanupService, error) {
	return newGraphResultCleanupService(db, declarations)
}

func SetGraphResultCleanupTestBounds(service *GraphResultCleanupService, maximumResults int, maximumDuration time.Duration) {
	service.maximumResults = maximumResults
	service.maximumDuration = maximumDuration
}
