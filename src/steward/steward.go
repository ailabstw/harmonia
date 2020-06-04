package main

import (
	"sync"

	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/gocraft/work"
	"go.uber.org/zap"

	"harmonia.com/steward/operator"
	"harmonia.com/steward/operator/util"
)

// Make an enqueuer with a particular namespace
var enqueuer *work.Enqueuer

func startHTTPServer(wg *sync.WaitGroup, address string, operator *operator.Operator) *http.Server {
	srv := &http.Server{Addr: address}
	http.HandleFunc("/", operator.HttpHandleFunc())

	go func() {
		defer wg.Done()

		zap.L().Debug("Steward starts")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			zap.L().Fatal("Cannot start server on the address",
				zap.String("service", "http"),
				zap.String("address", address),
				zap.Error(err))
		}
	}()

	return srv
}

func startGrpcServer(wg *sync.WaitGroup, address string, operator *operator.Operator) *grpc.Server {

	lis, err := net.Listen("tcp", address)
	if err != nil {
		zap.L().Fatal("Cannot listen on the address",
			zap.String("service", "grpc"),
			zap.String("address", address),
			zap.Error(err))
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	go func() {
		defer wg.Done()

		zap.L().Debug("Grpc server listen the address",
			zap.String("service", "grpc"),
			zap.String("address", address))
		operator.GrpcServerRegister(grpcServer)
		if err := grpcServer.Serve(lis); err != nil {
			zap.L().Fatal("Cannot start server on the address",
				zap.String("service", "grpc"),
				zap.String("address", address),
				zap.Error(err))
		}
	}()

	return grpcServer
}

func setupGitConfig() {
	if util.Config.AggregatedModelRepo != nil {
		util.CloneRepository(util.Config.AggregatedModelRepo.GitHttpURL)
	}

	if util.Config.EdgeModelRepo != nil {
		util.CloneRepository(util.Config.EdgeModelRepo.GitHttpURL)
	}

	if util.Config.EdgeModelRepos != nil {
		for _, edgeModelRepo := range util.Config.EdgeModelRepos {
			util.CloneRepository(edgeModelRepo.GitHttpURL)
		}
	}

	if util.Config.TrainPlanRepo != nil {
		util.CloneRepository(util.Config.TrainPlanRepo.GitHttpURL)
	}

	zap.L().Debug("git configuration setup succeeds")
}

func init() {
	// setupEnqueuer()
	setupGitConfig()
	zap.L().Info("Init Finished")
}

func main() {
	defer zap.L().Sync()

	wg := &sync.WaitGroup{}

	operator := operator.NewOperator[util.Config.Type]()

	wg.Add(1)
	startGrpcServer(wg, util.Config.OperatorGrpcServerURI, operator)

	wg.Add(1)
	startHTTPServer(wg, util.Config.StewardServerURI, operator)

	zap.L().Info("steward startup",
		zap.String("Steward Listen", util.Config.StewardServerURI),
		zap.String("Grpc Listen", util.Config.OperatorGrpcServerURI),
	)

	wg.Wait()
	zap.L().Info("main: done. exiting")
}
