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
func WebhookToAction(webhook *util.Webhook, operator util.AbstractOperator) (util.Action, error) {
	if repoName, _ := util.GitHttpURLToRepoFullName(operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL); repoName == webhook.Repo.FullName {
		return &aggregatedModelReceivedAction{}, nil
	}

	if repoName, _ := util.GitHttpURLToRepoFullName(operator.GetPayload().(Payload).TrainPlanRepoGitHttpURL); repoName == webhook.Repo.FullName {
		plan, err := util.GetTrainPlanData(operator.GetPayload().(Payload).TrainPlanRepoGitHttpURL)
		if err != nil {
			return nil, err
		}

		return &trainStartAction{
			trainPlan: *plan,
		}, nil
	}

	return nil, fmt.Errorf("invalid webhook")
}
