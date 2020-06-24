package aggregateServer

import (
	"context"

	"google.golang.org/grpc"

	"harmonia.com/steward/operator/util"
	"harmonia.com/steward/protos"
)

func sendAggregateMessage(appGrpcServerURI string, EdgeModelRepoGitHttpURLs []string, aggregatedModelRepoGitHttpURL string) {
	util.EmitEvent(
		appGrpcServerURI,
		func(conn *grpc.ClientConn) interface{} {
			return protos.NewAggregateServerAppClient(conn)
		},
		func(ctx context.Context, client interface{}) (interface{}, error) {
			edgeModelPaths := make([]string, len(EdgeModelRepoGitHttpURLs))
			for i, edgeModelRepoGitURL := range EdgeModelRepoGitHttpURLs {
				edgeModelPaths[i], _ = util.GitHttpURLToRepoFullName(edgeModelRepoGitURL)
			}

			aggregatedModelPath, _ := util.GitHttpURLToRepoFullName(aggregatedModelRepoGitHttpURL)
			return client.(protos.AggregateServerAppClient).Aggregate(ctx, &protos.AggregateParams{
				InputModelPaths: edgeModelPaths,
				OutputModelPath: aggregatedModelPath,
			})
		},
	)
}
