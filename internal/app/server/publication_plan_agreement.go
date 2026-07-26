package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const (
	publicationPlanAgreementAdvisoryKey int64 = 4850189438622597895
	publicationPlanMarkerPrefix               = "cartulary/runtime/publication-plans/"
	publicationPlanMarkerMaximumBytes   int64 = 4096
	publicationPlanMarkerLifetime             = 2 * time.Minute
)

type publicationPlanMarker struct {
	SchemaID          string    `json:"schema_id"`
	PlanSHA256        string    `json:"plan_sha256"`
	ServiceInstanceID string    `json:"service_instance_id"`
	ExpiresAt         time.Time `json:"expires_at"`
}

type publicationPlanAgreement struct {
	store     objectstore.Store
	markerKey string
	marker    publicationPlanMarker
	now       func() time.Time
	cancel    context.CancelFunc
	done      chan struct{}
	closeOnce sync.Once
}

func admitPublicationPlanAgreement(
	ctx context.Context,
	pool *pgxpool.Pool,
	store objectstore.Store,
	summary extensions.PublicationPlanSummary,
	serviceInstanceID string,
	now func() time.Time,
) (*publicationPlanAgreement, error) {
	if pool == nil || store == nil || serviceInstanceID == "" {
		return nil, errors.New("replicated publication-plan agreement dependencies are incomplete")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	planJSON, err := json.Marshal(summary)
	if err != nil {
		return nil, fmt.Errorf("encode publication plan identity: %w", err)
	}
	sum := sha256.Sum256(planJSON)
	digest := hex.EncodeToString(sum[:])
	agreement := &publicationPlanAgreement{
		store:     store,
		markerKey: publicationPlanMarkerPrefix + digest + "/" + serviceInstanceID + ".json",
		marker: publicationPlanMarker{
			SchemaID:          "cartulary.runtime_publication_plan_marker.v1",
			PlanSHA256:        digest,
			ServiceInstanceID: serviceInstanceID,
		},
		now:  now,
		done: make(chan struct{}),
	}
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire publication-plan agreement session: %w", err)
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, `SELECT pg_advisory_lock($1)`, publicationPlanAgreementAdvisoryKey); err != nil {
		return nil, fmt.Errorf("lock publication-plan agreement: %w", err)
	}
	defer func() {
		_, _ = connection.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, publicationPlanAgreementAdvisoryKey)
	}()
	if err := agreement.rejectConflictingMarkers(ctx); err != nil {
		return nil, err
	}
	if err := agreement.writeMarker(ctx); err != nil {
		return nil, err
	}
	renewCtx, cancel := context.WithCancel(context.Background())
	agreement.cancel = cancel
	go agreement.renew(renewCtx)
	return agreement, nil
}

func (agreement *publicationPlanAgreement) rejectConflictingMarkers(ctx context.Context) error {
	objects, err := agreement.store.ListObjects(ctx, publicationPlanMarkerPrefix)
	if err != nil {
		return fmt.Errorf("list publication-plan agreement markers: %w", err)
	}
	now := agreement.now().UTC()
	for _, object := range objects {
		marker, err := agreement.readMarker(ctx, object.Key)
		if err != nil {
			return err
		}
		if !marker.ExpiresAt.After(now) {
			if err := agreement.store.DeleteObject(ctx, object.Key); err != nil {
				return fmt.Errorf("prune expired publication-plan marker: %w", err)
			}
			continue
		}
		if marker.PlanSHA256 != agreement.marker.PlanSHA256 {
			return errors.New("replicated publication-plan digest conflicts with an active process")
		}
	}
	return nil
}

func (agreement *publicationPlanAgreement) readMarker(ctx context.Context, key string) (publicationPlanMarker, error) {
	reader, info, err := agreement.store.ReadObject(ctx, key, objectstore.ReadOptions{})
	if err != nil {
		return publicationPlanMarker{}, fmt.Errorf("read publication-plan agreement marker: %w", err)
	}
	defer reader.Close()
	if info.Size < 1 || info.Size > publicationPlanMarkerMaximumBytes {
		return publicationPlanMarker{}, errors.New("publication-plan agreement marker has an invalid size")
	}
	payload, err := io.ReadAll(io.LimitReader(reader, publicationPlanMarkerMaximumBytes+1))
	if err != nil {
		return publicationPlanMarker{}, fmt.Errorf("load publication-plan agreement marker: %w", err)
	}
	var marker publicationPlanMarker
	if len(payload) > int(publicationPlanMarkerMaximumBytes) ||
		json.Unmarshal(payload, &marker) != nil ||
		marker.SchemaID != "cartulary.runtime_publication_plan_marker.v1" ||
		len(marker.PlanSHA256) != 64 ||
		marker.ServiceInstanceID == "" ||
		marker.ExpiresAt.IsZero() {
		return publicationPlanMarker{}, errors.New("publication-plan agreement marker is malformed")
	}
	return marker, nil
}

func (agreement *publicationPlanAgreement) writeMarker(ctx context.Context) error {
	agreement.marker.ExpiresAt = agreement.now().UTC().Add(publicationPlanMarkerLifetime)
	payload, err := json.Marshal(agreement.marker)
	if err != nil {
		return fmt.Errorf("encode publication-plan agreement marker: %w", err)
	}
	if err := agreement.store.PutObject(
		ctx,
		agreement.markerKey,
		bytes.NewReader(payload),
		int64(len(payload)),
		"application/json",
	); err != nil {
		return fmt.Errorf("write publication-plan agreement marker: %w", err)
	}
	return nil
}

func (agreement *publicationPlanAgreement) renew(ctx context.Context) {
	defer close(agreement.done)
	ticker := time.NewTicker(publicationPlanMarkerLifetime / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			renewCtx, cancel := context.WithTimeout(ctx, sharedPublicationOperationTimeout)
			_ = agreement.writeMarker(renewCtx)
			cancel()
		}
	}
}

func (agreement *publicationPlanAgreement) Close() {
	if agreement == nil {
		return
	}
	agreement.closeOnce.Do(func() {
		if agreement.cancel != nil {
			agreement.cancel()
			<-agreement.done
		}
		ctx, cancel := context.WithTimeout(context.Background(), sharedPublicationOperationTimeout)
		_ = agreement.store.DeleteObject(ctx, agreement.markerKey)
		cancel()
	})
}
