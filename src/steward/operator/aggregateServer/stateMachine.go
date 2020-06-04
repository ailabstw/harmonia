package aggregateServer

import (
	"reflect"

	"harmonia.com/steward/operator/util"

	"go.uber.org/zap"
)

var StateTransit = util.StateTransit {
	reflect.TypeOf(idleState{}): {
		reflect.TypeOf(&roundStartAction{}): func (state *util.State, action util.Action) {
			// TODO: send onRoundStart
			roundStartAction := action.(*roundStartAction)
			*state = localTrainState{
				trainPlan: roundStartAction.trainPlan,
				roundRemain: roundStartAction.trainPlan.RoundCount,
				edgeModels: []string{},
			}
		},
	},
	reflect.TypeOf(localTrainState{}): {
		reflect.TypeOf(&localTrainFinishAction{}): func (state *util.State, action util.Action) {
			// TODO: model verification
			aggLocalTrainFinish := action.(*localTrainFinishAction)
			currentLocalTrainState := (*state).(localTrainState)

			inArray := func (arr []string, needle string) bool {
				for _, item := range arr {
					if item == needle {
						return true
					}
				}
				return false
			}

			if inArray(currentLocalTrainState.edgeModels, aggLocalTrainFinish.gitHttpURL) {
				return
			}

			err := util.PullData(aggLocalTrainFinish.gitHttpURL)

			currentLocalTrainState = (*state).(localTrainState) // State may change after pull
			if err != nil {
				zap.L().Error("pull edge model error", zap.Error(err))
			}
		
			if len(currentLocalTrainState.edgeModels) + 1 == currentLocalTrainState.trainPlan.EdgeCount {
				sendAggregateMessage()
				*state = aggregateState {
					trainPlan: currentLocalTrainState.trainPlan,
					roundRemain: currentLocalTrainState.roundRemain,
				}
			} else {
				*state = localTrainState{
					trainPlan: currentLocalTrainState.trainPlan,
					roundRemain: currentLocalTrainState.roundRemain,
					edgeModels: append(currentLocalTrainState.edgeModels, aggLocalTrainFinish.gitHttpURL),
				}
			}
		},
	},
	reflect.TypeOf(aggregateState{}): {
		reflect.TypeOf(&aggregateFinishAction{}): func(state *util.State, action util.Action) {
			aggregateState := (*state).(aggregateState)
			if aggregateState.roundRemain - 1 == 0 {
				hash := aggregateState.trainPlan.PlanHash
				err := util.PushUpdates(util.Config.AggregatedModelRepo.GitHttpURL, "inference-" + hash)
				if err != nil {
					zap.L().Error("push aggregated model error", zap.Error(err))
				}
				// Finish a round
				*state = idleState{}
			} else {
				err := util.PushUpdates(util.Config.AggregatedModelRepo.GitHttpURL, "")
				if err != nil {
					zap.L().Error("push aggregated model error", zap.Error(err))
				}
				*state = localTrainState {
					trainPlan: aggregateState.trainPlan,
					roundRemain: aggregateState.roundRemain - 1,
					edgeModels: []string{},
				}
			}
		},
	},
}

var InitState = idleState{}
