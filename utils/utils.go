package utils

import (
	"os"
	"path/filepath"
)

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
