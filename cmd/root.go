package cmd

import (
	"os"

	"github.com/bestruirui/octopus/internal/conf"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   conf.APP_NAME,
	Short: conf.APP_DESC,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
