package edge

import (
	"context"

	"google.golang.org/grpc"

	"harmonia.com/steward/operator/util"
	"harmonia.com/steward/protos"
)

func sendLocalTrainMessage(epochCount int, appGrpcServerURI string, edgeModelRepoGitURL string, aggregatedModelRepoGitHttpURL string) {
	util.EmitEvent(
		appGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewEdgeAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			aggregatedModelPath, _ := util.GitHttpURLToRepoFullName(aggregatedModelRepoGitHttpURL)
			edgeModelPath, _ := util.GitHttpURLToRepoFullName(edgeModelRepoGitURL)

			return client.(protos.EdgeAppClient).LocalTrain(ctx, &protos.LocalTrainParams{
				InputModelPath:  aggregatedModelPath,
				OutputModelPath: edgeModelPath,
				EpochCount:      int32(epochCount),
			})
		},
	)
}
