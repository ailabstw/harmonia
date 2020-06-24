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
type pullFinishAction struct {
	util.Action
	gitHttpURL string
}
type aggregateFinishAction struct {
	util.Action
}

// WebhookToAction : perform an action according to the content of the webhook
func WebhookToAction(webhook *util.Webhook, operator util.AbstractOperator) (util.Action, error) {
	operatorPayload := operator.GetPayload().(Payload)
	for _, edgeRepoGitURL := range operatorPayload.EdgeModelRepoGitHttpURLs {
		repoFullName, err := util.GitHttpURLToRepoFullName(edgeRepoGitURL)
		if err != nil {
			continue
		}
		if repoFullName == webhook.Repo.FullName {
			return &localTrainFinishAction{
				gitHttpURL: edgeRepoGitURL,
			}, nil
		}
	}

	repoFullName, _ := util.GitHttpURLToRepoFullName(operatorPayload.TrainPlanRepoGitHttpURL)
	if repoFullName == webhook.Repo.FullName {
		plan, err := util.GetTrainPlanData(operatorPayload.TrainPlanRepoGitHttpURL)
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
