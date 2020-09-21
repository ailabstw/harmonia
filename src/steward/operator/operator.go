package operator

import (
	"encoding/json"
	"sync"
	"fmt"
	"time"
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

	// Notification
	webhookToAction util.WebhookToAction
	pullRepoNotification util.PullRepoNotification

	grpcServerRegister util.GrpcServerRegisterFunc
	payload interface{}
	stateMux *sync.Mutex
	trainFinish func()
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

var remoteNoticationFunc = map[reflect.Type] func(operator *Operator, notificationParam util.NotificationParam) {
	reflect.TypeOf(util.PushNotificationParam{}): func(operator *Operator, notificationParam util.NotificationParam)  {
		pushNotificationParam := notificationParam.(util.PushNotificationParam)
		srv := &http.Server{Addr: pushNotificationParam.WebhookURL}
		http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			defer zap.L().Sync()

			zap.L().Info("Receive http request")
			zap.L().Debug(fmt.Sprintf("payload [%v]", req))
			webhook, err := httpRequestToWebhook(req)
			if err != nil {
				http.Error(w, err.Error(), 500)
				zap.L().Error("webhook to action error", zap.Error(err))
				return
			}
	
			zap.L().Debug(fmt.Sprintf("Receive webhook [%v]", webhook))
			action, err := operator.webhookToAction(webhook, operator)
			if err != nil {
				http.Error(w, err.Error(), 500)
				zap.L().Error("webhook to action error", zap.Error(err))
				return
			}
	
			zap.L().Debug(fmt.Sprintf("Perform action [%v] [%v]", reflect.TypeOf(action), action))
			go operator.Dispatch(action)
	
			w.WriteHeader(http.StatusOK)
		})
		zap.L().Debug("Steward starts")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			zap.L().Fatal("Cannot start server on the address",
				zap.String("service", "http"),
				zap.String("address", pushNotificationParam.WebhookURL),
				zap.Error(err))
		}
	},
	reflect.TypeOf(util.PullNotificationParam{}): func(operator *Operator, notificationParam util.NotificationParam) {
		pullNotificationParam := notificationParam.(util.PullNotificationParam)
		for true {
			operator.stateMux.Lock()
			actions, err := operator.pullRepoNotification(operator)
			operator.stateMux.Unlock()

			if err != nil {
				zap.L().Error(fmt.Sprintf("pullRepoNotification error [%v]", err))
			}
			for _, action := range(actions) {
				go operator.Dispatch(action)
			}
			time.Sleep(time.Duration(pullNotificationParam.PullPeriod) * time.Second)
		}
	},
}

func (operator *Operator) RemoteNotificationRegister(notificationParam util.NotificationParam) {
	if _, ok := remoteNoticationFunc[reflect.TypeOf(notificationParam)]; !ok {
		zap.L().Fatal(fmt.Sprintf("Invalid notification.type [%v]", reflect.TypeOf(notificationParam)))
	}

	remoteNoticationFunc[reflect.TypeOf(notificationParam)](operator, notificationParam)
}

// GrpcServerRegister : register grpc server
func (operator *Operator) GrpcServerRegister(grpcServer *grpc.Server) {
	operator.grpcServerRegister(grpcServer, operator)
}

func (operator *Operator) Dispatch(action util.Action) {
	operator.stateMux.Lock()
	defer operator.stateMux.Unlock()

	zap.L().Debug(" --- State Before Dispatching --- ",
		zap.String("type", fmt.Sprintf("%v", reflect.TypeOf(operator.state))),
		zap.String("payload", fmt.Sprintf("%v", operator.state)))

	zap.L().Debug(" --- Action --- ",
		zap.String("type", fmt.Sprintf("%v", reflect.TypeOf(action))),
		zap.String("payload", fmt.Sprintf("%v", action)))
	defer zap.L().Sync()

	if _, ok := operator.stateTransit[reflect.TypeOf(operator.state)][reflect.TypeOf(action)]; !ok {
		zap.L().Error(" *** Invalid transit *** ")
		return
	}

	nextState, thens := operator.stateTransit[reflect.TypeOf(operator.state)][reflect.TypeOf(action)](operator.state, action, operator)
	operator.state = nextState
	for _, then := range(thens) {
		go then()
	}
	zap.L().Debug(" --- State After Dispatching --- ", zap.String("type", fmt.Sprintf("%v", reflect.TypeOf(operator.state))), zap.String("payload", fmt.Sprintf("%v", operator.state)))
}

func (operator *Operator) GetPayload() interface{} {
	return operator.payload
}

func (operator *Operator) TrainFinish() {
	operator.trainFinish()
}

func newAggregateServerOperator(
	appGrpcServerURI string,
	trainPlanRepoGitHttpURL string,
	aggregatedModelRepoGitHttpURL string,
	_ string,
	edgeModelRepoGitHttpURLs []string,
	trainFinish func(),
) *Operator {
	return &Operator {
		aggregateServer.InitState,
		aggregateServer.StateTransit,
		aggregateServer.WebhookToAction,
		aggregateServer.PullRepoNotification,
		aggregateServer.GrpcServerRegister,
		aggregateServer.Payload {
			GrpcServerURI: appGrpcServerURI,
			TrainPlanRepoGitHttpURL: trainPlanRepoGitHttpURL,
			AggregatedModelRepoGitHttpURL: aggregatedModelRepoGitHttpURL,
			EdgeModelRepoGitHttpURLs: edgeModelRepoGitHttpURLs,
		},
		&sync.Mutex{},
		trainFinish,
	}
}

func newEdgeOperator(
	appGrpcServerURI string,
	trainPlanRepoGitHttpURL string,
	aggregatedModelRepoGitHttpURL string,
	edgeModelRepoGitHttpURL string,
	_ []string,
	trainFinish func(),
) *Operator {
	return &Operator {
		edge.InitState,
		edge.StateTransit,
		edge.WebhookToAction,
		edge.PullRepoNotification,
		edge.GrpcServerRegister,
		edge.Payload {
			GrpcServerURI: appGrpcServerURI,
			TrainPlanRepoGitHttpURL: trainPlanRepoGitHttpURL,
			AggregatedModelRepoGitHttpURL: aggregatedModelRepoGitHttpURL,
			EdgeModelRepoGitHttpURL: edgeModelRepoGitHttpURL,
		},
		&sync.Mutex{},
		trainFinish,
	}
}

var NewOperator = map[string] func(
	appGrpcServerURI string,
	trainPlanRepoGitHttpURL string,
	aggregatedModelRepoGitHttpURL string,
	edgeModelRepoGitHttpURL string,
	edgeModelRepoGitHttpURLs []string,
	trainFinish func(),
) *Operator {
	"aggregator": newAggregateServerOperator,
	"edge": newEdgeOperator,
}
