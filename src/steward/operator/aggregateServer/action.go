package aggregateServer

import (
	"fmt"

	"harmonia.com/steward/operator/util"
)

type trainPlanAction struct {
	trainPlan util.TrainPlan
	util.Action
}
type trainStartAction struct {
	util.Action
}
type localTrainFinishAction struct {
	util.Action
	gitHttpURL string
}
type pullFinishAction struct {
	util.Action
	gitHttpURL string
	datasetSize int
	metadata map[string]string
	metrics map[string]float64
}
type aggregateFinishAction struct {
	util.Action
	errCode int
	metadata map[string]string
	metrics map[string]float64
}

// WebhookToAction : perform an action according to the content of the webhook
func WebhookToAction(webhook *util.Webhook, operator util.AbstractOperator) (util.Action, error) {
	operatorPayload := operator.GetPayload().(Payload)
	for _, edgeRepoGitURL := range operatorPayload.EdgeModelRepoGitHttpURLs {
		repoFullname, err := util.GitHttpURLToRepoFullName(edgeRepoGitURL)
		if err != nil {
			continue
		}
		if repoFullname == webhook.Repo.FullName {
			return &localTrainFinishAction{
				gitHttpURL: edgeRepoGitURL,
			}, nil
		}
	}

	if repoFullname, _ := util.GitHttpURLToRepoFullName(operatorPayload.TrainPlanRepoGitHttpURL); repoFullname == webhook.Repo.FullName {
		plan, err := util.GetTrainPlanData(operatorPayload.TrainPlanRepoGitHttpURL)
		if err != nil {
			return nil, err
		}

		plan.CommitID = webhook.LatestCommit[0:6]
		return &trainPlanAction {
			trainPlan: *plan,
		}, nil
	}

	return nil, fmt.Errorf("invalid webhook")
}
