package commands

import (
	"fmt"
	"io/ioutil"
	"os"
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
	CloneNew bool
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
	flags.BoolVarP(&h.CloneNew, "clone-new", "n", false, "also clone new repositories in the group")
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
		output, result.err = utils.GitCommand(h.Progress, filepath.Join(groupDir, target), "pull")
		if result.err != nil {
			result.output = string(output)
		}
		h.outputChannel <- result
	}
}

// Run ...
func (h *PullCommand) Run(args []string) int {
	err := h.parseArgs(args)
	h.Meta.FatalError(err)

	groupDir, err := h.LoadConfig(h.WorkDir)
	if os.IsNotExist(err) {
		h.Meta.Fatal("fatal: .gitgroup not found in current directory  (or any of the parent directories)")
	}
	h.Meta.FatalError(err)

	fileInfo, err := ioutil.ReadDir(groupDir)
	h.Meta.FatalError(err)

	var cloneErrors []string

	h.StartWorkers(func() {
		h.worker(groupDir)
	})

	// not just as set() but also to review which ones no longer exist in the group
	repos := make(map[string]bool)

	for _, file := range fileInfo {
		if file.IsDir() {
			repos[file.Name()] = false
		}
	}

	var newOnes []utils.ListedProject

	if h.CloneNew {
		// find new repos
		username, password, err := h.Meta.AskCredentials()
		h.Meta.FatalError(err)

		client, err := utils.NewGitlabClient(h.GroupURL)
		h.Meta.FatalError(err)

		err = client.Authenticate(username, password)
		h.Meta.FatalError(err)

		// FIXME: refactor list method to proper iterator!!!
		x := make(chan interface{}, 1)
		go func() {
			err = client.ListGroupProjects(x)
			close(x)
		}()
		for y := range x {
			yy := y.(utils.ListedProject)
			_, ok := repos[yy.Project.Name]
			if ok {
				// track for non-deletion
				repos[yy.Project.Name] = true
			} else {
				newOnes = append(newOnes, yy)
			}
		}
	}

	skipped := 0
	for k, v := range repos {
		if !v {
			skipped++
			h.UI.Warn(fmt.Sprintf("%s no longer exists (or has been archived)", k))
		}
	}

	h.FeedWorkers(func() {
		for repo, v := range repos {
			if v {
				h.inputChannel <- repo
				break
			}
		}
	})

	h.Start(len(repos) - skipped)
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

	if len(newOnes) > 0 {
		// FIXME: refactor this to more re-usable code (with CloneCommand)
		h.StartWorkers(func() {
			cloneWorker(groupDir, &h.MultiThread, &h.WorkDirConfig, h.Progress)
		})
		h.FeedWorkers(func() {
			for _, p := range newOnes {
				h.inputChannel <- p
			}
		})
		h.Start(len(newOnes))
		var result cloneResult
		for resultInt := range h.outputChannel {
			result = resultInt.(cloneResult)
			h.bar.Increment()
			if result.err != nil {
				h.UI.Error(fmt.Sprintf("%v failed (%v)\n", result.project.Project.SSHURL, result.err))
				h.UI.Error(result.output)
				cloneErrors = append(cloneErrors, fmt.Sprintf("%v (%v)", result.project.Project.Name, result.project.Project.SSHURL))
			}
		}
		h.bar.Finish()
	}

	if len(cloneErrors) > 0 {
		h.UI.Error("Failed to pull/clone some repositories")
		for _, err := range cloneErrors {
			h.Meta.UI.Error(fmt.Sprintf("- %v", err))
		}
		return 1
	}

	return 0
}
