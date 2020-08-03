package edge

import (
	"reflect"
	"fmt"

	"go.uber.org/zap"

	"harmonia.com/steward/operator/util"
)

var StateTransit = util.StateTransit{
	reflect.TypeOf(idleState{}): {
		reflect.TypeOf(&trainPlanAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			trainPlanAction := action.(*trainPlanAction)
			if (util.TrainPlan{}) == trainPlanAction.trainPlan {
				zap.L().Warn("train plan is empty")
			}

			return trainInitState {
				init: false,
				pretrainedModel: false,
				trainPlan: trainPlanAction.trainPlan,
			}, func() {
				// Synchonous message
				sendInitMessage(operator.GetPayload().(Payload).GrpcServerURI)
				operator.Dispatch(&initMessageResponseAction{})
			}
		},
	},
	reflect.TypeOf(trainInitState{}): {
		reflect.TypeOf(&trainStartAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			trainInitState := state.(trainInitState)
			return localTrainState {
				roundCount: 0,
				trainPlan: trainInitState.trainPlan,
			}, func() {
				repoMetadata, err := util.ReadMetadata(operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL)
				if err != nil {
					zap.L().Fatal("Cannot read repoMetadata", zap.Error(err))
					return
				}

				metadata := map[string]string{}
				for key, val := range repoMetadata["metadata"].(map[string]interface{}) {
					metadata[key] = val.(string)
				}
				metrics := map[string]float64{}
				for key, val := range repoMetadata["metrics"].(map[string]interface{}) {
					metrics[key] = val.(float64)
				}

				sendLocalTrainMessage(
					operator.GetPayload().(Payload).GrpcServerURI,
					trainInitState.trainPlan.EpochCount,
					baseModel {
						gitHttpURL: operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
						metadata: metadata,
						metrics: metrics,
					},
					operator.GetPayload().(Payload).EdgeModelRepoGitHttpURL,
				)
			}
		},
		reflect.TypeOf(&initMessageResponseAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			currentState := state.(trainInitState)
			if !currentState.pretrainedModel {
				return trainInitState {
					init: true,
					pretrainedModel: currentState.pretrainedModel,
					trainPlan: currentState.trainPlan,
				}, nil
			} else {
				return state, func() {
					operator.Dispatch(&trainStartAction{})
				}
			}
		},
		reflect.TypeOf(&baseModelReceivedAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			trainInitState := state.(trainInitState)
			baseModelReceivedAction := action.(*baseModelReceivedAction)

			return state, func() {
				zap.L().Debug(fmt.Sprintf("Webhook ref [%s]. Expect [%s]",
					baseModelReceivedAction.ref,
					fmt.Sprintf("refs/heads/%s", util.TrainBranch(trainInitState.trainPlan.Name, trainInitState.trainPlan.CommitID)),
				))

				if baseModelReceivedAction.ref != fmt.Sprintf("refs/heads/%s", util.TrainBranch(trainInitState.trainPlan.Name, trainInitState.trainPlan.CommitID)) {
					return 
				}

				util.CheckoutPretrainedModel(
					operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
					trainInitState.trainPlan.Name,
					trainInitState.trainPlan.CommitID,
				)
				operator.Dispatch(&pretrainedModelReadyAction{})
			}
		},
		reflect.TypeOf(&pretrainedModelReadyAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			currentState := state.(trainInitState)
			if !currentState.init {
				return trainInitState {
					init: currentState.init,
					pretrainedModel: true,
					trainPlan: currentState.trainPlan,
				}, nil
			} else {
				return state, func() {
					operator.Dispatch(&trainStartAction{})
				}
			}
		},
	},
	reflect.TypeOf(localTrainState{}): {
		reflect.TypeOf(&trainFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			localTrainState := state.(localTrainState)
			trainFinishAction := action.(*trainFinishAction)

			pushModel := func () {
				err := util.WriteMetadata(operator.GetPayload().(Payload).EdgeModelRepoGitHttpURL, map[string]interface{} {
					"trainPlanID": localTrainState.trainPlan.CommitID,
					"datasetSize": trainFinishAction.datasetSize,
					"metadata": trainFinishAction.metadata,
					"metrics": trainFinishAction.metrics,
					"roundNumber": localTrainState.roundCount + 1,
					"pretrainedModel": localTrainState.trainPlan.PretrainedModelCommitID,
					// TODO
					// "baseModel"
				})
				if err != nil {
					zap.L().Error("push edge model error", zap.Error(err))
					return
				}

				err = util.PushUpdates(operator.GetPayload().(Payload).EdgeModelRepoGitHttpURL, "")
				if err != nil {
					zap.L().Error("push edge model error", zap.Error(err))
				}
			}

			if localTrainState.roundCount + 1 == localTrainState.trainPlan.RoundCount {
				return idleState{}, pushModel
			} else {
				if (util.TrainPlan{}) == localTrainState.trainPlan {
					zap.L().Warn("train plan is empty")
				}
				return aggregateState{
					roundCount: localTrainState.roundCount + 1,
					trainPlan:   localTrainState.trainPlan,
				}, pushModel
			}
		},
	},
	reflect.TypeOf(aggregateState{}): {
		reflect.TypeOf(&baseModelReceivedAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			aggregateState := state.(aggregateState)
			if (util.TrainPlan{}) == aggregateState.trainPlan {
				zap.L().Warn("train plan is empty")
			}
			return localTrainState{
				roundCount: aggregateState.roundCount,
				trainPlan:   aggregateState.trainPlan,
			}, func() {
				zap.L().Debug("pull aggregated model")
				util.PullData(operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL)

				repoMetadata, err := util.ReadMetadata(operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL)
				if err != nil {
					zap.L().Fatal("Cannot read repoMetadata", zap.Error(err))
					return
				}

				metadata := map[string]string{}
				for key, val := range repoMetadata["metadata"].(map[string]interface{}) {
					metadata[key] = val.(string)
				}
				metrics := map[string]float64{}
				for key, val := range repoMetadata["metrics"].(map[string]interface{}) {
					metrics[key] = val.(float64)
				}

				sendLocalTrainMessage(
					operator.GetPayload().(Payload).GrpcServerURI,
					aggregateState.trainPlan.EpochCount,
					baseModel {
						gitHttpURL: operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
						metadata: metadata,
						metrics: metrics,
					},
					operator.GetPayload().(Payload).EdgeModelRepoGitHttpURL,
				)
			}
		},
	},
}

var InitState = idleState{}
