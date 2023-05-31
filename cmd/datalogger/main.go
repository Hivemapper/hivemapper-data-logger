package main

import (
	"fmt"

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
