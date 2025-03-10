package workflow

import (
	"context"
	"encoding/json"
	"reflect"
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
	IWorkflowStore interface {
		SetWorkflowState(ctx context.Context, request *SetWorkflowStateRequest) (*SetWorkflowStateRequestResponse, error)
		GetWorkflowState(ctx context.Context, request *GetWorkflowStateRequest) (*GetWorkflowStateRequestResponse, error)
	}
	NodeInput struct {
		Data     []byte `json:"data,omitempty"`
		HintType string `json:"hintType,omitempty"`
	}
	NodeExecutionResult struct {
		Error    string `json:"error,omitempty"`
		Data     []byte `json:"data,omitempty"`
		HintType string `json:"hintType,omitempty"`
	}
	Node struct {
		// ID would be the workflow ID for the root node
		// omit for all other nodes if not present
		ID              string               `json:"id,omitempty"`
		Name            string               `json:"name"`
		Type            NodeType             `json:"type"`
		Nodes           []*Node              `json:"nodes,omitempty"`
		Status          NodeStatus           `json:"status"`
		ExecutionResult *NodeExecutionResult `json:"executionResult,omitempty"`
		Input           *NodeInput           `json:"input,omitempty"`
		// NodeHandler is a lookup key to a INodeFunc
		NodeHandler string `json:"nodeHandler,omitempty"`
		// ParentNode is put in when we unmarshal the json
		ParentNode *Node `json:"-"`
	}

	IExecutionResult interface {
		State() ExecutionResultState
	}
	INodeFunc interface {
		Func(ctx context.Context, node INode) error
	}
	INode interface {
		INodeFunc
		ExecutionResult() IExecutionResult
		SetState(state ExecutionResultState)
		GetState() ExecutionResultState
	}

	IDecisionNode interface {
		INode
	}
	ITaskNode interface {
		INode
	}
)

func NewExecutionResult[T any](data *T, err error) (*NodeExecutionResult, error) {
	nr := &NodeExecutionResult{}
	if data != nil {

		hintType := reflect.TypeOf(data).Elem().Name()

		jsonB, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		nr.Data = jsonB
		nr.HintType = hintType
	}
	if err != nil {
		nr.Error = err.Error()
	}

	return nr, nil
}
func NewInput[T any](data *T) (*NodeInput, error) {
	nr := &NodeInput{}
	if data != nil {
		hintType := reflect.TypeOf(data).Elem().Name()

		jsonB, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		nr.Data = jsonB
		nr.HintType = hintType
	}

	return nr, nil
}
func (w *Node) SetExecutionResult(result *NodeExecutionResult) {
	w.ExecutionResult = result
}
func (w *Node) SetInput(input *NodeInput) {
	w.Input = input
}
func (w *Node) FixupParentNode() {
	for _, n := range w.Nodes {
		n.ParentNode = w
		n.FixupParentNode()
	}
}

func (w *Node) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

func WorkflowFromJson(jsonB []byte) (*Node, error) {
	n := &Node{}
	err := json.Unmarshal(jsonB, n)
	if err != nil {
		return nil, err
	}
	n.FixupParentNode()
	return n, nil
}

type NodePredicate func(n *Node) bool
type NodePredicateOption struct {
	Name   *string
	Type   *NodeType
	Status *NodeStatus
}
type WithNodePredicateOption func(o *NodePredicateOption)

func NewNodePredicationOption(b ...WithNodePredicateOption) *NodePredicateOption {
	o := &NodePredicateOption{}
	for _, f := range b {
		f(o)
	}
	return o
}

func WithNodePredicateOptionName(name string) WithNodePredicateOption {
	return func(o *NodePredicateOption) {
		o.Name = &name
	}
}

func WithNodePredicateOptionType(nodeType NodeType) WithNodePredicateOption {
	return func(o *NodePredicateOption) {
		o.Type = &nodeType
	}
}
func WithNodePredicateOptionStatus(status NodeStatus) WithNodePredicateOption {
	return func(o *NodePredicateOption) {
		o.Status = &status
	}
}

func NodePredicateFactory(o *NodePredicateOption) NodePredicate {
	return func(n *Node) bool {
		if o.Name != nil && *o.Name != n.Name {
			return false
		}
		if o.Type != nil && *o.Type != n.Type {
			return false
		}
		if o.Status != nil && *o.Status != n.Status {
			return false
		}
		return true
	}
}
func FindNode(w *Node, predicate NodePredicate) *Node {
	if predicate(w) {
		return w
	}
	for _, n := range w.Nodes {
		if found := FindNode(n, predicate); found != nil {
			return found
		}
	}
	return nil
}
