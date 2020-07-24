package commands

import "fmt"

type PullCommand struct {
}

func (*PullCommand) Synopsis() string {
	return "Run git pull on currently cloned repositories"
}

func (h *PullCommand) Help() string {
	return h.Synopsis()
}

func (*PullCommand) Run(args []string) int {
	fmt.Printf("hello, %v", args)
	return 0
}
