package edge

import (
	"fmt"

	"harmonia.com/steward/operator/util"
)

type trainPlanAction struct {
	util.Action
}
type trainInvitationReplyAction struct {
	util.Action
}
type trainStartAction struct {
	util.Action
	trainPlan util.TrainPlan
}
type roundStartAction struct {
	util.Action
}
type trainFinishAction struct {
	util.Action
}
type aggregatedModelReceivedAction struct {
	util.Action
}

// WebhookToAction : perform an action according to the content of the webhook
func WebhookToAction(webhook *util.Webhook) (util.Action, error) {
	if util.GitHttpURLToRepoFullName(util.Config.AggregatedModelRepo.GitHttpURL) == webhook.Repo.FullName {
		return &aggregatedModelReceivedAction{}, nil
	}

	if util.GitHttpURLToRepoFullName(util.Config.TrainPlanRepo.GitHttpURL) == webhook.Repo.FullName {
		plan, err := util.GetTrainPlanData()
		if err != nil {
			return nil, err
		}

		return &trainStartAction{
			trainPlan: *plan,
		}, nil
	}

	return nil, fmt.Errorf("invalid webhook")
}
