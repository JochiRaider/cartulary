package networkflow_test

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Decorate the existing consumer port only in the isolated owner test runtime.
// The after-create failure occurs inside the real borrowed transaction.
type linkFixtureDependencyProbe struct {
	IndicatorParticipationPort
	mu               sync.Mutex
	boundary         string
	calls            int
	created, visible bool
	result           *indicators.IndicatorFindOrCreateParticipantResult
}

func (p *linkFixtureDependencyProbe) arm(boundary string) error {
	switch boundary {
	case "", "before_core_create", "after_core_create":
	default:
		return fmt.Errorf("unknown Core fixture boundary")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.boundary = boundary
	p.calls = 0
	p.created = false
	p.visible = false
	p.result = nil
	return nil
}
func (p *linkFixtureDependencyProbe) FindOrCreateIndicatorParticipantTx(ctx context.Context, tx pgx.Tx, command indicators.IndicatorFindOrCreateParticipantCommand) (indicators.IndicatorFindOrCreateParticipantResult, error) {
	p.mu.Lock()
	boundary := p.boundary
	if boundary != "" {
		p.calls++
	}
	p.mu.Unlock()
	if boundary == "before_core_create" {
		return indicators.IndicatorFindOrCreateParticipantResult{}, errors.New("fixture Core dependency failure")
	}
	result, err := p.IndicatorParticipationPort.FindOrCreateIndicatorParticipantTx(ctx, tx, command)
	if err != nil || boundary == "" {
		return result, err
	}
	retained, err := p.IndicatorParticipationPort.GetActiveIndicatorParticipantTx(ctx, tx, command.IncidentID, result.Indicator.RecordID)
	if err != nil {
		return result, err
	}
	p.mu.Lock()
	p.result = &result
	p.created = result.Status == "created"
	p.visible = result.Indicator.RecordID != uuid.Nil && retained.RecordID == result.Indicator.RecordID
	p.mu.Unlock()
	return indicators.IndicatorFindOrCreateParticipantResult{}, errors.New("fixture failure after Core indicator creation")
}
func (p *linkFixtureDependencyProbe) observation() (map[string]any, any, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.calls != 1 || p.boundary == "after_core_create" && (!p.created || !p.visible) {
		return nil, nil, fmt.Errorf("Core fixture fault did not reach its declared transaction boundary")
	}
	return map[string]any{"boundary": p.boundary, "calls": p.calls, "core_created": p.created, "created_indicator_visible_in_tx": p.visible}, p.result, nil
}
