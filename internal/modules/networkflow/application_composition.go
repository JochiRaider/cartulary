package networkflow

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type incidentAdmissionChecker interface {
	Check(context.Context, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
	CheckTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
}

// activeApplications is published only after complete, quiescent construction.
// Core publication controls whether its contributions are ever requested.
type activeApplications struct {
	savedGraphReads *savedGraphReadApplication
	graphQueries    *graphQueryApplication
	tableQueries    *tableQueryApplication
	tables          *tableApplication
	links           *indicatorLinkApplication
	savedGraphs     *savedGraphApplication
	importOwner     imports.ExtensionImportFacade
}

func (m *Module) prepareActiveApplications() error {
	if m == nil || m.store == nil || m.store.pool == nil || m.transactions == nil ||
		m.importSources == nil || m.incidentAccess == nil || m.authStore == nil ||
		m.now == nil || m.safeDigester == nil || m.cursorProtector == nil ||
		m.graphComposer == nil || m.graphVerifier == nil || m.graphProjection == nil ||
		m.reportingSource == nil || m.graphViewJobs == nil || m.jobManager == nil ||
		m.jobRunner == nil || m.jobFinalizer == nil {
		return errors.New("network flow active composition dependencies unavailable")
	}
	if m.active != nil {
		return nil
	}
	reader, err := postgresresult.NewReader(m.store.pool)
	if err != nil {
		return err
	}
	m.active = &activeApplications{
		savedGraphReads: &savedGraphReadApplication{declarations: m.store, results: reader, jobs: m.jobManager, incidentAccess: m.incidentAccess, composer: m.graphComposer, verifier: m.graphVerifier, cursorProtector: m.cursorProtector, limits: m.limits},
		graphQueries:    &graphQueryApplication{store: m.store, incidentAccess: m.incidentAccess, composer: m.graphComposer, cursorProtector: m.cursorProtector, safeDigester: m.safeDigester},
		tableQueries:    &tableQueryApplication{store: m.store, incidentAccess: m.incidentAccess, cursorProtector: m.cursorProtector},
		tables:          &tableApplication{store: m.store, incidentAccess: m.incidentAccess, safeDigester: m.safeDigester, now: m.now},
		links: &indicatorLinkApplication{store: m.store, incidentAccess: m.incidentAccess,
			receipts:     indicatorLinkReceiptAdapter{reader: m.authStore, store: m.store},
			safeDigester: m.safeDigester, now: m.now, transactions: m.transactions, graphComposer: m.graphComposer},
		savedGraphs: &savedGraphApplication{store: m.store, incidentAccess: m.incidentAccess,
			receipts: savedGraphReceiptAdapter{reader: m.authStore}, graphViewJobs: m.graphViewJobs, jobRunner: m.jobRunner, now: m.now},
		importOwner: newImportFacade(m.store, m.importSources, m.limits, m.now, m.safeDigester),
	}
	return nil
}
