package edge

import (
	// "fmt"
	"reflect"
	"google.golang.org/grpc"

	"testing"
	// "github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"

	"harmonia.com/steward/operator/util"
)

type mockOperator struct{}
	 
func (*mockOperator) RemoteNotificationRegister(util.NotificationParam) {}
func (*mockOperator) GrpcServerRegister(*grpc.Server) {}
func (*mockOperator) Dispatch(util.Action) {}
func (*mockOperator) GetPayload() interface{} {
	return nil
}
func (*mockOperator) TrainFinish() {}

// TODO: transit
func TestStateTransit(t *testing.T) {
	mockTrainPlan := util.TrainPlan{
		Name: "fakePlan",
		RoundCount: 1,
		EdgeCount: 2,
		EpochCount: 3,
		CommitID: "foo",
		PretrainedModelCommitID: "bar",
	}

	testCases := map[string]struct{
		fromState util.State
		action util.Action
		operator util.AbstractOperator
		toState util.State
	} {
		"idleState_tranPlanAction": {
			idleState{},
			&trainPlanAction {
				trainPlan: mockTrainPlan,
			},
			&mockOperator{},
			trainInitState {
				init: false,
				pretrainedModel: false,
				trainPlan: mockTrainPlan,
			},
		},
		"trainInitState_initMessageResponseAction": {
			trainInitState {
				init: false,
				pretrainedModel: false,
				trainPlan: mockTrainPlan,
			},
			&initMessageResponseAction {},
			// TODO: trainInitState Dispatch trainStartAction
			&mockOperator{},
			trainInitState {
				init: true,
				pretrainedModel: false,
				trainPlan: mockTrainPlan,
			},
		},
		"trainInitState_baseModelReceivedAction": {
			trainInitState {
				init: false,
				pretrainedModel: false,
				trainPlan: mockTrainPlan,
			},
			&baseModelReceivedAction {},
			// TODO: trainInitState Dispatch trainStartAction
			&mockOperator{},
			trainInitState {
				init: false,
				pretrainedModel: false,
				trainPlan: mockTrainPlan,
			},
		},
		"trainInitState_pretrainedModelReadyAction": {
			trainInitState {
				init: false,
				pretrainedModel: false,
				trainPlan: mockTrainPlan,
			},
			&pretrainedModelReadyAction {},
			// TODO: trainInitState Dispatch trainStartAction
			&mockOperator{},
			trainInitState {
				init: false,
				pretrainedModel: true,
				trainPlan: mockTrainPlan,
			},
		},
		"trainInitState_trainStartAction": {
			trainInitState {
				init: true,
				pretrainedModel: true,
				trainPlan: mockTrainPlan,
			},
			&trainStartAction {},
			&mockOperator{},
			localTrainState {
				trainPlan: mockTrainPlan,
				roundCount: 0,
			},
		},
	}

	for caseName, testCase := range testCases {
		_, ok := StateTransit[reflect.TypeOf(testCase.fromState)][reflect.TypeOf(testCase.action)]
		assert.True(t, ok, "Case [%s] fails.", caseName)

		toState, _ := StateTransit[reflect.TypeOf(testCase.fromState)][reflect.TypeOf(testCase.action)](
			testCase.fromState,
			testCase.action,
			testCase.operator,
		)
		assert.Equal(t, testCase.toState, toState, "Case [%s] fails.", caseName)
	}
}