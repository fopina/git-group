package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fopina/git-group/utils"
)

// PullCommand implements the cli.Command interface to run git pull on every cloned project in group
type PullCommand struct {
}

// Synopsis ...
func (*PullCommand) Synopsis() string {
	return "Run git pull on currently cloned repositories"
}

// Help ...
func (h *PullCommand) Help() string {
	return `
Usage: git-group pull

  Retrieves list of projects for GROUP_URL and clones each to current directory or DIRECTORY (if specified)
`
}

// Run ...
func (h *PullCommand) Run(args []string) int {
	groupConf, err := utils.FindConfig(".")
	if os.IsNotExist(err) {
		log.Fatal("fatal: .gitgroup not found in current directory  (or any of the parent directories)")
	}
	if err != nil {
		log.Fatal(err)
	}
	groupDir := filepath.Dir(groupConf)
	fileInfo, err := ioutil.ReadDir(groupDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range fileInfo {
		if file.IsDir() {
			fmt.Printf("[%v / %v] Pulling %v\n", 1, 1, file.Name())
			cmd := exec.Command("git", "-C", filepath.Join(groupDir, file.Name()), "pull")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return 0
}
