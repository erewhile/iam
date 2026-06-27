package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop iam server",
	Run: func(cmd *cobra.Command, args []string) {
		stop()
	},
}

func stop() {
	pid, err := getRunningPID()
	if err != nil {
		fmt.Println("iam server is not running.")
		return
	}

	fmt.Printf("Stopping iam server (PID: %d)... ", pid)
	if err := killProcess(pid); err != nil {
		fmt.Printf("failed to stop: %v\n", err)
		return
	}

	for i := 0; i < 10; i++ {
		if !isProcessRunning(pid) {
			removePID()
			fmt.Println("stopped.")
			return
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("failed to verify stop status (timeout).")
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
