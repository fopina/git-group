package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fopina/git-group/utils"
)

// CloneCommand implements the cli.Command interface to clone a group
type CloneCommand struct {
}

// Synopsis ...
func (h *CloneCommand) Synopsis() string {
	return "Clone repositories in a group/org"
}

// Help ...
func (h *CloneCommand) Help() string {
	return `
Usage: git-group clone <GROUP_URL> [DIRECTORY]

  Retrieves list of projects for GROUP_URL and clones each to current directory or DIRECTORY (if specified)
`
}

// Run ...
func (h *CloneCommand) Run(args []string) int {
	if len(args) < 1 {
		fmt.Println(h.Help())
		return 1
	}

	client, err := utils.NewGitlabClient(args[0])
	if err != nil {
		log.Fatal(err)
	}

	var clonePath string

	if len(args) > 1 {
		clonePath = args[1]
	} else {
		clonePath = strings.TrimLeft(client.Group, "/")
	}

	/*
		if _, err := os.Stat(clonePath); !os.IsNotExist(err) {
			log.Fatalf("fatal: destination path '%v' already exists and is not an empty directory.", clonePath)
		}
	*/
	conf := WorkDirConfig{GroupURL: args[0]}
	confPath := filepath.Join(clonePath, ".gitgroup")

	file, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(clonePath, 0700)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(confPath, file, 0600)
	if err != nil {
		log.Fatal(err)
	}

	username, password, err := utils.AskCredentials()
	if err != nil {
		log.Fatal(err)
	}

	err = client.Authenticate(username, password)
	if err != nil {
		log.Fatal(err)
	}

	projects := make(chan utils.ListedProject)
	done := make(chan bool)
	go func() {
		for {
			project, ok := <-projects
			if !ok {
				done <- true
				break
			}
			fmt.Printf("[%v / %v] Cloning %v", project.Index, project.Total, project.Project.Name)
			cmd := exec.Command("git", "-C", clonePath, "clone", project.Project.SSHURL)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	err = client.ListGroupProjects(projects)
	if err != nil {
		log.Fatal(err)
	}
	close(projects)
	<-done

	return 0
}
