package aggregateServer

import (
	"reflect"

	"harmonia.com/steward/operator/util"

	"go.uber.org/zap"
)

var StateTransit = util.StateTransit {
	reflect.TypeOf(idleState{}): {
		reflect.TypeOf(&roundStartAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			// TODO: send onRoundStart
			roundStartAction := action.(*roundStartAction)
			return localTrainState{
				trainPlan: roundStartAction.trainPlan,
				roundRemain: roundStartAction.trainPlan.RoundCount,
				edgeModels: []string{},
				webhooks: []string{},
			}, nil
		},
	},
	reflect.TypeOf(localTrainState{}): {
		reflect.TypeOf(&localTrainFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			// TODO: model verification
			aggLocalTrainFinish := action.(*localTrainFinishAction)
			currentLocalTrainState := state.(localTrainState)

			inArray := func (arr []string, needle string) bool {
				for _, item := range arr {
					if item == needle {
						return true
					}
				}
				return false
			}

			if inArray(currentLocalTrainState.webhooks, aggLocalTrainFinish.gitHttpURL) {
				zap.L().Debug("this webhook is already in webhooks array")
				return currentLocalTrainState, nil
			}

			return localTrainState{
				trainPlan:   currentLocalTrainState.trainPlan,
				roundRemain: currentLocalTrainState.roundRemain,
				edgeModels:  currentLocalTrainState.edgeModels,
				webhooks:    append(currentLocalTrainState.webhooks, aggLocalTrainFinish.gitHttpURL),
			}, func() {
				err := util.PullData(aggLocalTrainFinish.gitHttpURL)
				if err != nil {
					zap.L().Fatal("pull repository error",
						zap.String("git http url", aggLocalTrainFinish.gitHttpURL),
						zap.Error(err))
				} else {
					operator.Dispatch(&pullFinishAction{gitHttpURL: aggLocalTrainFinish.gitHttpURL})
				}
			}
		},
		reflect.TypeOf(&pullFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			aggPullFinishAction := action.(*pullFinishAction)
			currentLocalTrainState := state.(localTrainState)
			if len(currentLocalTrainState.edgeModels) + 1 == currentLocalTrainState.trainPlan.EdgeCount {
				return aggregateState{
					trainPlan:   currentLocalTrainState.trainPlan,
					roundRemain: currentLocalTrainState.roundRemain,
				}, func() {
					sendAggregateMessage(
						operator.GetPayload().(Payload).GrpcServerURI,
						operator.GetPayload().(Payload).EdgeModelRepoGitHttpURLs,
						operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
					)
				}
			} else {
				return localTrainState{
					trainPlan:   currentLocalTrainState.trainPlan,
					roundRemain: currentLocalTrainState.roundRemain,
					edgeModels:  append(currentLocalTrainState.edgeModels, aggPullFinishAction.gitHttpURL),
					webhooks:    currentLocalTrainState.webhooks,
				}, nil
			}
		},
	},
	reflect.TypeOf(aggregateState{}): {
		reflect.TypeOf(&aggregateFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			aggregateState := state.(aggregateState)
			operatorPayload := operator.GetPayload().(Payload)
			if aggregateState.roundRemain - 1 == 0 {
				// Finish a round
				return idleState{}, func() {
					hash := aggregateState.trainPlan.PlanHash
					err := util.PushUpdates(operatorPayload.AggregatedModelRepoGitHttpURL, "inference-" + hash)
					if err != nil {
						zap.L().Error("push aggregated model error", zap.Error(err))
					}
				}
			} else {
				return localTrainState{
					trainPlan:   aggregateState.trainPlan,
					roundRemain: aggregateState.roundRemain - 1,
					edgeModels:  []string{},
					webhooks:    []string{},
				}, func() {
					err := util.PushUpdates(operatorPayload.AggregatedModelRepoGitHttpURL, "")
					if err != nil {
						zap.L().Error("push aggregated model error", zap.Error(err))
					}
				}
			}
		},
	},
}

var InitState = idleState{}
