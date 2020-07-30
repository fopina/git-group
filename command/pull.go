package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fopina/git-group/utils"
	flag "github.com/spf13/pflag"
)

// PullCommand implements the cli.Command interface to run git pull on every cloned project in group
type PullCommand struct {
	Meta
	That string
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
	flags.StringVarP(&h.That, "that", "t", "", "profile to use")
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
