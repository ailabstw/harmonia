package util

import (
	"reflect"
	"net/http"

	"google.golang.org/grpc"
)

type State interface {}
type Action interface {}
// type StateTransit map[reflect.Type] map[reflect.Type] func(State, Action) State
type StateTransit map[reflect.Type] map[reflect.Type] func(*State, Action) // This is workaround

type WebhookToAction func(*Webhook) (Action, error)
type HttpHandleFunc func(http.ResponseWriter, *http.Request)
type GrpcServerRegisterFunc func(*grpc.Server, AbstractOperator)

type AbstractOperator interface {
	HttpHandleFunc() HttpHandleFunc
	GrpcServerRegister(*grpc.Server)
	Dispatch(Action)
}

type Repository struct {
	FullName string `json:"full_name"`
}

type Webhook struct {
	Repo         Repository `json:"repository"`
	LatestCommit string     `json:"after"`
}

type TrainPlan struct {
	RoundCount     int `json:"roundCount"`
	EdgeCount      int `json:"edgeCount"`
	PlanHash       string
}
