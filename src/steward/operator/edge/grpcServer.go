package edge

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"harmonia.com/steward/operator/util"
	"harmonia.com/steward/protos"
)

type EdgeOperatorServer struct {
	operator util.AbstractOperator
}

// LocalTrainFinish : event on finishing local training
func (server *EdgeOperatorServer) LocalTrainFinish(context.Context, *protos.Msg) (*protos.Msg, error) {
	zap.L().Debug(" --- On Local Train Finish --- ", zap.String("server", fmt.Sprintf("%v", server)))
	server.operator.Dispatch(&trainFinishAction{})

	return &protos.Msg{
		Message: "ok",
	}, nil
}

// GrpcServerRegister : register grpc server
func GrpcServerRegister(grpcServer *grpc.Server, operator util.AbstractOperator) {
	protos.RegisterEdgeOperatorServer(grpcServer, &EdgeOperatorServer {
		operator: operator, 
	})
}
