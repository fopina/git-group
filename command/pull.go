package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fopina/git-group/utils"
	flag "github.com/spf13/pflag"
)

// PullCommand implements the cli.Command interface to run git pull on every cloned project in group
type PullCommand struct {
	Meta
	WorkDirConfig
	WorkDir string
}

// Synopsis ...
func (*PullCommand) Synopsis() string {
	return "Run git pull on currently cloned repositories"
}

// Help ...
func (h *PullCommand) Help() string {
	return `
Usage: git-group pull [<OPTIONS>]

  Retrieves list of projects for GROUP_URL and clones each to current directory or DIRECTORY (if specified)

Options:
` + h.flagSet().FlagUsages()
}

func (h *PullCommand) flagSet() *flag.FlagSet {
	flags := flag.FlagSet{}
	flags.Usage = func() {
		h.UI.Output(h.Help())
	}
	flags.StringVarP(&h.WorkDir, "work-dir", "w", ".", "existing git group working directory, default to current dir")
	return &flags
}

func (h *PullCommand) parseArgs(args []string) error {
	err := h.flagSet().Parse(args)

	return err
}

// Run ...
func (h *PullCommand) Run(args []string) int {
	err := h.parseArgs(args)
	h.Meta.FatalError(err)

	groupConf, err := utils.FindConfig(h.WorkDir)
	if os.IsNotExist(err) {
		h.Meta.Fatal("fatal: .gitgroup not found in current directory  (or any of the parent directories)")
	}
	h.Meta.FatalError(err)

	groupDir := filepath.Dir(groupConf)
	fileInfo, err := ioutil.ReadDir(groupDir)
	h.Meta.FatalError(err)

	var targets []string

	for _, file := range fileInfo {
		if file.IsDir() {
			targets = append(targets, file.Name())
		}
	}

	var cloneErrors []string
	total := len(targets)
	for i, file := range targets {
		fmt.Printf("[%v / %v] Pulling %v\n", i+1, total, file)
		cmd := exec.Command("git", "-C", filepath.Join(groupDir, file), "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			// no concurrency at the moment, all good
			cloneErrors = append(cloneErrors, file)
		}
	}

	if len(cloneErrors) > 0 {
		h.UI.Error("Failed to pull for some repositories")
		for _, err := range cloneErrors {
			h.Meta.UI.Error(fmt.Sprintf("- %v", err))
		}
		return 1
	}

	return 0
}
