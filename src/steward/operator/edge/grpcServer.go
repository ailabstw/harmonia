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
func (server *EdgeOperatorServer) LocalTrainFinish(_ context.Context, localTrainResult *protos.LocalTrainResult) (*protos.Empty, error) {
	zap.L().Debug(" --- On Local Train Finish --- ", zap.String("server", fmt.Sprintf("%v", server)))
	zap.L().Debug(fmt.Sprintf("Receive localTrainResult.Metadata [%v]", localTrainResult.Metadata))
	zap.L().Debug(fmt.Sprintf("Receive localTrainResult.Metrics [%v]", localTrainResult.Metrics))

	var metadata map[string]string
	if localTrainResult.Metadata == nil {
		metadata = map[string]string {}
	} else {
		metadata = localTrainResult.Metadata
	}

	var metrics map[string]float64
	if localTrainResult.Metrics == nil {
		metrics = map[string]float64 {}
	} else {
		metrics = localTrainResult.Metrics
	}

	server.operator.Dispatch(&trainFinishAction{
		errCode: int(localTrainResult.Error),
		datasetSize: int(localTrainResult.DatasetSize),
		metadata: metadata,
		metrics: metrics,
	})

	return &protos.Empty{}, nil
}

// GrpcServerRegister : register grpc server
func GrpcServerRegister(grpcServer *grpc.Server, operator util.AbstractOperator) {
	protos.RegisterEdgeOperatorServer(grpcServer, &EdgeOperatorServer {
		operator: operator, 
	})
}
