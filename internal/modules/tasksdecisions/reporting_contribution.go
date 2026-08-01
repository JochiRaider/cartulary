package tasksdecisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	taskdecisionreporting "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/reporting"
)

type ReportingContribution struct{}

func NewReportingContribution() ReportingContribution {
	return ReportingContribution{}
}

func (ReportingContribution) ProviderKey() string {
	return "tasksdecisions"
}

func (ReportingContribution) CollectFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	supportRefs map[string][]string,
) (exportprovider.ProviderOutput, error) {
	return taskdecisionreporting.CollectFactsTx(ctx, tx, incidentID, supportRefs)
}

var _ exportprovider.FieldProvider = ReportingContribution{}
