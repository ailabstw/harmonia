package main

import (
	"sync"

	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/gocraft/work"
	"go.uber.org/zap"

	"harmonia.com/steward/config"
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
	util.GitSetup(util.GitUser{
		"Harmonia Operator",
		"operator@harmonia",
		config.Config.GitUserToken,
	})

	if config.Config.AggregatedModelRepo != nil {
		util.CloneRepository(config.Config.AggregatedModelRepo.GitHttpURL)
	}

	if config.Config.EdgeModelRepo != nil {
		util.CloneRepository(config.Config.EdgeModelRepo.GitHttpURL)
	}

	if config.Config.EdgeModelRepos != nil {
		for _, edgeModelRepo := range config.Config.EdgeModelRepos {
			util.CloneRepository(edgeModelRepo.GitHttpURL)
		}
	}

	if config.Config.TrainPlanRepo != nil {
		util.CloneRepository(config.Config.TrainPlanRepo.GitHttpURL)
	}

	zap.L().Debug("git configuration setup succeeds")
}

func init() {
	setupGitConfig()
	zap.L().Info("Init Finished")
}

func main() {
	defer zap.L().Sync()

	wg := &sync.WaitGroup{}

	var edgeModelRepoGitHttpURLs []string
	var edgeModelRepoGitHttpURL string

	if config.Config.EdgeModelRepos != nil {	
		edgeModelRepoGitHttpURLs = make([]string, len(config.Config.EdgeModelRepos))
		for i, edgeModelRepo := range config.Config.EdgeModelRepos {
			edgeModelRepoGitHttpURLs[i] = edgeModelRepo.GitHttpURL
		}
	}
	if config.Config.EdgeModelRepo != nil {
		edgeModelRepoGitHttpURL = config.Config.EdgeModelRepo.GitHttpURL
	}
	operator := operator.NewOperator[config.Config.Type](
		config.Config.AppGrpcServerURI,
		config.Config.TrainPlanRepo.GitHttpURL,
		config.Config.AggregatedModelRepo.GitHttpURL,
		edgeModelRepoGitHttpURL,
		edgeModelRepoGitHttpURLs,
	)

	wg.Add(1)
	startGrpcServer(wg, config.Config.OperatorGrpcServerURI, operator)

	wg.Add(1)
	startHTTPServer(wg, config.Config.StewardServerURI, operator)

	zap.L().Info("steward startup",
		zap.String("Steward Listen", config.Config.StewardServerURI),
		zap.String("Grpc Listen", config.Config.OperatorGrpcServerURI),
	)

	wg.Wait()
	zap.L().Info("main: done. exiting")
}
