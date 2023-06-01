package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "datalogger",
	Short: "Hivemapper HDC data logger",
	RunE:  rootRun,
}

func rootRun(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		panic(err)
	}

	fmt.Println("Goodbye")
}

func mustGetString(cmd *cobra.Command, flagName string) string {
	val, err := cmd.Flags().GetString(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}

func mustGetDuration(cmd *cobra.Command, flagName string) time.Duration {
	val, err := cmd.Flags().GetDuration(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}

func mustGetInt64(cmd *cobra.Command, flagName string) int64 {
	val, err := cmd.Flags().GetInt64(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}
