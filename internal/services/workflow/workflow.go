package workflow

import (
	"context"
	"encoding/json"
	"errors"
	contracts_workflow "natsauth/internal/contracts/workflow"
	"reflect"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	status "github.com/gogo/status"
	zerolog "github.com/rs/zerolog"
	codes "google.golang.org/grpc/codes"
)

type (
	ExecutionResponse struct {
		FuncName string         `json:"funcName,omitempty"` // the name of the function that was executed
		Data     interface{}    `json:"data,omitempty"`
		Error    *WorkflowError `json:"error,omitempty"`
	}
	service struct {
		WorkflowID   string `json:"id,omitempty"`
		WorkflowType string `json:"workflowType"`

		// A workflow would have a deterministic number of activities, so we can cache the results in order of execution.
		ExecutionResponses            []*ExecutionResponse              `json:"executionTracker,omitempty"`
		CurrentExecutionResponseIndex int                               `json:"currentExecutionResponseIndex,omitempty"`
		WorkflowStatus                contracts_workflow.WorkflowStatus `json:"status"`
		WorkflowInput                 interface{}                       `json:"input,omitempty"`
		WorkflowResponse              interface{}                       `json:"response,omitempty"`
		WorkflowError                 *WorkflowError                    `json:"error,omitempty"`

		store contracts_workflow.IWorkflowCache `json:"-"`
	}
	NewWorkflowRequest struct {
		ID           string                            `json:"id,omitempty"`
		WorkflowType string                            `json:"workflowType"`
		Store        contracts_workflow.IWorkflowCache `json:"-"`
	}
	WorkflowOption func(w *service)
)

var stemService = (*service)(nil)
var _ contracts_workflow.IWorkflow = (*service)(nil)

func (s *service) Ctor(store contracts_workflow.IWorkflowCache) (contracts_workflow.IWorkflow, error) {
	return &service{
		store: store,
	}, nil
}

func AddTransientWorkflow(builder di.ContainerBuilder) {
	di.AddTransient[contracts_workflow.IWorkflow](
		builder,
		stemService.Ctor,
	)
}

func WithWorkflowInputOption(input interface{}) WorkflowOption {
	return func(w *service) {
		w.WorkflowInput = input
	}
}

