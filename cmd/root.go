package cmd

import (
	"os"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "iam",
	Short: "Identity and Access Management.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVar(&flags.Data, "data", "data", "Data directory")
	rootCmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Enable debug mode")
}
