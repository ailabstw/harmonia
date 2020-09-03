package util

import (
	"fmt"
	"regexp"
	"strings"

	"io/ioutil"
	"path/filepath"
	"encoding/json"

	"go.uber.org/zap"
)

var baseDir = "/repos/"
var maxRetryCount = 5

// TODO: rollback
func retryOperation(operation func() error) error {
	var err error
	for tryCount := 0 ; tryCount < maxRetryCount ; tryCount++ {
		if err = operation(); err != nil {
			continue
		}
		return nil
	}
	return err
}

func trainNameFormat(trainName string) string {
	re := regexp.MustCompile(`[./?@*\[\]{}\\]+`)
	return re.ReplaceAllString(trainName, "_")
}

func InferenceTag(trainName string, trainPlanID string) string {
	if trainName == "" {
		return strings.Join([]string{"inference", trainPlanID}, "-")
	}
	return strings.Join([]string{"inference", trainNameFormat(trainName), trainPlanID}, "-")
}

func TrainBranch(trainName string, trainPlanID string) string {
	if trainName == "" {
		return strings.Join([]string{"AnonymousTask", trainPlanID}, "-")
	}
	return strings.Join([]string{trainNameFormat(trainName), trainPlanID}, "-")
}

func GitHttpURLToRepoFullName(gitHttpURL string) (string, error) {
	// Modified https://regex101.com/library/BuA5xF
	re := regexp.MustCompile(`(?P<method>https?):\/\/(?:[\w_-]+@)(?P<provider>.*?(?P<port>\:\d+)?)(?:\/|:)(?P<handle>(?P<owner>.+?)\/(?P<repo>.+?))(?:\.git|\/)?$`)
	if !re.MatchString(gitHttpURL) {
		return "", fmt.Errorf("Unsupported git URL: [%s]", gitHttpURL)
	}
	return re.ReplaceAllString(gitHttpURL, "${owner}/${repo}"), nil
}

func getRepoPath(gitHttpURL string) (string, error) {
	repoFullName, err := GitHttpURLToRepoFullName(gitHttpURL)
	if err != nil {
		return "", err
	}

	return filepath.Join(baseDir, repoFullName), nil
}

func GetTrainPlanData(gitHttpURL string) (*TrainPlan, error) {
	zap.L().Info("get train plan data...")
	PullData(gitHttpURL)
	repoFullName, err := GitHttpURLToRepoFullName(gitHttpURL)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(
		// TODO: get filename of plan by argument ?
		filepath.Join(baseDir, repoFullName, "plan.json"),
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
	zap.L().Debug(fmt.Sprintf("get train plan [%v]", plan))
	plan.CommitID, _ = execGitCommand([]string{"rev-parse", "--short", "HEAD"}, filepath.Join(baseDir, repoFullName))
	plan.CommitID = strings.TrimSpace(plan.CommitID)
	return &plan, nil
}

func CloneRepository(gitHttpURL string) error {
	repoPath, err := getRepoPath(gitHttpURL)
	if err != nil {
		return err
	}

	if err = retryOperation(func() error {
		return gitCloneRepository(repoPath, gitHttpURL)
	}); err != nil {
		return err
	}

	return nil
}

func PullData(gitHttpURL string) error {
	repoPath, err := getRepoPath(gitHttpURL)
	if err != nil {
		return err
	}

	if err = retryOperation(func() error {
		return gitPull(repoPath)
	}); err != nil {
		return err
	}

	return nil
}

func PushUpdates(gitHttpURL string, tag string) error {
	repoPath, err := getRepoPath(gitHttpURL)
	if err != nil {
		return err
	}

	if err = retryOperation(func() error {
		return gitCommitAll(repoPath, "Harmonia Commit")
	}); err != nil {
		return err
	}

	if err = retryOperation(func() error {
		if tag != "" {
			return gitTag(repoPath, tag)
		}
		return nil
	}); err != nil {
		return err
	}

	if err = retryOperation(func() error {
		return gitPushAll(repoPath)
	}); err != nil {
		return err
	}

	if err = retryOperation(func() error {
		return gitPushTags(repoPath)
	}); err != nil {
		return err
	}

	return nil
}

func CreateGlobalModelBranch(gitHttpURL string, trainName string, trainPlanID string, pretrainedModelID string) error {
	zap.L().Info("Initial Training Branch...")
	
	zap.L().Debug(fmt.Sprintf("gitHttpURL: [%v]", gitHttpURL))
	repoPath, err := getRepoPath(gitHttpURL)
	if err != nil {
		return err
	}

	if err = retryOperation(func() error {
		zap.L().Debug("fetch")
		return gitFetch(repoPath)
	}); err != nil {
		return err
	}

	if err = retryOperation(func() error {
		_, err := execGitCommand([]string{"checkout", "--detach"}, repoPath)
		return err
	}); err != nil {
		return err
	}

	removeAllLocalBranch(repoPath)

	if err = retryOperation(func() error {
		zap.L().Debug(fmt.Sprintf("checkout: [%v]", pretrainedModelID))
		return gitCheckout(repoPath, pretrainedModelID, "")
	}); err != nil {
		return err
	}

	if err = retryOperation(func() error {
		zap.L().Debug(fmt.Sprintf("checkout -b [%v]", TrainBranch(trainName, trainPlanID)))
		return gitCheckout(repoPath, "", TrainBranch(trainName, trainPlanID))
	}); err != nil {
		return err
	}

	if err = retryOperation(func() error {
		zap.L().Debug(fmt.Sprintf("push: [%v]", []string{TrainBranch(trainName, trainPlanID)}))
		return gitPushRefs(repoPath, []string{TrainBranch(trainName, trainPlanID)})
	}); err != nil {
		return err
	}

	return nil
}

func CheckoutPretrainedModel(gitHttpURL string, trainName string, trainPlanID string) error {
	zap.L().Info("Checking out to Pretrained Model...")

	zap.L().Debug(fmt.Sprintf("gitHttpURL: [%v]", gitHttpURL))
	repoPath, err := getRepoPath(gitHttpURL)
	if err != nil {
		return err
	}

	if err = retryOperation(func() error {
		zap.L().Debug("fetch")
		return gitFetch(repoPath)
	}); err != nil {
		return err
	}

	if err = retryOperation(func() error {
		zap.L().Debug(fmt.Sprintf("checkout: [%v]", TrainBranch(trainName, trainPlanID)))
		return gitCheckout(repoPath, TrainBranch(trainName, trainPlanID), "")
	}); err != nil {
		return err
	}

	return nil
}

func CheckUpdatedBranches(gitHttpURL string) (map[string]struct{}, error) {
	repoPath, err := getRepoPath(gitHttpURL)
	if err != nil {
		return nil, err
	}
	return gitCheckUpdatedBranches(repoPath)
}