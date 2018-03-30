package exec

import (
	"errors"
	"os/exec"
	"runtime"
)

// OpenBrowser opens the given URL with the user's default browser.
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return errors.New("Unsupported GOOS")
	}
}
