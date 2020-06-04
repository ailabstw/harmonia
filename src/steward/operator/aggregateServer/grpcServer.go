package aggregateServer

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"harmonia.com/steward/operator/util"
	"harmonia.com/steward/protos"
)

type AggregateServerOperatorServer struct {
	operator util.AbstractOperator
}

// AggregateFinish : event on finishing aggregating
func (server *AggregateServerOperatorServer) AggregateFinish(context.Context, *protos.Msg) (*protos.Msg, error) {
	zap.L().Debug(" --- On Aggregate Finish --- ", zap.String("server", fmt.Sprintf("%v", server)))
	server.operator.Dispatch(&aggregateFinishAction{})

	return &protos.Msg{
		Message: "uploaded",
	}, nil
}

// GrpcServerRegister : register grpc server
func GrpcServerRegister(grpcServer *grpc.Server, operator util.AbstractOperator) {
	protos.RegisterAggregateServerOperatorServer(grpcServer, &AggregateServerOperatorServer {
		operator: operator, 
	})
}
