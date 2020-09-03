package edge

import (
	"fmt"

	"harmonia.com/steward/operator/util"
)

type trainPlanAction struct {
	util.Action
	trainPlan util.TrainPlan
}
type initMessageResponseAction struct {
	util.Action
}
type pretrainedModelReadyAction struct {
	util.Action
}
type trainStartAction struct {
	util.Action
}
type trainCleanupAction struct {
	util.Action
	roundCount int
}
type trainFinishAction struct {
	util.Action
	errCode int
	datasetSize int
	metadata map[string]string
	metrics map[string]float64
}
type baseModelReceivedAction struct {
	util.Action
	ref string
}

// WebhookToAction : perform an action according to the content of the webhook
func WebhookToAction(webhook *util.Webhook, operator util.AbstractOperator) (util.Action, error) {
	if repoFullname, _ := util.GitHttpURLToRepoFullName(operator.GetPayload().(Payload).AggregatedModelRepoGitHttpURL); repoFullname == webhook.Repo.FullName {
		return &baseModelReceivedAction {
			ref: webhook.Ref,
		}, nil
	}

	if repoFullname, _ := util.GitHttpURLToRepoFullName(operator.GetPayload().(Payload).TrainPlanRepoGitHttpURL); repoFullname == webhook.Repo.FullName {
		plan, err := util.GetTrainPlanData(operator.GetPayload().(Payload).TrainPlanRepoGitHttpURL)
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
	actions := []util.Action{}
	operatorPayload := operator.GetPayload().(Payload)

	if branchSet, err := util.CheckUpdatedBranches(operatorPayload.TrainPlanRepoGitHttpURL); err == nil {
		if _, update := branchSet["master"]; update {
			plan, err := util.GetTrainPlanData(operatorPayload.TrainPlanRepoGitHttpURL)
			if err == nil {
				actions = append(actions, &trainPlanAction {
					trainPlan: *plan,
				})
			}
		}
	}

	if branchSet, err := util.CheckUpdatedBranches(operatorPayload.AggregatedModelRepoGitHttpURL); err == nil {
		for branch, _ := range(branchSet) {
			actions = append(actions, &baseModelReceivedAction {
				ref: fmt.Sprintf("refs/heads/%s", branch),
			})
		}
	}

	return actions, nil
}