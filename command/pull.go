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
	ProgressBar
	MultiThread
	Progress bool
	WorkDir  string
}

type pullResult struct {
	repo   string
	err    error
	output string
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
	flags.AddFlagSet(h.MultiThread.FlagSet())
	flags.Usage = func() {
		h.UI.Output(h.Help())
	}
	flags.StringVarP(&h.WorkDir, "work-dir", "w", ".", "existing git group working directory, default to current dir")
	flags.BoolVarP(&h.Progress, "progress", "p", false, "show output from git pull")
	return &flags
}

func (h *PullCommand) parseArgs(args []string) error {
	flags := h.flagSet()
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	extra := flags.Args()
	if len(extra) > 0 {
		return fmt.Errorf("this command does not have any positional parameters (%v)", extra)
	}
	return nil
}

func (h *PullCommand) worker(groupDir string) {
	var output []byte
	var target string

	for targetInt := range h.inputChannel {
		target = targetInt.(string)
		var result pullResult
		result.repo = target
		cmd := exec.Command("git", "-C", filepath.Join(groupDir, target), "pull")
		if h.Progress {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			result.err = cmd.Run()
		} else {
			output, result.err = cmd.CombinedOutput()
			if result.err != nil {
				result.output = string(output)
			}
		}
		h.outputChannel <- result
	}
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

	var cloneErrors []string

	h.StartWorkers(func() {
		h.worker(groupDir)
	})

	var repos []string

	for _, file := range fileInfo {
		if file.IsDir() {
			repos = append(repos, file.Name())
		}
	}

	h.FeedWorkers(func() {
		for _, repo := range repos {
			h.inputChannel <- repo
		}
	})

	h.Start(len(repos))
	var result pullResult
	for resultInt := range h.outputChannel {
		result = resultInt.(pullResult)
		h.bar.Increment()
		if result.err != nil {
			h.UI.Error(fmt.Sprintf("%v failed (%v)\n", result.repo, result.err))
			h.UI.Error(result.output)
			cloneErrors = append(cloneErrors, result.repo)
		}
	}
	h.bar.Finish()

	if len(cloneErrors) > 0 {
		h.UI.Error("Failed to pull for some repositories")
		for _, err := range cloneErrors {
			h.Meta.UI.Error(fmt.Sprintf("- %v", err))
		}
		return 1
	}

	return 0
}
