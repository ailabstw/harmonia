package edge

import (
	"reflect"

	"harmonia.com/steward/operator/util"

	"go.uber.org/zap"
)

var StateTransit = util.StateTransit{
	reflect.TypeOf(idleState{}): {
		reflect.TypeOf(&trainStartAction{}): func(state *util.State, action util.Action) {
			sendLocalTrainMessage()
			trainStartAction := action.(*trainStartAction)
			*state = localTrainState{
				roundRemain: trainStartAction.trainPlan.RoundCount}
		},
		reflect.TypeOf(&aggregatedModelReceivedAction{}): func(state *util.State, action util.Action) {
			err := util.PullData(util.Config.AggregatedModelRepo.GitHttpURL)
			if err != nil {
				zap.L().Error("pull aggregated model error", zap.Error(err))
			}
			*state = idleState{}
		},
	},
	reflect.TypeOf(localTrainState{}): {
		reflect.TypeOf(&trainFinishAction{}): func(state *util.State, action util.Action) {
			err := util.PushUpdates(util.Config.EdgeModelRepo.GitHttpURL, "")
			if err != nil {
				zap.L().Error("push edge model error", zap.Error(err))
			}
			localTrainState := (*state).(localTrainState)
			if localTrainState.roundRemain - 1 == 0 {
				*state = idleState{}
			} else {
				*state = aggregateState{
					roundRemain: localTrainState.roundRemain - 1}
			}
		},
	},
	reflect.TypeOf(aggregateState{}): {
		reflect.TypeOf(&aggregatedModelReceivedAction{}): func(state *util.State, action util.Action) {
			err := util.PullData(util.Config.AggregatedModelRepo.GitHttpURL)
			if err != nil {
				zap.L().Error("pull aggregated model error", zap.Error(err))
			}
			sendLocalTrainMessage()
			aggregateState := (*state).(aggregateState)
			*state = localTrainState{
				roundRemain: aggregateState.roundRemain}
		},
	},
}

var InitState = idleState{}
