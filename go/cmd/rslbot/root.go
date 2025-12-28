package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rslbot",
	Short: "RaidBot API server and test client",
}

func init() {
	rootCmd.AddCommand(adminCmd)
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(cliCmd)
}
