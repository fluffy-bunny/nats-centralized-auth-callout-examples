package workflow

import (
	"fmt"
	"testing"

	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	require "github.com/stretchr/testify/require"
)

func TestNodeWorkflowJson(t *testing.T) {
	type MyResult struct {
		Name string
		Age  int
	}
	u := &MyResult{
		Name: "John",
		Age:  30,
	}
	input, err := NewInput(u)
	require.NoError(t, err)

	executionResult, err := NewExecutionResult(u, nil)
	require.NoError(t, err)

	workflow := &Node{
		ID:     "0001",
		Name:   "test",
		Type:   NodeType_Root | NodeType_Execution,
		Status: NodeStatus_UNSPECIFIED,
		Nodes:  []*Node{},
	}
	workflow.SetInput(input)
	workflow.SetExecutionResult(executionResult)
	step2Node := &Node{
		Name:   "step2",
		Type:   NodeType_Execution,
		Status: NodeStatus_UNSPECIFIED,
		Nodes:  []*Node{},
	}
	workflow.Nodes = append(workflow.Nodes, step2Node)
	step3Node := &Node{
		Name:   "step3",
		Type:   NodeType_Execution | NodeType_Final,
		Status: NodeStatus_UNSPECIFIED,
		Nodes:  []*Node{},
	}
	step2Node.Nodes = append(step2Node.Nodes, step3Node)

	jsonB, err := workflow.Marshal()
	require.NoError(t, err)
	require.NotEmpty(t, jsonB)
	workflow.FixupParentNode()
	fmt.Println(fluffycore_utils.PrettyJSON(workflow))

	workflow2, err := WorkflowFromJson(jsonB)
	require.NoError(t, err)
	require.NotNil(t, workflow2)

	require.Equal(t, workflow.ID, workflow2.ID)

	require.Equal(t, workflow.ExecutionResult.Data, workflow2.ExecutionResult.Data)
	require.Equal(t, workflow.ExecutionResult.HintType, workflow2.ExecutionResult.HintType)

	fmt.Println(fluffycore_utils.PrettyJSON(workflow2))

}
