package inmemory_store

import (
	"context"
	contracts_workflow "natsauth/internal/contracts/workflow"
	"sync"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	status "github.com/gogo/status"
	zerolog "github.com/rs/zerolog"
	codes "google.golang.org/grpc/codes"
)

type (
	service struct {
		memmap sync.Map
	}
)

var stemService = (*service)(nil)
var _ contracts_workflow.IWorkflowCache = (*service)(nil)

func (s *service) Ctor() (contracts_workflow.IWorkflowCache, error) {
	return &service{}, nil
}
func AddSingletonWorkflowCache(builder di.ContainerBuilder) {
	di.AddSingleton[contracts_workflow.IWorkflowCache](
		builder,
		stemService.Ctor,
	)
}

func (s *service) validateSetWorkflowStateRequest(request *contracts_workflow.SetWorkflowStateRequest) error {

	if fluffycore_utils.IsNil(request) {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if fluffycore_utils.IsEmptyOrNil(request.WorkflowID) {
		return status.Error(codes.InvalidArgument, "WorkflowID is required")
	}
	if fluffycore_utils.IsEmptyOrNil(request.State) {
		return status.Error(codes.InvalidArgument, "State is required")
	}
	return nil
}
func (s *service) SetWorkflowState(ctx context.Context, request *contracts_workflow.SetWorkflowStateRequest) (*contracts_workflow.SetWorkflowStateRequestResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("service", "WorkflowStore").Logger()
	err := s.validateSetWorkflowStateRequest(request)
	if err != nil {
		log.Error().Err(err).Msg("failed to validate request")
		return nil, err
	}
	s.memmap.Store(request.WorkflowID, request.State)

	return &contracts_workflow.SetWorkflowStateRequestResponse{
		WorkflowID: request.WorkflowID,
	}, nil
}
func (s *service) validateGetWorkflowStateRequest(request *contracts_workflow.GetWorkflowStateRequest) error {

	if fluffycore_utils.IsNil(request) {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if fluffycore_utils.IsEmptyOrNil(request.WorkflowID) {
		return status.Error(codes.InvalidArgument, "WorkflowID is required")
	}
	return nil
}
func (s *service) GetWorkflowState(ctx context.Context, request *contracts_workflow.GetWorkflowStateRequest) (*contracts_workflow.GetWorkflowStateRequestResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("service", "WorkflowStore").Logger()
	err := s.validateGetWorkflowStateRequest(request)
	if err != nil {
		log.Error().Err(err).Msg("failed to validate request")
		return nil, err
	}
	state, ok := s.memmap.Load(request.WorkflowID)
	if !ok {
		return nil, status.Error(codes.NotFound, "WorkflowID not found")
	}
	resp := &contracts_workflow.GetWorkflowStateRequestResponse{
		WorkflowID: request.WorkflowID,
		State:      state.([]byte),
	}
	return resp, nil
}
