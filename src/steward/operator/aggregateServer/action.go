package aggregateServer

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
type roundStartAction struct {
	util.Action
	trainPlan util.TrainPlan
}
type localTrainFinishAction struct {
	util.Action
	gitHttpURL string
}
type aggregateFinishAction struct {
	util.Action
}

// WebhookToAction : perform an action according to the content of the webhook
func WebhookToAction(webhook *util.Webhook) (util.Action, error) {
	edgeRepos := util.Config.EdgeModelRepos
	for _, edgeRepo := range edgeRepos {
		if util.GitHttpURLToRepoFullName(edgeRepo.GitHttpURL) == webhook.Repo.FullName {
			return &localTrainFinishAction{
				gitHttpURL: edgeRepo.GitHttpURL,
			}, nil
		}
	}

	trainPlanRepo := util.Config.TrainPlanRepo
	if util.GitHttpURLToRepoFullName(trainPlanRepo.GitHttpURL) == webhook.Repo.FullName {
		plan, err := util.GetTrainPlanData()
		if err != nil {
			return nil, err
		}

		plan.PlanHash = webhook.LatestCommit[0:6]
		return &roundStartAction{
			trainPlan: *plan,
		}, nil
	}

	return nil, fmt.Errorf("invalid webhook")
}
