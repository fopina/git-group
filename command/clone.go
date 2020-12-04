package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fopina/git-group/utils"
	flag "github.com/spf13/pflag"
)

// CloneCommand implements the cli.Command interface to clone a group
type CloneCommand struct {
	Meta
	WorkDirConfig
	ProgressBar
	MultiThread
	Progress bool
	Args     []string
}

type cloneResult struct {
	project utils.ListedProject
	err     error
	output  string
}

// Synopsis ...
func (h *CloneCommand) Synopsis() string {
	return "Clone repositories in a group/org"
}

// Help ...
func (h *CloneCommand) Help() string {
	return `
Usage: git-group clone [<OPTIONS>] <GROUP_URL> [DIRECTORY]

  Retrieves list of projects for GROUP_URL and clones each to current directory or DIRECTORY (if specified)

  Options:
` + h.flagSet().FlagUsages()
}

func (h *CloneCommand) flagSet() *flag.FlagSet {
	flags := flag.FlagSet{}
	flags.Usage = func() {
		h.UI.Output(h.Help())
	}
	flags.AddFlagSet(h.MultiThread.FlagSet())
	flags.BoolVarP(&h.Progress, "progress", "p", false, "show output from git clone")
	flags.IntVarP(&h.SampleSize, "sample", "s", 0, "number of repos to clone, useful to quickly take samples of larger groups")
	flags.IntVarP(&h.Depth, "depth", "", 0, "create a shallow clone of that depth (passed to git clone)")
	flags.BoolVarP(&h.Recursive, "recursive", "", false, "initialize submodules in the clone (passed to git clone)")
	return &flags
}

func (h *CloneCommand) parseArgs(args []string) error {
	flags := h.flagSet()
	err := flags.Parse(args)
	h.Args = flags.Args()
	return err
}

func (h *CloneCommand) worker(clonePath string) {
	var output []byte

	for targetInt := range h.inputChannel {
		var result cloneResult
		result.project = targetInt.(utils.ListedProject)
		cmdArgs := []string{"-C", clonePath, "clone"}
		if h.Depth > 0 {
			cmdArgs = append(cmdArgs, "--depth", strconv.Itoa(h.Depth))
		}
		if h.Recursive {
			cmdArgs = append(cmdArgs, "--recursive")
		}
		cmdArgs = append(cmdArgs, result.project.Project.SSHURL)
		cmd := exec.Command("git", cmdArgs...)
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
func (h *CloneCommand) Run(args []string) int {
	err := h.parseArgs(args)
	h.Meta.FatalError(err)
	if len(h.Args) < 1 {
		h.Meta.Fatal("GROUP_URL is required")
	}
	h.GroupURL = h.Args[0]

	client, err := utils.NewGitlabClient(h.Args[0])
	h.Meta.FatalError(err)

	var clonePath string

	if len(h.Args) > 1 {
		clonePath = h.Args[1]
	} else {
		clonePath = strings.TrimLeft(client.Group, "/")
	}

	if _, err := os.Stat(clonePath); !os.IsNotExist(err) {
		h.Meta.Fatalf("destination path '%v' already exists and is not an empty directory.", clonePath)
	}

	username, password, err := h.Meta.AskCredentials()
	h.Meta.FatalError(err)

	err = client.Authenticate(username, password)
	h.Meta.FatalError(err)

	confPath := filepath.Join(clonePath, ".gitgroup")

	file, err := json.MarshalIndent(h.WorkDirConfig, "", " ")
	h.Meta.FatalError(err)

	err = os.MkdirAll(clonePath, 0700)
	h.Meta.FatalError(err)

	err = ioutil.WriteFile(confPath, file, 0600)
	h.Meta.FatalError(err)

	var cloneErrors []string

	h.StartWorkers(func() {
		h.worker(clonePath)
	})

	h.FeedWorkers(func() {
		err = client.ListGroupProjectsWithMax(h.inputChannel, h.SampleSize)
		h.Meta.FatalError(err)
	})

	h.Start(-1)
	var result cloneResult
	for resultInt := range h.outputChannel {
		result = resultInt.(cloneResult)
		h.bar.SetTotal(int64(result.project.Total.Int))
		h.bar.Increment()
		if result.err != nil {
			h.UI.Error(fmt.Sprintf("%v failed (%v)\n", result.project.Project.SSHURL, result.err))
			h.UI.Error(result.output)
			cloneErrors = append(cloneErrors, fmt.Sprintf("%v (%v)", result.project.Project.Name, result.project.Project.SSHURL))
		}
	}
	h.bar.Finish()

	if len(cloneErrors) > 0 {
		h.UI.Error("Failed to clone some repositories")
		for _, err := range cloneErrors {
			h.Meta.UI.Error(fmt.Sprintf("- %v", err))
		}
		return 1
	}

	return 0
}
