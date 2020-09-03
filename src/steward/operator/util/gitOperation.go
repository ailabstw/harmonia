package util

import (
	"bytes"
	"fmt"
	"io"
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
	out, err := execGitCommand([]string{"config", "--global", "user.email", gitUser.Email}, ".")
	if err != nil {
		zap.L().Warn("git config setup error", zap.String("output", out), zap.Error(err))
	}
	out, err = execGitCommand([]string{"config", "--global", "user.name", gitUser.Name}, ".")
	if err != nil {
		zap.L().Warn("git config setup error", zap.String("output", out), zap.Error(err))
	}
}

func createCredHelperScript(userToken string) {
	txt := []byte("printf '%s\\n' " + userToken)
	err := ioutil.WriteFile(credentialHelperScript, txt, 0710)
	if err != nil {
		zap.L().Error("create credential helper error", zap.Error(err))
	}
	zap.L().Debug("create credential helper script complete")
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
	_, err := execGitCommand([]string{"add", "--all"}, repoPath)
	if err != nil {
		return err
	}
	_, err = execGitCommand([]string{"commit", "-a", "-m", message, "--allow-empty"}, repoPath)
	if err != nil {
		return err
	}
	return nil
}

func gitTag(repoPath string, tagName string) error {
	_, err := execGitCommand([]string{"tag", tagName}, repoPath)
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
	_, err := execGitCommand([]string{"branch", branchName}, repoPath)
	if err != nil {
		return err
	}
	return nil
}

func gitCheckout(repoPath string, ref string, branch string) error {
	args := []string{"checkout"}

	if ref != "" {
		args = append(args, ref)
	}

	if branch != "" {
		args = append(args, "-b", branch)
	}

	_, err := execGitCommand(args, repoPath)
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
	_, err := execGitCommand([]string{"lfs", "track", filename}, repoPath)
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

func execGitCommand(args []string, cwd string) (string, error) {
	return execCommand("git", args, cwd, []string{})
}

func removeAllLocalBranch(cwd string) (string, error) {
	// git for-each-ref --format '%(refname:short)' refs/heads | xargs git branch -D
	cmd1 := exec.Command("git", "for-each-ref", "--format", "%(refname:short)", "refs/heads")
	cmd2 := exec.Command("xargs", "git", "branch", "-D")

	if cwd != "" {
		cmd1.Dir = cwd
		cmd2.Dir = cwd
	}

	read, write := io.Pipe() 
	cmd1.Stdout = write
	cmd2.Stdin = read

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd2.Stdout = &stdout
	cmd2.Stderr = &stderr

	cmd1.Start()
	cmd2.Start()
	cmd1.Wait()
	write.Close()
	cmd2.Wait()

	return stdout.String(), nil
}

func gitCheckUpdatedBranches(repoPath string) (map[string]struct{}, error) {
	if err := retryOperation(func() error {
		zap.L().Debug("fetch")
		return gitFetch(repoPath)
	}); err != nil {
		return nil, err
	}

	branchOutputToSet := func (output string, prefix string) map[string]string {
		branchSet := map[string]string{}
		for _, branch := range(strings.Split(output, "\n")) {
			tmp := strings.Split(branch, " ")
			if len(tmp) != 2 {
				continue
			}
			if tmp[0][len(prefix):] == "HEAD" {
				continue
			}
			branchSet[tmp[0][len(prefix):]] = tmp[1]
		}

		return branchSet
	}

	getBranchSet := func (pattern string) (map[string]string, error) {
		branchOutput, err := execGitCommand([]string{"for-each-ref", "--format", "%(refname) %(objectname)", pattern}, repoPath)
		if err != nil {
			return nil, err
		}

		return branchOutputToSet(branchOutput, pattern), nil
	}

	remoteBranchSet, err := getBranchSet("refs/remotes/origin/")
	if err != nil {
		return nil, err
	}

	localBranchSet, err := getBranchSet("refs/heads/")
	if err != nil {
		return nil, err
	}

	updatedBranches := map[string]struct{}{}
	for branch, remoteCommit := range(remoteBranchSet) {
		localCommit, branchExist := localBranchSet[branch]
		if !branchExist || localCommit != remoteCommit {
			updatedBranches[branch] = struct{}{}
		}
	}

	return updatedBranches, nil
}