package commands

import (
	"log"

	"github.com/mitchellh/cli"
)

// Run is the CLI entrypoint
func Run(args []string) int {
	c := &cli.CLI{
		Name:         "git-group",
		Version:      version,
		HelpFunc:     cli.BasicHelpFunc("git-group"),
		Autocomplete: true,
		Commands: map[string]cli.CommandFactory{
			"clone": func() (cli.Command, error) {
				return &CloneCommand{}, nil
			},
			"pull": func() (cli.Command, error) {
				return &PullCommand{}, nil
			},
			"update": func() (cli.Command, error) {
				return &UpdateCommand{}, nil
			},
		},
		Args: args,
	}
	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
		return 1
	}

	return exitStatus
}
