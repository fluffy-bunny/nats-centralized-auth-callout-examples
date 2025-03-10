package workflow

import (
	"fmt"
	"testing"

	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	require "github.com/stretchr/testify/require"
)

var dd = `
{
    "id": "0001",
    "name": "wf_test",
    "rootNode": {
        "name": "test",
        "type": 0,
        "nodes": [
            {
                "name": "step2",
                "type": 0,
                "nodes": [
                    {
                        "name": "step3",
                        "type": 0,
                        "status": 0
                    }
                ],
                "status": 0
            }
        ],
        "status": 0,
        "executionResult": {
            "data": "eyJOYW1lIjoiSm9obiIsIkFnZSI6MzB9",
            "hintType": "MyResult"
        },
        "input": {
            "data": "eyJOYW1lIjoiSm9obiIsIkFnZSI6MzB9",
            "hintType": "MyResult"
        },
        "metadata": {
            "a": 1,
            "b": "test",
            "c": [
                "a",
                "b"
            ],
            "d": {
                "e": 1,
                "f": "test"
            }
        }
    }
}`

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

	workflow := &Workflow{
		ID:   "0001",
		Name: "wf_test",
	}
	rootNode := &Node{
		Name:   "test",
		Type:   NodeType_Root | NodeType_Execution,
		Status: NodeStatus_UNSPECIFIED,
		Nodes:  []*Node{},
		Metadata: map[string]interface{}{
			"a": 1,
			"b": "test",
			"c": []string{"a", "b"},
			"d": map[string]interface{}{
				"e": 1,
				"f": "test",
			},
		},
	}
	workflow.RootNode = rootNode

	workflow.RootNode.SetInput(input)
	workflow.RootNode.SetExecutionResult(executionResult)
	step2Node := &Node{
		Name:   "step2",
		Type:   NodeType_Execution,
		Status: NodeStatus_UNSPECIFIED,
		Nodes:  []*Node{},
	}
	workflow.RootNode.Nodes = append(workflow.RootNode.Nodes, step2Node)
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
	workflow.RootNode.FixupParentNode()
	fmt.Println(fluffycore_utils.PrettyJSON(workflow))

	workflow2, err := WorkflowFromJson(jsonB)
	require.NoError(t, err)
	require.NotNil(t, workflow2)

	require.Equal(t, workflow.ID, workflow2.ID)

	require.Equal(t, workflow.RootNode.ExecutionResult.Data, workflow2.RootNode.ExecutionResult.Data)
	require.Equal(t, workflow.RootNode.ExecutionResult.HintType, workflow2.RootNode.ExecutionResult.HintType)

	fmt.Println(fluffycore_utils.PrettyJSON(workflow2))

}
