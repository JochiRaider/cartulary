package tasksdecisions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	taskdecisionreporting "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/reporting"
	taskprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionports"
)

type reportingContribution struct {
	reader taskprojection.ReportingReader
}

func NewReportingContribution(reader taskprojection.ReportingReader) (exportprovider.FieldProvider, error) {
	if reader == nil {
		return nil, fmt.Errorf("compose Tasks/Decisions reporting provider: projection reader is required")
	}
	return reportingContribution{reader: reader}, nil
}

func (reportingContribution) ProviderKey() string {
	return "tasksdecisions"
}

func (contribution reportingContribution) CollectFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	supportRefs map[string][]string,
) (exportprovider.ProviderOutput, error) {
	return taskdecisionreporting.CollectFactsTx(ctx, tx, incidentID, supportRefs, contribution.reader)
}

var _ exportprovider.FieldProvider = reportingContribution{}
