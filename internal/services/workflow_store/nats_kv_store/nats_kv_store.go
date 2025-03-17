package nats_kv_store

import (
	"context"
	contracts_nats "natsauth/internal/contracts/nats"
	contracts_workflow "natsauth/internal/contracts/workflow"
	"sync"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	nats_jetstream "github.com/nats-io/nats.go/jetstream"
	zerolog "github.com/rs/zerolog"
)

type (
	service struct {
		config         *contracts_workflow.NATSWorkflowStoreConfig
		natsConnection contracts_nats.INATSConnection
		js             nats_jetstream.JetStream
		store          nats_jetstream.KeyValue
		lock           sync.Mutex
	}
)

var stemService = (*service)(nil)
var _ contracts_workflow.IWorkflowCache = (*service)(nil)

func (s *service) Ctor(
	config *contracts_workflow.NATSWorkflowStoreConfig,
	natsConnection contracts_nats.INATSConnection) (contracts_workflow.IWorkflowCache, error) {
	return &service{
		natsConnection: natsConnection,
	}, nil
}
func AddSingletonWorkflowCache(builder di.ContainerBuilder) {
	di.AddSingleton[contracts_workflow.IWorkflowCache](
		builder,
		stemService.Ctor,
	)
}
func (s *service) ensureStore(ctx context.Context) error {
	//--~--~--~--~-- BARBED WIRE --~--~--~--~--//
	s.lock.Lock()
	defer s.lock.Unlock()
	//--~--~--~--~-- BARBED WIRE --~--~--~--~--//

	log := zerolog.Ctx(ctx).With().Str("service", "WorkflowStore").Logger()
	if s.store == nil {
		nc, err := s.natsConnection.Conn(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to get nats connection")
			return err
		}

		js, err := nats_jetstream.New(nc)
		if err != nil {
			log.Error().Err(err).Msg("failed to create JetStream context")
			return err
		}
		s.js = js

		store, err := js.KeyValue(ctx, s.config.Bucket)
		if err != nil {
			log.Error().Err(err).Msg("failed to get key value")
			return err
		}
		s.store = store
	}
	return nil
}
func (s *service) SetWorkflowState(ctx context.Context, request *contracts_workflow.SetWorkflowStateRequest) (*contracts_workflow.SetWorkflowStateRequestResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("service", "WorkflowStore").Logger()
	err := s.ensureStore(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to ensure store")
		return nil, err
	}
	_, err = s.store.Put(ctx, request.WorkflowID, request.State)
	if err != nil {
		return nil, err
	}
	return &contracts_workflow.SetWorkflowStateRequestResponse{
		WorkflowID: request.WorkflowID,
	}, nil
}
func (s *service) GetWorkflowState(ctx context.Context, request *contracts_workflow.GetWorkflowStateRequest) (*contracts_workflow.GetWorkflowStateRequestResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("service", "WorkflowStore").Logger()
	err := s.ensureStore(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to ensure store")
		return nil, err
	}
	entry, err := s.store.Get(ctx, request.WorkflowID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get key value")
		return nil, err
	}
	resp := &contracts_workflow.GetWorkflowStateRequestResponse{
		WorkflowID: request.WorkflowID,
		State:      entry.Value(),
	}

	return resp, nil
}
