package commands

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

type UpdateCommand struct {
}

func (*UpdateCommand) Synopsis() string {
	return "self-update (directly from github releases)"
}

func (h *UpdateCommand) Help() string {
	return h.Synopsis()
}

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
