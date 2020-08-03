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
func (server *AggregateServerOperatorServer) AggregateFinish(_ context.Context, aggregateResult *protos.AggregateResult) (*protos.Empty, error) {
	zap.L().Debug(" --- On Aggregate Finish --- ")
	zap.L().Debug(fmt.Sprintf("Receive aggregateResult.Metadata [%v]", aggregateResult.Metadata))
	zap.L().Debug(fmt.Sprintf("Receive aggregateResult.Metrics [%v]", aggregateResult.Metrics))
	
	var metadata map[string]string
	if aggregateResult.Metadata == nil {
		metadata = map[string]string {}
	} else {
		metadata = aggregateResult.Metadata
	}

	var metrics map[string]float64
	if aggregateResult.Metrics == nil {
		metrics = map[string]float64 {}
	} else {
		metrics = aggregateResult.Metrics
	}

	server.operator.Dispatch(&aggregateFinishAction{
		errCode: int(aggregateResult.Error),
		metadata: metadata,
		metrics: metrics,
	})

	return &protos.Empty{}, nil
}

// GrpcServerRegister : register grpc server
func GrpcServerRegister(grpcServer *grpc.Server, operator util.AbstractOperator) {
	protos.RegisterAggregateServerOperatorServer(grpcServer, &AggregateServerOperatorServer {
		operator: operator, 
	})
}
