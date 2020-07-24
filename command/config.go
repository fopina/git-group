package commands

var version string = "DEV"
var date string

const repo = "fopina/git-group"

// WorkDirConfig persists the options used to initially clone a group to re-use them for pull commands
type WorkDirConfig struct {
	GroupURL string
}
