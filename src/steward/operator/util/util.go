package util

import (
	"os"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"encoding/json"
	//"go.uber.org/zap"
)

func WriteMapToFile(_m map[string]interface{}, fullFilename string) error {
	json, err := json.Marshal(_m)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fullFilename, []byte(json), 0644)
	if err != nil {
		return err
	}
	return nil
}

func ReadMapToFile(fullFilename string) (map[string]interface{}, error) {
	stat, err := os.Stat(fullFilename)
    if os.IsNotExist(err) {
        return map[string]interface{} {}, nil
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("[%s] is directory, not file", fullFilename)
	}

	bytes, err := ioutil.ReadFile(fullFilename)
	if err != nil {
		return nil, err
	}

	var _m map[string]interface{}
	err = json.Unmarshal(bytes, &_m)
	if err != nil {
		return nil, err
	}
	return _m, nil
}

func WriteMetadata(gitHttpURL string, metadata map[string]interface{}) error {
	repoPath, err := getRepoPath(gitHttpURL)
	if err != nil {
		return err
	}
	return WriteMapToFile(metadata, filepath.Join(repoPath, ".harmonia"))
}

func ReadMetadata(gitHttpURL string) (map[string]interface{}, error) {
	repoPath, err := getRepoPath(gitHttpURL)
	if err != nil {
		return nil, err
	}
	
	repoMetadata, err := ReadMapToFile(filepath.Join(repoPath, ".harmonia"))
	if err != nil {
		return nil, err
	}

	if _, ok := repoMetadata["metadata"]; !ok {
		repoMetadata["metadata"] = map[string]interface{} {}
	}
	if _, ok := repoMetadata["metrics"]; !ok {
		repoMetadata["metrics"] = map[string]interface{} {}
	}
	return repoMetadata, nil
}