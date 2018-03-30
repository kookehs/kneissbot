package core

import (
	"errors"
	"os"
	"runtime"
	"strings"
)

// Home returns the path to the home directory
func Home() (string, error) {
	var home string

	switch runtime.GOOS {
	case "darwin":
		fallthrough
	case "linux":
		home = "~"
	case "windows":
		home = os.Getenv("USERPROFILE")

		if strings.Compare(home, "") == 0 {
			return "", errors.New("Home directory not found")
		}
	default:
		return "", errors.New("Unsupported GOOS")
	}

	return home, nil
}
