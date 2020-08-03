package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

type GitUser struct{
	Name string
	Email string
	Token string
}

var credentialHelperScript = "/tmp/credentialHelper.sh"

func GitSetup(gitUser GitUser) {
	createCredHelperScript(gitUser.Token)
	out, err := execCommand("git", []string{"config", "--global", "user.email", gitUser.Email}, ".", []string{})
	if err != nil {
		zap.L().Warn("git config setup error", zap.String("output", out), zap.Error(err))
	}
	out, err = execCommand("git", []string{"config", "--global", "user.name", gitUser.Name}, ".", []string{})
	if err != nil {
		zap.L().Warn("git config setup error", zap.String("output", out), zap.Error(err))
	}
}

func createCredHelperScript(userToken string) {
	txt := []byte("printf '%s\\n' " + userToken)
	err := ioutil.WriteFile(credentialHelperScript, txt, 0710)
	if err != nil {
		fmt.Println("create credential helper error", zap.Error(err))
		zap.L().Error("create credential helper error", zap.Error(err))
	}
	zap.L().Debug("create credential helper script complete")
	fmt.Println("create credential helper script complete")
}

func gitCloneRepository(repoPath string, gitURL string) error {
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); !os.IsNotExist(err) {
		zap.L().Debug(fmt.Sprintf("Repo [%v] existed, skipped clone [%v]", repoPath, gitURL))
		return nil
	}

	zap.L().Debug(fmt.Sprintf("Cloning Data from [%v] [%v]...", gitURL, repoPath))
	os.MkdirAll(repoPath, 0755)

	_, err := execGitPasswordCommand([]string{"clone", gitURL, repoPath}, repoPath)

	if err != nil {
		zap.L().Error(fmt.Sprintf("Clone fail [%v]", err))
		return err
	}

	return nil
}

func gitFetch(repoPath string) error {
	zap.L().Info(fmt.Sprintf("Fetch Data [%v]...", repoPath))
	_, err := execGitPasswordCommand([]string{"fetch"}, repoPath)
	if err != nil {
		return err
	}

	zap.L().Info("Fetch Succeed")
	return nil
}

func gitPull(repoPath string) error {
	zap.L().Info(fmt.Sprintf("Pulling Data [%v]...", repoPath))
	_, err := execGitPasswordCommand([]string{"pull"}, repoPath)
	if err != nil {
		return err
	}

	zap.L().Info("Pull Succeed")
	return nil
}

func gitCommitAll(repoPath string, message string) error {
	lfsCheck(repoPath)
	_, err := execCommand("git", []string{"add", "--all"}, repoPath, []string{})
	if err != nil {
		return err
	}
	_, err = execCommand("git", []string{"commit", "-a", "-m", message, "--allow-empty"}, repoPath, []string{})
	if err != nil {
		return err
	}
	return nil
}

func gitTag(repoPath string, tagName string) error {
	_, err := execCommand("git", []string{"tag", tagName}, repoPath, []string{})
	return err
}

func gitPushArgs(repoPath string, args []string) error {
	zap.L().Info(fmt.Sprintf("Pushing to [%v] args [%v]...", repoPath, args))

	_, err := execGitPasswordCommand(
		append([]string{"push", "origin"}, args...),
		repoPath,
	)
	if err != nil {
		return err
	}

	zap.L().Info("Push Succeed")
	return nil
}

func gitPushRefs(repoPath string, commits []string) error {
	return gitPushArgs(repoPath, commits)
}

func gitPushAll(repoPath string) error {
	return gitPushArgs(repoPath, []string{"--all"})
}

func gitPushTags(repoPath string) error {
	return gitPushArgs(repoPath, []string{"--tags"})
}

func gitBranch(repoPath string, branchName string) error {
	_, err := execCommand("git", []string{"branch", branchName}, repoPath, []string{})
	if err != nil {
		return err
	}
	return nil
}

func gitCheckout(repoPath string, commit string, branch string) error {
	args := []string{"checkout", commit}

	if branch != "" {
		args = append(args, "-b", branch)
	}

	_, err := execCommand("git", args, repoPath, []string{})
	if err != nil {
		return err
	}
	return nil
}

func lfsCheck(repoPath string) {
	filepath.Walk(repoPath, processFile(repoPath))
}

func processFile(repoPath string) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {
		if err != nil {
			zap.L().Fatal("", zap.Error(err))
		}

		// ignore .git directory
		if f.IsDir() && f.Name() == ".git" {
			return filepath.SkipDir
		}

		switch mode := f.Mode(); {
		case mode.IsDir():
			// skip directory
			return nil
		case mode.IsRegular():
			if !isTextFile(path) {
				err := lfsTrackFile(f.Name(), repoPath)
				if err != nil {
					zap.L().Fatal("git lfs track file error", zap.Error(err))
				}
			}
		}
		return nil
	}
}

// To Rewrite
func isTextFile(path string) bool {
	out, err := execCommand("file", []string{path}, ".", []string{})
	if err != nil {
		zap.L().Fatal("run file command fail", zap.Error(err))
	}
	return strings.Contains(out, "text") || strings.Contains(out, "JSON")
}

func lfsTrackFile(filename string, repoPath string) error {
	zap.L().Info("git lfs track file", zap.String("file", filename))
	_, err := execCommand("git", []string{"lfs", "track", filename}, repoPath, []string{})
	if err != nil {
		return err
	}
	return nil
}

func execGitPasswordCommand(args []string, path string) (string, error) {
	return execCommand(
		"git",
		args,
		path,
		append(os.Environ(), "GIT_ASKPASS=" + credentialHelperScript),
	)
}

func execCommand(name string, args []string, path string, env []string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if path != "" {
		cmd.Dir = path
	}
	if len(env) > 0 {
		cmd.Env = env
	}
	err := cmd.Run()
	zap.L().Debug("exec command",
		zap.String("command", name+" "+strings.Join(args, " ")),
		zap.String("output", stdout.String()))
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}
