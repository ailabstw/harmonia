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
type localModelPullFinishAction struct {
	util.Action
	gitHttpURL string
	datasetSize int
	metadata map[string]string
	metrics map[string]float64
}
type localTrainTimeoutAction struct {
	util.Action
	roundCount int
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

		return &trainPlanAction {
			trainPlan: *plan,
		}, nil
	}

	return nil, fmt.Errorf("invalid webhook")
}

func PullRepoNotification(operator util.AbstractOperator) ([]util.Action, error) {
	operatorPayload := operator.GetPayload().(Payload)

	isRepoUpdated := func(gitURL string) bool {
		if branchSet, err := util.CheckUpdatedBranches(gitURL); err == nil {
			_, updated := branchSet["master"]
			return updated
		}
		return false
	}

	actions := []util.Action{}
	if isRepoUpdated(operatorPayload.TrainPlanRepoGitHttpURL) {
		plan, err := util.GetTrainPlanData(operatorPayload.TrainPlanRepoGitHttpURL)
		if err == nil {
			actions = append(actions, &trainPlanAction {
				trainPlan: *plan,
			})
		}
	}

	for _, edgeRepoGitURL := range operatorPayload.EdgeModelRepoGitHttpURLs {
		if isRepoUpdated(edgeRepoGitURL) {
			actions = append(actions, &localTrainFinishAction {
				gitHttpURL: edgeRepoGitURL,
			})
		}
	}

	return actions, nil
}
