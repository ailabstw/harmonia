package aggregateServer

import (
	"context"

	"google.golang.org/grpc"

	"harmonia.com/steward/operator/util"
	"harmonia.com/steward/protos"
)

func sendAggregateMessage() {
	util.EmitEvent(
		util.Config.AppGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewAggregateServerAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			edgeModelPaths := make([]string, len(util.Config.EdgeModelRepos))
			for i, repo := range util.Config.EdgeModelRepos {
				edgeModelPaths[i] = util.GitHttpURLToRepoFullName(repo.GitHttpURL)
			}

			return client.(protos.AggregateServerAppClient).Aggregate(ctx, &protos.AggregateParams{
				InputModelPaths: edgeModelPaths,
				OutputModelPath: util.GitHttpURLToRepoFullName(util.Config.AggregatedModelRepo.GitHttpURL),
			})
		},
	)
}
