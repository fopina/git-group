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
	Progress bool
	Args     []string
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

	projects := make(chan utils.ListedProject)
	done := make(chan bool)
	var cloneErrors []string
	go func() {
		for {
			project, ok := <-projects
			if !ok {
				break
			}
			h.UI.Output(fmt.Sprintf("[%v / %v] Cloning %v", project.Index, project.Total, project.Project.Name))
			cmdArgs := []string{"-C", clonePath, "clone"}
			if h.Depth > 0 {
				cmdArgs = append(cmdArgs, "--depth", strconv.Itoa(h.Depth))
			}
			if h.Recursive {
				cmdArgs = append(cmdArgs, "--recursive")
			}
			cmdArgs = append(cmdArgs, project.Project.SSHURL)
			cmd := exec.Command("git", cmdArgs...)
			var err error
			if h.Progress {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
			} else {
				output, err := cmd.CombinedOutput()
				if err != nil {
					h.UI.Error(string(output))
				}
			}
			if err != nil {
				// no concurrency at the moment, all good
				cloneErrors = append(cloneErrors, fmt.Sprintf("%v (%v)", project.Project.Name, project.Project.SSHURL))
			}
		}
		done <- true
	}()
	err = client.ListGroupProjectsWithMax(projects, h.SampleSize)
	h.Meta.FatalError(err)
	close(projects)
	<-done
	if len(cloneErrors) > 0 {
		h.UI.Error("Failed to clone some repositories")
		for _, err := range cloneErrors {
			h.Meta.UI.Error(fmt.Sprintf("- %v", err))
		}
		return 1
	}

	return 0
}
