package edge

import (
	"context"

	"google.golang.org/grpc"

	"harmonia.com/steward/operator/util"
	"harmonia.com/steward/protos"
)

func sendLocalTrainMessage() {
	util.EmitEvent(
		util.Config.AppGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewEdgeAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			return client.(protos.EdgeAppClient).LocalTrain(ctx, &protos.LocalTrainParams{
				InputModelPath: util.GitHttpURLToRepoFullName(util.Config.AggregatedModelRepo.GitHttpURL),
				OutputModelPath: util.GitHttpURLToRepoFullName(util.Config.EdgeModelRepo.GitHttpURL),
			})
		},
	)
}
