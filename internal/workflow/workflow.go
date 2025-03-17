package workflow

import (
	"errors"
	"fmt"
	"math/rand"
)

// NodeType defines the type of a workflow node.
type NodeType string

const (
	NodeTypeTask     NodeType = "task"
	NodeTypeDecision NodeType = "decision"
	NodeTypeSuccess  NodeType = "success"
	NodeTypeFailure  NodeType = "failure"
)

// TaskFunction is a function type for task nodes.
// It takes a context map and returns an outcome string and an error.
// The outcome string is used to determine the next node in the workflow.
type TaskFunction func(context map[string]interface{}) (outcome string, err error)

// DecisionFunction is a function type for decision nodes.
// It takes a context map and returns an outcome string and an error.
// The outcome string is used to determine the next node in the workflow.
type DecisionFunction func(context map[string]interface{}) (outcome string, err error)

// Node represents a node in the workflow graph.
type Node struct {
	ID          string            `json:"id"`
	NodeType    NodeType          `json:"nodeType"`
	Description string            `json:"description,omitempty"`
	Task        TaskFunction      `json:"-"`                     // Exclude from JSON serialization, function pointers are not directly serializable
	Decision    DecisionFunction  `json:"-"`                     // Exclude from JSON serialization
	NextNodeIDs map[string]string `json:"nextNodeIDs,omitempty"` // Map of outcome/condition to next Node ID
}

// Workflow represents the workflow graph.
type Workflow struct {
	Nodes         map[string]*Node `json:"nodes"`         // Map of node IDs to Node objects
	StartNodeID   string           `json:"startNodeID"`   // ID of the starting node
	SuccessNodeID string           `json:"successNodeID"` // ID of the success node
	FailureNodeID string           `json:"failureNodeID"` // ID of the failure node
}

// NewWorkflow creates a new workflow.
func NewWorkflow(startNodeID, successNodeID, failureNodeID string) *Workflow {
	return &Workflow{
		Nodes:         make(map[string]*Node),
		StartNodeID:   startNodeID,
		SuccessNodeID: successNodeID,
		FailureNodeID: failureNodeID,
	}
}

// AddNode adds a node to the workflow.
func (wf *Workflow) AddNode(node *Node) {
	wf.Nodes[node.ID] = node
}

// ExecuteWorkflow executes the workflow starting from the StartNode.
func ExecuteWorkflow(wf *Workflow, initialContext map[string]interface{}) (string, error) {
	currentNodeID := wf.StartNodeID
	context := initialContext

	for {
		currentNode, ok := wf.Nodes[currentNodeID]
		if !ok {
			return wf.FailureNodeID, fmt.Errorf("node with ID '%s' not found", currentNodeID)
		}

		switch currentNode.NodeType {
		case NodeTypeTask:
			if currentNode.Task != nil {
				outcome, err := currentNode.Task(context)
				if err != nil {
					return wf.FailureNodeID, fmt.Errorf("task node '%s' failed: %w", currentNodeID, err)
				}
				if nextNodeID, ok := currentNode.NextNodeIDs[outcome]; ok {
					currentNodeID = nextNodeID
				} else {
					return wf.FailureNodeID, fmt.Errorf("no next node defined for outcome '%s' in task node '%s'", outcome, currentNodeID)
				}
			} else {
				return wf.FailureNodeID, fmt.Errorf("task node '%s' has no task function", currentNodeID)
			}

		case NodeTypeDecision:
			if currentNode.Decision != nil {
				outcome, err := currentNode.Decision(context)
				if err != nil {
					return wf.FailureNodeID, fmt.Errorf("decision node '%s' failed: %w", currentNodeID, err)
				}
				if nextNodeID, ok := currentNode.NextNodeIDs[outcome]; ok {
					currentNodeID = nextNodeID
				} else {
					return wf.FailureNodeID, fmt.Errorf("no next node defined for decision outcome '%s' in node '%s'", outcome, currentNodeID)
				}
			} else {
				return wf.FailureNodeID, fmt.Errorf("decision node '%s' has no decision function", currentNodeID)
			}

		case NodeTypeSuccess:
			return wf.SuccessNodeID, nil // Workflow Success!

		case NodeTypeFailure:
			return wf.FailureNodeID, nil // Workflow Failure!

		default:
			return wf.FailureNodeID, fmt.Errorf("unknown node type '%s' in node '%s'", currentNode.NodeType, currentNodeID)
		}
	}
}

func Test() {
	// Define Task Functions
	taskA := func(context map[string]interface{}) (string, error) {
		fmt.Println("Executing Task A...")
		// Simulate some work and outcome
		if randBool() {
			context["taskAResult"] = "success"
			return "success", nil
		} else {
			context["taskAResult"] = "failure"
			return "failure", errors.New("task A failed")
		}
	}

	taskB := func(context map[string]interface{}) (string, error) {
		fmt.Println("Executing Task B...")
		fmt.Println("Context:", context)
		return "success", nil
	}

	decisionNode := func(context map[string]interface{}) (string, error) {
		fmt.Println("Executing Decision Node...")
		if result, ok := context["taskAResult"].(string); ok && result == "success" {
			fmt.Println("Decision: Task A was successful.")
			return "task_a_success", nil
		} else {
			fmt.Println("Decision: Task A was not successful.")
			return "task_a_failure", nil
		}
	}

	// Create Workflow
	wf := NewWorkflow("start", "success", "failure")

	// Create Nodes
	startNode := &Node{ID: "start", NodeType: NodeTypeTask, Description: "Start Task", Task: taskA, NextNodeIDs: map[string]string{"success": "decision", "failure": "failure"}}
	decisionNodeObj := &Node{ID: "decision", NodeType: NodeTypeDecision, Description: "Decision after Task A", Decision: decisionNode, NextNodeIDs: map[string]string{"task_a_success": "task_b", "task_a_failure": "failure"}}
	taskBNode := &Node{ID: "task_b", NodeType: NodeTypeTask, Description: "Task B", Task: taskB, NextNodeIDs: map[string]string{"success": "success"}}
	successNode := &Node{ID: "success", NodeType: NodeTypeSuccess, Description: "Success Node"}
	failureNode := &Node{ID: "failure", NodeType: NodeTypeFailure, Description: "Failure Node"}

	// Add Nodes to Workflow
	wf.AddNode(startNode)
	wf.AddNode(decisionNodeObj)
	wf.AddNode(taskBNode)
	wf.AddNode(successNode)
	wf.AddNode(failureNode)

	// Execute Workflow
	initialContext := make(map[string]interface{})
	finalNodeID, err := ExecuteWorkflow(wf, initialContext)

	if err != nil {
		fmt.Printf("\nWorkflow Execution Finished in Node: %s with error: %v\n", finalNodeID, err)
	} else {
		fmt.Printf("\nWorkflow Execution Finished in Node: %s successfully.\n", finalNodeID)
	}
}

func randBool() bool {
	return rand.Float64() < 0.5
}
