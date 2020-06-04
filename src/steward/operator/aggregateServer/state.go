package aggregateServer

import "harmonia.com/steward/operator/util"

type idleState struct {
	util.State
}
type localTrainState struct {
	util.State
	trainPlan util.TrainPlan
	roundRemain int
	edgeModels []string
}
type aggregateState struct {
	util.State
	trainPlan util.TrainPlan
	roundRemain int
}