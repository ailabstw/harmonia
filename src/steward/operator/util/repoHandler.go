package util

import (
	"fmt"

	"io/ioutil"
	"path/filepath"
	"encoding/json"

	"go.uber.org/zap"
)

func GetTrainPlanData(gitHttpURL string) (*TrainPlan, error) {
	zap.L().Info("get train plan data...")
	PullData(gitHttpURL)
	repoFullName, err := GitHttpURLToRepoFullName(gitHttpURL)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(
		// TODO: get filename of plan by argument ?
		filepath.Join(baseDir + repoFullName, "plan.json"),
	)
	if err != nil {
		zap.L().Error("cannot read file", zap.Error(err))
		return nil, fmt.Errorf("No plan.json in repository")
	}
	var plan TrainPlan
	err = json.Unmarshal(data, &plan)
	if err != nil {
		zap.L().Error("unmarshal json error", zap.Error(err))
		return nil, err
	}
	zap.L().Debug("", zap.String("train plan", fmt.Sprintf("%v", plan)))
	return &plan, nil
}
