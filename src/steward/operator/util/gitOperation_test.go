package util

import (
	"testing"
	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"

	"fmt"

	"io/ioutil"
	"os"
	"path/filepath"
)

func TestExecCommand(t *testing.T) {
	currentDir, _ := os.Getwd()

	testCases := map[string] struct{
		cmd string
		args []string
		path string
		env []string
		output string
		errMsg string
	} {
		"echo": {
			"echo",
			[]string{},
			".",
			[]string{},
			"\n",
			"",
		},
		"echo a": {
			"echo",
			[]string{"a"},
			".",
			[]string{},
			"a\n",
			"",
		},
		"echo a b": {
			"echo",
			[]string{"a", "b"},
			".",
			[]string{},
			"a b\n",
			"",
		},
		"printenv AAA": {
			"printenv",
			[]string{"AAA"},
			".",
			[]string{"AAA=aaa"},
			"aaa\n",
			"",
		},
		"pwd": {
			"pwd",
			[]string{},
			".",
			[]string{},
			currentDir + "\n",
			"",
		},
	}

	for caseName, testCase := range testCases {
		output, err := execCommand(testCase.cmd, testCase.args, testCase.path, testCase.env)
		assert.Equal(t, testCase.output, output, "Case [%s] fails.", caseName)
		if testCase.errMsg != "" {
			assert.EqualError(t, err, testCase.errMsg, "Case [%s] fails.", caseName)
		}
	}
}

func TestGitSetup(t *testing.T) {
	repoDir, _ := ioutil.TempDir("", "*-repo")
	defer os.RemoveAll(repoDir)

	execCommand("git", []string{"init"}, repoDir, []string{})
	GitSetup(GitUser{
		"testName",
		"testMail",
		"testToken",
	})
	defer os.Remove(credentialHelperScript)

	userName, _ := execCommand("git", []string{"config", "user.name"}, repoDir, []string{})
	assert.Equal(t, "testName\n", userName)
	userMail, _ := execCommand("git", []string{"config", "user.email"}, repoDir, []string{})
	assert.Equal(t, "testMail\n", userMail)
	scriptContent, err := ioutil.ReadFile(credentialHelperScript)
	if err != nil {
		fmt.Println(err)
	}
	assert.Equal(t, "printf '%s\\n' testToken", string(scriptContent))

	// clean up
	GitSetup(GitUser{"", "", ""})
}

func TestGitClone(t *testing.T) {
	remoteDir, _ := ioutil.TempDir("", "*-remote")
	defer os.RemoveAll(remoteDir)

	localDir, _ := ioutil.TempDir("", "*-local")
	defer os.RemoveAll(localDir)

	execCommand("git", []string{"init"}, remoteDir, []string{})
	gitCloneRepository(localDir, remoteDir)

	remoteURL, err := execCommand("git", []string{"remote", "get-url", "origin"}, localDir, []string{})

	assert.NoError(t, err)
	assert.Equal(t, remoteDir + "\n", remoteURL)
}

func TestGitOperation(t *testing.T) {
	suite.Run(t, new(GitOperationTestSuite))
}

type GitOperationTestSuite struct{
	suite.Suite
	remoteRepoURL string
	localRepoURL string
}

func (suite *GitOperationTestSuite) SetupSuite() {
	GitSetup(GitUser{
		"testName",
		"testMail",
		"testToken",
	})
}

func (suite *GitOperationTestSuite) BeforeTest(_, testName string) {
	suite.remoteRepoURL, _ = ioutil.TempDir("", fmt.Sprintf("*-%s-remote", testName))
	suite.localRepoURL, _ = ioutil.TempDir("", fmt.Sprintf("*-%s-local", testName))
	execCommand("git", []string{"init"}, suite.remoteRepoURL, []string{})
	gitCloneRepository(suite.localRepoURL, suite.remoteRepoURL)
}

func (suite *GitOperationTestSuite) AfterTest(_, _ string) {
	os.RemoveAll(suite.remoteRepoURL)
	os.RemoveAll(suite.localRepoURL)
}

func (suite *GitOperationTestSuite) TestCommit() {
	files := []string{
		"aaa",
		"bbb",
		"ccc",
	}
	for _, file := range files {
		os.Create(filepath.Join(suite.remoteRepoURL, file))
	}

	err := gitCommitAll(suite.remoteRepoURL, "commit message")
	suite.Assert().NoError(err)

	commitMessage, err := execCommand(
		"git",
		[]string{"log", "-l1", "--pretty=format:%s"},
		suite.remoteRepoURL,
		[]string{},
	)
	suite.Assert().Equal("commit message", commitMessage)
}

func (suite *GitOperationTestSuite) TestBranch() {
	gitCommitAll(suite.remoteRepoURL, "first commit")

	err := gitBranch(suite.remoteRepoURL, "branch")
	suite.Assert().NoError(err)

	_, err = execCommand(
		"git",
		[]string{"log", "branch"},
		suite.remoteRepoURL,
		[]string{},
	)
	suite.Assert().NoError(err)
}

// TODO: Modify Parameter / Add Case
func (suite *GitOperationTestSuite) TestCheckout() {
	branchName := "otherBranch"
	gitCommitAll(suite.remoteRepoURL, "first commit")
	gitBranch(suite.remoteRepoURL, branchName)

	err := gitCheckout(suite.remoteRepoURL, branchName, "")
	suite.Assert().NoError(err)

	branch, err := execCommand(
		"git",
		[]string{"rev-parse", "--abbrev-ref", "HEAD"},
		suite.remoteRepoURL,
		[]string{},
	)
	suite.Assert().NoError(err)
	suite.Assert().Equal(branchName + "\n", branch)
}

// TODO: TestPush