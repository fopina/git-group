package commands

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// UpdateCommand implements the cli.Command interface to self update git-group binary
type UpdateCommand struct {
	Meta
}

// Synopsis ...
func (*UpdateCommand) Synopsis() string {
	return "Self-update (directly from github releases)"
}

// Help ...
func (h *UpdateCommand) Help() string {
	return `
Usage: git-group update

  Checks latest version from https://github.com/fopina/git-group releases and updates with latest, if there is one
`
}

// Run ...
func (*UpdateCommand) Run(args []string) int {
	previous := semver.MustParse(version)
	latest, err := selfupdate.UpdateSelf(previous, repo)

	if err != nil {
		panic(err)
	}

	if previous.Equals(latest.Version) {
		fmt.Println("Current binary is the latest version", version)
	} else {
		fmt.Println("Update successfully done to version", latest.Version)
		fmt.Println("Release note:\n", latest.ReleaseNotes)
	}
	return 0
}
