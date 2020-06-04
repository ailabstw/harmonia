package edge

import "harmonia.com/steward/operator/util"

type idleState struct {
	util.State
}
type localTrainState struct {
	util.State
	roundRemain int
}

type aggregateState struct {
	util.State
	roundRemain int
}