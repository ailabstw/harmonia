package aggregateServer

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"harmonia.com/steward/operator/util"
	"harmonia.com/steward/protos"

	"go.uber.org/zap"
)

func sendAggregateMessage(appGrpcServerURI string, localModels []localModel, aggregatedModelRepoGitHttpURL string) {
	util.EmitEvent(
		appGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewAggregateServerAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			localModelsMsg := make([]*protos.AggregateParams_LocalModel, len(localModels))
			for i, localModel := range localModels {
				repoPath, _ := util.GitHttpURLToRepoFullName(localModel.gitHttpURL)
				localModelsMsg[i] = &protos.AggregateParams_LocalModel {
					Path: repoPath,
					DatasetSize: int32(localModel.datasetSize),
					Metadata: localModel.metadata,
					Metrics: localModel.metrics,
				}
			}

			aggregatedModelPath, _ := util.GitHttpURLToRepoFullName(aggregatedModelRepoGitHttpURL)

			zap.L().Debug(fmt.Sprintf("Aggregate message [%v]", protos.AggregateParams{
				LocalModels: localModelsMsg,
				AggregatedModel: &protos.AggregateParams_AggregatedModel {
					Path: aggregatedModelPath,
				},
			}))

			return client.(protos.AggregateServerAppClient).Aggregate(ctx, &protos.AggregateParams{
				LocalModels: localModelsMsg,
				AggregatedModel: &protos.AggregateParams_AggregatedModel {
					Path: aggregatedModelPath,
				},
			})
		},
	)
}

func sendTrainFinishMessage(appGrpcServerURI string) {
	util.EmitEvent(
		appGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewAggregateServerAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			return client.(protos.AggregateServerAppClient).TrainFinish(ctx, &protos.Empty{})
		},
	)
}
