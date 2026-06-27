package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start iam server in background",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func start() {
	if pid, err := getRunningPID(); err == nil {
		fmt.Printf("iam server is already running with PID %d\n", pid)
		return
	}

	binary, err := os.Executable()
	if err != nil {
		log.Fatalf("failed to get executable path: %v", err)
	}

	subArgs := []string{"server"}
	if flags.Debug {
		subArgs = append(subArgs, "--debug")
	}

	if flags.Data != "" {
		subArgs = append(subArgs, "--data", flags.Data)
	}

	childCmd := exec.Command(binary, subArgs...)

	childCmd.Stdout = nil
	childCmd.Stderr = nil
	childCmd.Stdin = nil

	if err := childCmd.Start(); err != nil {
		log.Fatalf("failed to start background process: %v", err)
	}

	fmt.Printf("iam server started in background (PID: %d)\n", childCmd.Process.Pid)
}

func init() {
	rootCmd.AddCommand(startCmd)
}
