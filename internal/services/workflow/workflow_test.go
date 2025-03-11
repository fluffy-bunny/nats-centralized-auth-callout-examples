package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	contracts_workflow "natsauth/internal/contracts/workflow"
	services_workflow_store_inmemory_store "natsauth/internal/services/workflow_store/inmemory_store"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	require "github.com/stretchr/testify/require"
)

type MyInput struct {
	Name string
}
type MyResponse struct {
	Name string
}
type MyWorkflowResponse struct {
	Responses []interface{}
}

func myAcivity(_ context.Context, request interface{}) (interface{}, error) {
	myInput := request.(*MyInput)
	myResponse := &MyResponse{
		Name: myInput.Name,
	}
	return myResponse, nil
}
func myWF(ctx context.Context, wf contracts_workflow.IWorkflow, request interface{}) (interface{}, error) {
	myInput := request.(*MyInput)

	myWorkflowResponse := &MyWorkflowResponse{}
	mm, err := wf.ExecuteActivity(ctx, myAcivity, myInput)
	if err != nil {
		return nil, err
	}
	myWorkflowResponse.Responses = append(myWorkflowResponse.Responses, mm)

	mm, err = wf.ExecuteActivity(ctx, SideEffectPrimitiveActivity, 1234)
	if err != nil {
		return nil, err
	}

	myWorkflowResponse.Responses = append(myWorkflowResponse.Responses, mm)

	for i := 0; i < 10; i++ {
		mi := &MyInput{
			Name: fmt.Sprintf("John %d", i),
		}
		mm, err = wf.ExecuteActivity(ctx, myAcivity, mi)
		if err != nil {
			return nil, err
		}
		myWorkflowResponse.Responses = append(myWorkflowResponse.Responses, mm)
	}
	return myWorkflowResponse, nil
}
func TestWorkflowInstance(t *testing.T) {
	builder := di.Builder()
	AddTransientWorkflow(builder)
	services_workflow_store_inmemory_store.AddSingletonWorkflowStore(builder)
	ctn := builder.Build()

	wf, err := di.TryGet[contracts_workflow.IWorkflow](ctn)
	require.NoError(t, err)
	require.NotNil(t, wf)

	ctx := context.Background()
	wf.SetWorkflowID("0001")
	wf.SetWorkflowType("wf_test")
	wf.SetInput(&MyInput{
		Name: "John",
	})
	response, err := ExecuteWorkflow(ctx, wf, myWF)
	require.NoError(t, err)
	require.NotNil(t, response)

	jsonB, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotEmpty(t, jsonB)

	wfToGeneric := func(wf contracts_workflow.IWorkflow) (map[string]interface{}, error) {
		wfJsonB, err := wf.ToJson()
		if err != nil {
			return nil, err
		}

		generic := map[string]interface{}{}
		err = json.Unmarshal(wfJsonB, &generic)
		if err != nil {
			return nil, err
		}
		return generic, nil
	}

	generic, err := wfToGeneric(wf)
	require.NoError(t, err)
	fmt.Println(fluffycore_utils.PrettyJSON(generic))

	// we now have a wf that is fully run, lets rerun it and see if the caches all work.
	wf.ResetCurrentExecutionResponseIndex()
	response, err = ExecuteWorkflow(ctx, wf, myWF)
	require.NoError(t, err)
	require.NotNil(t, response)

	generic, err = wfToGeneric(wf)
	require.NoError(t, err)
	fmt.Println(fluffycore_utils.PrettyJSON(generic))

	jsonB, err = wf.ToJson()
	require.NoError(t, err)
	require.NotEmpty(t, jsonB)

	wf2 := di.Get[contracts_workflow.IWorkflow](ctn)
	err = wf2.FromJson(jsonB)
	require.NoError(t, err)
	wf2.ResetCurrentExecutionResponseIndex()
	response, err = ExecuteWorkflow(ctx, wf2, myWF)
	require.NoError(t, err)
	require.NotNil(t, response)
	generic, err = wfToGeneric(wf2)
	require.NoError(t, err)
	fmt.Println(fluffycore_utils.PrettyJSON(generic))
}
