package reporting

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	taskprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
)

func CollectFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	supportRefs map[string][]string,
	reader taskprojection.Reader,
) (exportprovider.ProviderOutput, error) {
	if reader == nil {
		return exportprovider.ProviderOutput{}, fmt.Errorf("collect Tasks/Decisions reporting facts: projection reader is required")
	}
	tasks, err := reader.CollectTaskDerivedFactsTx(ctx, tx, incidentID)
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	decisions, err := reader.CollectDecisionDerivedFactsTx(ctx, tx, incidentID)
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	fields := make([]exportprovider.FieldFact, 0, len(tasks)+len(decisions))
	for _, task := range tasks {
		recordID := task.RecordID.String()
		fields = append(fields, exportprovider.FieldFact{
			SchemaID:     exportprovider.FieldFactSchemaID,
			Path:         "/task_requests/" + recordID,
			ContentClass: "working_material",
			SourceFamily: "task_request",
			Value:        task.Value,
			SupportRefs:  exportprovider.CloneStrings(supportRefs[recordID]),
		})
	}
	for _, decision := range decisions {
		recordID := decision.RecordID.String()
		fields = append(fields, exportprovider.FieldFact{
			SchemaID:     exportprovider.FieldFactSchemaID,
			Path:         "/decisions/" + recordID,
			ContentClass: "working_material",
			SourceFamily: "decision",
			Value:        decision.Value,
			SupportRefs:  exportprovider.CloneStrings(supportRefs[recordID]),
		})
	}
	return exportprovider.NewProviderOutput("tasksdecisions", fields)
}
