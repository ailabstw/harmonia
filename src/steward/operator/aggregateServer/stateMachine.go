package aggregateServer

import (
	"fmt"
	"reflect"
	"strings"

	"harmonia.com/steward/operator/util"

	"go.uber.org/zap"
)

var StateTransit = util.StateTransit {
	reflect.TypeOf(idleState{}): {
		reflect.TypeOf(&trainPlanAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			trainPlanAction := action.(*trainPlanAction)
			return waitPretrainModelState {
				trainPlan: trainPlanAction.trainPlan,
			}, func() {
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
			}
		},
	},
	reflect.TypeOf(waitPretrainModelState{}): {
		reflect.TypeOf(&trainStartAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			waitPretrainModelState := state.(waitPretrainModelState)
			return localTrainState{
				trainPlan: waitPretrainModelState.trainPlan,
				roundCount: 0,
				edgeModels: []localModel{},
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
				roundCount:  currentLocalTrainState.roundCount,
				edgeModels:  currentLocalTrainState.edgeModels,
				webhooks:    append(currentLocalTrainState.webhooks, aggLocalTrainFinish.gitHttpURL),
			}, func() {
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

				operator.Dispatch(&pullFinishAction{
					gitHttpURL: aggLocalTrainFinish.gitHttpURL,
					datasetSize: datasetSize,
					metadata: metadata,
					metrics: metrics,
				})
			}
		},
		reflect.TypeOf(&pullFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			pullFinishAction := action.(*pullFinishAction)
			currentLocalTrainState := state.(localTrainState)

			edgeModels := append(currentLocalTrainState.edgeModels, localModel {
				gitHttpURL: pullFinishAction.gitHttpURL,
				datasetSize: pullFinishAction.datasetSize,
				metadata: pullFinishAction.metadata,
				metrics: pullFinishAction.metrics,
			})

			if len(currentLocalTrainState.edgeModels) + 1 == currentLocalTrainState.trainPlan.EdgeCount {
				return aggregateState{
					trainPlan: currentLocalTrainState.trainPlan,
					roundCount: currentLocalTrainState.roundCount,
				}, func() {
					sendAggregateMessage(
						operator.GetPayload().(Payload).GrpcServerURI,
						edgeModels,
						operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL,
					)
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
	},
	reflect.TypeOf(aggregateState{}): {
		reflect.TypeOf(&aggregateFinishAction{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, func()) {
			aggregateState := state.(aggregateState)
			operatorPayload := operator.GetPayload().(Payload)
			aggregateFinishAction := action.(*aggregateFinishAction)

			if aggregateState.roundCount + 1 == aggregateState.trainPlan.RoundCount {
				// Finish a round
				return idleState{}, func() {
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
						strings.Join([]string{"inference", aggregateState.trainPlan.Name, aggregateState.trainPlan.CommitID}, "-"),
					)
					if err != nil {
						zap.L().Error("push aggregated model error", zap.Error(err))
						return
					}
				}
			} else {
				return localTrainState{
					trainPlan:   aggregateState.trainPlan,
					roundCount:  aggregateState.roundCount + 1,
					edgeModels:  []localModel{},
					webhooks:    []string{},
				}, func() {
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
				}
			}
		},
	},
}

var InitState = idleState{}
