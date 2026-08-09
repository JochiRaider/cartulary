package tasksdecisions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	taskdecisionreporting "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/reporting"
	taskprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
)

type ReportingContribution struct {
	reader taskprojection.Reader
}

func NewReportingContribution(reader taskprojection.Reader) (ReportingContribution, error) {
	if reader == nil {
		return ReportingContribution{}, fmt.Errorf("compose Tasks/Decisions reporting provider: projection reader is required")
	}
	return ReportingContribution{reader: reader}, nil
}

func (ReportingContribution) ProviderKey() string {
	return "tasksdecisions"
}

func (contribution ReportingContribution) CollectFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	supportRefs map[string][]string,
) (exportprovider.ProviderOutput, error) {
	return taskdecisionreporting.CollectFactsTx(ctx, tx, incidentID, supportRefs, contribution.reader)
}

var _ exportprovider.FieldProvider = ReportingContribution{}
