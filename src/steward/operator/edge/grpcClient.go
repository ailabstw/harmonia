package edge

import (
	"context"

	"google.golang.org/grpc"

	"harmonia.com/steward/operator/util"
	"harmonia.com/steward/protos"
)

func sendInitMessage(appGrpcServerURI string) {
	util.EmitEvent(
		appGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewEdgeAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			return client.(protos.EdgeAppClient).TrainInit(ctx, &protos.Empty{})
		},
	)
}

func sendLocalTrainMessage(appGrpcServerURI string, epochPerRound int, baseModel baseModel, edgeModelRepoGitURL string) {
	util.EmitEvent(
		appGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewEdgeAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			aggregatedModelPath, _ := util.GitHttpURLToRepoFullName(baseModel.gitHttpURL)
			edgeModelPath, _ := util.GitHttpURLToRepoFullName(edgeModelRepoGitURL)

			return client.(protos.EdgeAppClient).LocalTrain(ctx, &protos.LocalTrainParams{
				BaseModel: &protos.LocalTrainParams_BaseModel {
					Path: aggregatedModelPath,
					Metadata: baseModel.metadata,
					Metrics: baseModel.metrics,
				},
				LocalModel: &protos.LocalTrainParams_LocalModel {
					Path: edgeModelPath,
				},
				EpR: int32(epochPerRound),
			})
		},
	)
}

func sendTrainInterruptMessage(appGrpcServerURI string) {
	util.EmitEvent(
		appGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewEdgeAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			return client.(protos.EdgeAppClient).TrainInterrupt(ctx, &protos.Empty{})
		},
	)
}

func sendTrainFinishMessage(appGrpcServerURI string) {
	util.EmitEvent(
		appGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewEdgeAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			return client.(protos.EdgeAppClient).TrainFinish(ctx, &protos.Empty{})
		},
	)
}
