package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Open(resource string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", resource)
	case "darwin":
		cmd = exec.Command("open", resource)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", resource)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
