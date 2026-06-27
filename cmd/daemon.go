package cmd

import (
	"errors"
	"fmt"
	
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/erewhile/iam/cmd/flags"
)

func getPIDFilePath() string {
	_ = os.MkdirAll(flags.Data, 0755)
	return filepath.Join(flags.Data, "daemon.pid")
}

func writePID() error {
	pid := os.Getpid()
	return os.WriteFile(getPIDFilePath(), []byte(strconv.Itoa(pid)), 0644)
}

func removePID() {
	_ = os.Remove(getPIDFilePath())
}

func getRunningPID() (int, error) {
	data, err := os.ReadFile(getPIDFilePath())
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	if isProcessRunning(pid) {
		return pid, nil
	}
	return 0, errors.New("process not running")
}

func isProcessRunning(pid int) bool {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid))
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(output), strconv.Itoa(pid))
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

func killProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		// Windows
		cmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
		return cmd.Run()
	}

	// Unix
	return proc.Signal(os.Interrupt)
}
