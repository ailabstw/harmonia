package util

import (
	"reflect"

	"google.golang.org/grpc"
)

type State interface {}
type Action interface {}
type StateTransit map[reflect.Type] map[reflect.Type] func(State, Action, AbstractOperator) (State, []func())

type WebhookToAction func(*Webhook, AbstractOperator) (Action, error)
type PullRepoNotification func(AbstractOperator) ([]Action, error)

type NotificationParam interface {}
type PushNotificationParam struct {
	NotificationParam
	WebhookURL string
}
type PullNotificationParam struct {
	NotificationParam
	PullPeriod int
}
type GrpcServerRegisterFunc func(*grpc.Server, AbstractOperator)

type AbstractOperator interface {
	RemoteNotificationRegister(NotificationParam)
	GrpcServerRegister(*grpc.Server)
	Dispatch(Action)
	GetPayload() interface{}
	TrainFinish()
}

type Repository struct {
	FullName string `json:"full_name"`
	Branch string `json:"default_branch"`
}

type Webhook struct {
	Repo         Repository `json:"repository"`
	Ref          string     `json:"ref"`
}

type TrainPlan struct {
	Name                    string `json:"name"`
	RoundCount              int `json:"round"`
	EdgeCount               int `json:"edge"`
	EpochCount     	        int `json:"EpR"`
	Timeout                 int `json:"timeout"`
    PretrainedModelCommitID string `json:"pretrainedModel"`

	CommitID string
}