func validateNewWorkflowRequest(request *NewWorkflowRequest) error {
	if fluffycore_utils.IsNil(request) {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if fluffycore_utils.IsEmptyOrNil(request.ID) {
		return errors.New("ID is required")
	}
	if fluffycore_utils.IsEmptyOrNil(request.WorkflowType) {
		return errors.New("WorkflowType is required")
	}
	return nil
}

func NewWorkflow(ctx context.Context, request *NewWorkflowRequest, opt ...WorkflowOption) (*service, error) {
	err := validateNewWorkflowRequest(request)
	if err != nil {
		return nil, err
	}
	wf := &service{
		WorkflowID:   request.ID,
		WorkflowType: request.WorkflowType,
		store:        request.Store,
	}
	for _, o := range opt {
		o(wf)
	}
	return wf, nil
}

func (s *service) PushExecutionResponse(ctx context.Context, er *ExecutionResponse) error {
	s.ExecutionResponses = append(s.ExecutionResponses, er)
	return s.StoreState(ctx)
}
func (s *service) PopExecutionResponse(ctx context.Context) {
	erLen := len(s.ExecutionResponses)
	if erLen == 0 {
		return
	}
	s.ExecutionResponses = s.ExecutionResponses[0 : erLen-1]

}
func ExecuteWorkflow(ctx context.Context, wf contracts_workflow.IWorkflow, fn contracts_workflow.WorkflowExecutionFunc) (resp interface{}, err error) {

	if wf.GetStatus() == contracts_workflow.WorkflowStatus_Complete {
		return wf.GetResponse(), nil
	}
	log := zerolog.Ctx(ctx).With().Logger()
	defer func() {
		err = wf.StoreState(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to store workflow state")
		}
	}()
	wf.SetStatus(contracts_workflow.WorkflowStatus_Running)
	resp, err = fn(ctx, wf, wf.GetInput())
	if err != nil {

		if !IsRetryableError(err) {
			wf.SetStatus(contracts_workflow.WorkflowStatus_Terminated)
		} else {
			wf.SetStatus(contracts_workflow.WorkflowStatus_PendingRetry)
		}
		wf.SetError(err)

		return nil, err
	}
	wf.SetStatus(contracts_workflow.WorkflowStatus_Complete)
	wf.SetResponse(resp)
	return resp, nil
}
func (s *service) StoreState(ctx context.Context) error {
	jsonB, err := s.ToJson()
	if err != nil {
		return err
	}
	_, err = s.store.SetWorkflowState(ctx,
		&contracts_workflow.SetWorkflowStateRequest{
			WorkflowID: s.WorkflowID,
			State:      jsonB,
		})
	return err
}

func (s *service) ExecuteActivity(ctx context.Context, fn contracts_workflow.ActivityExecutionFunc, request interface{}) (interface{}, error) {

	doFN := func(ctx context.Context, request interface{}) (interface{}, error) {
		response, err := fn(ctx, request)
		workflowError := NewErrorWrapperWorkflowError(err)
		funcName := ActivityName(fn)

		s.PushExecutionResponse(ctx,
			&ExecutionResponse{
				FuncName: funcName,
				Data:     response,
				Error:    workflowError,
			})
		s.CurrentExecutionResponseIndex = len(s.ExecutionResponses)
		if workflowError == nil {
			return response, nil
		}
		return response, workflowError
	}

	c := s.CurrentExecutionResponseIndex
	if c < len(s.ExecutionResponses) {
		// return the cached result
		er := s.ExecutionResponses[c]
		// is this a retryable error
		if fluffycore_utils.IsNil(er.Error) {
			// increment the index
			s.CurrentExecutionResponseIndex++
			return s.ExecutionResponses[c].Data, nil
		}
		if IsRetryableError(er.Error) {
			// clear out
			s.PopExecutionResponse(ctx)
			return doFN(ctx, request)
		}
		return s.ExecutionResponses[c].Data, er.Error
	}
	return doFN(ctx, request)
}

func (s *service) FromJson(jsonB []byte) error {
	store := s.store
	err := json.Unmarshal(jsonB, s)
	if err != nil {
		return err
	}
	s.store = store
	return nil
}

func (s *service) ToJson() ([]byte, error) {
	return json.Marshal(s)
}
func (s *service) ResetCurrentExecutionResponseIndex() {
	s.CurrentExecutionResponseIndex = 0
}
func (s *service) GetWorkflowID() string {
	return s.WorkflowID
}
func (s *service) SetWorkflowID(id string) {
	s.WorkflowID = id
}
func (s *service) GetWorkflowType() string {
	return s.WorkflowType
}
func (s *service) SetWorkflowType(workflowType string) {
	s.WorkflowType = workflowType
}
func (s *service) GetStatus() contracts_workflow.WorkflowStatus {
	return s.WorkflowStatus
}
func (s *service) SetStatus(status contracts_workflow.WorkflowStatus) {
	s.WorkflowStatus = status
}
func (s *service) GetResponse() interface{} {
	return s.WorkflowResponse
}
func (s *service) SetResponse(response interface{}) {
	s.WorkflowResponse = response
}
func (s *service) GetInput() interface{} {
	return s.WorkflowInput
}
func (s *service) SetInput(input interface{}) {
	s.WorkflowInput = input
}
func (s *service) GetError() error {
	return s.WorkflowError
}
func (s *service) SetError(err error) {
	if IsWorkflowError(err) {
		s.WorkflowError = err.(*WorkflowError)
	} else {
		s.WorkflowError = NewErrorWrapperWorkflowError(err)
	}
}

// SideEffectActivity is a generic function limited to PrimitiveConstraint types.
func SideEffectPrimitiveActivity(_ context.Context, request interface{}) (interface{}, error) {
	if !isPrimitive(request) {
		// TODO: This needs to be a non retryable error
		return nil, status.Error(codes.InvalidArgument, "request is not a primitive type")
	}
	return request, nil
}

// isPrimitive checks if the value is a Go primitive type.
func isPrimitive(value interface{}) bool {
	if value == nil { // nil is not considered a primitive in this context, but you can change if needed
		return false
	}

	switch value.(type) {
	// String
	case string:
		return true
	// Byte Slice
	case []byte:
		return true
	// Integers
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return true
	// Floats
	case float32, float64:
		return true
	// Boolean
	case bool:
		return true
	// Complex numbers (if you consider them primitives, otherwise remove)
	case complex64, complex128:
		return true
	default:
		// Check if it's a primitive kind using reflection for a more generic approach if needed.
		// Note: This might have performance implications if called very frequently.
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String,
			reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128:
			return true
		case reflect.Slice:
			if v.Type().Elem().Kind() == reflect.Uint8 { // Special case for []byte
				return true
			}
		}
		return false
	}
}
