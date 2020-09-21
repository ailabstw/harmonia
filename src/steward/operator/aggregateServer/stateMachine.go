package aggregateServer

import (
	"fmt"
	"reflect"
	"time"

	"harmonia.com/steward/operator/util"

	"go.uber.org/zap"
)

var StateTransit = util.StateTransit {
	reflect.TypeOf(idleState{}): {
		reflect.TypeOf(&trainPlanAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, []func()) {
			trainPlanAction := action.(*trainPlanAction)
			return waitPretrainModelState {
				trainPlan: trainPlanAction.trainPlan,
			}, []func() {
				func() {
					err := util.CreateGlobalModelBranch(
						operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
						trainPlanAction.trainPlan.Name,
						trainPlanAction.trainPlan.CommitID,
						trainPlanAction.trainPlan.PretrainedModelCommitID,
					)
					if err != nil {
						zap.L().Fatal("Cannot create train branch", zap.Error(err))
						return
					}
					operator.Dispatch(&trainStartAction{})
				},
			}
		},
	},
	reflect.TypeOf(waitPretrainModelState{}): {
		reflect.TypeOf(&trainStartAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, []func()) {
			waitPretrainModelState := state.(waitPretrainModelState)
			return localTrainState{
				trainPlan: waitPretrainModelState.trainPlan,
				roundCount: 0,
				edgeModels: []localModel{},
				webhooks: []string{},
			}, []func() {
				func () {
					time.Sleep(time.Duration(waitPretrainModelState.trainPlan.Timeout) * time.Second)
					operator.Dispatch(&localTrainTimeoutAction{
						roundCount: 0,
					})
				},
			}
		},
	},
	reflect.TypeOf(localTrainState{}): {
		reflect.TypeOf(&localTrainFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, []func()) {
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
				roundCount:  currentLocalTrainState.roundCount,
				edgeModels:  currentLocalTrainState.edgeModels,
				webhooks:    append(currentLocalTrainState.webhooks, aggLocalTrainFinish.gitHttpURL),
			}, []func() {
				func() {
					err := util.PullData(aggLocalTrainFinish.gitHttpURL)
					if err != nil {
						zap.L().Fatal("pull repository error",
							zap.String("git http url", aggLocalTrainFinish.gitHttpURL),
							zap.Error(err))
						return
					}

					repoMetadata, err := util.ReadMetadata(aggLocalTrainFinish.gitHttpURL)
					zap.L().Debug(fmt.Sprintf("metadata [%v]", repoMetadata))
					if err != nil {
						zap.L().Fatal("Cannot read metadata", zap.Error(err))
						return
					}

					datasetSize := int(repoMetadata["datasetSize"].(float64))
					metadata := map[string]string{}
					for key, val := range repoMetadata["metadata"].(map[string]interface{}) {
						metadata[key] = val.(string)
					}
					metrics := map[string]float64{}
					for key, val := range repoMetadata["metrics"].(map[string]interface{}) {
						metrics[key] = val.(float64)
					}

					operator.Dispatch(&localModelPullFinishAction{
						gitHttpURL: aggLocalTrainFinish.gitHttpURL,
						datasetSize: datasetSize,
						metadata: metadata,
						metrics: metrics,
					})
				},
			}
		},
		reflect.TypeOf(&localModelPullFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, []func()) {
			localModelPullFinishAction := action.(*localModelPullFinishAction)
			currentLocalTrainState := state.(localTrainState)

			edgeModels := append(currentLocalTrainState.edgeModels, localModel {
				gitHttpURL: localModelPullFinishAction.gitHttpURL,
				datasetSize: localModelPullFinishAction.datasetSize,
				metadata: localModelPullFinishAction.metadata,
				metrics: localModelPullFinishAction.metrics,
			})

			if len(currentLocalTrainState.edgeModels) + 1 == currentLocalTrainState.trainPlan.EdgeCount {
				return aggregateState{
					trainPlan: currentLocalTrainState.trainPlan,
					roundCount: currentLocalTrainState.roundCount,
				}, []func() {
					func() {
						sendAggregateMessage(
							operator.GetPayload().(Payload).GrpcServerURI,
							edgeModels,
							operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
						)
					},
				}
			} else {
				return localTrainState{
					trainPlan:   currentLocalTrainState.trainPlan,
					roundCount:  currentLocalTrainState.roundCount,
					edgeModels:  edgeModels,
					webhooks:    currentLocalTrainState.webhooks,
				}, nil
			}
		},
		reflect.TypeOf(&localTrainTimeoutAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, []func()) {
			currentState := state.(localTrainState)
			currentAction := action.(*localTrainTimeoutAction)

			if currentAction.roundCount != currentState.roundCount {
				return currentState, nil
			}

			zap.L().Debug(fmt.Sprintf("Round [%d] timeout", currentState.roundCount))
			return aggregateState {
				trainPlan: currentState.trainPlan,
				roundCount: currentState.roundCount,
			}, []func() {
				func() {
					sendAggregateMessage(
						operator.GetPayload().(Payload).GrpcServerURI,
						currentState.edgeModels,
						operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
					)
				},
			}
		},
	},
	reflect.TypeOf(aggregateState{}): {
		reflect.TypeOf(&aggregateFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, []func()) {
			aggregateState := state.(aggregateState)
			operatorPayload := operator.GetPayload().(Payload)
			aggregateFinishAction := action.(*aggregateFinishAction)

			if aggregateState.roundCount + 1 == aggregateState.trainPlan.RoundCount {
				// Finish a round
				return idleState{}, []func() {
					func() {
						err := util.WriteMetadata(operatorPayload.AggregatedModelRepoGitHttpURL ,map[string]interface{} {
							"metadata": aggregateFinishAction.metadata,
							"metrics": aggregateFinishAction.metrics,
							"trainPlanID": aggregateState.trainPlan.CommitID,
							"roundNumber": aggregateState.roundCount + 1,
							"pretrainedModel": aggregateState.trainPlan.PretrainedModelCommitID,
							// TODO
							// "baseModel"
						})
						if err != nil {
							zap.L().Error("push aggregated model error", zap.Error(err))
							return
						}

						err = util.PushUpdates(
							operatorPayload.AggregatedModelRepoGitHttpURL,
							util.InferenceTag(aggregateState.trainPlan.Name, aggregateState.trainPlan.CommitID),
						)
						if err != nil {
							zap.L().Error("push aggregated model error", zap.Error(err))
							return
						}

						sendTrainFinishMessage(
							operator.GetPayload().(Payload).GrpcServerURI,
						)
						operator.TrainFinish()
					},
				}
			} else {
				return localTrainState{
					trainPlan:   aggregateState.trainPlan,
					roundCount:  aggregateState.roundCount + 1,
					edgeModels:  []localModel{},
					webhooks:    []string{},
				}, []func() {
					func() {
						zap.L().Debug(fmt.Sprintf("aggregateFinishAction [%v]", aggregateFinishAction))
						
						err := util.WriteMetadata(operatorPayload.AggregatedModelRepoGitHttpURL, map[string]interface{} {
							"metadata": aggregateFinishAction.metadata,
							"metrics": aggregateFinishAction.metrics,
							"trainPlanID": aggregateState.trainPlan.CommitID,
							"roundNumber": aggregateState.roundCount + 1,
							"pretrainedModel": aggregateState.trainPlan.PretrainedModelCommitID,
							// TODO
							// "baseModel"
						})
						if err != nil {
							zap.L().Error("push aggregated model error", zap.Error(err))
							return
						}

						err = util.PushUpdates(operatorPayload.AggregatedModelRepoGitHttpURL, "")
						if err != nil {
							zap.L().Error("push aggregated model error", zap.Error(err))
						}

						// ------

						time.Sleep(time.Duration(aggregateState.trainPlan.Timeout) * time.Second)
						operator.Dispatch(&localTrainTimeoutAction{
							roundCount: aggregateState.roundCount + 1,
						})
					},
				}
			}
		},
	},
}

var InitState = idleState{}
