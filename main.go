package main

import (
	"os"

	cmd "github.com/fopina/git-group/command"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
