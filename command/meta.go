package commands

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
)

// Meta ...
type Meta struct {
	UI cli.BasicUi
}

// Fatal outputs message in the UI and exits with exit code 1
func (m *Meta) Fatal(v ...interface{}) {
	m.UI.Error(fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf outputs message in the UI and exits with exit code 1
func (m *Meta) Fatalf(format string, v ...interface{}) {
	m.UI.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

// FatalError outputs error in the UI and exits with exit code 1, noop if error is nil
func (m *Meta) FatalError(v error) {
	if v != nil {
		m.Fatal(v)
	}
}

// AskCredentials prompts for username and password
func (m *Meta) AskCredentials() (string, string, error) {
	username, err := m.UI.Ask("Username (use \"token\" if you're providing an API token as password):")
	if err != nil {
		return "", "", err
	}
	password, err := m.UI.AskSecret("Password:")
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}
