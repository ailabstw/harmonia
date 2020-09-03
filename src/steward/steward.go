package main

import (
	"fmt"
	"sync"
	"net"

	"google.golang.org/grpc"

	"github.com/gocraft/work"
	"go.uber.org/zap"

	"harmonia.com/steward/config"
	"harmonia.com/steward/operator"
	"harmonia.com/steward/operator/util"
)

// Make an enqueuer with a particular namespace
var enqueuer *work.Enqueuer

func startNotificationServer(wg *sync.WaitGroup, operator *operator.Operator) {
	go func() {
		defer wg.Done()

		notificationParam := func() util.NotificationParam {
			switch config.Config.Notification.Type {
			case "push":
				return util.PushNotificationParam {
					WebhookURL: config.Config.Notification.StewardServerURI,
				}
			case "pull":
				return util.PullNotificationParam {
					PullPeriod: config.Config.Notification.PullPeriod,
				}
			default:
				zap.L().Fatal(fmt.Sprintf("Invalid config.Config.Notification.Type [%v]", config.Config.Notification.Type))
				return nil
			}
		}()

		operator.RemoteNotificationRegister(notificationParam)
	}()
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

	zap.L().Info(fmt.Sprintf("Grpc Listen [%v]", config.Config.OperatorGrpcServerURI))
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
	startNotificationServer(wg, operator)

	zap.L().Info("steward startup")

	wg.Wait()
	zap.L().Info("main: done. exiting")
}
