package commands

import (
	"log"
	"os"

	"github.com/mitchellh/cli"
)

// Run is the CLI entrypoint
func Run(args []string) int {
	meta := &Meta{UI: cli.BasicUi{Reader: os.Stdin, Writer: os.Stdout, ErrorWriter: os.Stderr}}
	c := &cli.CLI{
		Name:         "git-group",
		Version:      version,
		HelpFunc:     cli.BasicHelpFunc("git-group"),
		Autocomplete: true,
		Commands: map[string]cli.CommandFactory{
			"clone": func() (cli.Command, error) {
				return &CloneCommand{Meta: *meta}, nil
			},
			"pull": func() (cli.Command, error) {
				return &PullCommand{Meta: *meta}, nil
			},
			"update": func() (cli.Command, error) {
				return &UpdateCommand{Meta: *meta}, nil
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
