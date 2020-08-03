package aggregateServer

import "harmonia.com/steward/operator/util"

type idleState struct {
	util.State
}
type waitPretrainModelState struct {
	util.State
	trainPlan util.TrainPlan
}
type localTrainState struct {
	util.State
	trainPlan util.TrainPlan
	roundCount int
	webhooks []string
	edgeModels []localModel
	// TODO
	// baseModelCommitID string
}
type aggregateState struct {
	util.State
	trainPlan util.TrainPlan
	roundCount int
}

//-----

type localModel struct {
	gitHttpURL string
	datasetSize int
	metadata map[string]string
	metrics map[string]float64
}
