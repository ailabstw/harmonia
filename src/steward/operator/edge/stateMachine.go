package edge

import (
	"reflect"

	"go.uber.org/zap"

	"harmonia.com/steward/operator/util"
)

var StateTransit = util.StateTransit{
	reflect.TypeOf(idleState{}): {
		reflect.TypeOf(&trainStartAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			trainStartAction := action.(*trainStartAction)
			if (util.TrainPlan{}) == trainStartAction.trainPlan {
				zap.L().Warn("train plan is empty")
			}
			return localTrainState{
				roundRemain: trainStartAction.trainPlan.RoundCount,
				trainPlan:   trainStartAction.trainPlan,
			}, func() {
				sendLocalTrainMessage(
					trainStartAction.trainPlan.EpochCount,
					operator.GetPayload().(Payload).GrpcServerURI,
					operator.GetPayload().(Payload).EdgeModelRepoGitHttpURL,
					operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
				)
			}
		},
		reflect.TypeOf(&aggregatedModelReceivedAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			return idleState{}, func() {
				zap.L().Debug("pull aggregated model")
				util.PullData(operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL)
			}
		},
	},
	reflect.TypeOf(localTrainState{}): {
		reflect.TypeOf(&trainFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			pushModel := func () {
				err := util.PushUpdates(operator.GetPayload().(Payload).EdgeModelRepoGitHttpURL, "")
				if err != nil {
					zap.L().Error("push edge model error", zap.Error(err))
				}
			}

			localTrainState := state.(localTrainState)
			if localTrainState.roundRemain-1 == 0 {
				return idleState{}, pushModel
			} else {
				if (util.TrainPlan{}) == localTrainState.trainPlan {
					zap.L().Warn("train plan is empty")
				}
				return aggregateState{
					roundRemain: localTrainState.roundRemain - 1,
					trainPlan:   localTrainState.trainPlan,
				}, pushModel
			}
		},
	},
	reflect.TypeOf(aggregateState{}): {
		reflect.TypeOf(&aggregatedModelReceivedAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			aggregateState := state.(aggregateState)
			if (util.TrainPlan{}) == aggregateState.trainPlan {
				zap.L().Warn("train plan is empty")
			}
			return localTrainState{
				roundRemain: aggregateState.roundRemain,
				trainPlan:   aggregateState.trainPlan,
			}, func() {
				zap.L().Debug("pull aggregated model")
				util.PullData(operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL)
				sendLocalTrainMessage(
					aggregateState.trainPlan.EpochCount,
					operator.GetPayload().(Payload).GrpcServerURI,
					operator.GetPayload().(Payload).EdgeModelRepoGitHttpURL,
					operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
				)
			}
		},
	},
}

var InitState = idleState{}
