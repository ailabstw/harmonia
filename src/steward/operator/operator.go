package operator

import (
	"encoding/json"
	"sync"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"harmonia.com/steward/operator/aggregateServer"
	"harmonia.com/steward/operator/edge"
	"harmonia.com/steward/operator/util"
)

type Operator struct {
	state util.State
	stateTransit util.StateTransit

	webhookToAction util.WebhookToAction
	grpcServerRegister util.GrpcServerRegisterFunc
	payload interface{}
	stateMux *sync.Mutex
}

func httpRequestToWebhook(req *http.Request) (*util.Webhook, error) {
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		return nil, err
	}

	var webhook util.Webhook
	err = json.Unmarshal(body, &webhook)
	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

// HttpHandleFunc : Handle http request
func (operator *Operator) HttpHandleFunc() util.HttpHandleFunc {
	defer zap.L().Sync()
	return func(w http.ResponseWriter, req *http.Request) {
		zap.L().Info("received request", zap.String("service", "http server"))
		zap.L().Debug("received request", zap.String("service", "http server"), zap.String("request", fmt.Sprintf("%v", req)))
		webhook, err := httpRequestToWebhook(req)
		if err != nil {
			http.Error(w, err.Error(), 500)
			zap.L().Error("webhook to action error", zap.Error(err))
			return
		}

		zap.L().Debug("received webhook", zap.String("service", "http server"), zap.String("webhook", fmt.Sprintf("%v", webhook)))
		action, err := operator.webhookToAction(webhook, operator)
		if err != nil {
			http.Error(w, err.Error(), 500)
			zap.L().Error("webhook to action error", zap.Error(err))
			return
		}

		zap.L().Debug("perform action", zap.String("service", "http server"), zap.String("action", fmt.Sprintf("%v", action)))
		go operator.Dispatch(action)

		w.WriteHeader(http.StatusOK)
	}
}

// GrpcServerRegister : register grpc server
func (operator *Operator) GrpcServerRegister(grpcServer *grpc.Server) {
	operator.grpcServerRegister(grpcServer, operator)
}

func (operator *Operator) Dispatch(action util.Action) {
	operator.stateMux.Lock()
	defer zap.L().Sync()
	zap.L().Debug(" --- State Before Dispatching --- ",
		zap.String("type", fmt.Sprintf("%v", reflect.TypeOf(operator.state))),
		zap.String("payload", fmt.Sprintf("%v", operator.state)))

	zap.L().Debug(" --- Action --- ",
		zap.String("type", fmt.Sprintf("%v", reflect.TypeOf(action))),
		zap.String("payload", fmt.Sprintf("%v", action)))

	defer zap.L().Debug(" --- State After Dispatching --- ", zap.String("type", fmt.Sprintf("%v", reflect.TypeOf(operator.state))), zap.String("payload", fmt.Sprintf("%v", operator.state)))

	if _, ok := operator.stateTransit[reflect.TypeOf(operator.state)][reflect.TypeOf(action)]; !ok {
		zap.L().Error(" *** Invalid transit *** ")
		return
	}

	nextState, then := operator.stateTransit[reflect.TypeOf(operator.state)][reflect.TypeOf(action)](operator.state, action, operator)
	operator.state = nextState
	if then != nil {
		go then()
	}
	operator.stateMux.Unlock()
}

func (operator *Operator) GetPayload() interface{} {
	return operator.payload
}

func newAggregateServerOperator(
	appGrpcServerURI string,
	trainPlanRepoGitHttpURL string,
	aggregatedModelRepoGitHttpURL string,
	_ string,
	edgeModelRepoGitHttpURLs []string,
) *Operator {
	return &Operator {
		aggregateServer.InitState,
		aggregateServer.StateTransit,
		aggregateServer.WebhookToAction,
		aggregateServer.GrpcServerRegister,
		aggregateServer.Payload {
			GrpcServerURI: appGrpcServerURI,
			TrainPlanRepoGitHttpURL: trainPlanRepoGitHttpURL,
			AggregatedModelRepoGitHttpURL: aggregatedModelRepoGitHttpURL,
			EdgeModelRepoGitHttpURLs: edgeModelRepoGitHttpURLs,
		},
		&sync.Mutex{},
	}
}

func newEdgeOperator(
	appGrpcServerURI string,
	trainPlanRepoGitHttpURL string,
	aggregatedModelRepoGitHttpURL string,
	edgeModelRepoGitHttpURL string,
	_ []string,
) *Operator {
	return &Operator {
		edge.InitState,
		edge.StateTransit,
		edge.WebhookToAction,
		edge.GrpcServerRegister,
		edge.Payload {
			GrpcServerURI: appGrpcServerURI,
			TrainPlanRepoGitHttpURL: trainPlanRepoGitHttpURL,
			AggregatedModelRepoGitHttpURL: aggregatedModelRepoGitHttpURL,
			EdgeModelRepoGitHttpURL: edgeModelRepoGitHttpURL,
		},
		&sync.Mutex{},
	}
}

var NewOperator = map[string] func(
	appGrpcServerURI string,
	trainPlanRepoGitHttpURL string,
	aggregatedModelRepoGitHttpURL string,
	edgeModelRepoGitHttpURL string,
	edgeModelRepoGitHttpURLs []string,
) *Operator {
	"aggregator": newAggregateServerOperator,
	"edge": newEdgeOperator,
}
