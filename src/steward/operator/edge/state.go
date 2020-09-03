package edge

import "harmonia.com/steward/operator/util"

type idleState struct {
	util.State
}

type trainInitState struct {
	util.State
	init bool
	pretrainedModel bool
	trainPlan util.TrainPlan
}

type localTrainState struct {
	util.State
	trainPlan   util.TrainPlan
	roundCount int
}

type localTrainInterruptedState struct {
	util.State
	trainPlan util.TrainPlan
	roundCount int
}

type aggregateState struct {
	util.State
	trainPlan   util.TrainPlan
	roundCount int
}

// ----

type baseModel struct {
	gitHttpURL string
	metadata map[string]string
	metrics map[string]float64
}
