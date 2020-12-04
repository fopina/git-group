package commands

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

var version string = "DEV"
var date string

const repo = "fopina/git-group"

// WorkDirConfig persists the options used to initially clone a group to re-use them for pull commands
type WorkDirConfig struct {
	GroupURL   string
	Depth      int
	SampleSize int `json:"SampleSize,omitempty"`
	Recursive  bool
}

// SaveConfig saves current group config to .gitgroup
func (h *WorkDirConfig) SaveConfig(groupDir string) error {
	confPath := filepath.Join(groupDir, ".gitgroup")

	file, err := json.MarshalIndent(h, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(confPath, file, 0600)
	return err
}

// FindConfig will return the first .gitgroup file that it finds going up the directory tree
func FindConfig(path string) (string, error) {
	x, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	y := ""
	var conf string
	for y != x {
		y = x
		conf = filepath.Join(y, ".gitgroup")
		if _, err := os.Stat(conf); !os.IsNotExist(err) {
			return conf, nil
		}
		x = filepath.Dir(y)
	}
	return "", os.ErrNotExist
}

// LoadConfig loads group config from the "closest" .gitgroup
func (h *WorkDirConfig) LoadConfig(currentDir string) (string, error) {
	groupConf, err := FindConfig(currentDir)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadFile(groupConf)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(data, h)
	return filepath.Dir(groupConf), err
}
