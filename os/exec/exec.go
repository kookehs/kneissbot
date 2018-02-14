package exec

import (
	"errors"
	"os/exec"
	"runtime"
)

func OpenBrowser(s string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", s).Start()
	case "linux":
		return exec.Command("xdg-open", s).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", s).Start()
	default:
		return errors.New("unsupported GOOS")
	}
}
