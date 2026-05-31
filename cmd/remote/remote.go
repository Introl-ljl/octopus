package remote

import (
	"github.com/bestruirui/octopus/cmd"
	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
)

var (
	profileName string
	serverURL   string
	outputMode  string
	force       bool
)

var RemoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage Octopus server via remote API",
	Long: `Remote management commands for Octopus server.

Connect to a running Octopus server and manage channels, groups, API keys,
models, settings, view statistics and logs — all from the command line.

Requires a configured profile and authenticated session.
Use 'octopus remote login' to authenticate after configuring a profile.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if profileName != "" {
			cli.SetActiveProfile(profileName)
		}
		return nil
	},
}

func init() {
	RemoteCmd.PersistentFlags().StringVar(&profileName, "profile", "", "local connection profile name")
	RemoteCmd.PersistentFlags().StringVar(&serverURL, "server", "", "override server URL for this invocation")
	RemoteCmd.PersistentFlags().StringVar(&outputMode, "output", "", "output mode: table (default) or json")
	RemoteCmd.PersistentFlags().BoolVar(&force, "force", false, "bypass confirmation for destructive commands")
	cmd.RootCmd.AddCommand(RemoteCmd)
}
