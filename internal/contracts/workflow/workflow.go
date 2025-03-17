package workflow

import (
	"context"
)

type ExecutionResultState uint

const (
	ExecutionResultState_Undefined ExecutionResultState = 0
	ExecutionResultState_Complete  ExecutionResultState = 1
	ExecutionResultState_Running   ExecutionResultState = 2
	ExecutionResultState_Failed    ExecutionResultState = 3
)

// NodeType defines the type of a workflow node.

type NodeStatus int

const (
	NodeStatus_UNSPECIFIED NodeStatus = 0
	NodeStatus_Complete    NodeStatus = 1
	NodeStatus_Running     NodeStatus = 2
	NodeStatus_Terminated  NodeStatus = 3
)

type WorkflowStatus int

const (
	WorkflowStatus_UNSPECIFIED  WorkflowStatus = 0
	WorkflowStatus_Complete     WorkflowStatus = 1
	WorkflowStatus_Running      WorkflowStatus = 2
	WorkflowStatus_Terminated   WorkflowStatus = 3
	WorkflowStatus_PendingRetry WorkflowStatus = 4
)

type NodeType int

const (
	NodeType_UNSPECIFIED NodeType = 0 << iota // 0 (binary 0000)
	NodeType_Root                             // 1 (binary 0001)
	NodeType_Final                            // 2 (binary 0010)
	NodeType_Execution                        // 4 (binary 0100)
)

type (
	NATSWorkflowStoreConfig struct {
		Bucket string
	}
	GetWorkflowStateRequest struct {
		WorkflowID string
	}
	GetWorkflowStateRequestResponse struct {
		WorkflowID string
		State      []byte
	}
	SetWorkflowStateRequest struct {
		WorkflowID string
		State      []byte
	}
	SetWorkflowStateRequestResponse struct {
		WorkflowID string
	}
	IWorkflowCache interface {
		SetWorkflowState(ctx context.Context, request *SetWorkflowStateRequest) (*SetWorkflowStateRequestResponse, error)
		GetWorkflowState(ctx context.Context, request *GetWorkflowStateRequest) (*GetWorkflowStateRequestResponse, error)
	}
	IWorkflowManager interface {
		CreateWorkflowType(ctx context.Context, request interface{}) (interface{}, error)
		ExecuteWorkflow(ctx context.Context, request interface{}) (interface{}, error)
	}

	WorkflowExecutionFunc func(ctx context.Context, wf IWorkflow, request interface{}) (interface{}, error)
	ActivityExecutionFunc func(ctx context.Context, request interface{}) (interface{}, error)

	IWorkflow interface {
		// ResetCurrentExecutionResponseIndex sets the current execution response index to 0
		// so we can run the workflow again from the top
		ResetCurrentExecutionResponseIndex()
		GetWorkflowID() string
		SetWorkflowID(id string)
		GetWorkflowType() string
		SetWorkflowType(workflowType string)
		GetStatus() WorkflowStatus
		SetStatus(status WorkflowStatus)
		GetResponse() interface{}
		SetResponse(response interface{})
		GetInput() interface{}
		SetInput(input interface{})
		GetError() error
		SetError(err error)
		ToJson() ([]byte, error)
		FromJson(jsonB []byte) error
		ExecuteActivity(ctx context.Context, fn ActivityExecutionFunc, request interface{}) (interface{}, error)
		StoreState(ctx context.Context) error
	}
	IWorkflowExecutorFunc interface {
		WorkflowExecutionFunc
	}
)
