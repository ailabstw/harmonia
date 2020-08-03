package util

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGitHttpURLToRepoFullName(t *testing.T) {
	testCases := map[string] struct{
		gitURL string
		fullName string
		errMsg string
	} {
		"Http": {
			"http://username@github.com/owner/git.git",
			"owner/git",
			"",
		},
		"Https": {
			"https://username@github.com/owner/git.git",
			"owner/git",
			"",
		},
		"Without suffix": {
			"http://username@github.com/owner/git",
			"owner/git",
			"",
		},
		"Without username": {
			"https://github.com/owner/git",
			"",
			fmt.Sprintf("Unsupported git URL: [https://github.com/owner/git]"),
		},
	}

	for caseName, testCase := range testCases {
		fullName, err := GitHttpURLToRepoFullName(testCase.gitURL)

		if testCase.errMsg != "" {
			assert.EqualError(
				t,
				err,
				testCase.errMsg,
				"Case [%s] fails.", caseName,
			)
		} else {
			assert.Equal(
				t,
				testCase.fullName,
				fullName,
				"Case [%s] fails.", caseName,
			)
		}
	}
}
