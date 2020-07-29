package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// AskCredentials prompts for username and password
func AskCredentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println() // newline on the password prompt
	if err != nil {
		return "", "", err
	}
	password := string(bytePassword)

	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

// FindConfig will return the first .gitgroup file that it finds going up the directory tree
func FindConfig(path string) (string, error) {
	x, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	y := ""
	var conf string
	for y != x {
		y = x
		conf = filepath.Join(y, ".gitgroup")
		if _, err := os.Stat(conf); !os.IsNotExist(err) {
			return conf, nil
		}
		x = filepath.Dir(y)
	}
	return "", os.ErrNotExist
}
